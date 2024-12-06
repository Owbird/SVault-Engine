// Package handlers provides the functionalities for
// file web server hosting
package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Owbird/SVault-Engine/internal/config"
	"github.com/Owbird/SVault-Engine/internal/utils"
	"github.com/Owbird/SVault-Engine/pkg/models"
)

type Handlers struct {
	logCh        chan models.ServerLog
	dir          string
	serverConfig *config.ServerConfig
	notifConfig  *config.NotifConfig
}

type File struct {
	// The name of the file
	Name string `json:"name"`

	// Whether it's a file or directory
	IsDir bool `json:"is_dir"`

	// Size of the file in bytes
	Size string `json:"size"`
}

type IndexHTMLConfig struct {
	Name         string
	AllowUploads bool
}

// IndexHTML defines the data passed to the index.html
// template file
type IndexHTML struct {
	Files        []File
	CurrentPath  string
	ServerConfig IndexHTMLConfig
}

var tmpl *template.Template

func getCwd() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalln("Failed to get templates dir")
	}

	cwd := filepath.Dir(filename)

	return cwd
}

func NewHandlers(
	logCh chan models.ServerLog,
	dir string,
	serverConfig *config.ServerConfig,
	notifConfig *config.NotifConfig,
) *Handlers {
	cwd := getCwd()

	tpl, err := template.ParseGlob(filepath.Join(cwd, "templates/*.html"))
	if err != nil {
		log.Fatal(err)
	}

	tmpl = tpl

	return &Handlers{
		logCh:        logCh,
		dir:          dir,
		serverConfig: serverConfig,
		notifConfig:  notifConfig,
	}
}

func (h *Handlers) GetFileUpload(w http.ResponseWriter, r *http.Request) {
	h.logCh <- models.ServerLog{
		Message: "Receiving files",
		Type:    models.API_LOG,
	}

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

		uploadDir := filepath.Join(h.dir, uploadDir, fileHeader.Filename)

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

		h.logCh <- models.ServerLog{
			Message: fmt.Sprintf("File received at %v", uploadDir),
			Type:    models.API_LOG,
		}

		h.notifConfig.SendNotification(models.Notification{
			Title: "File received",
			Body:  fmt.Sprintf("File %v received", fileHeader.Filename),
		})

	}

	return
}

func (h *Handlers) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	if len(query["file"]) > 0 {
		if filepath.Dir(query["file"][0]) == ".." || filepath.Base(query["file"][0]) == ".." {
			http.Error(w, "Failed to download file", http.StatusInternalServerError)
			return

		}

		file := filepath.Join(h.dir, query["file"][0])

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%v", filepath.Base(file)))
		w.Header().Set("Content-Type", "application/octet-stream")

		h.logCh <- models.ServerLog{
			Message: fmt.Sprintf("Downloading %v", file),
			Type:    models.API_LOG,
		}

		http.ServeFile(w, r, file)
		return
	}

	http.Error(w, "Failed to download file", http.StatusBadRequest)
	return
}

func (h *Handlers) GetFilesHandler(w http.ResponseWriter, r *http.Request) {
	files := []File{}

	query := r.URL.Query()

	var fullPath string
	var currentPath string

	if len(query["dir"]) > 0 {
		currentPath = query["dir"][0]

		if filepath.Base(currentPath) == ".." {
			http.Error(w, "Failed to list files", http.StatusInternalServerError)
			return

		}

		fullPath = filepath.Join(h.dir, currentPath)

	} else {
		currentPath = "/"
		fullPath = h.dir
	}

	h.logCh <- models.ServerLog{
		Message: fmt.Sprintf("Getting files for %v", fullPath),
		Type:    models.API_LOG,
	}

	dirFiles, err := os.ReadDir(fullPath)
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

		fmtedFile := File{
			Name:  file.Name(),
			IsDir: file.IsDir(),
		}

		if !fmtedFile.IsDir {
			fmtedFile.Size = utils.FmtBytes(info.Size())
		}

		files = append(files, fmtedFile)
	}

	tmpl.ExecuteTemplate(w, "index.html", IndexHTML{
		Files:       files,
		CurrentPath: currentPath,
		ServerConfig: IndexHTMLConfig{
			Name:         h.serverConfig.GetName(),
			AllowUploads: h.serverConfig.GetAllowUploads(),
		},
	})
}

func (h *Handlers) GetAssets(w http.ResponseWriter, r *http.Request) {
	cwd := getCwd()

	path := r.URL.Path
	data, err := os.ReadFile(filepath.Join(cwd, "templates", path))
	if err != nil {
		fmt.Print(err)
	}
	if strings.HasSuffix(path, "js") {
		w.Header().Set("Content-Type", "text/javascript")
	}
	_, err = w.Write(data)
	if err != nil {
		fmt.Print(err)
	}
}
