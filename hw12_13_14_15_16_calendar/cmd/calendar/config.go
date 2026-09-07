package main

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Logger  LoggerConf  `yaml:"logger"`
	Storage StorageConf `yaml:"storage"`
	Server  ServerConf  `yaml:"server"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

type StorageConf struct {
	Type string `yaml:"type"` // "in-memory"|"sql"
	DSN  string `yaml:"dsn"`
}

type ServerConf struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func NewConfig() Config {
	return Config{
		Logger:  LoggerConf{Level: "info"},
		Storage: StorageConf{Type: "in-memory"},
		Server:  ServerConf{Host: "127.0.0.1", Port: "8080"},
	}
}

func LoadConfig(path string) (Config, error) {
	cfg := NewConfig()
	file, err := os.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("loading cfg file error: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	err = yaml.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("decoding yaml cfg file error: %w", err)
	}
	return cfg, nil
}
