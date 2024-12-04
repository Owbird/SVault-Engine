package models

import "time"

const (
	// File Server Log Types
	API_LOG               = "api_log"
	WEB_UI_DOWNLOAD       = "web_ui_download"
	WEB_UI_VERSION_CHECK  = "web_ui_version_check"
	WEB_UI_VERSION_UPDATE = "web_ui_version_update"
	WEB_DEPS_INSTALLATION = "web_deps_installation"
	WEB_UI_BUILD          = "web_ui_build"
	SERVE_WEB_UI_LOCAL    = "serve_web_ui_local"
	SERVE_WEB_UI_NETWORK    = "serve_web_ui_network"
	SERVE_WEB_UI_REMOTE   = "serve_web_ui_remote"
)

type Notification struct {
	// The title of the notification
	Title string

	// The message of the notification
	Body string

	// The text to be copied to the clipboard
	ClipboardText string
}

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

type FileShareProgress struct {
	Bytes      int64
	Total      int64
	Percentage int
}
