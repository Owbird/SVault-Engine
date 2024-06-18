package cmd

import (
	"log"

	"github.com/Owbird/SVault-Engine/pkg/models"
	"github.com/Owbird/SVault-Engine/pkg/server"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage file server",
	Long:  `Manage file server`,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long:  `Start the server`,
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := cmd.Flags().GetString("dir")
		if err != nil {
			log.Fatalf("Failed to get 'dir' flag: %v", err)
		}

		logCh := make(chan models.ServerLog)

		defer close(logCh)

		go func() {
			for l := range logCh {
				switch l.Type {
				case "api_log":
					log.Printf("[+] API Log: %v", l.Message)

				case "serve_web_ui_local":
					log.Printf("[+] Local Web Running: %v", l.Message)

				case "serve_web_ui_remote":
					log.Printf("[+] Remote Web Running: %v", l.Message)

				}
			}
		}()

		server := server.NewServer(dir, logCh)
		server.Start()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.AddCommand(startCmd)

	startCmd.Flags().StringP("dir", "d", "", "Directory to serve")

	startCmd.MarkFlagRequired("dir")
}
