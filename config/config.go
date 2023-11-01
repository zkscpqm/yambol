package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

var configFilePath, _ = filepath.Abs("config.json")

func init() {
	val, ok := syscall.Getenv("YAMBOL_CONFIG")
	if ok {
		configFilePath = val
	}
}

type QueueMap map[string]QueueState

func (qm QueueMap) Copy() QueueMap {
	rv := make(QueueMap)
	for k, v := range qm {
		rv[k] = v
	}
	return rv
}

type QueueState struct {
	MinLength    int64 `json:"min_length"`
	MaxLength    int64 `json:"max_length"`
	MaxSizeBytes int64 `json:"max_size_bytes"`
	TTL          int64 `json:"ttl"`
}

type BrokerState struct {
	DefaultMinLength    int64    `json:"default_min_length"`
	DefaultMaxLength    int64    `json:"default_max_length"`
	DefaultMaxSizeBytes int64    `json:"default_max_size_bytes"`
	DefaultTTL          int64    `json:"default_ttl"`
	Queues              QueueMap `json:"queues"`
}

func (s BrokerState) Copy() (rv BrokerState) {
	q := s.Queues
	if q == nil {
		q = make(QueueMap)
	}
	return BrokerState{
		DefaultMinLength:    s.DefaultMinLength,
		DefaultMaxLength:    s.DefaultMaxLength,
		DefaultMaxSizeBytes: s.DefaultMaxSizeBytes,
		DefaultTTL:          s.DefaultTTL,
		Queues:              q.Copy(),
	}
}

type Server struct {
	Enabled    bool `json:"enabled,omitempty"`
	Port       int  `json:"port,omitempty"`
	TlsEnabled bool `json:"tls_enabled,omitempty"`
}

type ApiConfig struct {
	REST        Server `json:"rest,omitempty"`
	GRPC        Server `json:"grpc,omitempty"`
	HTTP        Server `json:"http,omitempty"`
	Certificate string `json:"certificate,omitempty"`
	Key         string `json:"key,omitempty"`
}

type LogConfig struct {
	Level string `json:"level,omitempty"`
	File  string `json:"file,omitempty"`
}

type Configuration struct {
	DisableAutoSave bool        `json:"disable_auto_save,omitempty"`
	API             ApiConfig   `json:"api,omitempty"`
	Broker          BrokerState `json:"broker,omitempty"`
	Log             LogConfig   `json:"log,omitempty"`
}

func Empty() Configuration {
	return Configuration{
		Broker: BrokerState{
			Queues: make(QueueMap),
		},
	}
}

func FromFile() (*Configuration, error) {
	logger.Debug("Loading config from file: %s", configFilePath)
	file, err := os.Open(configFilePath)
	if err != nil {
		logger.Debug("Failed to open config file `%s`: %v", configFilePath, err)
		return nil, fmt.Errorf("failed to open config file `%s`: %v", configFilePath, err)
	}
	defer file.Close()
	logger.Debug("Loaded config from file: %s", configFilePath)

	var config Configuration
	if err = json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file `%s`: %v", configFilePath, err)
	}

	return &config, nil
}

func CopyRunningConfigToStartupConfig() error {
	f, err := os.OpenFile(configFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open state log file %s: %v", configFilePath, err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(_config)
}

func (c *Configuration) Copy() Configuration {
	return Configuration{
		DisableAutoSave: c.DisableAutoSave,
		API:             c.API,
		Broker:          c.Broker.Copy(),
		Log:             c.Log,
	}
}

func (c *Configuration) String() string {
	b, _ := json.MarshalIndent(c, "", "    ")
	return string(b)
}
