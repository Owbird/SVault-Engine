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
