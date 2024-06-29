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
    get_users
)
import json

# Datos de configuración
GATEWAY_IP = '10.20.12.162'
DOMAIN_ID = 'default'
ADMIN_USER_USERNAME = 'admin'
ADMIN_USER_ID = '54702871434c4554aff7e36f48031ddc'
ADMIN_USER_PASSWORD = '88393102cc9e5ae0ec8247080fc0849d'
ADMIN_ROLE_ID = '5ddf01a82a9e49f8b4641295e81b1597'
MEMBER_ROLE_ID = '91f9765154b44489ab090a9cb523b2a9'
ADMIN_USER_DOMAIN_NAME = 'Default'
ADMIN_PROJECT_NAME = 'admin'
PROJECT_NAME = 'prueba'  # Nombre del proyecto específico
PROJECT_ID = '96c08a83b3bc46e6b29e0fcddeff0101'  # ID del proyecto específico
IMAGE_ID = '377a37ee-a633-4623-b05d-a35f0c6ed6f1'
FLAVOR_ID = '923cb049-1fad-4caf-b802-feb00719daf1'

# Endpoints
KEYSTONE_ENDPOINT = f'http://{GATEWAY_IP}:5000/v3'
NOVA_ENDPOINT = f'http://{GATEWAY_IP}:8774/v2.1'
NEUTRON_ENDPOINT = f'http://{GATEWAY_IP}:9696/v2.0'

# Función para autenticación y obtención de token de administrador
def authenticate_admin():
    resp = password_authentication_with_scoped_authorization(
        KEYSTONE_ENDPOINT, ADMIN_USER_DOMAIN_NAME, ADMIN_USER_USERNAME,
        ADMIN_USER_PASSWORD, DOMAIN_ID, ADMIN_PROJECT_NAME
    )
    if resp.status_code == 201:
        return resp.headers['X-Subject-Token']
    else:
        raise Exception('Failed to authenticate as admin')

# Función para obtener token del proyecto
def authenticate_project(admin_token, project_name):
    resp = token_authentication_with_scoped_authorization(
        KEYSTONE_ENDPOINT, admin_token, DOMAIN_ID, project_name
    )
    if resp.status_code == 201:
        return resp.headers['X-Subject-Token']
    else:
        raise Exception('Failed to authenticate project')
    

# Función para crear una red y subred
def create_network_and_subnet(token, network_name, subnet_cidr):
    # Crear red
    network_resp = create_network(NEUTRON_ENDPOINT, token, network_name)
    if network_resp.status_code == 201:
        network_data = network_resp.json()
        network_id = network_data['network']['id']
        print(f' {network_name} created successfully with ID {network_id}')
        # Crear subred
        subnet_resp = create_subnet(
            NEUTRON_ENDPOINT, token, network_id, network_name + '_subnet',
            '4', subnet_cidr
        )
        if subnet_resp.status_code == 201:
            subnet_data = subnet_resp.json()
            subnet_id = subnet_data['subnet']['id']
            return network_id, subnet_id
        else:
            # Borrar la red si la subred falla
            delete_network(NEUTRON_ENDPOINT, token, network_id)
            raise Exception(f'Failed to create subnet for network {network_name}')
    else:
        raise Exception(f'Failed to create network {network_name}')

# Función para crear un puerto en una red específica
def create_instance_port(token, network_id, port_name, project_id):
    port_resp = create_port(
        NEUTRON_ENDPOINT, token, port_name, network_id, project_id
    )
    if port_resp.status_code == 201:
        port_data = port_resp.json()
        print(f'Port {port_name} created successfully with ID {port_data["port"]["id"]}')
        return port_data['port']['id']
    else:
        raise Exception(f'Failed to create port {port_name} in network {network_id}')

# Función para crear una instancia (VM)
def launch_instance(token, instance_name, image_name, flavor_name, networks):
    resp = create_instance(
        NOVA_ENDPOINT, token, instance_name, flavor_name, image_name, networks
    )
    if resp.status_code == 202:
        instance_data = resp.json()
        print(f'Instance {instance_name} created successfully with ID {instance_data["server"]["id"]}')
    else:
        raise Exception(f'Failed to launch instance {instance_name}')




