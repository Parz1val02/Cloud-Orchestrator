import asyncio
import time
from openstack_sdk import (
    password_authentication_with_scoped_authorization,
    create_instance,
    create_network,
    create_subnet,
    create_port,
    token_authentication_with_scoped_authorization,
    delete_network,
    delete_port,
    delete_server,
    delete_subnet,
    list_instances,
    list_ports,
    list_subnets,
    get_instance_detail,
    get_network_by_id,
    list_networks,
    create_project,
    assign_role_to_user,
    get_users,
    list_projects_general,
    delete_project,
    get_instance_console
)
import json

# Datos de configuración
GATEWAY_IP = "10.20.12.162"
DOMAIN_ID = "default"
ADMIN_USER_USERNAME = "admin"
ADMIN_USER_ID = "54702871434c4554aff7e36f48031ddc"
ADMIN_USER_PASSWORD = "88393102cc9e5ae0ec8247080fc0849d"
ADMIN_ROLE_ID = "5ddf01a82a9e49f8b4641295e81b1597"
MEMBER_ROLE_ID = "91f9765154b44489ab090a9cb523b2a9"
ADMIN_USER_DOMAIN_NAME = "Default"
ADMIN_PROJECT_NAME = "admin"
PROJECT_NAME = "prueba"  # Nombre del proyecto específico
PROJECT_ID = "96c08a83b3bc46e6b29e0fcddeff0101"  # ID del proyecto específico
IMAGE_ID = "377a37ee-a633-4623-b05d-a35f0c6ed6f1"
FLAVOR_ID = "923cb049-1fad-4caf-b802-feb00719daf1"

# Endpoints
KEYSTONE_ENDPOINT = f"http://{GATEWAY_IP}:5000/v3"
NOVA_ENDPOINT = f"http://{GATEWAY_IP}:8774/v2.1"
NEUTRON_ENDPOINT = f"http://{GATEWAY_IP}:9696/v2.0"


# Función para autenticación y obtención de token de administrador
def authenticate_admin():
    resp = password_authentication_with_scoped_authorization(
        KEYSTONE_ENDPOINT,
        ADMIN_USER_DOMAIN_NAME,
        ADMIN_USER_USERNAME,
        ADMIN_USER_PASSWORD,
        DOMAIN_ID,
        ADMIN_PROJECT_NAME,
    )
    if resp.status_code == 201:
        return resp.headers["X-Subject-Token"]
    else:
        raise Exception("Failed to authenticate as admin")


# Función para obtener token del proyecto
def authenticate_project(admin_token, project_name):
    resp = token_authentication_with_scoped_authorization(
        KEYSTONE_ENDPOINT, admin_token, DOMAIN_ID, project_name
    )
    if resp.status_code == 201:
        return resp.headers["X-Subject-Token"]
    else:
        raise Exception("Failed to authenticate project")


# Función para crear una red y subred
def create_network_and_subnet(token, network_name, subnet_cidr):
    # Crear red
    network_resp = create_network(NEUTRON_ENDPOINT, token, network_name)
    if network_resp.status_code == 201:
        network_data = network_resp.json()
        network_id = network_data["network"]["id"]
        print(f" {network_name} created successfully with ID {network_id}")
        # Crear subred
        subnet_resp = create_subnet(
            NEUTRON_ENDPOINT,
            token,
            network_id,
            network_name + "_subnet",
            "4",
            subnet_cidr,
        )
        if subnet_resp.status_code == 201:
            subnet_data = subnet_resp.json()
            subnet_id = subnet_data["subnet"]["id"]
            return network_id, subnet_id
        else:
            # Borrar la red si la subred falla
            delete_network(NEUTRON_ENDPOINT, token, network_id)
            raise Exception(f"Failed to create subnet for network {network_name}")
    else:
        raise Exception(f"Failed to create network {network_name}")


# Función para crear un puerto en una red específica
def create_instance_port(token, network_id, port_name, project_id):
    port_resp = create_port(NEUTRON_ENDPOINT, token, port_name, network_id, project_id)
    if port_resp.status_code == 201:
        port_data = port_resp.json()
        port_id = port_data["port"]["id"]
        print(
            f"Port {port_name} created successfully with ID {port_id}"
        )
        return port_data["port"]["id"]
    else:
        raise Exception(f"Failed to create port {port_name} in network {network_id}")


