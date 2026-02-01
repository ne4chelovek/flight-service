package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func LoadConfig() (*Config, error) {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local" // значение по умолчанию
	}

	viper.SetConfigName(env)
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

	log.Printf("Loaded config: %s", env)
	return &cfg, nil
}
