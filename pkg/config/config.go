package config

import (
	"fmt"
	"log"
	"os"

	"github.com/Owbird/SVault-Engine/internal/config"
	"github.com/Owbird/SVault-Engine/internal/utils"
	"github.com/spf13/viper"
)

// AppConfig holds the server configuration
type AppConfig struct {
	// The server configuration
	server *config.ServerConfig
}

// Gets the app configuration from
// svault.toml with default values
// if absent
func NewAppConfig() *AppConfig {
	userDir, err := utils.GetSVaultDir()
	if err != nil {
		log.Fatalln("Failed to get user dir")
	}

	viper.SetConfigName("svault")
	viper.SetConfigType("toml")

	viper.AddConfigPath(userDir)

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln(err)
	}

	viper.SetDefault("server.name", fmt.Sprintf("%v's Server", hostname))
	viper.SetDefault("server.allowUploads", false)

	err = viper.ReadInConfig()
	if err != nil {
		viper.SafeWriteConfig()
	}

	config := &AppConfig{
		server: config.NewServerConfig(),
	}

	config.server.SetName(viper.GetString("server.name"))
	config.server.SetAllowUploads(viper.GetBool("server.allowUploads"))

	return config
}

// GetSeverConfig returns the server configuration
func (ac *AppConfig) GetSeverConfig() *config.ServerConfig {
	return ac.server
}

// Save saves the server configuration to svault.toml
func (ac *AppConfig) Save() error {
	viper.Set("server.name", ac.server.GetName())
	viper.Set("server.allowUploads", ac.server.GetAllowUploads())

	return viper.WriteConfig()
}

// ToJson returns the server configuration as json
func (ac *AppConfig) ToJson() map[string]interface{} {
	return viper.AllSettings()
}
