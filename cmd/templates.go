/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	crud "github.com/Parz1val02/cloud-cli/crud_functions"
	simplelist "github.com/Parz1val02/cloud-cli/simplelist"
	simpletable "github.com/Parz1val02/cloud-cli/simpletable"
	tabs "github.com/Parz1val02/cloud-cli/tabs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initialModelCRUD1() simplelist.Model {
	return simplelist.Model{
		Choices:  []string{"List templates", "Create template", "Import template"},
		Selected: make(map[int]struct{}),
	}
}

func initialModelCRUD2() simplelist.Model {
	return simplelist.Model{
		Choices:  []string{"List template configuration", "Edit template", "Delete template", "Graph template", "Export template"},
		Selected: make(map[int]struct{}),
	}
}

func listTemplates() {
	templateId, err := simpletable.MainTable()
	if err != nil {
		fmt.Println(err)
	}
	if templateId != "" {
		p := tea.NewProgram(initialModelCRUD2())
		m, err := p.Run()
		if err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
		if m, ok := m.(simplelist.Model); ok && m.Choices[m.Cursor] != "" {
			if m.Quit {
				fmt.Printf("\n---\nQuitting!\n")
			} else {
				fmt.Printf("\n---\nYou chose %s!\n", m.Choices[m.Cursor])
				switch m.Cursor {
				case 0:
					tabs.MainTabs(templateId)
				default:

				}
			}
		}
	}
}

// topologiesCmd represents the topologies command
var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage CRUD operations related to templates",
	Long:  `Manage CRUD operations to templates`,
	Run: func(cmd *cobra.Command, args []string) {
		myFigure := figure.NewFigure("PUCP Private Cloud Orchestrator", "doom", true)
		myFigure.Print()
		fmt.Println()
		p := tea.NewProgram(initialModelCRUD1())
		m, err := p.Run()
		if err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
		if m, ok := m.(simplelist.Model); ok && m.Choices[m.Cursor] != "" {
			if m.Quit {
				fmt.Printf("\n---\nQuitting!\n")
			} else {
				fmt.Printf("\n---\nYou chose %s!\n", m.Choices[m.Cursor])
				fmt.Print("\n---\nSelect a template to execute CRUD operation on\n")
				switch m.Cursor {
				case 0:
					listTemplates()
				case 1:
					crud.CreateTemplate() // funcion crear plantilla
				default:

				}
			}
		}
	},
}

func init() {
	initConfig()
	err := viper.ReadInConfig()
	if err == nil {
		rootCmd.AddCommand(templatesCmd)
	}
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// topologiesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// topologiesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