# Función para crear una instancia (VM)
def launch_instance(token, instance_name, image_name, flavor_name, networks):
    resp = create_instance(
        NOVA_ENDPOINT, token, instance_name, flavor_name, IMAGE_ID, networks
    )
    if resp.status_code == 202:
        instance_data = resp.json()
        instance_id = instance_data["server"]["id"]
        print(
            f"Instance {instance_name} created successfully with ID {instance_id}"
        )
    else:
        raise Exception(f"Failed to launch instance {instance_name}")


def create_slice_topology(topology_json, project_token, project_id):
    # admin_token = authenticate_admin()
    # project_token = authenticate_project(admin_token)

    subnet_cidr_template = "10.0.{subnet_index}.0/24"

    links = topology_json["links"]
    links_temp = {}

    for i, link in enumerate(links, start=1):
        subnet_cidr = subnet_cidr_template.format(subnet_index=i)

        network_name = link["link_id"]
        network_id, subnet_id = create_network_and_subnet(
            project_token, network_name, subnet_cidr
        )

        port_name0 = link["source"]
        puerto0_id = create_instance_port(
            project_token, network_id, port_name0, project_id
        )

        port_name1 = link["target"]
        puerto1_id = create_instance_port(
            project_token, network_id, port_name1, project_id
        )

        links_temp[link["link_id"]] = {
            "network": network_id,
            "subnet": subnet_id,
            "puerto0": puerto0_id,
            "puerto1": puerto1_id,
        }
    print("links temp: ",links_temp)

    nodes = topology_json["nodes"]

    for node in nodes:
        instance_name = node["name"]
        flavor_id = node["flavor"]["id"]
        image_id = node["image"]
        networks = []
        for port in node["ports"]:
            port_id = port["node_id"]    #antes "id"
            link = find_link_by_source_port(port_id, links)  
            if link is None:
                link = find_link_by_target_port(port_id, links)
                port = links_temp[link["link_id"]]["puerto1"]
            else:
                port = links_temp[link["link_id"]]["puerto0"]

            networks.append({"port": port})
        print("networks: ", networks)
        launch_instance(project_token, instance_name, image_id, flavor_id, networks)
        time.sleep(2)



def find_link_by_source_port(port_id, links):
    for link in links:
        if link["source_port"] == port_id:
            return link
    return None


def find_link_by_target_port(port_id, links):
    for link in links:
        if link["target_port"] == port_id:
            return link
    return None


# create_slice_topology()


async def delete_instances_by_networkid(token, network_id, project_id):
    instances_resp = list_instances(NOVA_ENDPOINT, token, project_id)
    network_original_info = get_network_by_id(NEUTRON_ENDPOINT, token, network_id)
    if instances_resp.status_code == 200:
        instances = instances_resp.json()["servers"]
        # print(instances)
        for instance in instances:
            # Verificar si la instancia está asociada a la red específica
            instance_detail_resp = get_instance_detail(
                NOVA_ENDPOINT, token, instance["id"]
            )
            if instance_detail_resp.status_code == 200:
                instance_detail = instance_detail_resp.json()["server"]
                for network in instance_detail["addresses"]:
                    # print(network)
                    if network == network_original_info["name"]:
                        instance_id = instance["id"]
                        delete_resp = delete_server(
                            NOVA_ENDPOINT, token, instance_id
                        )
                        print(f"Deleted instance: {instance_id} for {network}")
                        if delete_resp.status_code != 204:
                            raise Exception(
                                f"Failed to delete instance {instance_id}"
                            )
    else:
        raise Exception("Failed to list instances")


# Función para listar y borrar puertos
async def delete_ports(token, network_id):
    ports_resp = list_ports(NEUTRON_ENDPOINT, token, network_id)
    if ports_resp.status_code == 200:
        ports = ports_resp.json()["ports"]
        for port in ports:
            port_id = port["id"]
            delete_resp = delete_port(NEUTRON_ENDPOINT, token, port_id)
            if delete_resp.status_code != 204:
                raise Exception(f"Failed to delete port {port_id}")
            else:
                # print("Deleted port: ",port_id)
                pass
    else:
        raise Exception("Failed to list ports")


