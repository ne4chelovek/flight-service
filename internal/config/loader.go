package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"path/filepath"
	"strings"
)

func LoadConfig() (*Config, error) {
	viper.SetConfigName("local")
	viper.SetConfigType("yml")

	absPath, err := filepath.Abs("configs")
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for configs: %w", err)
	}
	viper.AddConfigPath(absPath)

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AllowEmptyEnv(true)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	log.Println("Loaded config")
	return &cfg, nil
}
