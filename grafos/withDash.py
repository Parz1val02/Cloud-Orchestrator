from dash import Dash, html, dcc, Input, Output, State
import dash_cytoscape as cyto
import networkx as nx
import json
import requests


app = Dash(__name__)


app.layout = html.Div([

    

    html.H1("Network Topology Generator"),
    
    

    html.Label("Escoja la topologia a generar:"),
    dcc.Dropdown(
        id='topology-type-dropdown',
        options=[
            {'label': 'Lineal', 'value': 'Lineal'},
            {'label': 'Malla', 'value': 'Malla'},
            {'label': 'Arbol', 'value': 'Arbol'},
            {'label': 'Anillo', 'value': 'Anillo'},
            {'label': 'Bus', 'value': 'Bus'}
        ],
        value='Escoja una topologia a armar' 
    ),
    
    html.Label("Ingrese el numero de nodos: "),
    dcc.Input(id='num-nodes-input', type='number', value="Numero de nodos"),

   
    
    html.Button('Generate Topology', id='generate-button', n_clicks=0 ),
    
    html.Div(id='cytoscape-container'),

    html.Div(id='json-output'),
    html.Button('Save JSON Data', id='save-button', n_clicks=0)
    
    



    
])


@app.callback(
    
    Output('cytoscape-container', 'children'),
    
    [Input('generate-button', 'n_clicks')],
    [State('topology-type-dropdown', 'value'),
     State('num-nodes-input', 'value')]
)
def generate_topology(n_clicks, topology_type, num_nodes):
    if n_clicks > 0:
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
        elif topology_type == 'Bus':
            for i in range(1, num_nodes + 1):
                for j in range(i + 1, num_nodes + 1):
                    G.add_edge(i, j)
        
        elements = [{'data': {'id': str(node), 'label': str(node)}} for node in G.nodes()]#nodos
        elements += [{'data': {'source': str(edge[0]), 'target': str(edge[1])}} for edge in G.edges()]#links
        print(elements)

        #------ Json hardcodeado para guardar nodes y links
        json_data = {
            "_id": {"$oid": "6640688d53c1187a6899a5c5"},
            "user_id": "6640550a53c1187a6899a5a9",
            "deployed": False,
            "name": "Template Estrella",
            "description": "Template con topología tipo estrella",
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

        return cyto.Cytoscape(
            id='network-graph',
            elements=elements,
            layout={'name': 'grid'},
            style={'width': '100%', 'height': '600px'}
        )
    else:
        return html.Div()

@app.callback(
    Output('json-output', 'children'),
    [Input('save-button', 'n_clicks')],
    [State('json-output', 'children')]  
)
def save_json(n_clicks, json_data):
    if n_clicks > 0:
        
        endpoint = 'http://127.0.0.1:27017/templates/6640688d53c1187a6899a5c5'
        
       
        response = requests.put(endpoint, json=json_data)
        
        if response.status_code == 200:
            return "JSON data saved successfully!"
        else:
            return "Error: Failed to save JSON data"
    else:
        return ""


if __name__ == '__main__':
    app.run_server(debug=True)


