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
	"strings"

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

var webUIPath string

func NewServer(dir string) *Server {
	userDir, err := utils.GetSVaultDir()
	if err != nil {
		log.Fatalln("Failed to get user dir")
	}

	webUIPath = filepath.Join(userDir, "web_ui")

	return &Server{
		Dir: dir,
	}
}

func runCmd(cmd string, args ...string) string {
	res, err := exec.Command(cmd, args...).Output()
	if err != nil {
		log.Fatalln(err)
	}

	return string(res)
}

func buildUI() {
	commands := []map[string]interface{}{
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

		runCmd(command["command"].(string), command["args"].([]string)...)
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
	_, err := os.Stat(webUIPath)
	if err != nil {
		log.Printf("[+] Cloning web UI. This will only happen once")

		runCmd("git", "https://github.com/Owbird/SVault-Engine-File-Server-Web.git", webUIPath)

		buildUI()
	}

	res := runCmd("git", "-C", webUIPath, "show", "--summary")

	firstLine := strings.Split(string(res), "\n")[0]

	currentCommit := strings.Split(firstLine, " ")[1]

	resp, err := http.Get("https://api.github.com/repos/owbird/svault-engine-file-server-web/commits")
	if err != nil {
		log.Fatalln(err)
	}

	var commitsRes []map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&commitsRes)

	remoteCommit := commitsRes[0]["sha"]

	if remoteCommit != currentCommit {
		log.Println("[!] UI update available. Fetching updates")

		runCmd("git", "-C", webUIPath, "pull")

		buildUI()
	}

	go (func() {
		log.Println("[+] Running web UI")
		runCmd("npm", "run", "start", "--prefix", webUIPath)
	})()

	mux := http.NewServeMux()

	mux.HandleFunc("/", s.getFilesHandler)

	log.Printf("[+] Starting API on port %v from %v", PORT, s.Dir)

	err = http.ListenAndServe(fmt.Sprintf(":%v", PORT), mux)
	if err != nil {
		log.Fatalf("Couldn't start server: %v", err)
	}
}
