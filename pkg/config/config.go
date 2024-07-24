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
	Server *config.ServerConfig `json:"server"`
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
		Server: config.NewServerConfig(),
	}

	viper.Unmarshal(&config)

	return config
}

// GetSeverConfig returns the server configuration
func (ac *AppConfig) GetSeverConfig() *config.ServerConfig {
	return ac.Server
}

// Save saves the server configuration to svault.toml
func (ac *AppConfig) Save() error {
	return viper.WriteConfig()
}
