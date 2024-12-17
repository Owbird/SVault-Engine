// Package filesystem provides functions for the SVault VFS
// such as mounting based on [fuse]
package filesystem

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Owbird/SVault-Engine/internal/crypto"
	"github.com/Owbird/SVault-Engine/internal/database"
	"github.com/winfsp/cgofuse/fuse"
)

type SVFileSystem struct {
	fuse.FileSystemBase
	db *database.Database

	vault    string
	password string
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

func (svfs *SVFileSystem) Read(path string, buff []byte, ofst int64, fh uint64) int {
	log.Printf("Read called on %v, offset: %v, buffer size: %v", path, ofst, len(buff))

	file := filepath.Base(path)

	akshualFile, err := svfs.db.GetVaultFile(svfs.vault, file)
	if err != nil {
		return -fuse.ENOENT
	}

	vaultKey, err := svfs.db.GetVaultKey(svfs.vault, svfs.password)
	if err != nil {
		return -fuse.ENOENT
	}

	crypFunc := crypto.NewCrypto()
	decryptedData, err := crypFunc.Decrypt(akshualFile.Data, vaultKey)
	if err != nil {
		return -fuse.ENOENT
	}

	if ofst >= int64(len(decryptedData)) {
		return 0 // EOF
	}

	newBuff := copy(buff, decryptedData[ofst:])
	return newBuff
}

func (svfs *SVFileSystem) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64,
) int {
	log.Printf("Readdir called on %v, is dir: %v ", path, isDir(path))

	fill(".", nil, 0)
	fill("..", nil, 0)

	files, err := svfs.db.ListVaultFiles(svfs.vault)
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

	return 0
}

func (svfs *SVFileSystem) Getattr(path string, stat *fuse.Stat_t, fh uint64) int {
	log.Printf("Getattr called on %v, is dir: %v ", path, isDir(path))

	if isDir(path) && path == "/" {
		stat.Mode = fuse.S_IFDIR
		return 0
	}

	stat.Mode = fuse.S_IFREG

	file := filepath.Base(path)

	akshualFile, err := svfs.db.GetVaultFile(svfs.vault, file)
	if err != nil {
		return -fuse.ENOENT
	}

	stat.Size = akshualFile.Size
	stat.Mtim = fuse.NewTimespec(akshualFile.ModTime)

	return 0
}

// Mount mounts the vfs with in a temp directory
func Mount(vault, password string) {
	vaultDir := filepath.Join(os.TempDir(), fmt.Sprintf("svault-%s-%s", vault, time.Now().Format("20060102150405")))

	if runtime.GOOS != "windows" {
		err := os.MkdirAll(vaultDir, 0755)
		if err != nil {
			log.Fatalln(err)
		}
	}

	log.Println("Mounting bank dir", vaultDir)

	fs := &SVFileSystem{
		db:       database.NewDatabase(),
		vault:    vault,
		password: password,
	}

	host := fuse.NewFileSystemHost(fs)

	defer host.Unmount()

	var command string
	switch runtime.GOOS {
	case "darwin":
		command = "open"
	case "windows":
		command = "explorer"
	case "linux":
		command = "xdg-open"
	default:
		return
	}

	cmd := exec.Command(command, vaultDir)
	cmd.Run()

	host.Mount(vaultDir, []string{})
}
