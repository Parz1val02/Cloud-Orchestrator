/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	simplelist "github.com/Parz1val02/cloud-cli/simplelist"
	simpletable "github.com/Parz1val02/cloud-cli/simpletable"
	tabs "github.com/Parz1val02/cloud-cli/tabs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func createSliceRequestHttp(templateId string, token string, avZone string, internet bool, deploymentType string) {
	serverPort := 4444
	requestURL := fmt.Sprintf("http://localhost:%d/sliceservice/slices", serverPort)

	now := time.Now().UTC()
	//formattedTime := now.Format("2006-01-02 15:04:05")
	// Parámetros de la solicitud en formato JSON
	jsonData := map[string]interface{}{
		"template_id":       templateId,
		"availability_zone": avZone,
		"deployment_type":   deploymentType,
		"internet":          internet,
		"created_at":        now,
	}

	// Codificar los parámetros como JSON
	jsonValue, err := json.Marshal(jsonData)
	//fmt.Printf("Generated JSON:\n%s\n", string(jsonValue))

	if err != nil {
		fmt.Println("Error al codificar parámetros:", err)
		return
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonValue))
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

	// Estructura para deserializar la respuesta
	type ResponseCreateSlice struct {
		Result string `json:"result"`
		Msg    string `json:"msg"`
	}

	// Leer la respuesta
	var result ResponseCreateSlice

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Printf("Error decoding response body create slice http: %v", err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d, Error: %s\n", resp.StatusCode, result.Msg)
		os.Exit(1)
	}
	// Mostrar la respuesta
	fmt.Println("Respuesta:", result)

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

func selectDeploymentType() string {
	deployment_type := []string{"linux", "openstack"}
	for i, name := range deployment_type {
		fmt.Printf("%d. %s\n", i+1, name)
	}
	var choice int
	for {
		choice = promptInt("Enter the number of the chosen deployment type: ")
		if choice > 0 && choice <= len(deployment_type) {
			break
		}
		fmt.Println("Invalid choice. Please enter a valid number.")
	}
	return deployment_type[choice-1]
}

func selectInternet() bool {
	internet := false
	for {
		user_selection := promptString("Connection to internet (y/n): ")
		if user_selection == "y" {
			internet = true
			break
		} else if user_selection == "n" {
			break
		}
	}

	return internet
}

type Server struct {
	Name string `json:"name"`
}

// Estructura para las zonas de disponibilidad
type AvailabilityZone struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Servers []Server `json:"servers"`
}

func selectAvailabilityZone() string {
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
	// Solicitar al usuario que ingrese el número correspondiente a la av elegida
	var choice int
	for {
		choice = promptInt("Enter the number of the availability zone: ")
		if choice > 0 && choice <= len(availabilityZones) {
			break
		}
		fmt.Println("Invalid choice. Please enter a valid number.")
	}
	// Devolver la imagen seleccionada
	return availabilityZones[choice-1].Name
}
func fetchAvailabilityZone() ([]AvailabilityZone, error) {
	url := "http://localhost:4444/templateservice/templates/avz"

	var token = viper.GetString("token")

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}
	req.Header.Set("X-API-Key", token)

	//fmt.Println("TOKEN", token)
	resp, err := client.Do(req)
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

func createSlice() {
	token := viper.GetString("token")
	templateId, err := simpletable.MainTable(token)
	if err != nil {
		fmt.Println(err)
	}
	for {
		if templateId != "" {
			deployment_type := selectDeploymentType()
			internet := selectInternet()
			av_zone := selectAvailabilityZone()
			createSliceRequestHttp(templateId, token, av_zone, internet, deployment_type)
			break
		} else {
			break
		}
	}
}

func sliceModelCRUD2() simplelist.Model {
	return simplelist.Model{
		Choices:  []string{"List slice configuration", "Edit slice", "Delete slice", "Graph slice"},
		Selected: make(map[int]struct{}),
	}
}

func listSlices() {
	token := viper.GetString("token")
	sliceId, err := simpletable.SliceTable(token)
	if err != nil {
		fmt.Println(err)
	}
	for {
		if sliceId != "" {
			p := tea.NewProgram(sliceModelCRUD2())
			m, err := p.Run()
			if err != nil {
				fmt.Printf("Alas, there's been an error: %v", err)
				os.Exit(1)
			}
			if m, ok := m.(simplelist.Model); ok && m.Choices[m.Cursor] != "" {
				if m.Quit {
					fmt.Printf("\n---\nQuitting!\n")
					break
				} else {
					fmt.Printf("\n---\nYou chose %s!\n", m.Choices[m.Cursor])
					switch m.Cursor {
					case 0:
						tabs.SliceInfoTabs(sliceId, token)
					case 1:
						//crud.EditSlice(sliceId, token)

					case 2:
						/*var option string
						fmt.Printf(">Are you sure you want to delete slice with id %s? (y/N): ", sliceId)
						fmt.Scanf("%s\n", &option)
						if option != "" && option == "y" || option == "Y" {
							error := crud.DeleteSlice(sliceId, token)
							if error != nil {
								fmt.Println("Error:", err)
								os.Exit(1)
							}
							break
						}*/
					default:

					}
				}
			}
		} else {
			break
		}
	}
}

// slicesCmd represents the slices command
var slicesCmd = &cobra.Command{
	Use:   "slices",
	Short: "Manage CRUD operations related to slices",
	Long:  `Manage CRUD operations related to slices`,
	Run: func(cmd *cobra.Command, args []string) {
		myFigure := figure.NewFigure("PUCP Private Cloud Orchestrator", "doom", true)
		myFigure.Print()
		fmt.Println()
		for {
			p := tea.NewProgram(initialModelSlices())
			m, err := p.Run()
			if err != nil {
				fmt.Printf("Alas, there's been an error: %v", err)
				os.Exit(1)
			}
			if m, ok := m.(simplelist.Model); ok && m.Choices[m.Cursor] != "" {
				if m.Quit {
					fmt.Printf("\n---\nQuitting!\n")
					break
				} else {
					fmt.Printf("\n---\nYou chose %s!\n", m.Choices[m.Cursor])
					switch m.Cursor {
					case 0:
						createSlice()
					case 1:
						fmt.Print("\n---\nSelect a slice to execute CRUD operation on\n")
						listSlices()
					default:

					}
				}
			}
		}
	},
}

func initialModelSlices() simplelist.Model {
	return simplelist.Model{
		Choices:  []string{"Create slice", "List slices"},
		Selected: make(map[int]struct{}),
	}
}

func init() {
	initConfig()
	err := viper.ReadInConfig()
	if err == nil {
		rootCmd.AddCommand(slicesCmd)
	}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// slicesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// slicesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
