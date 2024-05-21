/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	simplelist "github.com/Parz1val02/cloud-cli/simplelist"
	simpletable "github.com/Parz1val02/cloud-cli/simpletable"
	tabs "github.com/Parz1val02/cloud-cli/tabs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// topologiesCmd represents the topologies command
var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage CRUD operations related to templates",
	Long:  `Manage CRUD operations to templates`,
	Run: func(cmd *cobra.Command, args []string) {
		myFigure := figure.NewFigure("PUCP Private Cloud Orchestrator", "doom", true)
		myFigure.Print()
		fmt.Println()
		p := tea.NewProgram(initialModelTopologies())
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
					listTemplates()
				default:

				}
			}
		}
	},
}

func initialModelTopologies() simplelist.Model {
	return simplelist.Model{
		Choices:  []string{"List templates", "Create template", "Edit template", "Delete template", "Graph template", "Import template", "Export template"},
		Selected: make(map[int]struct{}),
	}
}

func listTemplates() {
	templateId, err := simpletable.MainTable()
	if err != nil {
		fmt.Println(err)
	}
    tabs.MainTabs(templateId)
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
