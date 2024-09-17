package cmd

import (
	"log"
	"os"

	"github.com/Owbird/SVault-Engine/pkg/vault"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

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

		vault := vault.NewVault()
		vault.Create(name, password)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List created vaults",
	Long:  `List all created vaults`,
	Run: func(cmd *cobra.Command, args []string) {
		vault := vault.NewVault()
		vaults, err := vault.List()
		if err != nil {
			log.Fatalf("Failed to fetch vaults: %v", err)
		}

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"#", "Vault", "Created At"})
		for index, vault := range vaults {
			t.AppendRows([]table.Row{
				{index + 1, vault.Name, vault.CreatedAt},
			})
			t.AppendSeparator()
		}
		t.Render()
	},
}

var fileCmd = &cobra.Command{
	Use:   "files",
	Short: "Manage vault files",
	Long:  `Manage vault files`,
	Run: func(cmd *cobra.Command, args []string) {
		v := vault.NewVault()
		vault, err := cmd.Flags().GetString("vault")
		if err != nil {
			log.Fatalf("Failed to get 'vault' flag: %v", err)
		}

		password, err := cmd.Flags().GetString("password")
		if err != nil {
			log.Fatalf("Failed to get 'password' flag: %v", err)
		}

		fileToAdd, err := cmd.Flags().GetString("add")
		if err != nil {
			log.Fatalf("Failed to get 'file' flag: %v", err)
		}

		if fileToAdd != "" {
			err = v.Add(fileToAdd, vault, password)
			if err != nil {
				log.Fatalf("Failed to add file to vault: %v", err)
			}
		}

		fileToDelete, err := cmd.Flags().GetString("delete")
		if err != nil {
			log.Fatalf("Failed to get 'file' flag: %v", err)
		}

		if fileToDelete != "" {
			err = v.DeleteFile(fileToDelete, vault, password)
			if err != nil {
				log.Fatalf("Failed to delete file from vault: %v", err)
			}
		}

		listFiles, err := cmd.Flags().GetBool("list")
		if err != nil {
			log.Fatalf("Failed to get 'list' flag: %v", err)
		}

		if listFiles {
			files, err := v.ListFileVaults(vault, password)
			if err != nil {
				log.Fatalf("Failed to get vault files: %v", err)
			}

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"#", "File", "Size", "Mode", "ModTime"})
			for index, file := range files {
				t.AppendRows([]table.Row{
					{index + 1, file.Name, file.Size, file.Mode, file.ModTime},
				})
				t.AppendSeparator()
			}
			t.Render()

		}
	},
}

func init() {
	rootCmd.AddCommand(vaultCmd)

	vaultCmd.AddCommand(createCmd)
	vaultCmd.AddCommand(listCmd)
	vaultCmd.AddCommand(fileCmd)

	createCmd.Flags().StringP("name", "n", "", "Name of new vault")
	createCmd.Flags().StringP("password", "p", "", "Password of the vault")

	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("password")

	fileCmd.Flags().StringP("add", "a", "", "Add file to the vault")
	fileCmd.Flags().StringP("delete", "d", "", "Delete file from the vault")
	fileCmd.Flags().StringP("password", "p", "", "Password of the vault")
	fileCmd.Flags().StringP("vault", "v", "", "Name of the vault")
	fileCmd.Flags().BoolP("list", "l", false, "List files in the vault")

	fileCmd.MarkFlagRequired("vault")
	fileCmd.MarkFlagRequired("password")
}
