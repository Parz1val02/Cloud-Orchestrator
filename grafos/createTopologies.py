import networkx as nx
import matplotlib.pyplot as plt


print("Ingrese la topologia a crear: \n a) Lineal \n b) Malla  \n c) Arbol  \n d) Anillo  \n e) Bus" )
opcion = input("Ingrese su opcion: ")

if((opcion) != "c"):
    numero_nodos = int(input("Ingrese el numero de nodos para su topologia: "))

G = nx.Graph()


if opcion == "a":
    edge_list = []
    for  i in range(numero_nodos-1):
        edge_list.append((i+1, i+2))
    print(edge_list)
    print("Lineal")
    G.add_edges_from(edge_list)
    nx.draw_spring(G, with_labels=True)
    plt.show()
elif opcion == "b":
    print("Malla")


    for i in range(numero_nodos):
        for j in range(numero_nodos):
            if i != j:
                G.add_edge(i+1,j+1)


    nx.draw_spring(G, with_labels=True)
    plt.show()
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
            if(dicc.get(x+2) == y):
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

    nx.draw(G, with_labels=True)
    plt.show()

    print("Arbol")
elif opcion == "d":
    edge_list = []
    for  i in range(numero_nodos-1):
        edge_list.append((i+1, i+2))
    print(edge_list)
    
    edge_list.append((1,numero_nodos))

    G.add_edges_from(edge_list)
    nx.draw_spring(G, with_labels=True)
    plt.show()


    print("Anillo")
elif opcion == "e":
    print("Bus")



