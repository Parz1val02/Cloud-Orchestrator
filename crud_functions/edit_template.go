package crud_functions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

// Estructuras de datos

type TemplateByID struct {
	TemplateID   string    `json:"template_id"`
	CreatedAt    time.Time `json:"created_at"`
	Description  string    `json:"description"`
	Name         string    `json:"name"`
	Topology     Topology  `json:"topology"`
	UserID       string    `json:"user_id"`
	TopologyType string    `json:"topology_type"`
}

type Modification struct {
	CreatedNodes    []Node `json:"created_nodes,omitempty"`
	DeletedNodes    []Node `json:"deleted_nodes,omitempty"`
	CreatedLinks    []Link `json:"created_links,omitempty"`
	DeletedLinks    []Link `json:"deleted_links,omitempty"`
	ModifiedFlavors []struct {
		NodeID string `json:"node_id"`
		Flavor Flavor `json:"flavor"`
	} `json:"modified_flavors,omitempty"`
}

type NormalResponse struct {
	Msg string `json:"msg"`
}

type ListTemplateById struct {
	Template Template `json:"template"`
}

// Funciones para manipular el template

// Helper function to get the next available port for a node
func getNextPort(nodeID string, portMap map[string]int) string {
	portNumber := portMap[nodeID]
	portID := fmt.Sprintf("%s_port%d", nodeID, portNumber)
	portMap[nodeID]++
	return portID
}

func createNode(template *Template, modification *Modification, name, image string, flavor Flavor) {
	newID := fmt.Sprintf("nd%d", len(template.Topology.Nodes)+1)
	newNode := Node{
		NodeID: newID,
		Name:   name,
		Image:  image,
		Flavor: flavor,
		Ports:  []Port{}, // Initialize with an empty slice of Ports
	}

	template.Topology.Nodes = append(template.Topology.Nodes, newNode)
	modification.CreatedNodes = append(modification.CreatedNodes, newNode)
}

func deleteNode(template *Template, modification *Modification, nodeID string) {
	var updatedNodes []Node
	var deletedNode Node
	found := false
	// Normalizar el ID del nodo para asegurar consistencia en las comparaciones
	nodeID = strings.TrimSpace(nodeID)
	for _, node := range template.Topology.Nodes {
		fmt.Println("node: ", node.NodeID)
		nodeIDTemplate := strings.TrimSpace(node.NodeID)
		if nodeIDTemplate == nodeID {
			deletedNode = node
			found = true
		} else {
			updatedNodes = append(updatedNodes, node)
		}
	}

	if !found {
		fmt.Printf("Node %s not found in template\n", nodeID)
		return
	}

	template.Topology.Nodes = updatedNodes
	modification.DeletedNodes = append(modification.DeletedNodes, deletedNode)

	var updatedLinks []Link
	for _, link := range template.Topology.Links {
		if link.Source != nodeID && link.Target != nodeID {
			updatedLinks = append(updatedLinks, link)
		} else {
			modification.DeletedLinks = append(modification.DeletedLinks, link)
		}
	}
	template.Topology.Links = updatedLinks
}

func editFlavor(template *Template, modification *Modification, nodeID string, newFlavor Flavor) {
	var foundNode *Node

	for i := range template.Topology.Nodes {
		if template.Topology.Nodes[i].NodeID == nodeID {
			foundNode = &template.Topology.Nodes[i]
			break
		}
	}

	if foundNode == nil {
		fmt.Printf("Node %s not found", nodeID)
		return
	}

	foundNode.Flavor = newFlavor

	modification.ModifiedFlavors = append(modification.ModifiedFlavors, struct {
		NodeID string `json:"node_id"`
		Flavor Flavor `json:"flavor"`
	}{NodeID: nodeID, Flavor: newFlavor})
}

func createLink(template *Template, modification *Modification, source, target string) {
	for _, link := range template.Topology.Links {
		if (link.Source == source && link.Target == target) || (link.Source == target && link.Target == source) {
			fmt.Println("Link already exists between ", source, " and ", target)
			return
		}
		if source == target {
			fmt.Println("Link cannot be created between ", source, " and ", target, ". Source and Target are the same.")
			return
		}
	}

	portMap := make(map[string]int)
	// Initialize the portMap with the current port count for each node
	for _, node := range template.Topology.Nodes {
		portMap[node.NodeID] = len(node.Ports)
	}

	sourcePortID := getNextPort(source, portMap)
	targetPortID := getNextPort(target, portMap)

	newID := fmt.Sprintf("%s_%s", source, target)
	newLink := Link{
		LinkID:     newID,
		Source:     source,
		Target:     target,
		SourcePort: sourcePortID,
		TargetPort: targetPortID,
	}

	// Add ports to the source and target nodes
	for i := range template.Topology.Nodes {
		if template.Topology.Nodes[i].NodeID == source {
			template.Topology.Nodes[i].Ports = append(template.Topology.Nodes[i].Ports, Port{PortID: sourcePortID})
		}
		if template.Topology.Nodes[i].NodeID == target {
			template.Topology.Nodes[i].Ports = append(template.Topology.Nodes[i].Ports, Port{PortID: targetPortID})
		}
	}

	template.Topology.Links = append(template.Topology.Links, newLink)
	modification.CreatedLinks = append(modification.CreatedLinks, newLink)
	fmt.Println("Link created between ", source, " and ", target)
}