"""
topology_json = {
    "name": "arbol_binario",
    "topology": {
        "links": [
            {
                "link_id": "nd1_nd2",
                "source": "nd1",
                "target": "nd2",
                "source_port": "nd1_port0",
                "target_port": "nd2_port0"
            },
            {
                "link_id": "nd1_nd3",
                "source": "nd1",
                "target": "nd3",
                "source_port": "nd1_port1",
                "target_port": "nd3_port0"
            },
            {
                "link_id": "nd2_nd4",
                "source": "nd2",
                "target": "nd4",
                "source_port": "nd2_port1",
                "target_port": "nd4_port0"
            },
            {
                "link_id": "nd2_nd5",
                "source": "nd2",
                "target": "nd5",
                "source_port": "nd2_port2",
                "target_port": "nd5_port0"
            },
            {
                "link_id": "nd3_nd6",
                "source": "nd3",
                "target": "nd6",
                "source_port": "nd3_port1",
                "target_port": "nd6_port0"
            },
            {
                "link_id": "nd3_nd7",
                "source": "nd3",
                "target": "nd7",
                "source_port": "nd3_port2",
                "target_port": "nd7_port0"
            }
        ],
        "nodes": [
            {
                "node_id": "nd1",
                "name": "node_1",
                "flavor": {
                    "id": FLAVOR_ID,
                    "name": "c4.2xlarge",
                    "cpu": 4,
                    "memory": 2,
                    "storage": 8
                },
                "image": IMAGE_ID,
                "security_rules": [22],
                "ports": [
                    { "id": "nd1_port0" },
                    { "id": "nd1_port1" }
                ]
            },
            {
                "node_id": "nd2",
                "name": "node_2",
                "flavor": {
                    "id": FLAVOR_ID,
                    "name": "c4.2xlarge",
                    "cpu": 4,
                    "memory": 2,
                    "storage": 8
                },
                "image": IMAGE_ID,
                "security_rules": [22],
                "ports": [
                    { "id": "nd2_port0" },
                    { "id": "nd2_port1" },
                    { "id": "nd2_port2" }
                ]
            },
            {
                "node_id": "nd3",
                "name": "node_3",
                "flavor": {
                    "id": FLAVOR_ID,
                    "name": "c4.2xlarge",
                    "cpu": 4,
                    "memory": 2,
                    "storage": 8
                },
                "image": IMAGE_ID,
                "security_rules": [22],
                "ports": [
                    { "id": "nd3_port0" },
                    { "id": "nd3_port1" },
                    { "id": "nd3_port2" }
                ]
            },
            {
                "node_id": "nd4",
                "name": "node_4",
                "flavor": {
                    "id": FLAVOR_ID,
                    "name": "c4.2xlarge",
                    "cpu": 4,
                    "memory": 2,
                    "storage": 8
                },
                "image": IMAGE_ID,
                "security_rules": [22],
                "ports": [
                    { "id": "nd4_port0" }
                ]
            },
            {
                "node_id": "nd5",
                "name": "node_5",
                "flavor": {
                    "id": FLAVOR_ID,
                    "name": "c4.2xlarge",
                    "cpu": 4,
                    "memory": 2,
                    "storage": 8
                },
                "image": IMAGE_ID,
                "security_rules": [22],
                "ports": [
                    { "id": "nd5_port0" }
                ]
            },
            {
                "node_id": "nd6",
                "name": "node_6",
                "flavor": {
                    "id": FLAVOR_ID,
                    "name": "c4.2xlarge",
                    "cpu": 4,
                    "memory": 2,
                    "storage": 8
                },
                "image": IMAGE_ID,
                "security_rules": [22],
                "ports": [
                    { "id": "nd6_port0" }
                ]
            },
            {
                "node_id": "nd7",
                "name": "node_7",
                "flavor": {
                    "id": FLAVOR_ID,
                    "name": "c4.2xlarge",
                    "cpu": 4,
                    "memory": 2,
                    "storage": 8
                },
                "image": IMAGE_ID,
                "security_rules": [22],
                "ports": [
                    { "id": "nd7_port0" }
                ]
            }
        ]
    }
}"""



