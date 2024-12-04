// Package server handles the file hosting server
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/Owbird/SVault-Engine/internal/utils"
	"github.com/Owbird/SVault-Engine/pkg/config"
	"github.com/Owbird/SVault-Engine/pkg/models"
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

type File struct {
	// The name of the file
	Name string `json:"name"`

	// Whether it's a file or directory
	IsDir bool `json:"is_dir"`

	// Size of the file in bytes
	Size int64 `json:"size"`
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

var (
	webUIPath string
	appConfig = config.NewAppConfig()
)

func sendNotification(notif models.Notification) {
	appConfig.GetNotifConfig().SendNotification(models.Notification{
		Title:         notif.Title,
		Body:          notif.Body,
		ClipboardText: notif.ClipboardText,
	})
}

func NewServer(dir string, logCh chan models.ServerLog) *Server {
	userDir, _ := utils.GetSVaultDir()

	webUIPath = filepath.Join(userDir, "web_ui")

	return &Server{
		Dir:   dir,
		logCh: logCh,
	}
}

func (s *Server) getFileUpload(w http.ResponseWriter, r *http.Request) {
	s.logCh <- models.ServerLog{
		Message: "Receiving files",
		Type:    models.API_LOG,
	}

	// TODO: Make limit configurable
	// 100MB Limit
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	files := r.MultipartForm.File["file"]

	uploadDir := r.FormValue("uploadDir")

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		uploadDir := filepath.Join(s.Dir, uploadDir, fileHeader.Filename)

		dst, err := os.Create(uploadDir)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.logCh <- models.ServerLog{
			Message: fmt.Sprintf("File received at %v", uploadDir),
			Type:    models.API_LOG,
		}

	}

	return
}

func (s *Server) downloadFileHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	if len(query["file"]) > 0 {
		if filepath.Dir(query["file"][0]) == ".." || filepath.Base(query["file"][0]) == ".." {
			http.Error(w, "Failed to download file", http.StatusInternalServerError)
			return

		}

		file := filepath.Join(s.Dir, query["file"][0])

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%v", filepath.Base(file)))
		w.Header().Set("Content-Type", "application/octet-stream")

		s.logCh <- models.ServerLog{
			Message: fmt.Sprintf("Downloading %v", file),
			Type:    models.API_LOG,
		}

		http.ServeFile(w, r, file)
		return
	}

	http.Error(w, "Failed to download file", http.StatusBadRequest)
	return
}

func (s *Server) getServerConfig(w http.ResponseWriter, _ *http.Request) {
	serverConfig := appConfig.ToJson()["server"]

	configJson, err := json.Marshal(serverConfig)
	if err != nil {
		http.Error(w, "Failed to get server", http.StatusInternalServerError)
		return
	}

	s.logCh <- models.ServerLog{
		Message: "Getting server config",
		Type:    models.API_LOG,
	}

	w.Write(configJson)
	return
}

func (s *Server) getFilesHandler(w http.ResponseWriter, r *http.Request) {
	files := []File{}

	query := r.URL.Query()

	var dir string

	if len(query["dir"]) > 0 {
		if filepath.Base(query["dir"][0]) == ".." {
			http.Error(w, "Failed to list files", http.StatusInternalServerError)
			return

		}

		dir = filepath.Join(s.Dir, query["dir"][0])

	} else {
		dir = s.Dir
	}

	s.logCh <- models.ServerLog{
		Message: fmt.Sprintf("Getting files for %v", dir),
		Type:    models.API_LOG,
	}

	dirFiles, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, "Failed to list files", http.StatusInternalServerError)
		return
	}

	for _, file := range dirFiles {

		info, err := file.Info()
		if err != nil {
			http.Error(w, "Failed to list files", http.StatusInternalServerError)
			return
		}

		files = append(files, File{
			Name:  file.Name(),
			IsDir: file.IsDir(),
			Size:  info.Size(),
		})
	}

	filesJson, err := json.Marshal(files)
	if err != nil {
		http.Error(w, "Failed to list files", http.StatusInternalServerError)
		return
	}

	w.Write(filesJson)
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

	mux.HandleFunc("/", s.getFilesHandler)
	mux.HandleFunc("/config", s.getServerConfig)
	mux.HandleFunc("/download", s.downloadFileHandler)
	mux.HandleFunc("/upload", s.getFileUpload)

	corsOpts := cors.New(cors.Options{
		AllowedOrigins: []string{"https://*.loca.lt", "http://localhost:3000", "http://localhost:3001"},
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
