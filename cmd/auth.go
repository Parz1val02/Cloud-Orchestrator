/* Copyright © 2024 NAME HERE <EMAIL ADDRESS>
 */
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type User struct {
	Username string `yaml: "username"`
	Password string `yaml: "password"`
	Role     string `yaml: "role"`
}

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage credentials for PUCP private cloud",
	Long:  `Manage authentication crentials for the PUCP private cloud platform`,
}

// loginCmd
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authorize cloud-cli to access the platform",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Viper successfully read the user configuration file
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found; ignore error if desired
				var username, password string
				fmt.Print(">Enter username: ")
				fmt.Scanf("%s", &username)
				fmt.Print(">Enter password: ")
				fmt.Scanf("%s", &password)

				user := User{
					Username: username,
					Password: password,
					Role:     "administrator",
				}
				// Write user's credentials to YAML file
				yamlData, err := yaml.Marshal(&user)
				if err != nil {
					fmt.Println(">Error marshalling struct to YAML:", err)
					return
				}
				home, err := os.UserHomeDir()
				cobra.CheckErr(err)
				file, err := os.Create(home + "/.cloud-cli.yaml")
				if err != nil {
					fmt.Println(">Error creating file:", err)
					return
				}
				defer file.Close()

				_, err = file.Write(yamlData)
				if err != nil {
					fmt.Println(">Error writing to file:", err)
					return
				}
				fmt.Println(">User logged in successfully.")
			} else {
				// Config file was found but another error was produced
			}
		} else {
			username := viper.GetString("username")
			fmt.Printf(">User %s already authenticated\n", username)
		}
	},
}

// logoutCmd
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Revoke access credentials",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		filePath := home + "/.cloud-cli.yaml"

		// Check if the file exists
		if _, err := os.Stat(filePath); err == nil {
			// File exists, attempt to remove it
			if err := os.Remove(filePath); err != nil {
				fmt.Println(">Error deleting file:", err)
				return
			}
			fmt.Println(">User logged out successfully.")
		} else if os.IsNotExist(err) {
			// File does not exist, print a message
			fmt.Println(">No user authenticated")
		} else {
			// Error occurred while checking file status
			fmt.Println(">Error checking file status:", err)
		}
	},
}

// passwordCmd
var passwordCmd = &cobra.Command{
	Use:   "password",
	Short: "Change password for the authenticated user",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("password")
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd, logoutCmd, passwordCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// authCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// authCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
