/* Copyright © 2024 NAME HERE <EMAIL ADDRESS>
 */
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"syscall"

	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

type User struct {
	ID       string `yaml:"id" json:"ID"`
	Username string `yaml:"username" json:"Username"`
	Role     string `yaml:"role" json:"Role"`
	Token    string `yaml:"token" json:"Token"`
}

func PasswordPrompt(label string) string {
	var s string
	for {
		fmt.Fprint(os.Stderr, label)
		b, _ := term.ReadPassword(int(syscall.Stdin))
		s = string(b)
		if s != "" {
			break
		}
	}
	fmt.Println()
	return s
}

func login(username, password string) (User, error) {
	authData := map[string]string{
		"username": username,
		"password": password,
	}

	jsonData, _ := json.Marshal(authData)
	resp, err := http.Post("http://localhost:4444/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error in authentication", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	var data struct {
		User User `json:"user"`
	}

	var user User
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
		return user, err
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		err = fmt.Errorf("Error decoding response body: %v", err)
		return user, err
	}
	return data.User, nil
}

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage credentials for PUCP private cloud",
	Long:  `Manage authentication crentials for the PUCP private cloud platform`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			myFigure := figure.NewFigure("PUCP Private Cloud Orchestrator", "doom", true)
			myFigure.Print()
			fmt.Println()
			err := cmd.Help()
			if err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		}
	},
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
				var username string
				var user User
				fmt.Printf(">Enter username: ")
				fmt.Scanf("%s\n", &username)
				password := PasswordPrompt(">Enter password: ")
				user, err = login(username, password)
				if err != nil {
					fmt.Println("Error: " + err.Error())
					return
				}

				// Write user's credentials to YAML file
				yamlData, err := yaml.Marshal(&user)
				if err != nil {
					fmt.Println(">Error marshalling struct to YAML", err)
					return
				}
				home, err := os.UserHomeDir()
				cobra.CheckErr(err)
				file, err := os.Create(home + "/.cloud-cli.yaml")
				if err != nil {
					fmt.Println(">Error creating file", err)
					return
				}
				defer file.Close()

				_, err = file.Write(yamlData)
				if err != nil {
					fmt.Println(">Error writing to file", err)
					return
				}
				fmt.Println(">UseR logged in successfully.")
			} else {
				// Config file was found but another error was produced
				fmt.Println(">Configurations loaded but other error happened")
				return
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

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				fmt.Println(">No user authenticated")
			} else {
				// Config file was found but another error was produced
				fmt.Println(">Configurations loaded but other error happened")
				return
			}
		} else {
			if err := os.Remove(filePath); err != nil {
				fmt.Println(">Error deleting file:", err)
				return
			}
			fmt.Println(">User logged out successfully.")
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
