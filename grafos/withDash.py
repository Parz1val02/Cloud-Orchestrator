from dash import Dash, html, dcc, Input, Output, State
import dash_cytoscape as cyto
import networkx as nx
import json
from flask import request, jsonify
import requests

app = Dash(__name__)
server = app.server

latest_topology_data = None

app.layout = html.Div([

    

    html.H1("Network Topology Generator"),
    html.Div(id='cytoscape-container')


    
])


@app.callback(
    
    Output('cytoscape-container', 'children'),
     Input('cytoscape-container', 'id')
   
)
def generate_topology(_):
    #print(json_data)
    global latest_topology_data
    if latest_topology_data:
        data = latest_topology_data
        topology_type = data.get('topology_type')
        num_nodes = data.get('num_nodes')

        if topology_type is None or num_nodes is None:
            return html.Div("Error: Missing topology_type or num_nodes in JSON data")
        G = nx.Graph()
        
        if topology_type == 'Lineal':
            edge_list = [(i, i + 1) for i in range(num_nodes - 1)]
            G.add_edges_from(edge_list)
        elif topology_type == 'Malla':

            for i in range(num_nodes):
                for j in range (num_nodes):
                    if i != j:
                        G.add_edge(i, j)

        
        elif topology_type == 'Arbol':
            
            G = nx.balanced_tree(2, int(num_nodes ** 0.5))
        elif topology_type == 'Anillo':
            edge_list = [(i, i + 1) for i in range(num_nodes - 1)]
            edge_list.append((0, num_nodes - 1))
            G.add_edges_from(edge_list)
        
        elements = [{'data': {'id': str(node), 'label': str(node)}} for node in G.nodes()]#nodos
        elements += [{'data': {'source': str(edge[0]), 'target': str(edge[1])}} for edge in G.edges()]#links
        print(elements)

        #------ Json hardcodeado para guardar nodes y links
        json_data = {
            "_id": {"$oid": "6640688d53c1187a6899a5c5"},
            "user_id": "6640550a53c1187a6899a5a9",
            "deployed": False,
            "name": "Template Estrella",
            "description": "Template con topologia tipo estrellaaa",
            "vlan_id": "vlan_id_estrella",
            "topology": {
                "nodes": [],
                "links": [],
                "specifications": {
                    "central_node": {
                        "cpu": 4,
                        "memory": 8,
                        "storage": 100,
                        "image": "Ubuntu",
                        "access_protocol": "SSH",
                        "security_rules": [22]
                    },
                    "peripheral_node1": {
                        "cpu": 2,
                        "memory": 4,
                        "storage": 50,
                        "image": "Ubuntu",
                        "access_protocol": "SSH",
                        "security_rules": [22]
                    },
                    "peripheral_node2": {
                        "cpu": 2,
                        "memory": 4,
                        "storage": 50,
                        "image": "Ubuntu",
                        "access_protocol": "SSH",
                        "security_rules": [22]
                    },
                    "peripheral_node3": {
                        "cpu": 2,
                        "memory": 4,
                        "storage": 50,
                        "image": "Ubuntu",
                        "access_protocol": "SSH",
                        "security_rules": [22]
                    }
                }
            },
            "availability_zone": "zone_estrella"
        }


        # Nodos y links que iran al json
        nodes = []
        links = []
        for element in elements:
            if 'source' not in element['data']:
                nodes.append(element['data']['id'])
            else:
                links.append({'source': element['data']['source'], 'target': element['data']['target']})

        # Update the JSON structure with nodes and links
        json_data['topology']['nodes'] = nodes
        json_data['topology']['links'] = links

        # Convert the dictionary to JSON
        json_string = json.dumps(json_data, indent=4)
        print(json_string)
        #------
        
        #json_string = json.dumps(json_data, indent=4)

        return cyto.Cytoscape(
            id='network-graph',
            elements=elements,
            layout={'name': 'grid'},
            style={'width': '100%', 'height': '500px'}
        )
    
        
    else:
        return html.Div()


'''
@app.callback(
    Output('json-output', 'children'),
    [Input('save-button', 'n_clicks')]
    
)
def save_json(n_clicks):
    if n_clicks > 0:
        
        #json_data = json_string

        json_data = {
                    "_id": {
                        "$oid": "6640688d53c1187a6899a5c5"
                    },
                    "user_id": "6640550a53c1187a6899a5a9",
                    "deployed": 0,
                    "name": "Template Estrella",
                    "description": "Template con topologia tipo estrellaaa",
                    "vlan_id": "vlan_id_estrella",
                    "topology": {
                        "nodes": [
                            "0",
                            "1",
                            "2"
                        ],
                        "links": [
                            {
                                "source": "0",
                                "target": "1"
                            },
                            {
                                "source": "1",
                                "target": "2"
                            }
                        ],
                        "specifications": {
                            "central_node": {
                                "cpu": 4,
                                "memory": 8,
                                "storage": 100,
                                "image": "Ubuntu",
                                "access_protocol": "SSH",
                                "security_rules": [
                                    22
                                ]
                            },
                            "peripheral_node1": {
                                "cpu": 2,
                                "memory": 4,
                                "storage": 50,
                                "image": "Ubuntu",
                                "access_protocol": "SSH",
                                "security_rules": [
                                    22
                                ]
                            },
                            "peripheral_node2": {
                                "cpu": 2,
                                "memory": 4,
                                "storage": 50,
                                "image": "Ubuntu",
                                "access_protocol": "SSH",
                                "security_rules": [
                                    22
                                ]
                            },
                            "peripheral_node3": {
                                "cpu": 2,
                                "memory": 4,
                                "storage": 50,
                                "image": "Ubuntu",
                                "access_protocol": "SSH",
                                "security_rules": [
                                    22
                                ]
                            }
                        }
                    }
        }

        endpoint = 'http://127.0.0.1:5000/templates/6640688d53c1187a6899a5c5'

        json_string = json.dumps(json_data, indent=4)

        response = requests.put(endpoint, json=json_string)
        
        if response.status_code == 200:
            return "JSON data saved successfully!"
        else:
            print(response.text)
            return "Error: Failed to save JSON data"
    else:
        return ""
'''

@server.route('/api/generate_topology', methods=['POST'])
def generate_topology_api():
    global latest_topology_data
    if request.method == 'POST':
        data = request.get_json()
        #print(data)
        topology_type = data.get('topology_type')
        num_nodes = data.get('num_nodes')
        if not topology_type or not num_nodes:
            return jsonify({"error": "Invalid input"}), 400
        latest_topology_data = data
        return jsonify(data), 200



if __name__ == '__main__':
    app.run_server(debug=True)


