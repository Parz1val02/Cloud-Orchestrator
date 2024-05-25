package crud_functions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
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
	topologies := []string{"malla", "arbol", "lineal", "anillo", "bus", "estrella"}
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
	case "arbol":
		return createTreeTopology()
	case "bus":
		return createBusTopology()
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
			LinkID: fmt.Sprintf("link_id_%d", i+1),
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
				LinkID: fmt.Sprintf("link_id_%d_%d", i+1, j+1),
				Source: nodes[i].Name,
				Target: nodes[j].Name,
			})
		}
	}
	return Topology{Nodes: nodes, Links: links}
}

func createTreeTopology() Topology {
	numLevels := promptInt("Enter the number of levels: ")
	numNodes := (1 << numLevels) - 1 // 2^levels - 1
	nodes := createNodes(numNodes)
	var links []Link
	for i := 0; i < (1<<(numLevels-1))-1; i++ {
		leftChild := 2*i + 1
		rightChild := 2*i + 2
		if leftChild < len(nodes) {
			links = append(links, Link{
				LinkID: fmt.Sprintf("link_id_%d_%d", i+1, leftChild+1),
				Source: nodes[i].Name,
				Target: nodes[leftChild].Name,
			})
		}
		if rightChild < len(nodes) {
			links = append(links, Link{
				LinkID: fmt.Sprintf("link_id_%d_%d", i+1, rightChild+1),
				Source: nodes[i].Name,
				Target: nodes[rightChild].Name,
			})
		}
	}
	return Topology{Nodes: nodes, Links: links}
}

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
}

func createStarTopology() Topology {
	numNodes := promptInt("Enter the number of peripheral nodes: ") + 1 // Include central node
	nodes := createNodes(numNodes)
	var links []Link
	for i := 1; i < numNodes; i++ {
		links = append(links, Link{
			LinkID: fmt.Sprintf("link_id_%d", i),
			Source: nodes[0].Name,
			Target: nodes[i].Name,
		})
	}
	return Topology{Nodes: nodes, Links: links}
}

func createNodes(numNodes int) []Node {
	nodes := make([]Node, numNodes)
	for i := 0; i < numNodes; i++ {
		nodeName := fmt.Sprintf("node_%d", i+1)
		nodes[i] = Node{
			NodeID:         fmt.Sprintf("node_id_%d", i+1),
			Name:           nodeName,
			AccessProtocol: "SSH",
			CPU:            promptInt(fmt.Sprintf("Enter CPU for %s: ", nodeName)),
			Image:          promptString(fmt.Sprintf("Enter Image for %s: ", nodeName)),
			Memory:         promptInt(fmt.Sprintf("Enter Memory (GB) for %s: ", nodeName)),
			SecurityRules:  []int{22},
			Storage:        promptInt(fmt.Sprintf("Enter Storage (GB) for %s: ", nodeName)),
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

func CreateTemplate() {
	name := promptString("Enter template name: ")
	description := promptString("Enter template description: ")
	topologyType := promptString("Do you want to create a predefined or custom topology? (predefined/custom): ")

	var topology Topology
	if topologyType == "predefined" {
		topoType := promptTopology()
		topology = createPredefinedTopology(topoType)
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
		UserID:           "6640550a53c1187a6899a5a9",
	}

	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	fmt.Printf("Generated JSON:\n%s\n", string(templateJSON))

	/*
		// Graficar la topología y guardarla como un archivo HTML
		if err := graphTemplate(topology.Nodes, topology.Links); err != nil {
			fmt.Println("Error:", err)
			return
		}*/

	// Implement HTTP request to send JSON to the server as needed

}
