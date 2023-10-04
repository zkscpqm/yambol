package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Todo: Hot reloading?

type Configuration struct {
	API struct {
		HTTP struct {
			Port int `json:"port"`
		} `json:"http,omitempty"`
		GRPC struct {
			Port int `json:"port"`
		} `json:"grpc,omitempty"`
		Certificate string `json:"certificate,omitempty"`
		Key         string `json:"key,omitempty"`
	} `json:"api,omitempty"`
	Broker struct {
		DefaultMinLength    int `json:"default_min_length"`
		DefaultMaxLength    int `json:"default_max_length"`
		DefaultMaxSizeBytes int `json:"default_max_size_bytes"`
		DefaultTTL          int `json:"default_ttl"`
		Queues              map[string]struct {
			MinLength    int `json:"min_length"`
			MaxLength    int `json:"max_length"`
			MaxSizeBytes int `json:"max_size_bytes"`
			TTL          int `json:"ttl"`
		} `json:"queues"`
	} `json:"broker"`
}

func FromFile(path string) (*Configuration, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file `%s`: %v", path, err)
	}
	defer file.Close()

	var config Configuration
	if err = json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file `%s`: %v", path, err)
	}

	return &config, nil
}
