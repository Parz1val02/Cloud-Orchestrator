import paramiko
import subprocess
import random

# import math
# import ipaddress

SLICE_ID = "667790af65d36d25bb0779f6"


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

    stdin, stdout, stderr = ssh_client.exec_command(script, get_pty=True)
    stdin.write(password + "\n")
    stdin.flush()
    print(stderr.read().decode("utf-8"))
    print(stdout.read().decode("utf-8"))
    ssh_client.close()


# def calculate_subnet_mask(number_of_nodes):
#    # Include the reserved addresses in the count
#    total_required_addresses = number_of_nodes + 2  # +2 for the reserved addresses
#
#    # Calculate the number of bits needed for the total required addresses
#    required_bits = math.ceil(math.log2(total_required_addresses))
#
#    # Calculate the number of host bits
#    host_bits = max(0, required_bits)
#
#    # Calculate the number of network bits
#    network_bits = 32 - host_bits
#
#    # Calculate the subnet mask
#    subnet_mask = (0xFFFFFFFF >> host_bits) << host_bits
#
#    # Format the subnet mask in the familiar dot-decimal notation
#    subnet_mask = (
#        subnet_mask >> 24 & 0xFF,
#        subnet_mask >> 16 & 0xFF,
#        subnet_mask >> 8 & 0xFF,
#        subnet_mask & 0xFF,
#    )
#    subnet_mask_str = ".".join(map(str, subnet_mask))
#
#    return network_bits, subnet_mask_str
#
#
# def calculate_ip_range(network_bits):
#    # Define the base network address
#    network = ipaddress.IPv4Network(f"192.168.0.0/{network_bits}", strict=False)
#    # Extract all the IP addresses in the network
#    all_ips = list(network.hosts())
#
#    # Assign the first and last IP addresses
#    first_ip = all_ips[0]
#    last_ip = all_ips[-1]
#
#    return str(first_ip), str(last_ip)


def main():
    json_data = {
        "slice_id": "667790af65d36d25bb0779f6",
        "created_at": "2024-06-24T02:35:37.1113996Z",
        "description": "descr",
        "name": "name",
        "topology": {
            "links": [
                {"link_id": "nd1_nd2", "source": "node_1", "target": "node_2"},
                {"link_id": "nd2_nd3", "source": "node_2", "target": "node_3"},
                {"link_id": "nd3_nd1", "source": "node_3", "target": "node_1"},
            ],
            "nodes": [
                {
                    "node_id": "nd1",
                    "name": "node_1",
                    "image": "Ubuntu 20.04 LTS",
                    "flavor": {
                        "id": "665275b98c45f0c2b8a2e230",
                        "name": "t2.micro",
                        "cpu": 1,
                        "memory": 0.5,
                        "storage": 1,
                    },
                    "security_rules": [22],
                },
                {
                    "node_id": "nd2",
                    "name": "node_2",
                    "image": "Ubuntu 20.04 LTS",
                    "flavor": {
                        "id": "665275b98c45f0c2b8a2e230",
                        "name": "t2.micro",
                        "cpu": 1,
                        "memory": 0.5,
                        "storage": 1,
                    },
                    "security_rules": [22],
                },
                {
                    "node_id": "nd3",
                    "name": "node_3",
                    "image": "Ubuntu 20.04 LTS",
                    "flavor": {
                        "id": "665275b98c45f0c2b8a2e230",
                        "name": "t2.micro",
                        "cpu": 1,
                        "memory": 0.5,
                        "storage": 1,
                    },
                    "security_rules": [22],
                },
            ],
        },
        "user_id": "6640550a53c1187a6899a5a9",
        "topology_type": "anillo",
        "availability_zone": "ga",
        "deployment_type": "linux",
        "internet": False,
    }

    # network_bits, subnet_mask = calculate_subnet_mask(number_of_nodes)
    # first_ip, last_ip = calculate_ip_range(network_bits)

    # print(f"First IP: {first_ip}")
    # print(f"Last IP: {last_ip}")
    # print(f"Subnet Mask: {subnet_mask} (/ {network_bits})")

    # Direcciones y credenciales de los nodos
    worker_addresses = ["10.0.0.30", "10.0.0.40", "10.0.0.50"]
    username = "ubuntu"
    password = "ubuntu"

    # Parámetros para los scripts
    headnode_ovs_name = "br-linux"
    headnode_interfaces = "ens5"  # Coloca las interfaces del HeadNode aquí
    worker_ovs_name = "br-linux"
    worker_interfaces = "ens4"  # Coloca las interfaces de los Workers aquí
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
    nodes = json_data["topology"]["nodes"]
    vm_parameters = []
    for i in nodes:
        vm_name = f"{i['node_id']}"
        bridge = "br-linux"
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
        print(f"bash init_orchestrator/internal_net_headnode.sh {' '.join(vlan_param)}")
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
        execute_on_worker(
            worker_address,
            f"sudo -S bash obtain_worker.sh {vlan_id}",
            username,
            password,
        )
    # if internet == 1:
    #    for vlan_param in vlan_parameters:
    #        vlan_id = vlan_param[1]
    #        print(f"implement_orchestrator/vlan_internet.sh {vlan_id}")
    #        execute_on_headnode(f"implement_orchestrator/vlan_internet.sh {vlan_id}")

    print("Orquestador de cómputo inicializado exitosamente.")


if __name__ == "__main__":
    main()
