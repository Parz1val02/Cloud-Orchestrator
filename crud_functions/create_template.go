package crud_functions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Node struct {
	NodeID         string `json:"node_id"`
	Name           string `json:"name"`
	AccessProtocol string `json:"access_protocol"`
	CPU            int    `json:"cpu"`
	Image          string `json:"image"`
	Memory         int    `json:"memory"`
	SecurityRules  []int  `json:"security_rules"`
	Storage        int    `json:"storage"`
}

type Link struct {
	LinkID string `json:"link_id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type Topology struct {
	Links []Link `json:"links"`
	Nodes []Node `json:"nodes"`
}

type Template struct {
	CreatedAt        time.Time `json:"created_at"`
	AvailabilityZone string    `json:"availability_zone"`
	Deployed         bool      `json:"deployed"`
	Description      string    `json:"description"`
	Name             string    `json:"name"`
	Topology         Topology  `json:"topology"`
	UserID           string    `json:"user_id"`
	TopologyType     string    `json:"topology_type"`
}

func promptString(promptText string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(promptText)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func promptInt(promptText string) int {
	var input int
	fmt.Print(promptText)
	fmt.Scanln(&input)
	return input
}

func promptTopology() string {
	topologies := []string{"malla", "arbol binario", "arbol general", "lineal", "anillo", "estrella"}
	fmt.Println("Choose a topology:")
	for i, topology := range topologies {
		fmt.Printf("%d. %s\n", i+1, topology)
	}
	choice := promptInt("Enter choice number: ")
	return topologies[choice-1]
}

func createPredefinedTopology(topologyType string) Topology {
	switch topologyType {
	case "lineal":
		return createLinealTopology()
	case "malla":
		return createMeshTopology()
	case "arbol binario":
		return createBinaryTreeTopology()
	case "arbol general":
		return createGeneralTreeTopology()
	case "anillo":
		return createRingTopology()
	case "estrella":
		return createStarTopology()
	}
	return Topology{}
}

func createLinealTopology() Topology {
	numNodes := promptInt("Enter the number of nodes: ")
	nodes := createNodes(numNodes)
	var links []Link
	for i := 0; i < numNodes-1; i++ {
		links = append(links, Link{
			LinkID: fmt.Sprintf("nd%d_nd%d", i+1, i+2),
			Source: nodes[i].Name,
			Target: nodes[i+1].Name,
		})
	}
	return Topology{Nodes: nodes, Links: links}
}

func createMeshTopology() Topology {
	numNodes := promptInt("Enter the number of nodes: ")
	nodes := createNodes(numNodes)
	var links []Link
	for i := 0; i < numNodes; i++ {
		for j := i + 1; j < numNodes; j++ {
			links = append(links, Link{
				LinkID: fmt.Sprintf("nd%d_nd%d", i+1, j+1),
				Source: nodes[i].Name,
				Target: nodes[j].Name,
			})
		}
	}
	return Topology{Nodes: nodes, Links: links}
}

// stric binary tree topology
func createBinaryTreeTopology() Topology {
	numLevels := promptInt("Enter the number of levels: ")
	numNodes := (1 << numLevels) - 1 // 2^levels - 1
	nodes := createNodes(numNodes)
	var links []Link
	for i := 0; i < (1<<(numLevels-1))-1; i++ {
		leftChild := 2*i + 1
		rightChild := 2*i + 2
		if leftChild < len(nodes) {
			links = append(links, Link{
				LinkID: fmt.Sprintf("nd%d_nd%d", i+1, leftChild+1),
				Source: nodes[i].Name,
				Target: nodes[leftChild].Name,
			})
		}
		if rightChild < len(nodes) {
			links = append(links, Link{
				LinkID: fmt.Sprintf("nd%d_nd%d", i+1, rightChild+1),
				Source: nodes[i].Name,
				Target: nodes[rightChild].Name,
			})
		}
	}
	return Topology{Nodes: nodes, Links: links}
}

// general tree node
func generalTreeNode(id int) Node {
	nodeName := fmt.Sprintf("node_%d", id)
	flavor := selectFlavor(nodeName)
	return Node{
		NodeID:         fmt.Sprintf("nd%d", id),
		Name:           nodeName,
		AccessProtocol: "SSH",
		CPU:            flavor.CPU,
		Image:          promptString(fmt.Sprintf("Enter Image for %s: ", nodeName)),
		Memory:         flavor.Memory,
		SecurityRules:  []int{22},
		Storage:        flavor.Storage,
	}
}

// general tree topology
func createGeneralTreeTopology() Topology {
	/*numLevels := promptInt("Enter the number of levels: ")
	nodes := []Node{}
	links := []Link{}
	nodeIDCounter := 1

	// Create root node

	root := generalTreeNode(nodeIDCounter)
	nodes = append(nodes, root)
	nodeIDCounter++

	// Track parent nodes in the current level
	currentLevelParents := []Node{root}

	for level := 1; level < numLevels; level++ {
		numNodesInLevel := promptInt(fmt.Sprintf("Enter the number of nodes in level %d: ", level+1))
		levelNodes := []Node{}
		for i := 0; i < numNodesInLevel; i++ {

			node := generalTreeNode(nodeIDCounter)
			nodes = append(nodes, node)
			levelNodes = append(levelNodes, node)
			nodeIDCounter++
		}

		// Create links between the current level parents and the level nodes
		for i, parent := range currentLevelParents {
			for j := 0; j < numNodesInLevel/len(currentLevelParents); j++ {
				childIndex := i*(numNodesInLevel/len(currentLevelParents)) + j
				if childIndex < len(levelNodes) {
					links = append(links, Link{
						LinkID: fmt.Sprintf("%s_%s", parent.NodeID, levelNodes[childIndex].NodeID),
						Source: parent.Name,
						Target: levelNodes[childIndex].Name,
					})
				}
			}
		}

		// Update current level parents to the nodes of the current level
		currentLevelParents = levelNodes
	}*/

	numLevels := promptInt("Enter the number of levels: ")
	nodes := []Node{}
	links := []Link{}
	nodeIDCounter := 1

	// Create root node
	root := generalTreeNode(nodeIDCounter)
	nodes = append(nodes, root)
	nodeIDCounter++

	// Track parent nodes in the current level
	currentLevelParents := []Node{root}

	for level := 1; level < numLevels; level++ {
		levelNodes := []Node{}
		newLevelParents := []Node{}

		for _, parent := range currentLevelParents {
			numChildren := promptInt(fmt.Sprintf("Enter the number of children for node %s: ", parent.Name))
			for i := 0; i < numChildren; i++ {
				node := generalTreeNode(nodeIDCounter)
				nodes = append(nodes, node)
				levelNodes = append(levelNodes, node)
				nodeIDCounter++

				// Create link from parent to child
				links = append(links, Link{
					LinkID: fmt.Sprintf("%s_%s", parent.NodeID, node.NodeID),
					Source: parent.Name,
					Target: node.Name,
				})
			}
			newLevelParents = append(newLevelParents, levelNodes...)
		}

		// Update current level parents to the nodes of the current level
		currentLevelParents = newLevelParents
	}

	return Topology{Nodes: nodes, Links: links}
}

/* ES LOS MISMO QUE LINEAL
func createBusTopology() Topology {
	numNodes := promptInt("Enter the number of nodes: ")
	nodes := createNodes(numNodes)
	var links []Link
	for i := 0; i < numNodes-1; i++ {
		links = append(links, Link{
			LinkID: fmt.Sprintf("link_id_%d", i+1),
			Source: nodes[i].Name,
			Target: nodes[i+1].Name,
		})
	}
	return Topology{Nodes: nodes, Links: links}
}*/

func createRingTopology() Topology {
	numNodes := promptInt("Enter the number of nodes: ")
	nodes := createNodes(numNodes)
	var links []Link
	for i := 0; i < numNodes; i++ {
		links = append(links, Link{
			LinkID: fmt.Sprintf("nd%d_nd%d", i+1, (i+1)%numNodes+1),
			Source: nodes[i].Name,
			Target: nodes[(i+1)%numNodes].Name,
		})
	}
	return Topology{Nodes: nodes, Links: links}
}

func createStarTopology() Topology {
	fmt.Println("For star topology, node_1 (id: nd1) is the central node.")
	numNodes := promptInt("Enter the number of peripheral nodes: ") + 1 // Include central node
	nodes := createNodes(numNodes)
	var links []Link
	for i := 1; i < numNodes; i++ {
		links = append(links, Link{
			LinkID: fmt.Sprintf("nd1_nd%d", i+1), //  i+1 for rest of the nodes
			Source: nodes[0].Name,                //  central node is node_1
			Target: nodes[i].Name,                //  array index for rest of the nodes
		})
	}
	return Topology{Nodes: nodes, Links: links}
}

type Flavor struct {
	Name    string
	CPU     int
	Memory  int // en GB
	Storage int // en GB
}

var flavors = []Flavor{
	{Name: "Small", CPU: 1, Memory: 2, Storage: 50},
	{Name: "Medium", CPU: 2, Memory: 4, Storage: 100},
	{Name: "Large", CPU: 4, Memory: 8, Storage: 200},
	// Agrega más flavors según sea necesario
}

func selectFlavor(nodeName string) Flavor {
	/*flavors, err := fetchFlavors()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}*/

	// Mostrar opciones de flavors al usuario
	fmt.Printf("Select a flavor for %s:\n", nodeName)
	for i, flavor := range flavors {
		fmt.Printf("%d. %s (CPU: %d, Memory: %dGB, Storage: %dGB)\n", i+1, flavor.Name, flavor.CPU, flavor.Memory, flavor.Storage)
	}
	// Solicitar al usuario que ingrese el número correspondiente al flavor elegido
	var choice int
	for {
		choice = promptInt("Enter the number of the flavor: ")
		if choice > 0 && choice <= len(flavors) {
			break
		}
		fmt.Println("Invalid choice. Please enter a valid number.")
	}
	// Devolver el flavor seleccionado
	return flavors[choice-1]
}

func fetchFlavors() ([]Flavor, error) {
	url := "http://localhost:5000/flavors"
	var flavors []Flavor
	// Realizar solicitud HTTP GET al API
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error al hacer la solicitud HTTP: %v", err)
	}
	defer resp.Body.Close()

	// Decodificar la respuesta JSON en la estructura de datos Flavor
	err = json.NewDecoder(resp.Body).Decode(&flavors)
	if err != nil {
		return nil, fmt.Errorf("error al decodificar JSON: %v", err)
	}

	return flavors, nil
}

func createNodes(numNodes int) []Node {
	nodes := make([]Node, numNodes)
	for i := 0; i < numNodes; i++ {
		nodeName := fmt.Sprintf("node_%d", i+1)
		flavor := selectFlavor(nodeName)
		nodes[i] = Node{
			NodeID:         fmt.Sprintf("nd%d", i+1),
			Name:           nodeName,
			AccessProtocol: "SSH", // Supongamos que siempre hay una regla de seguridad SSH por defecto
			CPU:            flavor.CPU,
			Image:          promptString(fmt.Sprintf("Enter Image for %s: ", nodeName)),
			Memory:         flavor.Memory,
			SecurityRules:  []int{22}, // Supongamos que siempre hay una regla de seguridad SSH por defecto
			Storage:        flavor.Storage,
		}
	}
	return nodes
}

func createCustomTopology() Topology {
	numNodes := promptInt("Enter the number of nodes: ")
	nodes := createNodes(numNodes)
	var links []Link
	for i := 0; i < numNodes; i++ {
		for j := i + 1; j < numNodes; j++ {
			if promptString(fmt.Sprintf("Create link between %s and %s? (y/n): ", nodes[i].Name, nodes[j].Name)) == "y" {
				links = append(links, Link{
					LinkID: fmt.Sprintf("link_id_%d_%d", i+1, j+1),
					Source: nodes[i].Name,
					Target: nodes[j].Name,
				})
			}
		}
	}
	return Topology{Nodes: nodes, Links: links}
}

/*
func graphTemplate(nodes []Node, links []Link) error {
	// Crear un nuevo gráfico
	graphData := graph.NewGraph()
	graphData.Nodes = make([]*graph.Node, len(nodes))
	graphData.Edges = make([]*graph.Edge, len(links))

	for i, node := range nodes {
		graphData.Nodes[i] = &graph.Node{ID: node.NodeID, Label: node.Name}
	}

	for i, link := range links {
		graphData.Edges[i] = &graph.Edge{Source: link.Source, Target: link.Target}
	}

	// Configurar el diseño del gráfico
	layout := graph.NewLayout()
	layout.Title = "Topology Graph"
	layout.Height = 600
	layout.Width = 800

	// Crear el archivo HTML con el gráfico interactivo
	file, err := os.Create("topology.html")
	if err != nil {
		return err
	}
	defer file.Close()

	return graphData.Render(file, layout)
}*/

func printTopologyTable(topology Topology) {
	fmt.Println("Topology:")
	fmt.Println("-------------------------------------------------------")
	fmt.Println("Nodes:")
	fmt.Printf("%-10s %-10s %-10s %-15s %-15s %-15s\n", "NodeID", "Name", "CPU", "Memory(GB)", "Storage(GB)", "Links")
	for _, node := range topology.Nodes {
		fmt.Printf("%-10s %-10s %-10d %-15d %-15d", node.NodeID, node.Name, node.CPU, node.Memory, node.Storage)
		linkedNodes := []string{}
		for _, link := range topology.Links {
			if link.Source == node.Name || link.Target == node.Name {
				if link.Source == node.Name {
					linkedNodes = append(linkedNodes, link.Target)
				} else {
					linkedNodes = append(linkedNodes, link.Source)
				}
			}
		}
		fmt.Printf("%-15s\n", strings.Join(linkedNodes, ", "))
	}
	fmt.Println("-------------------------------------------------------")
	fmt.Println("Links:")
	fmt.Printf("%-10s %-15s %-15s\n", "LinkID", "Source", "Target")
	for _, link := range topology.Links {
		fmt.Printf("%-10s %-15s %-15s\n", link.LinkID, link.Source, link.Target)
	}
}

func graphTemplateTopology(template Template) {
	templateDetails := `
			<strong>Template Name:</strong> ` + template.Name + `<br>
			<strong>Description:</strong> ` + template.Description + `<br>
			<strong>Availability Zone:</strong> ` + template.AvailabilityZone + `<br>
			`

	htmlContent := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Network Topology</title>
		<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">
		<script src="https://cdnjs.cloudflare.com/ajax/libs/cytoscape/3.29.2/cytoscape.min.js" integrity="sha512-yi5TwB0WBpzqlJXNLURNMtpFXJt4yxJhkOG8yqkVQYWhfMkAoDF93rZ/KjfoN1gADGr5uKXvr5/Bw6CC03YWpA==" crossorigin="anonymous" referrerpolicy="no-referrer"></script>
		<style>
			#cy {
				width: 100%;
				height: 400px;
				border: 1px solid #333; /* Borde sólido de 1 píxel de color gris oscuro */
				border-radius: 3px; /* Borde redondeado */
			}
			#info-container {
				padding: 10px;
				border: 1px solid #ccc;
				border-radius: 5px;
				background-color: #f9f9f9;
			}
			#node-info {
				font-family: Arial, sans-serif;
				font-size: 14px;
			}
			#template-details {
				margin-bottom: 10px;
			}
			#template-details strong {
				color: #333;
				margin-right: 5px;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<div class="row justify-content-center mt-5">
				<div class="col-md-8 text-center">
					<h1 class="mb-4">Network Topology</h1>
					<div id="template-details" class="text-left">
					` + templateDetails + `
					</div>
				</div>
			</div>
			<div class="row justify-content-center">
				<div class="col-md-8">
					<div id="cy"></div>
				</div>
				<div class="col-md-4">
					<div id="info-container">
						<div id="node-info"></div>
					</div>
				</div>
			</div>
		</div>
		<script>
			document.addEventListener('DOMContentLoaded', function() {
				var cy = cytoscape({
					container: document.getElementById('cy'),
					elements: [
`
	for _, node := range template.Topology.Nodes {
		htmlContent += fmt.Sprintf(`
                    { data: { id: '%s', name: '%s', cpu: '%d', memory: '%d', storage: '%d', image: '%s' } },
`, node.Name, node.Name, node.CPU, node.Memory, node.Storage, node.Image)
	}

	for _, link := range template.Topology.Links {
		htmlContent += fmt.Sprintf(`
                    { data: { id: '%s', source: '%s', target: '%s' } },
`, link.LinkID, link.Source, link.Target)
	}

	htmlContent += `
	],
	style: [
		{
			selector: 'node',
			style: {
				'label': 'data(name)',
				'width': '60px',
				'height': '60px',
				'background-color': '#349beb', // Azul suave
				'color': '#000', // Color de la etiqueta
				'text-valign': 'center',
				'text-halign': 'center'
			}
		},
		{
			selector: 'edge',
			style: {
				'width': 3,
				'line-color': '#000', // Negro
				'curve-style': 'bezier'
			}
		}
	],
	layout: {
		name: 'grid',
		rows: 1
	}
});

// Función para mostrar información del nodo
function showNodeInfo(node) {
	var nodeData = node.data();
	var nodeInfo = '<strong>Node:</strong> ' + nodeData.name + '<br>' +
				   '<strong>vCPU:</strong> ' + nodeData.cpu + '<br>' +
				   '<strong>Memory:</strong> ' + nodeData.memory + 'GB<br>' +
				   '<strong>Storage:</strong> ' + nodeData.storage + 'GB<br>' + 
				   '<strong>Image:</strong> ' + nodeData.image;
	document.getElementById('node-info').innerHTML = nodeInfo;
}

// Agregar evento de clic a los nodos
cy.on('tap', 'node', function(event) {
	var node = event.target;
	showNodeInfo(node);
});
});
</script>
</body>
</html>

`

	file, err := os.Create("topology.html")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(htmlContent)
	if err != nil {
		panic(err)
	}

	// Open the HTML file in the default browser
	openBrowser("topology.html")
}

