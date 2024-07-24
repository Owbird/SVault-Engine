package config

type ServerConfig struct {
	// The server label to be displayed
	Name string `json:"name"`

	// Should uploads be allowed
	AllowUploads bool `json:"allow_uploads"`
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{}
}

// SetName sets the server name
// Defaults to machine hostname
func (sc *ServerConfig) SetName(name string) *ServerConfig {
	sc.Name = name
	return sc
}

// SetAllowUploads sets if uploads are allowed
// Defaults to false
func (sc *ServerConfig) SetAllowUploads(allowUploads bool) *ServerConfig {
	sc.AllowUploads = allowUploads
	return sc
}
