package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"yambol/pkg/util"
)

var (
	configFilePath, _ = filepath.Abs("config.json")
)

type QueueMap map[string]QueueConfig

func (qm QueueMap) toQueueState() queueStateMap {
	rv := make(queueStateMap)
	for k, v := range qm {
		rv[k] = queueState{
			minLength:    v.MinLength,
			maxLength:    v.MaxLength,
			maxSizeBytes: v.MaxSizeBytes,
			ttl:          v.TTLDuration(),
		}
	}
	return rv
}

func (qm QueueMap) Copy() QueueMap {
	rv := make(QueueMap)
	for k, v := range qm {
		rv[k] = v
	}
	return rv
}

func init() {
	val, ok := syscall.Getenv("YAMBOL_CONFIG")
	if ok {
		configFilePath = val
	}
}

type QueueConfig struct {
	MinLength    int64 `json:"min_length"`
	MaxLength    int64 `json:"max_length"`
	MaxSizeBytes int64 `json:"max_size_bytes"`
	TTL          int64 `json:"ttl"`
}

func (qc QueueConfig) TTLDuration() time.Duration {
	return time.Duration(qc.TTL) * time.Second
}

func (qc QueueConfig) String() string {
	b, _ := json.MarshalIndent(qc, "", "    ")
	return string(b)
}

func (qc QueueConfig) state() queueState {
	return queueState{
		minLength:    qc.MinLength,
		maxLength:    qc.MaxLength,
		maxSizeBytes: qc.MaxSizeBytes,
		ttl:          qc.TTLDuration(),
	}
}

type BrokerConfig struct {
	DefaultMinLength    int64    `json:"default_min_length"`
	DefaultMaxLength    int64    `json:"default_max_length"`
	DefaultMaxSizeBytes int64    `json:"default_max_size_bytes"`
	DefaultTTLSeconds   int64    `json:"default_ttl"`
	Queues              QueueMap `json:"queues"`
}

func (s BrokerConfig) Copy() (rv BrokerConfig) {
	q := s.Queues
	if q == nil {
		q = make(QueueMap)
	}
	return BrokerConfig{
		DefaultMinLength:    s.DefaultMinLength,
		DefaultMaxLength:    s.DefaultMaxLength,
		DefaultMaxSizeBytes: s.DefaultMaxSizeBytes,
		DefaultTTLSeconds:   s.DefaultTTLSeconds,
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
	DisableAutoSave bool         `json:"disable_auto_save,omitempty"`
	API             ApiConfig    `json:"api,omitempty"`
	Broker          BrokerConfig `json:"broker,omitempty"`
	Log             LogConfig    `json:"log,omitempty"`
}

func Empty() Configuration {
	return Configuration{
		Broker: BrokerConfig{
			Queues: make(QueueMap),
		},
	}
}

func (c Configuration) state() state {
	return state{
		DisableAutoSave: c.DisableAutoSave,
		API: apiState{
			REST: serverState{
				Enabled:    c.API.REST.Enabled,
				Port:       c.API.REST.Port,
				TlsEnabled: c.API.REST.TlsEnabled,
			},
			GRPC: serverState{
				Enabled:    c.API.GRPC.Enabled,
				Port:       c.API.GRPC.Port,
				TlsEnabled: c.API.GRPC.TlsEnabled,
			},
			HTTP: serverState{
				Enabled:    c.API.HTTP.Enabled,
				Port:       c.API.HTTP.Port,
				TlsEnabled: c.API.HTTP.TlsEnabled,
			},
			Certificate: c.API.Certificate,
			Key:         c.API.Key,
		},
		Broker: brokerState{
			DefaultMinLength:    c.Broker.DefaultMinLength,
			DefaultMaxLength:    c.Broker.DefaultMaxLength,
			DefaultMaxSizeBytes: c.Broker.DefaultMaxSizeBytes,
			DefaultTTL:          util.Seconds(c.Broker.DefaultTTLSeconds),
			Queues:              c.Broker.Queues.toQueueState(),
		},
		Log: logState{
			Level: c.Log.Level,
			File:  c.Log.File,
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

func (c Configuration) Copy() Configuration {
	return Configuration{
		DisableAutoSave: c.DisableAutoSave,
		API:             c.API,
		Broker:          c.Broker.Copy(),
		Log:             c.Log,
	}
}

func (c Configuration) String() string {
	b, _ := json.MarshalIndent(c, "", "    ")
	return string(b)
}

func GetStartupConfig() (Configuration, error) {
	logger.Debug("Getting startup config")
	cfg, err := FromFile()
	if err != nil {
		logger.Debug("Failed to get startup config from file: %v", err)
		return Empty(), fmt.Errorf("failed to load startup config: %s", err)
	}
	return *cfg, nil
}
