package config

type ServerConfig struct {
	// The server label to be displayed
	name string

	// Should uploads be allowed
	allowUploads bool
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{}
}

// SetName sets the server name
// Defaults to machine hostname
func (sc *ServerConfig) SetName(name string) *ServerConfig {
	sc.name = name
	return sc
}

// SetAllowUploads sets if uploads are allowed
// Defaults to false
func (sc *ServerConfig) SetAllowUploads(allowUploads bool) *ServerConfig {
	sc.allowUploads = allowUploads
	return sc
}

// GetName returns the server name
func (sc *ServerConfig) GetName() string {
	return sc.name
}

// GetAllowUploads returns if uploads are allowed
func (sc *ServerConfig) GetAllowUploads() bool {
	return sc.allowUploads
}
