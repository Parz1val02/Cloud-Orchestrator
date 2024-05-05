/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// topologiesCmd represents the topologies command
var topologiesCmd = &cobra.Command{
	Use:   "topologies",
	Short: "Manage CRUD operations related to topologies",
	Long: `Manage CRUD operations to topologies`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("topologies called")
	},
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
