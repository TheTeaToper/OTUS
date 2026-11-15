package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`

	Limits struct {
		LoginLimit    int           `yaml:"login_limit"`    // N
		PasswordLimit int           `yaml:"password_limit"` // M
		IpLimit       int           `yaml:"ip_limit"`       // K
		BucketTTL     time.Duration `yaml:"bucket_ttl"`     // Inactive buckets TTL
	} `yaml:"limits"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{}

	// Default values
	cfg.Server.Port = ":8080"
	cfg.Limits.LoginLimit = 10
	cfg.Limits.PasswordLimit = 100
	cfg.Limits.IpLimit = 1000
	cfg.Limits.BucketTTL = 10 * time.Minute

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config YAML: %w", err)
	}

	return cfg, nil
}