# Función para listar y borrar subredes
async def delete_subnets(token, network_id):
    subnets_resp = list_subnets(NEUTRON_ENDPOINT, token, network_id)
    if subnets_resp.status_code == 200:
        subnets = subnets_resp.json()["subnets"]
        for subnet in subnets:
            subnet_id = subnet["id"]
            delete_resp = delete_subnet(NEUTRON_ENDPOINT, token, subnet_id)
            if delete_resp.status_code != 204:
                raise Exception(f"Failed to delete subnet {subnet_id}")
            else:
                # print("Deleted subnet: ",subnet["id"])
                pass
    else:
        raise Exception("Failed to list subnets")


# Función para borrar una red
async def delete_network_by_id(token, network_id):
    delete_resp = delete_network(NEUTRON_ENDPOINT, token, network_id)
    if delete_resp.status_code != 204:
        raise Exception(f"Failed to delete network {network_id}")
    else:
        print("Deleted link with ID: ", network_id)


def list_networks_slice(project_id, token):
    list_resp = list_networks(NEUTRON_ENDPOINT, token, project_id)
    if list_resp.status_code == 200:
        networks_list = list_resp.json()["networks"]
        print(f"Networks list of project {project_id} obtained")
        return networks_list
    else:
        raise Exception(f"Failed to list networks of project {project_id}")


# Función principal para borrar la topología
async def delete_slice_topology(project_token, project_id):
    #admin_token = authenticate_admin()
    #project_token = authenticate_project(admin_token)
    print(project_token)

    networks = list_networks_slice(project_id, project_token)
    # networks_id = []
    if networks:
        for network in networks:
            # networks_id.append(network["id"])
            network_id = network["id"]
            # Borrar instancias
            await delete_instances_by_networkid(project_token, network_id, project_id)

            # Borrar puertos
            await delete_ports(project_token, network_id)

            # Borrar subredes
            await delete_subnets(project_token, network_id)

            # Borrar red
            await delete_network_by_id(project_token, network_id)
            time.sleep(0.25)
        print("network deleted.")
    else:
        print(f"Project with id {project_id} do not have links")


# Llamar a la función principal para borrar la topología
# network_id = "551f8238-1464-44dc-bc19-d448fdb46eb2"  # Reemplaza con el ID de la red que deseas borrar
# asyncio.run(delete_slice_topology())


def asignarRoleUsuarioPorProject(admin_token, project_id, user_id, role_id):
    try:
        resp = assign_role_to_user(
            KEYSTONE_ENDPOINT, admin_token, project_id, user_id, role_id
        )
        print(resp.status_code)
        if resp.status_code == 204:
            print("ROLE ASSIGNED SUCCESSFULLY")
            return True
        else:
            print("ROLE NOT ASSIGNED")
            return None
    except:
        return None


def crearProject(admin_token, domain_id, project_name, project_description):
    try:
        resp = create_project(
            KEYSTONE_ENDPOINT, admin_token, domain_id, project_name, project_description
        )
        print(resp.status_code)
        if resp.status_code == 201:
            print("PROJECT CREATED SUCCESSFULLY")
            project_created = resp.json()
            print(json.dumps(project_created))
            return project_created
        else:
            print("FAILED PROJECT CREATION")
            return None
    except:
        return None


def obtainUserId(admin_token, user_name):
    try:
        resp = get_users(KEYSTONE_ENDPOINT, admin_token)
        print(resp.status_code)
        if resp.status_code == 200:
            print("USERS OBTAINED SUCCESSFULLY")
            users = resp.json()
            print(json.dumps(users))
            for user in users["users"]:
                if user["name"] == user_name:
                    print("ID del Usuario: ", user["id"])
                    return user["id"]
            print("USER NOT FOUND")
            return None
        else:
            print("FAILED USERS OBTAINMENT")
            return None
    except:
        return None


def crear_proyecto(admin_token, domain_id, project_name, project_description):
    try:
        resp = create_project(
            KEYSTONE_ENDPOINT, admin_token, domain_id, project_name, project_description
        )
        print(resp.status_code)
        if resp.status_code == 201:
            print("PROJECT CREATED SUCCESSFULLY")
            project_created = resp.json()
            print(json.dumps(project_created))
            return project_created
        else:
            print("FAILED PROJECT CREATION")
            return None
    except:
        return None
    
