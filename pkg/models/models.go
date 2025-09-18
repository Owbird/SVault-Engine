package models

const (
	// File Server Log Types
	API_LOG               = "api_log"
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

type ServerLog struct {
	// Type of log from the file server.
	// [api_log]: Log for the API
	// [serve_web_ui_local]: Contains local url
	// [serve_web_ui_remote]: Contains remote link
	Type string

	Message string

	Error error
}

type FileShareProgress struct {
	Bytes      int64
	Total      int64
	Percentage int
}
