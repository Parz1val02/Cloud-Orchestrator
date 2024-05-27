package crud_functions

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	structs "github.com/Parz1val02/cloud-cli/structs"
)

func promptForFilePath() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter the absolute path or file name to read the import the template (must be a .json file): ")
		filePath, _ := reader.ReadString('\n')
		filePath = strings.TrimSpace(filePath)
		if !strings.HasSuffix(filePath, ".json") {
			fmt.Println("Error: File name must end with .json")
			continue
		}
		if !filepath.IsAbs(filePath) {
			cwd, err := os.Getwd()
			if err != nil {
				return "", fmt.Errorf("error getting current directory: %v", err)
			}
			filePath = filepath.Join(cwd, filePath)
		}
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Println("Error: File does not exist")
			continue
		}
		return filePath, nil
	}
}

func readAndUnmarshalFile(filePath string) (*structs.NormalTemplate, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var normalTemplate structs.NormalTemplate
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&normalTemplate)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON file: %v", err)
	}

	return &normalTemplate, nil
}

func ImportTemplate() {
	filePath, err := promptForFilePath()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	normalTemplate, err := readAndUnmarshalFile(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	jsonData, _ := json.Marshal(*normalTemplate)
	resp, err := http.Post("http://localhost:5000/templates", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error at importing template: ", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	var jsonresp structs.ResponseCreate

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d", resp.StatusCode)
		os.Exit(1)
	}
	err = json.NewDecoder(resp.Body).Decode(&jsonresp)
	if err != nil {
		fmt.Printf("Error decoding response body: %v", err)
		os.Exit(1)
	}

	if jsonresp.Result == "success" {
		fmt.Printf("Successfully imported and created tempalte with id %s\n", jsonresp.TemplateID)
	}
}
