// Package server handles the file hosting server
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Owbird/SVault-Engine/internal/utils"
)

type Server struct {
	// The current directory being hosted
	Dir string
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

func NewServer(dir string) *Server {
	return &Server{
		Dir: dir,
	}
}

func (s *Server) getFilesHandler(w http.ResponseWriter, r *http.Request) {
	files := []File{}

	dirFiles, err := os.ReadDir(s.Dir)
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
	userDir, err := utils.GetSVaultDir()
	if err != nil {
		log.Fatalln("Failed to get user dir")
	}

	webUIPath := filepath.Join(userDir, "web_ui")

	_, err = os.Stat(webUIPath)
	if err != nil {
		commands := []map[string]interface{}{
			{
				"step":    "Cloning web UI. This will only happen once",
				"command": "git",
				"args":    []string{"clone", "https://github.com/Owbird/SVault-Engine-File-Server-Web.git", webUIPath},
			},
			{
				"step":    "Installing dependencies",
				"command": "npm",
				"args":    []string{"install", "--prefix", webUIPath},
			},
			{
				"step":    "Building",
				"command": "npm",
				"args":    []string{"run", "build", "--prefix", webUIPath},
			},
		}

		for _, command := range commands {
			log.Printf("[+] %s\n", command["step"])

			_, err = exec.Command(command["command"].(string), command["args"].([]string)...).Output()
			if err != nil {
				log.Fatalln(err)
			}
		}

	}

	go (func() {
		log.Println("[+] Running web UI")
		_, err = exec.Command("npm", "run", "start", "--prefix", webUIPath).Output()
		if err != nil {
			log.Fatalln(err)
		}
	})()

	mux := http.NewServeMux()

	mux.HandleFunc("/", s.getFilesHandler)

	log.Printf("[+] Starting API on port %v from %v", PORT, s.Dir)

	err = http.ListenAndServe(fmt.Sprintf(":%v", PORT), mux)
	if err != nil {
		log.Fatalf("Couldn't start server: %v", err)
	}
}
