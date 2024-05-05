/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// slicesCmd represents the slices command
var slicesCmd = &cobra.Command{
	Use:   "slices",
	Short: "Manage CRUD operations related to slices",
	Long:  `Manage CRUD operations related to slices`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("slices called")
	},
}

func init() {
	rootCmd.AddCommand(slicesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// slicesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// slicesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
