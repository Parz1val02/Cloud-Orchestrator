package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Image struct {
	ImageID     string `bson:"id"`
	Name        string `bson:"name"`
	Version     string `bson:"version"`
	Description string `bson:"description"`
	ImageURL    string `bson:"image_url"`
}

type Port struct {
	PortID string `bson:"node_id"`
}

type Node struct {
	NodeID        string `bson:"node_id"`
	Name          string `bson:"name"`
	Image         string `bson:"image"`
	Flavor        Flavor `bson:"flavor"`
	SecurityRules []int  `bson:"security_rules"`
	Ports         []Port `bson:"ports"`
}

type Flavor struct {
	FlavorID string  `bson:"id"`
	Name     string  `bson:"name"`
	CPU      int     `bson:"cpu"`
	Memory   float32 `bson:"memory"`  // en GB
	Storage  float32 `bson:"storage"` // en GB
}

type Link struct {
	LinkID     string `bson:"link_id"`
	Source     string `bson:"source"`
	Target     string `bson:"target"`
	SourcePort string `bson:"source_port"`
	TargetPort string `bson:"target_port"`
}

type Topology struct {
	Links []Link `bson:"links"`
	Nodes []Node `bson:"nodes"`
}

type Template struct {
	TemplateID   string    `bson:"_id"`
	CreatedAt    time.Time `bson:"created_at"`
	Description  string    `bson:"description"`
	Name         string    `bson:"name"`
	Topology     Topology  `bson:"topology"`
	UserID       string    `bson:"user_id"`
	TopologyType string    `bson:"topology_type"`
}

var (
	mongoClient2 *mongo.Client
	collection2  *mongo.Collection
)

func mongoInit2() {
	uri := "mongodb://localhost:27017"

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	mongoClient2 = client
	collection2 = client.Database("cloud").Collection("templates")
}

func graphTemplateTopology(template Template) error {
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

	file, err := os.Create(template.TemplateID + ".html")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(htmlContent)
	if err != nil {
		return err
	}
	return nil
}

func graphHandler(c *gin.Context) {
	id := c.Param("id")
	fmt.Printf("%s id received\n", id)
	mongoInit2()
	defer func() {
		if err := mongoClient2.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID}

	var template Template
	err := collection2.FindOne(context.Background(), filter).Decode(&template)
	fmt.Printf("%s id received\n", template.TemplateID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			return
		}
		fmt.Printf("Error finding template: %s\n", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	err = graphTemplateTopology(template)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error with graph"})
		return
	}
	c.File(template.TemplateID + ".html")
}
