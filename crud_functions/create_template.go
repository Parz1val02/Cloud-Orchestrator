package crud_functions

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Image struct {
	ImageID     string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

type Port struct {
	PortID string `json:"node_id"`
}

type Node struct {
	NodeID        string `json:"node_id"`
	Name          string `json:"name"`
	Image         string `json:"image"`
	Flavor        Flavor `json:"flavor"`
	SecurityRules []int  `json:"security_rules"`
	Ports         []Port `json:"ports"`
}

type Flavor struct {
	FlavorID string  `json:"id"`
	Name     string  `json:"name"`
	CPU      int     `json:"cpu"`
	Memory   float32 `json:"memory"`  // en GB
	Storage  float32 `json:"storage"` // en GB
}

type Link struct {
	LinkID     string `json:"link_id"`
	Source     string `json:"source"`
	Target     string `json:"target"`
	SourcePort string `json:"source_port"`
	TargetPort string `json:"target_port"`
}

type Topology struct {
	Links []Link `json:"links"`
	Nodes []Node `json:"nodes"`
}

type Template struct {
	CreatedAt    time.Time `json:"created_at"`
	Description  string    `json:"description"`
	Name         string    `json:"name"`
	Topology     Topology  `json:"topology"`
	UserID       string    `json:"user_id"`
	TopologyType string    `json:"topology_type"`
}

func initConfig() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	// Check if the YAML file exists
	if _, err := os.Stat(home + "/.cloud-cli.yaml"); err == nil {
		// YAML file exists, assume user is authenticated
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cloud-cli")
	}
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

	// Map to keep track of the next available port number for each node
	portMap := make(map[string]int)

	for i := 0; i < numNodes-1; i++ {
		sourceNodeID := nodes[i].NodeID
		targetNodeID := nodes[i+1].NodeID

		sourcePortNumber := portMap[sourceNodeID]
		targetPortNumber := portMap[targetNodeID]

		sourcePortID := fmt.Sprintf("%s_port%d", sourceNodeID, sourcePortNumber)
		targetPortID := fmt.Sprintf("%s_port%d", targetNodeID, targetPortNumber)

		// Añadir puertos a los nodos
		nodes[i].Ports = append(nodes[i].Ports, Port{PortID: sourcePortID})
		nodes[i+1].Ports = append(nodes[i+1].Ports, Port{PortID: targetPortID})

		// Incrementar el contador de puertos para el próximo puerto disponible
		portMap[sourceNodeID]++
		portMap[targetNodeID]++

		links = append(links, Link{
			LinkID:     fmt.Sprintf("nd%d_nd%d", i+1, i+2),
			Source:     sourceNodeID,
			Target:     targetNodeID,
			SourcePort: sourcePortID,
			TargetPort: targetPortID,
		})

	}
	return Topology{Nodes: nodes, Links: links}
}