func deleteLink(template *Template, modification *Modification, linkID string) {
	var updatedLinks []Link
	var deletedLink Link
	found := false

	for _, link := range template.Topology.Links {
		if link.LinkID == linkID {
			deletedLink = link
			found = true
		} else {
			updatedLinks = append(updatedLinks, link)
		}
	}

	if !found {
		fmt.Printf("Link %s not found\n", linkID)
		return
	}

	template.Topology.Links = updatedLinks
	modification.DeletedLinks = append(modification.DeletedLinks, deletedLink)
}

func editTemplateTable(template *Template) {
	nodes := table.NewWriter()
	nodes.AppendHeader(table.Row{"ID", "Name", "Image", "CPU", "Memory", "Storage"})
	for _, v := range template.Topology.Nodes {
		nodes.AppendRow(table.Row{v.NodeID, v.Name, v.Image, strconv.Itoa(v.Flavor.CPU), strconv.FormatFloat(float64(v.Flavor.Memory), 'f', 1, 32), strconv.FormatFloat(float64(v.Flavor.Storage), 'f', 1, 32)})
	}
	links := table.NewWriter()
	links.AppendHeader(table.Row{"ID", "Source", "Target", "SourcePort", "TargetPort"})
	for _, v := range template.Topology.Links {
		links.AppendRow(table.Row{v.LinkID, v.Source, v.Target, v.SourcePort, v.TargetPort})
	}

	nodes.SetOutputMirror(os.Stdout)
	links.SetOutputMirror(os.Stdout)
	nodes.Render()
	links.Render()
}

// Función para hacer una solicitud PUT con las modificaciones
func sendPUTRequest(serverPort int, templateId, token string, template *Template) error {
	requestURL := fmt.Sprintf("http://localhost:%d/templateservice/templates/%s", serverPort, templateId)

	// Convertir Template a JSON
	templateJSON, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("error marshalling template to JSON: %w", err)
	}

	// Crear solicitud PUT
	client := &http.Client{}
	req, err := http.NewRequest("PUT", requestURL, bytes.NewBuffer(templateJSON))
	if err != nil {
		return fmt.Errorf("error creating PUT request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", token)

	// Realizar la solicitud PUT
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making PUT request: %w", err)
	}
	defer resp.Body.Close()

	// Verificar el código de estado de la respuesta
	if resp.StatusCode != http.StatusOK {
		var jsonresp NormalResponse
		err = json.NewDecoder(resp.Body).Decode(&jsonresp)
		if err != nil {
			return fmt.Errorf("error decoding response body: %w", err)
		}
		return fmt.Errorf("unexpected status code: %d, Error: %s", resp.StatusCode, jsonresp.Msg)
	}

	fmt.Println("Template updated successfully")
	return nil
}

func EditTemplate(templateId string, token string) {
	/*template := Template{
		CreatedAt:   time.Now().UTC(),
		Description: "Example Template",
		Name:        "Example",
		Topology: Topology{
			Nodes: []Node{
				{NodeID: "nd1", Name: "Node 1", Image: "cirros-0.5.1", Flavor: Flavor{FlavorID: "1", Name: "t2.micro", CPU: 1, Memory: 1.0, Storage: 8.0}},
				{NodeID: "nd2", Name: "Node 2", Image: "cirros-0.5.1", Flavor: Flavor{FlavorID: "2", Name: "t2.small", CPU: 1, Memory: 2.0, Storage: 16.0}},
			},
			Links: []Link{
				{LinkID: "link_nd1_nd2", Source: "nd1", Target: "nd2"},
			},
		},
		UserID:       "user_1",
		TopologyType: "example",
	}*/

	serverPort := 4444
	var templateById ListTemplateById
	var jsonresp NormalResponse
	requestURL := fmt.Sprintf("http://localhost:%d/templateservice/templates/%s", serverPort, templateId)
	client := &http.Client{}
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}
	req.Header.Set("X-API-Key", token)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&jsonresp)
		if err != nil {
			fmt.Printf("Error decoding response body: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Unexpected status code: %d, Error: %s\n", resp.StatusCode, jsonresp.Msg)
		os.Exit(1)
	}
	err = json.NewDecoder(resp.Body).Decode(&templateById)
	if err != nil {
		fmt.Printf("Error decoding response body: %v", err)
		os.Exit(1)
	}

	template := templateById.Template

	// Menú interactivo para editar la plantilla
	modification := Modification{}
