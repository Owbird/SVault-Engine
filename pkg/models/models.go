package models

import "time"

type Vault struct {
	// Name is the name of the vault
	Name string

	// Password is the password to the vault
	Password string

	// Time the vault was created. Automatically added when creating a vault
	CreatedAt time.Time
}

type File struct {
	// Vault is the parent name of vault the file belongs to
	Vault string

	// Name is the name of the file
	Name string

	// Data is the byte content of the file
	Data []byte

	// Size is the size of the file
	Size int64

	// Mode is the file mode
	Mode uint32

	// ModTime is the modification time of the file
	ModTime time.Time
}

type ServerLog struct {
	// Type of log from the file server.
	// [api_log]: Log for the API
	// [serve_web_ui_local]: Contains local url
	// [serve_web_ui_remote]: Contains remote link
	// [web_deps_installation]: Web UI dependency installation logs
	// [web_ui_build]: Building the web ui logs
	// [web_ui_download]: Fresh download of the web ui
	// [web_ui_version_check]: Check current version of web UI
	// [web_ui_version_update]: Update web UI version
	Type string

	Message string

	Error error
}