def create_slice_topology(topology_json, project_token, project_id):
    #admin_token = authenticate_admin()
    #project_token = authenticate_project(admin_token)

    
    subnet_cidr_template = '10.0.{subnet_index}.0/24'

    
    
    links = topology_json['links']
    links_temp = {}

    for i,link in enumerate(links, start=1):
        subnet_cidr = subnet_cidr_template.format(subnet_index=i)
       
        network_name = link['link_id']
        network_id, subnet_id = create_network_and_subnet(project_token, network_name,subnet_cidr)
    
       
        port_name0 = link['source']
        puerto0_id = create_instance_port(project_token,network_id, port_name0,project_id)

        port_name1 = link['target']
        puerto1_id = create_instance_port(project_token,network_id, port_name1,project_id)
     

        links_temp[link['link_id']] = {
            "network": network_id,
            "subnet": subnet_id,
            "puerto0": puerto0_id,
            "puerto1": puerto1_id
        }
    
        
    nodes = topology_json['nodes']

    for node in nodes:
        instance_name = node['name']
        flavor_id = node['flavor']['id']
        image_id = node['image']
        networks = []
        for port in node['ports']:
            link = find_link_by_source_port(port['id'], links)
            if link is None:
                link = find_link_by_target_port(port['id'], links)
                port = links_temp[link['link_id']]["puerto1"]
            else:
                port = links_temp[link['link_id']]["puerto0"]
                
            networks.append({"port": port})
        print("networks: ",networks)
        launch_instance(project_token, instance_name, image_id, flavor_id, networks)
        time.sleep(3)
        

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

#create_slice_topology()





async def delete_instances_by_networkid(token, network_id):
    instances_resp = list_instances(NOVA_ENDPOINT, token, PROJECT_ID)
    network_original_info = get_network_by_id(NEUTRON_ENDPOINT, token, network_id)
    if instances_resp.status_code == 200:
        instances = instances_resp.json()['servers']
        #print(instances)
        for instance in instances:
            # Verificar si la instancia está asociada a la red específica
            instance_detail_resp = get_instance_detail(NOVA_ENDPOINT, token, instance['id'])
            if instance_detail_resp.status_code == 200:
                instance_detail = instance_detail_resp.json()['server']
                for network in instance_detail['addresses']:
                    #print(network)
                    if network == network_original_info['name']:
                        delete_resp = delete_server(NOVA_ENDPOINT, token, instance['id'])
                        print(f'Deleted instance: {instance['id']} for {network}')
                        if delete_resp.status_code != 204:
                            raise Exception(f'Failed to delete instance {instance["id"]}')
    else:
        raise Exception('Failed to list instances')

# Función para listar y borrar puertos
async def delete_ports(token, network_id):
    ports_resp = list_ports(NEUTRON_ENDPOINT, token, network_id)
    if ports_resp.status_code == 200:
        ports = ports_resp.json()['ports']
        for port in ports:
            delete_resp = delete_port(NEUTRON_ENDPOINT, token, port['id'])
            if delete_resp.status_code != 204:
                raise Exception(f'Failed to delete port {port["id"]}')
            else:
                #print('Deleted port: ',port['id'])
                pass
    else:
        raise Exception('Failed to list ports')

# Función para listar y borrar subredes
async def delete_subnets(token, network_id):
    subnets_resp = list_subnets(NEUTRON_ENDPOINT, token, network_id)
    if subnets_resp.status_code == 200:
        subnets = subnets_resp.json()['subnets']
        for subnet in subnets:
            delete_resp = delete_subnet(NEUTRON_ENDPOINT, token, subnet['id'])
            if delete_resp.status_code != 204:
                raise Exception(f'Failed to delete subnet {subnet["id"]}')
            else:
                #print('Deleted subnet: ',subnet['id'])
                pass
    else:
        raise Exception('Failed to list subnets')

# Función para borrar una red
async def delete_network_by_id(token, network_id):
    delete_resp = delete_network(NEUTRON_ENDPOINT, token, network_id)
    if delete_resp.status_code != 204:
        raise Exception(f'Failed to delete network {network_id}')
    else:
        print('Deleted link with ID: ',network_id)

