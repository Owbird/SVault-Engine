// Package filesystem provides functions for the SVault VFS
// such as mounting based on [fuse]
package filesystem

import (
	"log"

	"github.com/Owbird/SVault-Engine/internal/utils"
	"github.com/winfsp/cgofuse/fuse"
)

type SVFileSystem struct {
	fuse.FileSystemBase
}

// Mount mounts the vfs with .svault as the mount
// point at the user's home directory
func Mount() {
	bankDir, err := utils.GetSVaultDir()
	if err != nil {
		log.Fatal("Could not get user directory")
	}

	log.Println("Mounting bank dir", bankDir)

	fs := &SVFileSystem{}

	host := fuse.NewFileSystemHost(fs)

	defer host.Unmount()

	host.Mount(bankDir, []string{})
}
