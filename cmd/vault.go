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
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		action := args[0]

		if action == "create" {

			name, err := cmd.Flags().GetString("name")
			if err != nil || name == "" {
				log.Fatalln("name required, use -n or --name")
				return
			}

			password, err := cmd.Flags().GetString("password")
			if err != nil || password == "" {
				log.Fatalln("password required, use -p or --password")
				return
			}

			v.Create(name, password)
		}

	},
}

func init() {
	rootCmd.AddCommand(vaultCmd)

	vaultCmd.Flags().StringP("name", "n", "", "Name of new vault")
	vaultCmd.Flags().StringP("password", "p", "", "Password of the vault")
}
