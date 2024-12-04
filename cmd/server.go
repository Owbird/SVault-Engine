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

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share file to device",
	Long:  `Share file to device`,
	Run: func(cmd *cobra.Command, args []string) {
		svr := server.NewServer("", nil)
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			log.Fatalf("Failed to get 'file' flag: %v", err)
		}

		svr.Share(file, server.ShareCallBacks{
			OnSendErr: func(err error) {
				log.Fatalf("Send error: %s", err)
			},
			OnFileSent: func() {
				log.Println("File sent!")
			},
			OnCodeReceive: func(code string) {
				log.Println("Code: ", code)
			},
			OnProgressChange: func(progress models.FileShareProgress) {
				log.Printf("Sent: %v/%v (%v%%)", progress.Bytes, progress.Total, progress.Percentage)
			},
		})
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
				case models.SERVE_WEB_UI_NETWORK:
					if l.Error != nil {
						log.Printf("[!] Network Web Run Error: %v", l.Error)
					} else {
						log.Printf("[+] Network Web Running: %v", l.Message)
					}
				case models.SERVE_WEB_UI_REMOTE:
					if l.Error != nil {
						log.Printf("[!] Remote Web Run Error: %v", l.Error)
					} else {
						log.Printf("[+] Remote Web Running: %v", l.Message)
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