flag:
	for {
		//printTopologyTable(template.Topology)
		modificationJSON, _ := json.MarshalIndent(modification, "", "  ")
		fmt.Printf("modification JSON:\n%s\n", string(modificationJSON))
		editTemplateTable(&template)
		fmt.Println("\nMenu:")
		fmt.Println("1. Create node")
		fmt.Println("2. Delete node")
		fmt.Println("3. Edit flavor")
		fmt.Println("4. Create link")
		fmt.Println("5. Delete link")
		fmt.Println("6. Save and exit")

		choice := promptInt("Enter your choice: ")

		switch choice {
		case 1:
			name := promptString("Enter the name of the node: ")

			images, err := fetchImages()
			if err != nil {
				fmt.Printf("Error fetching images: %v", err)
				continue flag
			}
			fmt.Println("Available images:")
			for i, image := range images {
				fmt.Printf("%d. %s (Version: %s, Description: %s)\n", i+1, image.Name, image.Version, image.Description)
			}
			imageChoice := promptInt("Enter the number of the image: ")
			image := fmt.Sprintf("%s %s", images[imageChoice-1].Name, images[imageChoice-1].Version)

			flavors, err := fetchFlavors()
			if err != nil {
				fmt.Printf("Error fetching flavors: %v", err)
				continue flag
			}
			fmt.Println("Available flavors:")
			for i, flavor := range flavors {
				fmt.Printf("%d. %s (CPU: %d, Memory: %.1fGB, Storage: %.1fGB)\n", i+1, flavor.Name, flavor.CPU, flavor.Memory, flavor.Storage)
				continue flag
			}
			flavorChoice := promptInt("Enter the number of the flavor: ")

			createNode(&template, &modification, name, image, flavors[flavorChoice-1])

		case 2:
			fmt.Println("Existing nodes:")
			for _, node := range template.Topology.Nodes {
				fmt.Printf("%s: %s\n", node.NodeID, node.Name)
			}
			nodeID := promptString("Enter the ID of the node to delete: ")
			deleteNode(&template, &modification, nodeID)

		case 3:
			fmt.Println("Existing nodes:")
			for _, node := range template.Topology.Nodes {
				fmt.Printf("%s: %s\n", node.NodeID, node.Name)
			}
			nodeID := promptString("Enter the ID of the node to edit the flavor: ")

			flavors, err := fetchFlavors()
			if err != nil {
				fmt.Printf("Error fetching flavors: %v", err)
				continue flag
			}
			fmt.Println("Available flavors:")
			for i, flavor := range flavors {
				fmt.Printf("%d. %s (CPU: %d, Memory: %.1fGB, Storage: %.1fGB)\n", i+1, flavor.Name, flavor.CPU, flavor.Memory, flavor.Storage)
			}
			flavorChoice := promptInt("Enter the number of the flavor: ")
			editFlavor(&template, &modification, nodeID, flavors[flavorChoice-1])

		case 4:
			fmt.Println("Existing nodes:")
			for _, node := range template.Topology.Nodes {
				fmt.Printf("%s: %s\n", node.NodeID, node.Name)
			}
			source := promptString("Enter the ID of the source node: ")
			target := promptString("Enter the ID of the target node: ")
			createLink(&template, &modification, source, target)
			/*if linkCreated {
				fmt.Printf("Link created between %s and %s", source, target)
			} else {
				fmt.Printf("Link cannot be created between %s and %s. Link already exists or source and target are the same", source, target)
			}*/

		case 5:
			fmt.Println("Existing links:")
			for _, link := range template.Topology.Links {
				fmt.Printf("%s: %s -> %s\n", link.LinkID, link.Source, link.Target)
			}
			linkID := promptString("Enter the ID of the link to delete: ")
			deleteLink(&template, &modification, linkID)
		case 6:
			// Espacio reservado para realizar la solicitud PUT
			fmt.Println("Modifications saved. You can now perform the PUT request to update the template.")

			// Envía la solicitud PUT con las modificaciones
			err := sendPUTRequest(serverPort, templateId, token, &template)
			if err != nil {
				fmt.Printf("Error updating template: %v\n", err)
				os.Exit(1)
			}

			// Romper el ciclo y salir del programa
			break flag
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
