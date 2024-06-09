package crud_functions

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	structs "github.com/Parz1val02/cloud-cli/structs"
)

func validateStruct(s interface{}) error {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := v.Type().Field(i).Name

		// Check if the field is exported
		if !field.CanInterface() {
			continue
		}

		if field.Kind() == reflect.Struct {
			if err := validateStruct(field.Interface()); err != nil {
				return err
			}
		} else if field.Kind() == reflect.Slice {
			for j := 0; j < field.Len(); j++ {
				if field.Index(j).Kind() == reflect.Struct {
					if err := validateStruct(field.Index(j).Interface()); err != nil {
						return err
					}
				} else if isEmptyValue(field.Index(j)) {
					return fmt.Errorf("the field '%s' must not be empty", fieldName)
				}
			}
		} else if isEmptyValue(field) {
			return fmt.Errorf("the field '%s' must not be empty", fieldName)
		}
	}
	return nil
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return false // Boolean field can be false, which is not considered empty
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Slice, reflect.Array:
		return v.Len() == 0
	case reflect.Struct:
		return false
	}
	return false
}

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

func ImportTemplate(userId, token string) {
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
	normalTemplate.UserID = userId
	normalTemplate.CreatedAt = time.Now().UTC()
	err = validateStruct(normalTemplate)
	if err != nil {
		fmt.Println("Failed validation: ", err)
		os.Exit(1)
	}
	jsonData, _ := json.Marshal(*normalTemplate)

	req, err := http.NewRequest("POST", "http://localhost:4444/templateservice/templates", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error at importing template: ", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	var jsonresp structs.NormalResponse
	err = json.NewDecoder(resp.Body).Decode(&jsonresp)
	if err != nil {
		fmt.Printf("Error decoding response body: %v", err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d, Error: %s\n", resp.StatusCode, jsonresp.Msg)
		os.Exit(1)
	}
	fmt.Printf("%s\n", jsonresp.Msg)
}