func createMeshTopology() Topology {
	numNodes := promptInt("Enter the number of nodes: ")
	nodes := createNodes(numNodes)
	var links []Link

	// Map to keep track of the next available port number for each node
	portMap := make(map[string]int)

	for i := 0; i < numNodes; i++ {
		for j := i + 1; j < numNodes; j++ {
			sourceNodeID := nodes[i].NodeID
			targetNodeID := nodes[j].NodeID

			sourcePortNumber := portMap[sourceNodeID]
			targetPortNumber := portMap[targetNodeID]

			sourcePortID := fmt.Sprintf("%s_port%d", sourceNodeID, sourcePortNumber)
			targetPortID := fmt.Sprintf("%s_port%d", targetNodeID, targetPortNumber)

			// Add ports to nodes
			nodes[i].Ports = append(nodes[i].Ports, Port{PortID: sourcePortID})
			nodes[j].Ports = append(nodes[j].Ports, Port{PortID: targetPortID})

			// Increment port numbers
			portMap[sourceNodeID]++
			portMap[targetNodeID]++

			links = append(links, Link{
				LinkID:     fmt.Sprintf("nd%d_nd%d", i+1, j+1),
				Source:     sourceNodeID,
				Target:     targetNodeID,
				SourcePort: sourcePortID,
				TargetPort: targetPortID,
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
	// Map to keep track of the next available port number for each node
	portMap := make(map[string]int)

	for i := 0; i < (1<<(numLevels-1))-1; i++ {
		leftChild := 2*i + 1
		rightChild := 2*i + 2
		if leftChild < len(nodes) {

			sourceNodeID := nodes[i].NodeID
			targetNodeID := nodes[leftChild].NodeID

			sourcePortNumber := portMap[sourceNodeID]
			targetPortNumber := portMap[targetNodeID]

			sourcePortID := fmt.Sprintf("%s_port%d", sourceNodeID, sourcePortNumber)
			targetPortID := fmt.Sprintf("%s_port%d", targetNodeID, targetPortNumber)

			// Add ports to nodes
			nodes[i].Ports = append(nodes[i].Ports, Port{PortID: sourcePortID})
			nodes[leftChild].Ports = append(nodes[leftChild].Ports, Port{PortID: targetPortID})

			// Increment port numbers
			portMap[sourceNodeID]++
			portMap[targetNodeID]++

			links = append(links, Link{
				LinkID:     fmt.Sprintf("nd%d_nd%d", i+1, leftChild+1),
				Source:     sourceNodeID,
				Target:     targetNodeID,
				SourcePort: sourcePortID,
				TargetPort: targetPortID,
			})
		}
		if rightChild < len(nodes) {

			sourceNodeID := nodes[i].NodeID
			targetNodeID := nodes[rightChild].NodeID

			sourcePortNumber := portMap[sourceNodeID]
			targetPortNumber := portMap[targetNodeID]

			sourcePortID := fmt.Sprintf("%s_port%d", sourceNodeID, sourcePortNumber)
			targetPortID := fmt.Sprintf("%s_port%d", targetNodeID, targetPortNumber)

			// Add ports to nodes
			nodes[i].Ports = append(nodes[i].Ports, Port{PortID: sourcePortID})
			nodes[rightChild].Ports = append(nodes[rightChild].Ports, Port{PortID: targetPortID})

			// Increment port numbers
			portMap[sourceNodeID]++
			portMap[targetNodeID]++

			links = append(links, Link{
				LinkID:     fmt.Sprintf("nd%d_nd%d", i+1, rightChild+1),
				Source:     sourceNodeID,
				Target:     targetNodeID,
				SourcePort: sourcePortID,
				TargetPort: targetPortID,
			})
		}
	}
	return Topology{Nodes: nodes, Links: links}
}

// general tree node
func generalTreeNode(id int) Node {
	nodeName := fmt.Sprintf("node_%d", id)
	flavor := selectFlavor(nodeName)
	image := selectImage(nodeName)
	return Node{
		NodeID:        fmt.Sprintf("nd%d", id),
		Name:          nodeName,
		Flavor:        flavor,
		Image:         image.ImageID,
		SecurityRules: []int{22},
		Ports:         []Port{},
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

	// Mapa para llevar la cuenta de los puertos disponibles para cada nodo
	portMap := make(map[string]int)

	// Create root node
	root := generalTreeNode(nodeIDCounter)
	nodes = append(nodes, root)
	nodeIDCounter++

	// Añadir el primer puerto para el nodo raíz
	portMap[root.NodeID] = 0

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

				// Crear puertos para los enlaces entre el padre y el hijo
				parentPortNumber := portMap[parent.NodeID]
				childPortNumber := portMap[node.NodeID]

				parentPortID := fmt.Sprintf("%s_port%d", parent.NodeID, parentPortNumber)
				childPortID := fmt.Sprintf("%s_port%d", node.NodeID, childPortNumber)

				// Añadir los puertos a los nodos
				parent.Ports = append(parent.Ports, Port{PortID: parentPortID})
				node.Ports = append(node.Ports, Port{PortID: childPortID})

				// Incrementar el número de puertos disponibles
				portMap[parent.NodeID]++
				portMap[node.NodeID]++

				// Create link from parent to child
				links = append(links, Link{
					LinkID:     fmt.Sprintf("%s_%s", parent.NodeID, node.NodeID),
					Source:     parent.NodeID,
					Target:     node.NodeID,
					SourcePort: parentPortID,
					TargetPort: childPortID,
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

	// Map to keep track of the next available port number for each node
	portMap := make(map[string]int)
	for i := 0; i < numNodes; i++ {

		sourceNodeID := nodes[i].NodeID
		targetNodeID := nodes[(i+1)%numNodes].NodeID

		sourcePortNumber := portMap[sourceNodeID]
		targetPortNumber := portMap[targetNodeID]

		sourcePortID := fmt.Sprintf("%s_port%d", sourceNodeID, sourcePortNumber)
		targetPortID := fmt.Sprintf("%s_port%d", targetNodeID, targetPortNumber)

		// Add ports to nodes
		nodes[i].Ports = append(nodes[i].Ports, Port{PortID: sourcePortID})
		nodes[(i+1)%numNodes].Ports = append(nodes[(i+1)%numNodes].Ports, Port{PortID: targetPortID})

		// Increment port numbers
		portMap[sourceNodeID]++
		portMap[targetNodeID]++

		links = append(links, Link{
			LinkID:     fmt.Sprintf("nd%d_nd%d", i+1, (i+1)%numNodes+1),
			Source:     sourceNodeID,
			Target:     targetNodeID,
			SourcePort: sourcePortID,
			TargetPort: targetPortID,
		})
	}
	return Topology{Nodes: nodes, Links: links}
}

func createStarTopology() Topology {
	fmt.Println("For star topology, node_1 (id: nd1) is the central node.")
	numNodes := promptInt("Enter the number of peripheral nodes: ") + 1 // Include central node
	nodes := createNodes(numNodes)
	var links []Link
	// Map to keep track of the next available port number for each node
	portMap := make(map[string]int)

	for i := 1; i < numNodes; i++ {

		sourceNodeID := nodes[0].NodeID
		targetNodeID := nodes[i].NodeID

		sourcePortNumber := portMap[sourceNodeID]
		targetPortNumber := portMap[targetNodeID]

		sourcePortID := fmt.Sprintf("%s_port%d", sourceNodeID, sourcePortNumber)
		targetPortID := fmt.Sprintf("%s_port%d", targetNodeID, targetPortNumber)

		// Add ports to nodes
		nodes[0].Ports = append(nodes[0].Ports, Port{PortID: sourcePortID})
		nodes[i].Ports = append(nodes[i].Ports, Port{PortID: targetPortID})

		// Increment port numbers
		portMap[sourceNodeID]++
		portMap[targetNodeID]++

		links = append(links, Link{
			LinkID:     fmt.Sprintf("nd1_nd%d", i+1), //  i+1 for rest of the nodes
			Source:     sourceNodeID,                 //  central node is node_1
			Target:     targetNodeID,                 //  array index for rest of the nodes
			SourcePort: sourcePortID,
			TargetPort: targetPortID,
		})
	}
	return Topology{Nodes: nodes, Links: links}
}

/*
var flavors = []Flavor{
	{Name: "Small", CPU: 1, Memory: 2, Storage: 50},
	{Name: "Medium", CPU: 2, Memory: 4, Storage: 100},
	{Name: "Large", CPU: 4, Memory: 8, Storage: 200},
	// Agrega más flavors según sea necesario
}*/

func selectImage(nodeName string) Image {
	var images []Image
	images, err := fetchImages()
	if err != nil {
		fmt.Printf("Error fetching images: %v\n", err)
	}

	// Mostrar opciones de imágenes al usuario
	fmt.Printf("Select an image for %s:\n", nodeName)
	for i, img := range images {
		fmt.Printf("%d. %s %s\n", i+1, img.Name, img.Version)
	}
	// Solicitar al usuario que ingrese el número correspondiente a la imagen elegida
	var choice int
	for {
		choice = promptInt("Enter the number of the image: ")
		if choice > 0 && choice <= len(images) {
			break
		}
		fmt.Println("Invalid choice. Please enter a valid number.")
	}
	// Devolver la imagen seleccionada
	return images[choice-1]
}

func fetchImages() ([]Image, error) {
	url := "http://localhost:4444/templateservice/templates/images"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}
	var token = viper.GetString("token")
	req.Header.Set("X-API-Key", token)
	//fmt.Println("TOKEN", token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching images: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var result struct {
		Result string  `json:"result"`
		Images []Image `json:"images"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return result.Images, nil
}

func selectFlavor(nodeName string) Flavor {
	var flavors []Flavor
	flavors, err := fetchFlavors()
	if err != nil {
		fmt.Printf("Error flavors: %v\n", err)
	}

	// Mostrar opciones de flavors al usuario
	fmt.Printf("Select a flavor for %s:\n", nodeName)
	for i, flavor := range flavors {
		fmt.Printf("%d. %s (CPU: %d, Memory: %.1fGB, Storage: %.1fGB)\n", i+1, flavor.Name, flavor.CPU, flavor.Memory, flavor.Storage)
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
	url := "http://localhost:4444/templateservice/templates/flavors"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}
	var token = viper.GetString("token")
	req.Header.Set("X-API-Key", token)
	//fmt.Println("token:", token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching flavors: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var result struct {
		Result  string   `json:"result"`
		Flavors []Flavor `json:"flavors"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return result.Flavors, nil
}

func createNodes(numNodes int) []Node {
	nodes := make([]Node, numNodes)
	for i := 0; i < numNodes; i++ {
		nodeName := fmt.Sprintf("node_%d", i+1)
		flavor := selectFlavor(nodeName)
		image := selectImage(nodeName)
		nodes[i] = Node{
			NodeID:        fmt.Sprintf("nd%d", i+1),
			Name:          nodeName,
			Flavor:        flavor,
			Image:         image.ImageID,
			SecurityRules: []int{22}, // Supongamos que siempre hay una regla de seguridad SSH por defecto
			Ports:         []Port{},
		}
	}
	return nodes
}

func createCustomTopology() Topology {
	numNodes := promptInt("Enter the number of nodes: ")
	nodes := createNodes(numNodes)
	var links []Link
	// Map to keep track of the next available port number for each node
	portMap := make(map[string]int)
	for i := 0; i < numNodes; i++ {
		for j := i + 1; j < numNodes; j++ {
			if promptString(fmt.Sprintf("Create link between %s and %s? (y/n): ", nodes[i].Name, nodes[j].Name)) == "y" {
				sourceNodeID := nodes[i].NodeID
				targetNodeID := nodes[j].NodeID

				sourcePortNumber := portMap[sourceNodeID]
				targetPortNumber := portMap[targetNodeID]

				sourcePortID := fmt.Sprintf("%s_port%d", sourceNodeID, sourcePortNumber)
				targetPortID := fmt.Sprintf("%s_port%d", targetNodeID, targetPortNumber)

				// Add ports to nodes
				nodes[i].Ports = append(nodes[i].Ports, Port{PortID: sourcePortID})
				nodes[j].Ports = append(nodes[j].Ports, Port{PortID: targetPortID})

				// Increment port numbers
				portMap[sourceNodeID]++
				portMap[targetNodeID]++

				links = append(links, Link{
					LinkID:     fmt.Sprintf("nd%d_nd%d", i+1, j+1),
					Source:     sourceNodeID,
					Target:     targetNodeID,
					SourcePort: sourcePortID,
					TargetPort: targetPortID,
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
	fmt.Printf("%-20s %-20s %-20s %-20s %-20s %-20s\n", "NodeID", "Name", "CPU", "Memory(GB)", "Storage(GB)", "Links")
	for _, node := range topology.Nodes {
		fmt.Printf("%-20s %-20s %-20d %-20.1f %-20.1f", node.NodeID, node.Name, node.Flavor.CPU, node.Flavor.Memory, node.Flavor.Storage)
		linkedNodes := []string{}
		for _, link := range topology.Links {
			if link.Source == node.NodeID || link.Target == node.NodeID {
				if link.Source == node.NodeID {
					linkedNodes = append(linkedNodes, link.Target)
				} else {
					linkedNodes = append(linkedNodes, link.Source)
				}
			}
		}
		fmt.Printf("%-20s\n", strings.Join(linkedNodes, ", "))
	}
	fmt.Println("-------------------------------------------------------")
	fmt.Println("Links:")
	fmt.Printf("%-20s %-20s %-20s\n", "LinkID", "Source", "Target")
	for _, link := range topology.Links {
		fmt.Printf("%-20s %-20s %-20s\n", link.LinkID, link.Source, link.Target)
	}
}

func graphTemplateTopology(template Template) {
	templateDetails := `
			<strong>Template Name:</strong> ` + template.Name + `<br>
			<strong>Description:</strong> ` + template.Description + `
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
                    { data: { id: '%s', name: '%s', cpu: '%d', memory: '%.1f', storage: '%.1f', image: '%s' } },
`, node.NodeID, node.Name, node.Flavor.CPU, node.Flavor.Memory, node.Flavor.Storage, node.Image)
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
		name: 'cose', // Layout for better node distribution
		fit: true, // Whether to fit the viewport to the graph
		padding: 30, // Padding around the graph
		animate: true, // Whether to animate the layout
		animationDuration: 1000 // Duration of animation in ms if enabled
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

func selectTopologyType() string {
	// Mostrar opciones de topology type al usuario
	topology_type := []string{"predefined", "custom"}

	for i, name := range topology_type {
		fmt.Printf("%d. %s\n", i+1, name)
	}
	// Solicitar al usuario que ingrese el número correspondiente al topologytype elegido
	var choice int
	for {
		choice = promptInt("Enter the number of the chosen topology type: ")
		if choice > 0 && choice <= len(topology_type) {
			break
		}
		fmt.Println("Invalid choice. Please enter a valid number.")
	}
	// Devolver el flavor seleccionado
	return topology_type[choice-1]
}

/*
func selectAvailabilityZone() AvailabilityZone {
	var availabilityZones []AvailabilityZone
	availabilityZones, err := fetchAvailabilityZone()
	if err != nil {
		fmt.Printf("Error fetching availabilityZones: %v\n", err)
	}

	// Mostrar opciones de imágenes al usuario
	fmt.Printf("Select an availability zone:\n")
	for i, az := range availabilityZones {
		fmt.Printf("%d. %s\n", i+1, az.Name)
	}
	// Solicitar al usuario que ingrese el número correspondiente a la imagen elegida
	var choice int
	for {
		choice = promptInt("Enter the number of the image: ")
		if choice > 0 && choice <= len(availabilityZones) {
			break
		}
		fmt.Println("Invalid choice. Please enter a valid number.")
	}
	// Devolver la imagen seleccionada
	return availabilityZones[choice-1]
}
func fetchAvailabilityZone() ([]AvailabilityZone, error) {
	url := "http://localhost:5000/templates/availabilityZones"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching availabilityZones: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var result struct {
		Result            string             `json:"result"`
		AvailabilityZones []AvailabilityZone `json:"availability_zones"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return result.AvailabilityZones, nil
}
*/

// Estructura para los servidores en una zona de disponibilidad
type Server struct {
	Name string `json:"name"`
}

// Estructura para las zonas de disponibilidad
type AvailabilityZone struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Servers []Server `json:"servers"`
}

// Crear 3 zonas de disponibilidad con diferentes combinaciones de servidores
/*var availabilityZones = []AvailabilityZone{
	{
		ID:   "zone_1",
		Name: "Zone 1",
		Servers: []Server{
			{Name: "Worker1"},
			{Name: "Worker2"},
			{Name: "Worker3"},
		},
	},
	{
		ID:   "zone_2",
		Name: "Zone 2",
		Servers: []Server{
			{Name: "Worker1"},
			{Name: "Worker3"},
		},
	},
	{
		ID:   "zone_3",
		Name: "Zone 3",
		Servers: []Server{
			{Name: "Worker2"},
			{Name: "Worker3"},
		},
	},
}*/

// Función para solicitar al usuario que seleccione una zona de disponibilidad
/*func selectorAvailabilityZone() string {
	fmt.Println("Select an Availability Zone:")

	for i, zone := range availabilityZones {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, zone.Name, zone.ID)
		fmt.Println("Servers:")
		for _, server := range zone.Servers {
			fmt.Printf("  - %s\n", server.Name)
		}
	}

	var choice int
	for {
		fmt.Print("Enter the number of the availability zone: ")
		fmt.Scan(&choice)
		if choice > 0 && choice <= len(availabilityZones) {
			break
		}
		fmt.Println("Invalid choice. Please enter a valid number.")
	}

	return availabilityZones[choice-1].ID
}*/

func sendTemplate(templateJSON []byte, token string) {

	serverPort := 4444
	requestURL := fmt.Sprintf("http://localhost:%d/templateservice/templates", serverPort)

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(templateJSON))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error at creating template: ", err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	// Estructura para deserializar la respuesta
	type ResponseCreateTemplate struct {
		Result string `json:"result"`
		Msg    string `json:"msg"`
	}

	// Leer la respuesta
	var result ResponseCreateTemplate

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Printf("Error decoding response body create template http: %v", err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d, Error: %s\n", resp.StatusCode, result.Msg)
		os.Exit(1)
	}
	// Mostrar la respuesta
	fmt.Println("Respuesta:", result)

}

func CreateTemplate(user_id string, token string) {

	initConfig()

	name := promptString("Enter template name: ")
	description := promptString("Enter template description: ")
	// topologyType := promptString("Do you want to create a predefined or custom topology? (predefined/custom): ")
	fmt.Println("Do you want to create a predefined or custom topology?:")
	topologyType := selectTopologyType()

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

	//availabilityZone := selectorAvailabilityZone()

	template := Template{
		CreatedAt:    time.Now().UTC(),
		Description:  description,
		Name:         name,
		Topology:     topology,
		TopologyType: topologyType,
		UserID:       user_id,
	}

	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	fmt.Printf("Generated JSON:\n%s\n", string(templateJSON))
	printTopologyTable(topology)

	graphTemplateTopology(template)

	sendTemplate(templateJSON, token)
	/*
		// Graficar la topología y guardarla como un archivo HTML
		if err := graphTemplate(topology.Nodes, topology.Links); err != nil {
			fmt.Println("Error:", err)
			return
		}*/

}
