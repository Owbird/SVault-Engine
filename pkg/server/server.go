// Package server handles the file hosting server
package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/Owbird/SVault-Engine/internal/utils"
	"github.com/Owbird/SVault-Engine/pkg/config"
	"github.com/Owbird/SVault-Engine/pkg/models"
	"github.com/Owbird/SVault-Engine/pkg/server/handlers"
	"github.com/localtunnel/go-localtunnel"
	"github.com/psanford/wormhole-william/wormhole"
	"github.com/rs/cors"
)

type Server struct {
	// The current directory being hosted
	Dir string

	// The channel to send the logs through
	logCh chan models.ServerLog
}

// ShareCallBacks defines a set of callback functions for handling file sharing events.
type ShareCallBacks struct {
	// OnFileSent is called when a file has been successfully sent.
	OnFileSent func()

	// OnSendErr is called when an error occurs during the file sending process.
	OnSendErr func(err error)

	// OnProgressChange is called to provide updates on the progress of the file sharing operation.
	OnProgressChange func(progress models.FileShareProgress)

	// OnCodeReceive is called when the code to initiate the file sharing process has been received.
	OnCodeReceive func(code string)
}

const (
	PORT = 8080
)

var appConfig = config.NewAppConfig()

func sendNotification(notif models.Notification) {
	appConfig.GetNotifConfig().SendNotification(models.Notification{
		Title:         notif.Title,
		Body:          notif.Body,
		ClipboardText: notif.ClipboardText,
	})
}

func NewServer(dir string, logCh chan models.ServerLog) *Server {
	return &Server{
		Dir:   dir,
		logCh: logCh,
	}
}

// Starts starts and serves the specified dir
func (s *Server) Start() {
	s.logCh <- models.ServerLog{
		Message: "Starting server",
		Type:    models.API_LOG,
	}

	host, err := utils.GetLocalIp()
	if err != nil {
		s.logCh <- models.ServerLog{
			Error: err,
			Type:  models.SERVE_WEB_UI_NETWORK,
		}
		return
	}

	s.logCh <- models.ServerLog{
		Message: fmt.Sprintf("http://%s:%s", host, strconv.Itoa(PORT)),
		Type:    models.SERVE_WEB_UI_NETWORK,
	}

	go (func() {
		tunnel, err := localtunnel.New(PORT, "localhost", localtunnel.Options{})
		if err != nil {
			s.logCh <- models.ServerLog{
				Error: err,
				Type:  models.SERVE_WEB_UI_REMOTE,
			}
			return
		}

		sendNotification(models.Notification{
			Title:         "Web Server Ready",
			Body:          "URL copied to clipboard",
			ClipboardText: tunnel.URL(),
		})

		s.logCh <- models.ServerLog{
			Message: tunnel.URL(),
			Type:    models.SERVE_WEB_UI_REMOTE,
		}
	})()

	mux := http.NewServeMux()

	serverConfig := appConfig.GetSeverConfig()

	handlerFuncs := handlers.NewHandlers(s.logCh, s.Dir, serverConfig, appConfig.GetNotifConfig())

	mux.HandleFunc("/", handlerFuncs.GetFilesHandler)
	mux.HandleFunc("/download", handlerFuncs.DownloadFileHandler)
	mux.HandleFunc("/upload", handlerFuncs.GetFileUpload)
	mux.HandleFunc("GET /assets/{file}", handlerFuncs.GetAssets)

	corsOpts := cors.New(cors.Options{
		AllowedOrigins: []string{"https://*.loca.lt"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodOptions,
			http.MethodHead,
		},

		AllowedHeaders: []string{
			"*",
		},
	})

	s.logCh <- models.ServerLog{
		Message: fmt.Sprintf("Starting API on port %v from %v", PORT, s.Dir),
		Type:    models.API_LOG,
	}

	err = http.ListenAndServe(fmt.Sprintf(":%v", PORT), corsOpts.Handler(mux))
	if err != nil {
		s.logCh <- models.ServerLog{
			Error: err,
			Type:  models.API_LOG,
		}
	}
}

// Send a file through a wormhole from a device
// TODO: Support directories
func (s *Server) Share(file string, callbacks ShareCallBacks) {
	f, err := os.Open(file)
	if err != nil {
		callbacks.OnSendErr(err)

		return
	}

	var c wormhole.Client
	ctx := context.Background()

	progressCh := make(chan models.FileShareProgress, 1)

	handleProgress := func(sentBytes int64, totalBytes int64) {
		progressCh <- models.FileShareProgress{
			Bytes:      sentBytes,
			Total:      totalBytes,
			Percentage: int((float64(sentBytes) / float64(totalBytes)) * 100),
		}
	}

	code, st, err := c.SendFile(ctx, file, f, wormhole.WithProgress(handleProgress))

	if err != nil && callbacks.OnSendErr != nil {
		callbacks.OnSendErr(err)

		return
	}

	if callbacks.OnCodeReceive != nil {
		callbacks.OnCodeReceive(code)

		sendNotification(models.Notification{
			Title:         "Share code received",
			Body:          "Code copied to clipboard.",
			ClipboardText: code,
		})
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		for {
			select {
			case status := <-st:
				if status.Error != nil && callbacks.OnSendErr != nil {
					callbacks.OnSendErr(status.Error)

					return
				}

				if !status.OK && status.Error != nil && callbacks.OnSendErr != nil {
					callbacks.OnSendErr(fmt.Errorf("unknown error occurred"))
					return

				} else {
					if callbacks.OnFileSent != nil {
						callbacks.OnFileSent()
					}

					wg.Done()
					return
				}

			case progress := <-progressCh:
				if callbacks.OnProgressChange != nil {
					callbacks.OnProgressChange(progress)
				}
			}
		}
	}()

	wg.Wait()
}

// Receive file from device through wormhole
// TODO: Support output dir
func (s *Server) Receive(code string) error {
	var c wormhole.Client

	ctx := context.Background()
	fileInfo, err := c.Receive(ctx, code)
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, fileInfo)
	if err != nil {
		return err
	}

	sendNotification(models.Notification{
		Title: "File received",
		Body:  fmt.Sprintf("File %v received", fileInfo.Name),
	})

	return nil
}
