package main

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Logger LoggerConf `yaml:"logger"`
	// Server                    ServerConf `yaml:"server"`
	Amqp AmqpConf `yaml:"amqp"`
	// ScanningIntervalInSeconds int64    `yaml:"scanningIntervalInSeconds"`
	// CleaningIntervalInHours   int64    `yaml:"cleaningIntervalInHours"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

// type ServerConf struct {
// 	Host string `yaml:"host"`
// 	Port string `yaml:"port"`
// }

type AmqpConf struct {
	Url       string `yaml:"url"`
	QueueName string `yaml:"queueName"`
}

//nolint:gosec
func NewConfig() Config {
	return Config{
		Logger: LoggerConf{Level: "info"},
		// Server:                    ServerConf{Host: "127.0.0.1", Port: "8080"},
		Amqp: AmqpConf{Url: "amqp://guest:guest@localhost:5672/", QueueName: "Notifications"},
		// ScanningIntervalInSeconds: 10,
		// CleaningIntervalInHours:   24,
	}
}

func LoadConfig(path string) (Config, error) {
	cfg := NewConfig()
	file, err := os.Open(path) //nolint:gosec
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
