// Package server handles the file hosting server
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.getFilesHandler)

	log.Printf("Starting server on port %v from %v", PORT, s.Dir)

	err := http.ListenAndServe(fmt.Sprintf(":%v", PORT), mux)
	if err != nil {
		log.Fatalf("Couldn't start server: %v", err)
	}
}
