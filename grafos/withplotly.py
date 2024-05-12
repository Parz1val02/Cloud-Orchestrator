import networkx as nx
import plotly.graph_objects as go

print("Ingrese la topologia a crear: \n a) Lineal \n b) Malla  \n c) Arbol  \n d) Anillo  \n e) Bus")
opcion = input("Ingrese su opcion: ")


if opcion != "c":
    numero_nodos = int(input("Ingrese el numero de nodos para su topologia: "))

G = nx.Graph()

if opcion == "a":
    edge_list = [(i+1, i+2) for i in range(numero_nodos-1)]
    print(edge_list)
    print("Lineal")
    topologia = "Lineal"
    G.add_edges_from(edge_list)
elif opcion == "b":
    print("Malla")
    topologia = "Malla"
    for i in range(numero_nodos):
        for j in range(numero_nodos):
            if i != j:
                G.add_edge(i+1, j+1)
elif opcion == "c":
    niveles = int(input("Ingrese el numero de niveles de su topologia arbol: "))
    ramas = int(input("Ingrese la cantidad de ramas por nodo de sus topologia arbol: "))

    node_levels = []
    node_names = ["a1"]
    number_nodes_for_level = [1]

    for n in range(niveles-1):
        number_nodes_level = number_nodes_for_level[-1]*ramas
        number_nodes_for_level.append(number_nodes_level)
    
    print(number_nodes_for_level)
            
    for i in range(niveles):
        node_level = chr(i+1+96)
        node_levels.append(node_level)

    print(node_levels)

    dicc = dict(zip(range(1,len(node_levels)+1), number_nodes_for_level))

    print(dicc)

    for x in range(len(node_levels)-1): 
        for y in number_nodes_for_level: 
            if dicc.get(x+2) == y:
                for z in range(y):
                    node_name = node_levels[x+1] + str(z+1) 
                    node_names.append(node_name)
    
    print(node_names)
    
    for node_name in node_names:
        G.add_node(node_name)

    for i in range(len(node_names)):
        level = node_names[i][0]  # Obtener el nivel del nombre del nodo
        if level != 'a':  # No agregar bordes para el primer nivel (raíz)
            parent_level = chr(ord(level) - 1)  # Nivel del nodo padre
            parent_nodes = [node for node in node_names if node.startswith(parent_level)]
            parent_node_index = (int(node_names[i][1:]) - 1) // ramas - 1  # El índice del nodo padre es el número en el nombre del nodo menos uno
            parent_node = parent_nodes[parent_node_index]
            G.add_edge(parent_node, node_names[i])

    print("Arbol")
    topologia = "Arbol"
elif opcion == "d":
    edge_list = [(i+1, i+2) for i in range(numero_nodos-1)]
    print(edge_list)
    edge_list.append((1, numero_nodos))
    G.add_edges_from(edge_list)
    print("Anillo")
    topologia = "Anillo"
elif opcion == "e":
    print("Bus")




# Drawing the graph
pos = nx.spring_layout(G)  # Compute the position of the nodes
edge_trace = go.Scatter(x=[], y=[], line=dict(width=0.5, color='#888'), hoverinfo='none', mode='lines')
for edge in G.edges():
    x0, y0 = pos[edge[0]]
    x1, y1 = pos[edge[1]]
    edge_trace['x'] += tuple([x0, x1, None])
    edge_trace['y'] += tuple([y0, y1, None])

node_trace = go.Scatter(x=[], y=[], text=[], mode='markers', hoverinfo='text', marker=dict(showscale=False, colorscale='YlGnBu', reversescale=True, color=[], size=10, colorbar=dict(thickness=15, title='Node Connections', xanchor='left', titleside='right'), line_width=2))

for node in G.nodes():
    x, y = pos[node]
    node_trace['x'] += tuple([x])
    node_trace['y'] += tuple([y])
    node_trace['text'] += tuple([node])

fig = go.Figure(data=[edge_trace, node_trace], layout=go.Layout(title='<br>'+topologia, titlefont_size=40, showlegend=False, hovermode='closest', margin=dict(b=20, l=5, r=5, t=40) ,annotations=[dict(text=" ", showarrow=False, xref="paper", yref="paper", x=0.005, y=-0.002)]))
fig.show()