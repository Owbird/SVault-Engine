// Package server handles the file hosting server
package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Owbird/SVault-Engine/internal/utils"
	"github.com/Owbird/SVault-Engine/pkg/models"
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		// The server label to be displayed
		Name string `json:"name"`

		// Should uploads be allowed
		AllowUploads bool `json:"allow_uploads"`
	} `json:"server"`
}

type Server struct {
	// The current directory being hosted
	Dir string

	// The channel to send the logs through
	logCh chan models.ServerLog

	// The server TOML configuration
	Config Config
}

type File struct {
	// The name of the file
	Name string `json:"name"`

	// Whether it's a file or directory
	IsDir bool `json:"is_dir"`

	// Size of the file in bytes
	Size int64 `json:"size"`
}

const (
	PORT = 8080
)

var webUIPath string

func NewServer(dir string, logCh chan models.ServerLog) *Server {
	userDir, err := utils.GetSVaultDir()
	if err != nil {
		log.Fatalln("Failed to get user dir")
	}

	webUIPath = filepath.Join(userDir, "web_ui")

	viper.SetConfigName("svault")
	viper.SetConfigType("toml")

	viper.AddConfigPath(userDir)

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln(err)
	}

	viper.SetDefault("server.name", fmt.Sprintf("%v's Server", hostname))
	viper.SetDefault("server.allowUploads", false)

	err = viper.ReadInConfig()
	if err != nil {
		viper.SafeWriteConfig()
	}

	var config Config

	viper.Unmarshal(&config)

	return &Server{
		Dir:    dir,
		logCh:  logCh,
		Config: config,
	}
}

func (s *Server) buildUI() (string, error) {
	commands := []map[string]interface{}{
		{
			"type":    "web_deps_installation",
			"step":    "Installing dependencies",
			"command": "npm",
			"args":    []string{"install", "--prefix", webUIPath},
		},
		{
			"type":    "web_ui_build",
			"step":    "Building",
			"command": "npm",
			"args":    []string{"run", "build", "--prefix", webUIPath},
		},
	}

	for _, command := range commands {
		if _, err := s.runCmd(command["type"].(string), command["command"].(string), command["args"].([]string)...); err != nil {
			return command["type"].(string), err
		}
	}

	return "", nil
}

func (s *Server) runCmd(logType, cmd string, args ...string) (string, error) {
	command := exec.Command(cmd, args...)

	stdout, err := command.StdoutPipe()
	if err != nil {
		return "", err
	}

	stderr, err := command.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := command.Start(); err != nil {
		return "", err
	}

	scanOutput := func(pipe *bufio.Scanner, output *string, isErrOutput bool) {
		for pipe.Scan() {
			line := pipe.Text()

			*output += line + "\n"

			if !isErrOutput {
				switch logType {
				case "serve_web_ui_local":
					if strings.Contains(*output, "Ready") {
						s.logCh <- models.ServerLog{
							Message: "http://localhost:3000",
							Type:    logType,
						}
					}

				case "serve_web_ui_remote":
					url := strings.Split(*output, "your url is: ")[1]

					s.logCh <- models.ServerLog{
						Message: url,
						Type:    logType,
					}

				default:
					s.logCh <- models.ServerLog{
						Message: *output,
						Type:    logType,
					}

				}
			} else {
				s.logCh <- models.ServerLog{
					Type:  logType,
					Error: fmt.Errorf(*output),
				}
			}
		}
	}

	var output string
	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	go scanOutput(stdoutScanner, &output, false)
	go scanOutput(stderrScanner, &output, true)

	if err := command.Wait(); err != nil {
		return "", errors.New(output)
	}

	return output, nil
}

func (s *Server) getFileUpload(w http.ResponseWriter, r *http.Request) {
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
			Type:    "api_log",
		}

		http.ServeFile(w, r, file)
		return
	}

	http.Error(w, "Failed to download file", http.StatusBadRequest)
	return
}

func (s *Server) getServerConfig(w http.ResponseWriter, _ *http.Request) {
	configJson, err := json.Marshal(s.Config)
	if err != nil {
		http.Error(w, "Failed to get server", http.StatusInternalServerError)
		return
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
		Type:    "api_log",
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
	_, err := os.Stat(webUIPath)
	if err != nil {
		_, err = s.runCmd("web_ui_download", "git", "clone", "https://github.com/Owbird/SVault-Engine-File-Server-Web.git", webUIPath)
		if err != nil {
			s.logCh <- models.ServerLog{
				Error: err,
				Type:  "web_ui_download",
			}
			return

		}
	} else {
		res, err := s.runCmd("web_ui_version_check", "git", "-C", webUIPath, "log", "--oneline", "-n", "1")
		if err != nil {
			s.logCh <- models.ServerLog{
				Error: err,
				Type:  "web_ui_version_check",
			}

			return
		}

		currentCommit := strings.Split(string(res), " ")[0]

		resp, err := http.Get("https://api.github.com/repos/owbird/svault-engine-file-server-web/commits")
		if err == nil && resp.StatusCode == 200 {

			var commitsRes []map[string]interface{}

			json.NewDecoder(resp.Body).Decode(&commitsRes)

			remoteCommit := commitsRes[0]["sha"].(string)[:7]

			if remoteCommit != currentCommit {
				_, err := s.runCmd("web_ui_version_update", "git", "-C", webUIPath, "pull")
				if err != nil {
					s.logCh <- models.ServerLog{
						Error: err,
						Type:  "web_ui_version_update",
					}

					return
				}

			}
		}
	}

	if cmd, err := s.buildUI(); err != nil {
		s.logCh <- models.ServerLog{
			Error: err,
			Type:  cmd,
		}

		return

	}

	go (func() {
		_, err := s.runCmd("serve_web_ui_local", "npm", "run", "start", "--prefix", webUIPath)
		if err != nil {
			s.logCh <- models.ServerLog{
				Error: err,
				Type:  "serve_web_ui_local",
			}
		}
	})()

	go (func() {
		_, err := s.runCmd("serve_web_ui_remote", "npx", "--yes", "localtunnel", "--port", "3000")
		if err != nil {
			s.logCh <- models.ServerLog{
				Error: err,
				Type:  "serve_web_ui_remote",
			}
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
		Type:    "api_log",
	}

	err = http.ListenAndServe(fmt.Sprintf(":%v", PORT), corsOpts.Handler(mux))
	if err != nil {
		s.logCh <- models.ServerLog{
			Error: err,
			Type:  "api_log",
		}
	}
}
