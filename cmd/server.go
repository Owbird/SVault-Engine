package cmd

import (
	"log"
	"sync"

	"github.com/Owbird/SVault-Engine/pkg/models"
	"github.com/Owbird/SVault-Engine/pkg/server"
	"github.com/psanford/wormhole-william/wormhole"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage file server",
	Long:  `Manage file server`,
}

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share file to device",
	Long:  `Share file to device`,
	Run: func(cmd *cobra.Command, args []string) {
		server := server.NewServer("", nil)
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			log.Fatalf("Failed to get 'file' flag: %v", err)
		}

		res := make(chan wormhole.SendResult)

		defer close(res)

		code, st, err := server.Share(file)

		log.Println("Code: ", code)

		status := <-st

		if status.Error != nil {
			log.Fatalf("Send error: %s", status.Error)
		}

		log.Println("File sent!")
	},
}

var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "Receive file from device",
	Long:  `Receive file from device`,
	Run: func(cmd *cobra.Command, args []string) {
		server := server.NewServer("", nil)
		code, err := cmd.Flags().GetString("code")
		if err != nil {
			log.Fatalf("Failed to get 'code' flag: %v", err)
		}
		server.Receive(code)
	},
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
				case models.API_LOG:
					if l.Error != nil {
						log.Printf("[!] API Error: %v", l.Error)
					} else {
						log.Printf("[+] API Log: %v", l.Message)
					}
				case models.SERVE_WEB_UI_LOCAL:
					if l.Error != nil {
						log.Printf("[!] Local Web Run Error: %v", l.Error)
					} else {
						log.Printf("[+] Local Web Running: %v", l.Message)
					}

				case models.SERVE_WEB_UI_REMOTE:
					if l.Error != nil {
						log.Printf("[!] Remote Web Run Error: %v", l.Error)
					} else {
						log.Printf("[+] Remote Web Running: %v", l.Message)
					}
				case models.WEB_DEPS_INSTALLATION:
					if l.Error != nil {
						log.Printf("[!] Web Dependencies Installation Error: %v", l.Error)
					} else {
						log.Printf("[+] Web Dependencies Installed: %v", l.Message)
					}
				case models.WEB_UI_BUILD:
					if l.Error != nil {
						log.Printf("[!] Web UI Build error: %v", l.Error)
					} else {
						log.Printf("[+] Web UI Built: %v", l.Message)
					}
				default:
					if l.Error != nil {
						log.Printf("[!] Server Error: %v", l.Error)
					} else {
						log.Printf("[+] Server Log: %v", l.Message)
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
	serverCmd.AddCommand(shareCmd)
	serverCmd.AddCommand(receiveCmd)

	startCmd.Flags().StringP("dir", "d", "", "Directory to serve")

	shareCmd.Flags().StringP("file", "f", "", "File to share")

	receiveCmd.Flags().StringP("code", "c", "", "Code from other device")

	startCmd.MarkFlagRequired("dir")
	shareCmd.MarkFlagRequired("file")
	receiveCmd.MarkFlagRequired("code")
}