def obtenerIdProyecto(project_token, project_name):
    try:
        resp = list_projects_general(KEYSTONE_ENDPOINT, project_token)
        print(resp.status_code)
        if resp.status_code == 200:
            print("PROJECTS OBTAINED SUCCESSFULLY")
            projects = resp.json()
            print(json.dumps(projects))
            for project in projects["projects"]:
                if project["name"] == project_name:
                    print("id del proyecto: ", project["id"])
                    return project["id"]
            print("PROJECT NOT FOUND")
            return None
        else:
            print("FAILED PROJECTS OBTAINMENT")
            return None
    except:
        return None

def openstackDeployment(slice_json, user_name):
    project_name = slice_json["name"]
    project_description = slice_json["description"]
    print(project_name)
    print(project_description)
    
    try:
        admin_token = authenticate_admin()
        if admin_token:
            # 2.- Crear el proyecto
            project = crear_proyecto(
                admin_token, DOMAIN_ID, project_name, project_description
            )
            if project:
                project_id = project["project"]["id"]
                rol_admin = asignarRoleUsuarioPorProject(
                    admin_token, project_id, ADMIN_USER_ID, ADMIN_ROLE_ID
                )
                if rol_admin:
                    project_token = authenticate_project(admin_token, project_name)
                    if project_token:
                        user_id = obtainUserId(project_token, user_name)
                        if user_id:
                            rol_user = asignarRoleUsuarioPorProject(
                                project_token, project_id, user_id, MEMBER_ROLE_ID
                            )
                            if rol_user:
                                create_slice_topology(
                                    slice_json["topology"], project_token, project_id
                                )
    except Exception as e:
        print(e)
        print("Slice deployment failed")
        #return False
    else:
        #return True
        pass

def deleteProject(project_token,project_id):
    try:
        resp = delete_project(KEYSTONE_ENDPOINT, project_token, project_id)
        print(resp.status_code)
        if resp.status_code == 204:
            print("PROJECT DELETED SUCCESSFULLY")
            return True
        else:
            print("FAILED TO DELETE PROJECT")
            return None
    except:
        return None



def openstackDeleteSlice(project_name,slice_id):
    
    print(project_name)
    deleted = False
    try:
        admin_token = authenticate_admin()
        if admin_token:
            project_token = authenticate_project(admin_token, project_name)
            if project_token:
                project_id = obtenerIdProyecto(project_token, project_name)
                if project_id:
                    asyncio.run(delete_slice_topology(project_token, project_id))
                    project_borrado = deleteProject(admin_token, project_id)
                    time.sleep(1.5)
                    if project_borrado:
                        deleted = True
                    else:
                        print("Slice elimination failed")
                    
    except Exception as e:
        print(e)
        print("Slice elimination failed")
        return deleted
    else:
        print(f"Project with id {project_id} deleted successfully")
        print(f"Slice with id {slice_id} deleted successfully")
        return deleted



def obtainVNCfromProject(project_name):
    vnc_urls = {}
    try:
        admin_token = authenticate_admin()
        if admin_token:
            project_token = authenticate_project(admin_token, project_name)
            if project_token:
                project_id = obtenerIdProyecto(project_token, project_name)
                if project_id:
                    instances_resp = list_instances(NOVA_ENDPOINT, project_token, project_id)
                    if instances_resp.status_code == 200:
                        instances = instances_resp.json()["servers"]
                        # print(instances)
                        for instance in instances:
                            server_id = instance["id"]
                            server_name = instance["name"]
                            vnc_resp = get_instance_console(NOVA_ENDPOINT,project_token,server_id)
                            if vnc_resp.status_code == 200:
                                vnc_console_url = vnc_resp.json()["console"]["url"]
                                vnc_console_url = vnc_console_url.replace("controller", "10.20.12.162")
                                vnc_urls[server_name] = vnc_console_url
                                print(f"Servidor: {server_name}, VNC URL: {vnc_console_url}")                                                                           
    except Exception as e:
        print(e)
        print("VNC links failed")
        return None
    else:
        print(f"VNCs of project id {project_id} obtained successfully")
        return vnc_urls
       
        