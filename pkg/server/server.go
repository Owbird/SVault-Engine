// Package server handles the file hosting server
package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
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

	err := filepath.WalkDir(s.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		files = append(files, File{Name: path, IsDir: d.IsDir()})
		return nil
	})
	if err != nil {
		http.Error(w, "Failed to list files", http.StatusInternalServerError)
		return
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
