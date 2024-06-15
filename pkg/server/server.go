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
	"github.com/rs/cors"
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

func runCmd(verbose bool, cmd string, args ...string) string {
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

	scanAndPrint := func(pipe *bufio.Scanner, output *string) {
		for pipe.Scan() {
			line := pipe.Text()

			if verbose {
				fmt.Println(line)
			}

			*output += line + "\n"
		}
	}

	var output string
	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	go scanAndPrint(stdoutScanner, &output)
	go scanAndPrint(stderrScanner, &output)

	if err := command.Wait(); err != nil {
		log.Fatalf("Command finished with error: %v", err)
	}

	return output
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

		runCmd(true, command["command"].(string), command["args"].([]string)...)
	}
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

	log.Println("[+] Getting files for", dir)

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
		log.Printf("[+] Cloning web UI. This will only happen once")

		runCmd(true, "git", "clone", "https://github.com/Owbird/SVault-Engine-File-Server-Web.git", webUIPath)

		buildUI()
	}

	res := runCmd(false, "git", "-C", webUIPath, "show", "--summary")

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

		runCmd(true, "git", "-C", webUIPath, "pull")

		buildUI()
	}

	go (func() {
		log.Println("[+] Running web UI")
		runCmd(true, "npm", "run", "start", "--prefix", webUIPath)
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

	log.Printf("[+] Starting API on port %v from %v", PORT, s.Dir)

	err = http.ListenAndServe(fmt.Sprintf(":%v", PORT), corsOpts.Handler(mux))
	if err != nil {
		log.Fatalf("Couldn't start server: %v", err)
	}
}
