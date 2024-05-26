// Package filesystem provides functions for the SVault VFS
// such as mounting based on [fuse]
package filesystem

import (
	"log"
	"os"
	"path/filepath"

	"github.com/winfsp/cgofuse/fuse"
)

type SVFileSystem struct {
	fuse.FileSystemBase
}

// Mount mounts the vfs with .svault as the mount
// point at the user's home directory
func Mount() {
	userDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Could not get user directory")
	}

	bankDir := filepath.Join(userDir, ".svault")

	log.Println("Mounting bank dir", bankDir)

	fs := &SVFileSystem{}

	host := fuse.NewFileSystemHost(fs)

	defer host.Unmount()

	host.Mount(bankDir, []string{})
}
