import json, requests

# KEYSTONE API
def password_authentication_with_scoped_authorization(auth_endpoint, user_domain_name, username, password, project_domain_id, project_name):
    url = auth_endpoint + '/auth/tokens'

    data = \
        {
            "auth": {
                "identity": {
                    "methods": [
                        "password"
                    ],
                    "password": {
                        "user": {
                            "name": username,
                            "domain": {
                                "name": user_domain_name
                            },
                            "password": password
                        }
                    }
                },
                "scope": {
                    "project": {
                        "domain": {
                            "id": project_domain_id
                        },
                        "name": project_name
                    }
                }
            }
        }
        
    r = requests.post(url=url, data=json.dumps(data))
    # status_code success = 201
    return r

def token_authentication_with_scoped_authorization(auth_endpoint, token, project_domain_id, project_name):
    url = auth_endpoint + '/auth/tokens'

    data = \
        {
            "auth": {
                "identity": {
                    "methods": [
                        "token"
                    ],
                    "token": {
                        "id": token
                    }
                },
                "scope": {
                    "project": {
                        "domain": {
                            "id": project_domain_id
                        },
                        "name": project_name
                    }
                }
            }
        }

    r = requests.post(url=url, data=json.dumps(data))
    # status_code success = 201
    return r


def password_authentication_with_unscoped_authorization(auth_endpoint, domain_id, username, password):
    url = auth_endpoint + '/auth/tokens'    

    data = \
        {
            "auth": {
                "identity": {
                    "methods": [
                        "password"
                    ],
                    "password": {
                        "user": {
                            "name": username,
                            "domain": {
                                "id": domain_id
                            },
                            "password": password
                        }
                    }
                }
            }
        }
    try:
        r = requests.post(url=url, data=json.dumps(data))
    except requests.exceptions.RequestException as e:
        print("Error en la solicitud: ",e)
        r=None
    return r

def get_users(auth_endpoint, token):
                
    url = auth_endpoint + '/users'
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}

    r = requests.get(url=url, headers=headers)
    # status_code success = 200
    return r

def create_project(auth_endpoint, token, domain_id, project_name, project_description):
        
    url = auth_endpoint + '/projects'
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}

    data = \
        {
            "project": {
                "name": project_name,
                "description": project_description,
                "domain_id": domain_id
            }
        }

    r = requests.post(url=url, headers=headers, data=json.dumps(data))
    # status_code success = 201
    return r

def list_projects(auth_endpoint, token, user_id):
    
    url = auth_endpoint + '/users/' + user_id + '/projects'
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}

    r = requests.get(url=url, headers=headers)
    return r



def assign_role_to_user(auth_endpoint, token, project_id, user_id, role_id):
            
    url = auth_endpoint + '/projects/' + project_id + '/users/' + user_id + '/roles/' + role_id
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}

    r = requests.put(url=url, headers=headers)
    # status_code success = 204
    return r



# NEUTRON API
#def create_network(auth_endpoint, token, name, network_type=None, segmentation_id=None):
def create_network(auth_endpoint, token, name):
    url = auth_endpoint + '/networks'
    data = \
        {
            "network": {
                "name": name,
                "port_security_enabled": "false",
            }
        }

    '''
    if network_type is not None:
        data['network']["provider:network_type"] = network_type

    if segmentation_id is not None:
        data["network"]["provider:segment"] = segmentation_id
    '''
    
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}
    r = requests.post(url=url, headers=headers, data=json.dumps(data))
    # status_code success = 201
    return r

def list_networks(neutron_endpoint, token, project_id):
    url = f"{neutron_endpoint}/networks"
    headers = {
        'Content-Type': 'application/json',
        'X-Auth-Token': token
    }
    
    params = {'tenant_id': project_id}
    response = requests.get(url, headers=headers,params=params)

    return response



def create_subnet(auth_endpoint, token, network_id, name, ip_version, cidr):
        
    url = auth_endpoint + '/subnets'
    data = \
        {
            "subnet": {
                "network_id": network_id,
                "name": name,
                "enable_dhcp": False,
                "gateway_ip": None,
                "ip_version": ip_version,
                "cidr": cidr
            }
        }

    data = data=json.dumps(data)

    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}
    r = requests.post(url=url, headers=headers, data=data)
    # status_code success = 201
    return r


def list_subnets(neutron_endpoint, token, network_id):
    url = neutron_endpoint + '/subnets'
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}

    params = {'network_id': network_id}

    r = requests.get(url=url, headers=headers, params=params)
    return r

def list_ports(neutron_endpoint, token, network_id):
    url = neutron_endpoint + '/ports'
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}

    params = {'network_id': network_id}

    r = requests.get(url=url, headers=headers, params=params)
    return r

