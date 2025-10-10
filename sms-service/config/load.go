package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

var C Config

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path) // path to look for the config file
	viper.AddConfigPath(".")  // optionally look in current directory

	// Allow environment variable overrides
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("fatal error reading config file: %w", err)
	}

	if err := viper.Unmarshal(&C); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	log.Println("✅ Config loaded successfully from:", viper.ConfigFileUsed())
	return &C, nil
}
