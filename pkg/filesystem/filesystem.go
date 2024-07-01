// Package filesystem provides functions for the SVault VFS
// such as mounting based on [fuse]
package filesystem

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Owbird/SVault-Engine/internal/database"
	"github.com/winfsp/cgofuse/fuse"
)

type SVFileSystem struct {
	fuse.FileSystemBase
	db *database.Database
}

func isDir(path string) bool {
	var paths []string

	if runtime.GOOS == "windows" {
		paths = strings.Split(path, "\\")
	} else {
		paths = strings.Split(path, "/")
	}

	return len(paths[1:]) == 1
}

func (svfs *SVFileSystem) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64,
) int {
	log.Printf("Readdir called on %v, is dir: %v ", path, isDir(path))

	fill(".", nil, 0)
	fill("..", nil, 0)

	if path == "/" {
		vaults, err := svfs.db.ListVaults()
		if err != nil {
			return -fuse.ENOENT
		}

		for _, vault := range vaults {
			fill(vault.Name, &fuse.Stat_t{
				Mtim: fuse.NewTimespec(vault.CreatedAt),
			}, 0)
		}
	}

	if isDir(path) && path != "/" {

		vault := filepath.Base(path)

		files, err := svfs.db.ListVaultFiles(vault)
		if err != nil {
			return -fuse.ENOENT
		}

		for _, file := range files {
			fill(filepath.Base(file.Name), &fuse.Stat_t{
				Size: file.Size,
				Mode: file.Mode,
				Mtim: fuse.NewTimespec(file.ModTime),
			}, 0)
		}

	}

	return 0
}

func (svfs *SVFileSystem) Getattr(path string, stat *fuse.Stat_t, fh uint64) int {
	log.Printf("Getattr called on %v, is dir: %v ", path, isDir(path))

	if isDir(path) {
		stat.Mode = fuse.S_IFDIR
	} else {
		stat.Mode = fuse.S_IFREG
	}

	return 0
}

// Mount mounts the vfs with in a temp directory
func Mount() {
	bankDir, err := os.MkdirTemp(os.TempDir(), "svault-")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting bank dir", bankDir)

	exec.Command("xdg-open", bankDir).Run()

	fs := &SVFileSystem{
		db: database.NewDatabase(),
	}

	host := fuse.NewFileSystemHost(fs)

	defer host.Unmount()

	host.Mount(bankDir, []string{})
}
