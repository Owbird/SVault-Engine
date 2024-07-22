package cmd

import (
	"log"
	"sync"

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

		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			for l := range logCh {
				switch l.Type {
				case "api_log":
					if l.Error != nil {
						log.Printf("[!] API Error: %v", l.Error)
					} else {
						log.Printf("[+] API Log: %v", l.Message)
					}
				case "serve_web_ui_local":
					if l.Error != nil {
						log.Printf("[!] Local Web Run Error: %v", l.Error)
					} else {
						log.Printf("[+] Local Web Running: %v", l.Message)
					}

				case "serve_web_ui_remote":
					if l.Error != nil {
						log.Printf("[!] Remote Web Run Error: %v", l.Error)
					} else {
						log.Printf("[+] Remote Web Running: %v", l.Message)
					}
				case "web_deps_installation":
					if l.Error != nil {
						log.Printf("[!] Web Dependencies Installation Error: %v", l.Error)
					} else {
						log.Printf("[+] Web Dependencies Installed: %v", l.Message)
					}
				case "web_ui_build":
					if l.Error != nil {
						log.Printf("[!] Web UI Build error: %v", l.Error)
					} else {
						log.Printf("[+] Web UI Built: %v", l.Message)
					}
				}
			}
		}()

		server := server.NewServer(dir, logCh)

		wg.Add(1)
		go server.Start()

		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.AddCommand(startCmd)

	startCmd.Flags().StringP("dir", "d", "", "Directory to serve")

	startCmd.MarkFlagRequired("dir")
}
