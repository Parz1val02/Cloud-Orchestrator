from dash import Dash, html, dcc, Input, Output, State
import dash_cytoscape as cyto
import networkx as nx

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

    
    html.Button('Generate Topology', id='generate-button', n_clicks=0),
    
    html.Div(id='cytoscape-container')
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
        return cyto.Cytoscape(
            id='network-graph',
            elements=elements,
            layout={'name': 'grid'},
            style={'width': '100%', 'height': '600px'}
        )
    else:
        return html.Div()


if __name__ == '__main__':
    app.run_server(debug=True)
