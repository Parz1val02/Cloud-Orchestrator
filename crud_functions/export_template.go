package crud_functions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	structs "github.com/Parz1val02/cloud-cli/structs"
)

func ExportTemplate(templateId, token string) error {
	serverPort := 4444
	var templateById structs.ListTemplateById
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
		fmt.Printf("Unexpected status code: %d\n", resp.StatusCode)
	}
	err = json.NewDecoder(resp.Body).Decode(&templateById)
	if err != nil {
		fmt.Printf("Error decoding response body: %v\n", err)
	}

	if templateById.Result == "success" && templateById.Template.TemplateID == templateId {
		exportTemplate := structs.NormalTemplate{
			CreatedAt:        templateById.Template.CreatedAt,
			AvailabilityZone: templateById.Template.AvailabilityZone,
			Deployed:         templateById.Template.Deployed,
			Description:      templateById.Template.Description,
			Name:             templateById.Template.Name,
			Topology:         templateById.Template.Topology,
			UserID:           "",
			VlanID:           templateById.Template.VlanID,
			TopologyType:     templateById.Template.TopologyType,
		}
		jsonData, err := json.MarshalIndent(exportTemplate, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling export template: %v", err)
		}
		reader := bufio.NewReader(os.Stdin)
		var filePath string
		for {
			fmt.Print("Enter the absolute path or file name to export the template (must be a .json file): ")
			filePath, _ = reader.ReadString('\n')
			filePath = strings.TrimSpace(filePath)

			if !strings.HasSuffix(filePath, ".json") {
				fmt.Println("Error: File name must end with .json")
				continue
			}
			break
		}
		if !filepath.IsAbs(filePath) {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("error getting current directory: %v", err)
			}
			filePath = filepath.Join(cwd, filePath)
		}
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error creating file: %v", err)
		}
		defer file.Close()

		_, err = file.Write(jsonData)
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}

		fmt.Println("Template exported successfully to ", filePath)
		return nil
	} else {
		return fmt.Errorf("Error in http request")
	}
}