func openBrowser(url string) {
	var err error

	switch os := runtime.GOOS; os {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
	}
}

func CreateTemplate() {
	name := promptString("Enter template name: ")
	description := promptString("Enter template description: ")
	topologyType := promptString("Do you want to create a predefined or custom topology? (predefined/custom): ")

	var topology Topology
	if topologyType == "predefined" {
		topologyType = promptTopology()
		topology = createPredefinedTopology(topologyType)
	} else {
		topology = createCustomTopology()
	}

	fmt.Println("Generated Topology:")
	fmt.Printf("Nodes: %+v\n", topology.Nodes)
	fmt.Printf("Links: %+v\n", topology.Links)

	availabilityZone := promptString("Enter availability zone: ")
	template := Template{
		CreatedAt:        time.Now().UTC(),
		AvailabilityZone: availabilityZone,
		Deployed:         false,
		Description:      description,
		Name:             name,
		Topology:         topology,
		TopologyType:     topologyType,
		UserID:           "6640550a53c1187a6899a5a9",
	}

	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	fmt.Printf("Generated JSON:\n%s\n", string(templateJSON))
	printTopologyTable(topology)

	graphTemplateTopology(template)
	/*
		// Graficar la topología y guardarla como un archivo HTML
		if err := graphTemplate(topology.Nodes, topology.Links); err != nil {
			fmt.Println("Error:", err)
			return
		}*/

	// Implement HTTP request to send JSON to the server as needed
}
