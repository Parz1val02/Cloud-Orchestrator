/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	configs "github.com/Parz1val02/cloud-cli/configs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// topologiesCmd represents the topologies command
var topologiesCmd = &cobra.Command{
	Use:   "topologies",
	Short: "Manage CRUD operations related to topologies",
	Long:  `Manage CRUD operations to topologies`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(initialModelTopologies())
		m, err := p.Run()
		if err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
		if m, ok := m.(configs.Model); ok && m.Choices[m.Cursor] != "" {
			if m.Quit {
				fmt.Printf("\n---\nQuitting!\n")
			} else {
				fmt.Printf("\n---\nYou chose %s!\n", m.Choices[m.Cursor])
			}
		}
	},
}

func initialModelTopologies() configs.Model {
	return configs.Model{
		Choices:  []string{"Create topology", "List topologies", "Edit topology", "Delete topoly"},
		Selected: make(map[int]struct{}),
	}
}

func init() {
	rootCmd.AddCommand(topologiesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// topologiesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// topologiesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
