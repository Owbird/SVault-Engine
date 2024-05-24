package cmd

import (
	"log"

	"github.com/Owbird/SVault-Engine/pkg/vault"
	"github.com/spf13/cobra"
)

var v = vault.NewVault()

// vaultCmd represents the vault command
var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage vaults",
	Long:  `Manage vaults`,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new vault",
	Long:  `Create a new vault with a specified name and password`,
	Run: func(cmd *cobra.Command, args []string) {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatalf("Failed to get 'name' flag: %v", err)
		}

		password, err := cmd.Flags().GetString("password")
		if err != nil {
			log.Fatalf("Failed to get 'password' flag: %v", err)
		}

		v.Create(name, password)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List created vaults",
	Long:  `List all created vaults`,
	Run: func(cmd *cobra.Command, args []string) {
		vaults, err := v.List()
		if err != nil {
			log.Fatalf("Failed to fetch vaults: %v", err)
		}

		log.Println(vaults)
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add file to vault",
	Long:  `Add file to vault`,
	Run: func(cmd *cobra.Command, args []string) {
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			log.Fatalf("Failed to get 'file' flag: %v", err)
		}

		vault, err := cmd.Flags().GetString("vault")
		if err != nil {
			log.Fatalf("Failed to get 'vault' flag: %v", err)
		}

		password, err := cmd.Flags().GetString("password")
		if err != nil {
			log.Fatalf("Failed to get 'password' flag: %v", err)
		}

		pwdMatch, err := v.Auth(vault, password)
		if err != nil {
			log.Fatalf("Failed to auth vault: %v", err)
		}

		if !pwdMatch {
			log.Fatal("Passwords do not match")
		}

		log.Println(file, vault, password)
	},
}

func init() {
	rootCmd.AddCommand(vaultCmd)

	vaultCmd.AddCommand(createCmd)
	vaultCmd.AddCommand(listCmd)
	vaultCmd.AddCommand(addCmd)

	createCmd.Flags().StringP("name", "n", "", "Name of new vault")
	createCmd.Flags().StringP("password", "p", "", "Password of the vault")

	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("password")

	addCmd.Flags().StringP("file", "f", "", "Path to the file")
	addCmd.Flags().StringP("vault", "v", "", "Name of vault to use")
	addCmd.Flags().StringP("password", "p", "", "Password to the vault")

	addCmd.MarkFlagRequired("file")
	addCmd.MarkFlagRequired("vault")
	addCmd.MarkFlagRequired("password")
}
