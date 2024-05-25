package models

import "time"

type Vault struct {
	Name     string
	Password string
}

type File struct {
	Vault   string
	Name    string
	Data    []byte
	Size    int64
	Mode    uint32
	ModTime time.Time
}