def create_port(auth_endpoint, token, name, network_id, project_id):
        
    url = auth_endpoint + '/ports'
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}

    data = \
        {
            'port': {
                'name': name,
                'tenant_id': project_id,
                'network_id': network_id,
                'port_security_enabled': 'false'
            }
        }

    r = requests.post(url=url, headers=headers, data=json.dumps(data))
    # status_code success = 201
    return r

# NOVA API
#def create_instance(auth_endpoint, token, name, flavorRef, imageRef=None, availability_zone=None, network_list=None, compute_version=None):

def get_instance_console(nova_endpoint, token, server_id, compute_api_version):
    url = nova_endpoint + '/servers/' + server_id + '/remote-consoles'
    headers = {
        'Content-type': 'application/json',
        'X-Auth-Token': token,
        "OpenStack-API-Version": "compute " + compute_api_version
    }
    
    data = \
        {
            "remote_console": {
                "protocol": "vnc",
                "type": "novnc"
                }
        }
    
    r = requests.post(url=url, headers=headers, data=json.dumps(data))
    # status_code success = 200
    return r


def get_instance_detail(nova_endpoint, token, instance_id):
    url = nova_endpoint + f'/servers/{instance_id}'
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}

    r = requests.get(url=url, headers=headers)
    return r


def create_instance(auth_endpoint, token, name, flavorRef, imageRef, network_list):
    url = auth_endpoint + '/servers'
    headers = {
        'Content-type': 'application/json',
        'X-Auth-Token': token,
    }
    '''
    if compute_version is not None:
        headers['OpenStack-API-Version'] = 'compute ' + compute_version
    '''
    
    data = \
        {
            'server': {
                'name': name,
                'flavorRef': flavorRef,
                'imageRef': imageRef,
                #'availability_zone': availability_zone,
                'networks': network_list,
                
            }
        }

    '''
    if imageRef is not None:
        data['server']['imageRef'] = imageRef

    if availability_zone is not None:
        data['server']['availability_zone'] = availability_zone

    if network_list is not None:
        data['server']['networks'] = network_list

    if volume_list is not None:
        data['server']['block_device_mapping'] = volume_list
    '''
    
    r = requests.post(url=url, headers=headers, data=json.dumps(data))
    # status_code success = 202
    return r

def list_instances(nova_endpoint, token, project_id=None):
    #url = nova_endpoint + '/servers'
    url = nova_endpoint + '/servers/detail'
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}

    params = {'all_tenants': True}
    if project_id:
        params = {'all_tenants': True, 'project_id': project_id}

    r = requests.get(url=url, headers=headers, params=params)
    return r




def get_network_by_id(neutron_endpoint, token, network_id):
    url = f"{neutron_endpoint}/networks/{network_id}"
    headers = {
        'Content-Type': 'application/json',
        'X-Auth-Token': token
    }
    
    response = requests.get(url, headers=headers)

    if response.status_code == 200:
        network = response.json()['network']
        return network
    else:
        print(f"Error al obtener la red: {response.status_code} - {response.text}")
        return None

def list_flavors(nova_endpoint, token):
    url = nova_endpoint + '/flavors/detail'
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}
    r = requests.get(url=url, headers=headers)
    # status_code success = 200
    return r

def list_images(auth_endpoint, token, limit=None):
    url = auth_endpoint + '/images'
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}

    if limit is not None:
        params = {'limit': limit}
        r = requests.get(url=url, headers=headers, params=params)
    else:
        r = requests.get(url=url, headers=headers)
    return r

def delete_subnet(auth_endpoint, token, subnet_id):
    url = auth_endpoint + '/subnets/' + subnet_id
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}

    r = requests.delete(url=url, headers=headers)
    # status_code success = 204
    return r

def delete_network(auth_endpoint, token, network_id):
    url = auth_endpoint + '/networks/' + network_id
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}

    r = requests.delete(url=url, headers=headers)
    # status_code success = 204
    return r

def monitor_server_assignment(auth_endpoint, token):
    url = auth_endpoint + '/os-hypervisors/detail'
    print(url)
    print(token)
    headers = {
        'Content-type': 'application/json',
        'X-Auth-Token': token,
        #'X-OpenStack-Nova-API-Version': '2.1',
        #"OpenStack-API-Version": "compute 2.1",
        
    }
    
    r = requests.get(url=url, headers=headers)
    # status_code success = 200
    return r

def delete_server(auth_endpoint, token, server_id):
    url = auth_endpoint + '/servers/' + server_id
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}

    r = requests.delete(url=url, headers=headers)
    # status_code success = 204
    return r

def delete_port(auth_endpoint, token, port_id):
    url = auth_endpoint + '/ports/' + port_id
    headers = {'Content-type': 'application/json', 'X-Auth-Token': token}

    r = requests.delete(url=url, headers=headers)
    # status_code success = 204
    return r