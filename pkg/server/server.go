// Package server handles the file hosting server
package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Owbird/SVault-Engine/internal/utils"
	"github.com/Owbird/SVault-Engine/pkg/models"
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

	return &Server{
		Dir:   dir,
		logCh: logCh,
	}
}

func (s *Server) buildUI() {
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
		s.runCmd(command["type"].(string), command["command"].(string), command["args"].([]string)...)
	}
}

func (s *Server) runCmd(logType, cmd string, args ...string) string {
	command := exec.Command(cmd, args...)

	stdout, err := command.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}

	stderr, err := command.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to get stderr pipe: %v", err)
	}

	if err := command.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
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
		log.Fatalf("Command: %v finished with error: %v", command.String(), err)
	}

	return output
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

		http.ServeFile(w, r, file)
		return
	}

	http.Error(w, "Failed to download file", http.StatusBadRequest)
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
		s.runCmd("web_ui_download", "git", "clone", "https://github.com/Owbird/SVault-Engine-File-Server-Web.git", webUIPath)

		s.buildUI()
	}

	res := s.runCmd("web_ui_version_check", "git", "-C", webUIPath, "log", "--oneline", "-n", "1")

	currentCommit := strings.Split(string(res), " ")[0]

	resp, err := http.Get("https://api.github.com/repos/owbird/svault-engine-file-server-web/commits")
	if err != nil {
		log.Fatalln(err)
	}

	var commitsRes []map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&commitsRes)

	remoteCommit := commitsRes[0]["sha"].(string)[:7]

	if remoteCommit != currentCommit {
		s.runCmd("web_ui_version_update", "git", "-C", webUIPath, "pull")

		s.buildUI()
	}

	go (func() {
		s.runCmd("serve_web_ui_local", "npm", "run", "start", "--prefix", webUIPath)
	})()

	go (func() {
		s.runCmd("serve_web_ui_remote", "npx", "--yes", "localtunnel", "--port", "3000")
	})()

	mux := http.NewServeMux()

	mux.HandleFunc("/", s.getFilesHandler)
	mux.HandleFunc("/download", s.downloadFileHandler)

	corsOpts := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
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
		log.Fatalf("Couldn't start server: %v", err)
	}
}
