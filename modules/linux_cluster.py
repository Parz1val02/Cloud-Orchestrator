import json
import paramiko
import subprocess
import random
import copy
from pymongo import MongoClient
from celery import shared_task
from bson.objectid import ObjectId

# import math
# import ipaddress

client = MongoClient("localhost", 27017)
db = client.cloud
collection = db.slices
# Direcciones y credenciales de los nodos
worker_addresses = ["10.0.0.30", "10.0.0.40", "10.0.0.50"]
username = "ubuntu"
password = "ubuntu"

# Parámetros para los scripts
headnode_ovs_name = "br-vlan"
headnode_interfaces = "ens5"  # Coloca las interfaces del HeadNode aquí
worker_ovs_name = "br-vlan"
worker_interfaces = "ens4"  # Coloca las interfaces de los Workers aquí


# ejecución de scripts en el HeadNode local
# Function to assign nodes to workers using a round-robin algorithm
def assign_nodes_to_workers(num_nodes, workers):
    nodes = [f"{i+1}" for i in range(num_nodes)]
    assignments = {}
    num_workers = len(workers)

    for i, node in enumerate(nodes):
        worker = workers[i % num_workers]
        if worker in assignments:
            assignments[worker].append(node)
        else:
            assignments[worker] = [node]

    return assignments


def execute_on_headnode(script):
    try:
        subprocess.run(script, shell=True, check=True)
    except subprocess.CalledProcessError as e:
        print("Error al ejecutar el script en el HeadNode:", e)


# Conexión SSH y ejecución de scripts en los Workers
def execute_on_worker(worker_address, script, username, password):
    ssh_client = paramiko.SSHClient()
    ssh_client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh_client.connect(hostname=worker_address, username=username, password=password)
    try:
        # Execute the command
        stdin, stdout, stderr = ssh_client.exec_command(script, get_pty=True)

        # Provide password for sudo if requested
        stdin.write(password + "\n")
        stdin.flush()

        # Read the output from stdout
        output = stdout.read().decode("utf-8")

        # Read any error output from stderr (if needed)
        error = stderr.read().decode("utf-8")
        if error:
            print(f"Error encountered: {error}")

        return output.strip()

    finally:
        ssh_client.close()


@shared_task
def create(slice_id):
    slice = collection.find_one({"_id": ObjectId(slice_id)})
    if slice:
        list_of_nodes = []
        vlan_id = str(random.randint(1, 500))
        vlan_parameters = [
            (
                "vlan" + vlan_id,
                vlan_id,
                "192.168.0.0/24",
                "192.168.0.3,192.168.0.100,255.255.255.255",
                headnode_ovs_name,
            )
        ]
        nodes = slice["topology"]["nodes"]
        vm_parameters = []
        for i in nodes:
            vm_name = f"{i['node_id']}"
            bridge = worker_ovs_name
            vlan_id = vlan_id
            portga = random.randint(1, 500)
            port = str(5900 + portga)
            vm_parameters.append([vm_name, bridge, vlan_id, port])

        # Ejecución de los scripts en el HeadNode
        print(
            f"bash init_orchestrator/init_headnode.sh {headnode_ovs_name} {headnode_interfaces}"
        )
        execute_on_headnode(
            f"bash init_orchestrator/init_headnode.sh {headnode_ovs_name} {headnode_interfaces}"
        )
        for vlan_param in vlan_parameters:
            print(
                f"bash init_orchestrator/internal_net_headnode.sh {' '.join(vlan_param)}"
            )
            execute_on_headnode(
                f"bash init_orchestrator/internal_net_headnode.sh {' '.join(vlan_param)}"
            )

        # Ejecución de los scripts en los Workers
        for worker_address in worker_addresses:
            print(f"sudo -S bash init_worker.sh {worker_ovs_name} {worker_interfaces}")
            execute_on_worker(
                worker_address,
                f"sudo -S bash init_worker.sh {worker_ovs_name} {worker_interfaces}",
                username,
                password,
            )

        assignments = assign_nodes_to_workers(len(nodes), worker_addresses)

        for worker, assigned_nodes in assignments.items():
            print(f"{worker} is assigned nodes: {', '.join(assigned_nodes)}")
            for i in assigned_nodes:
                print(f"sudo -S bash vm_script.sh {' '.join(vm_parameters[int(i)-1])}")
                execute_on_worker(
                    worker,
                    f"sudo -S bash vm_script.sh {' '.join(vm_parameters[int(i)-1])}",
                    username,
                    password,
                )

        for worker_address in worker_addresses:
            print(f"sudo -S bash obtain_worker.sh {vlan_id}")
            output = execute_on_worker(
                worker_address,
                f"sudo -S bash obtain_worker.sh {vlan_id}",
                username,
                password,
            )
            lines = output.strip().splitlines()
            print(lines)
            if lines:
                last_line = lines[-1]
                print(last_line)
                parts = last_line.split()  # Split the last line by whitespace

                if len(parts) == 4:
                    var1 = parts[0]  # node
                    var2 = parts[1]  # qemu process
                    var3 = parts[2]  # vnc port
                    var4 = parts[3]  # vnc port
                    node_info = {
                        "node_id": var1,
                        "process": var2,
                        "vnc": var3,
                        "worker": var4,
                    }
                    list_of_nodes.append(node_info)

                else:
                    print(
                        "Last line does not contain three strings separated by spaces."
                    )

            else:
                print("Empty string")

        # if internet == 1:
        #    for vlan_param in vlan_parameters:
        #        vlan_id = vlan_param[1]
        #        print(f"implement_orchestrator/vlan_internet.sh {vlan_id}")
        #        execute_on_headnode(f"implement_orchestrator/vlan_internet.sh {vlan_id}")

        slice_id_value = slice.pop("slice_id", None)

        updated_slice_data = copy.deepcopy(slice)

        for node in updated_slice_data["topology"]["nodes"]:
            for node2 in list_of_nodes:
                if node["node_id"] == node2["node_id"]:
                    node["process"] = node2["process"]
                    node["vnc"] = node2["vnc"]
                    node["worker"] = node2["worker"]
                    break
        updated_slice_data["vlan_id"] = vlan_id
        print(json.dumps(updated_slice_data, indent=2))
        result = collection.update_one(
            {"_id": ObjectId(slice_id_value)}, {"$set": updated_slice_data}
        )
        if result.modified_count == 1:
            print(f"Slice with slice id {slice_id_value} updated successfully")
        else:
            print(f"Slice with slice id {slice_id_value} not updated due to error")
        print("Orquestador de cómputo inicializado exitosamente.")
    else:
        print(f"Slice with slice id {slice_id} not found")


@shared_task
def delete(slice_id):
    slice = collection.find_one({"_id": ObjectId(slice_id)})
    if slice:
        vlan = slice["vlan_id"]
        execute_on_headnode(
            f"bash implement_orchestrator/delete_headnode.sh {vlan} {headnode_ovs_name}"
        )
        for node in slice["topology"]["nodes"]:
            node_id = node["node_id"]
            process = node["process"]
            worker = node["worker"]
            execute_on_worker(
                worker,
                f"sudo -S bash delete_worker.sh {vlan} {headnode_ovs_name} {node_id} {process}",
                username,
                password,
            )

    else:
        print(f"Slice with slice id {slice_id} not found")