def list_networks_slice(project_id,token):
    list_resp = list_networks(NEUTRON_ENDPOINT, token,project_id)
    if list_resp.status_code == 200:
        networks_list = list_resp.json()['networks']
        print(f'Networks list of project {project_id} obtained')
        return networks_list
    else:
        raise Exception(f'Failed to list networks of project {project_id}')


# Función principal para borrar la topología
async def delete_slice_topology():
    admin_token = authenticate_admin()
    project_token = authenticate_project(admin_token)
    print(project_token)
   
    networks = list_networks_slice(PROJECT_ID,project_token)
    #networks_id = []
    if networks:
        for network in networks:
            #networks_id.append(network['id'])  
            network_id = network['id']
            # Borrar instancias
            await delete_instances_by_networkid(project_token, network_id)

            # Borrar puertos
            await delete_ports(project_token, network_id)

            # Borrar subredes
            await delete_subnets(project_token, network_id)

            # Borrar red
            await delete_network_by_id(project_token, network_id)
            time.sleep(0.25)
        print("Ring network deleted.")
    else:
        print(f'Project with id {PROJECT_ID} do not have links')
  

# Llamar a la función principal para borrar la topología
#network_id = '551f8238-1464-44dc-bc19-d448fdb46eb2'  # Reemplaza con el ID de la red que deseas borrar
#asyncio.run(delete_slice_topology())


def asignarRoleUsuarioPorProject(admin_token, project_id, user_id, role_id):
    try:
        resp = assign_role_to_user(KEYSTONE_ENDPOINT, admin_token,project_id,user_id,role_id)
        print(resp.status_code)
        if resp.status_code == 204:
            print('ROLE ASSIGNED SUCCESSFULLY')
            return True
        else:
            print('ROLE NOT ASSIGNED')
            return None
    except:
        return None
    
def crearProject(admin_token, domain_id, project_name, project_description):
    try:
        resp = create_project(KEYSTONE_ENDPOINT, admin_token, domain_id, project_name, project_description)
        print(resp.status_code)
        if resp.status_code == 201:
            print('PROJECT CREATED SUCCESSFULLY')
            project_created = resp.json()
            print(json.dumps(project_created))
            return project_created
        else:
            print('FAILED PROJECT CREATION')
            return None
    except:
        return None
    

def obtainUserId(admin_token, user_name):
    try:
        resp = get_users(KEYSTONE_ENDPOINT, admin_token)
        print(resp.status_code)
        if resp.status_code == 200:
            print('USERS OBTAINED SUCCESSFULLY')
            users = resp.json()
            print(json.dumps(users))
            for user in users["users"]:
                if user['name'] == user_name:
                    print("ID del Usuario: ", user['id'])
                    return user['id']
            print('USER NOT FOUND')
            return None
        else:
            print('FAILED USERS OBTAINMENT')
            return None
    except:
        return None


def crear_proyecto(admin_token, domain_id, project_name, project_description):
    try:
        resp = create_project(KEYSTONE_ENDPOINT,admin_token, domain_id, project_name, project_description)
        print(resp.status_code)
        if resp.status_code == 201:
            print('PROJECT CREATED SUCCESSFULLY')
            project_created = resp.json()
            print(json.dumps(project_created))
            return project_created
        else:
            print('FAILED PROJECT CREATION')
            return None
    except:
        return None


def openstackDeployment(slice_json, user_name): 
    project_name = slice_json['name']
    project_description = slice_json['description']
    print(project_name)
    print(project_description)
    admin_token = authenticate_admin() 
    try:
        if admin_token:
            #2.- Crear el proyecto
            project = crear_proyecto(admin_token, DOMAIN_ID, project_name, project_description)
            if project:
                project_id = project["project"]["id"]
                rol_admin = asignarRoleUsuarioPorProject(admin_token, project_id, ADMIN_USER_ID, ADMIN_ROLE_ID)
                if rol_admin:
                    project_token = authenticate_project(admin_token, project_name)
                    if project_token:
                        user_id = obtainUserId(project_token, user_name)
                        if user_id:
                            rol_user = asignarRoleUsuarioPorProject(project_token, project_id, user_id, MEMBER_ROLE_ID)
                            if rol_user:
                                create_slice_topology(slice_json['topology'], project_token,project_id)
    except Exception as e:
        print(e)
        print("Slice deployment failed")
        


        