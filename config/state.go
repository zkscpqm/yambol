package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"yambol/pkg/util"
	"yambol/pkg/util/log"
)

var (
	activeState = defaultState()
	mx          = &sync.RWMutex{}
	logger      = log.New("CONFIG", log.LevelOff)
)

type queueStateMap map[string]queueState

func (qm queueStateMap) Copy() queueStateMap {
	rv := make(queueStateMap)
	for k, v := range qm {
		rv[k] = v
	}
	return rv
}

func (qm queueStateMap) toQueueConfig() QueueMap {
	rv := make(QueueMap)
	for k, v := range qm {
		rv[k] = QueueConfig{
			MinLength:    v.minLength,
			MaxLength:    v.maxLength,
			MaxSizeBytes: v.maxSizeBytes,
			TTL:          int64(v.ttl.Seconds()),
		}
	}
	return rv
}

type queueState struct {
	minLength    int64
	maxLength    int64
	maxSizeBytes int64
	ttl          time.Duration
}

type brokerState struct {
	DefaultMinLength    int64
	DefaultMaxLength    int64
	DefaultMaxSizeBytes int64
	DefaultTTL          time.Duration
	Queues              queueStateMap
}

func (s brokerState) Copy() (rv brokerState) {
	q := s.Queues
	if q == nil {
		q = make(queueStateMap)
	}
	return brokerState{
		DefaultMinLength:    s.DefaultMinLength,
		DefaultMaxLength:    s.DefaultMaxLength,
		DefaultMaxSizeBytes: s.DefaultMaxSizeBytes,
		DefaultTTL:          s.DefaultTTL,
		Queues:              q.Copy(),
	}
}

type serverState struct {
	Enabled    bool
	Port       int
	TlsEnabled bool
}

type apiState struct {
	REST        serverState
	GRPC        serverState
	HTTP        serverState
	Certificate string
	Key         string
}

type logState struct {
	Level string
	File  string
}

type state struct {
	DisableAutoSave bool
	API             apiState
	Broker          brokerState
	Log             logState
}

func (s state) asConfig() (rv Configuration) {
	return Configuration{
		DisableAutoSave: s.DisableAutoSave,
		API: ApiConfig{
			REST: Server{
				Enabled:    s.API.REST.Enabled,
				Port:       s.API.REST.Port,
				TlsEnabled: s.API.REST.TlsEnabled,
			},
			GRPC: Server{
				Enabled:    s.API.GRPC.Enabled,
				Port:       s.API.GRPC.Port,
				TlsEnabled: s.API.GRPC.TlsEnabled,
			},
			HTTP: Server{
				Enabled:    s.API.HTTP.Enabled,
				Port:       s.API.HTTP.Port,
				TlsEnabled: s.API.HTTP.TlsEnabled,
			},
			Certificate: s.API.Certificate,
			Key:         s.API.Key,
		},
		Broker: BrokerConfig{
			DefaultMinLength:    s.Broker.DefaultMinLength,
			DefaultMaxLength:    s.Broker.DefaultMaxLength,
			DefaultMaxSizeBytes: s.Broker.DefaultMaxSizeBytes,
			DefaultTTLSeconds:   int64(s.Broker.DefaultTTL.Seconds()),
			Queues:              s.Broker.Queues.toQueueConfig(),
		},
		Log: LogConfig{
			Level: s.Log.Level,
			File:  s.Log.File,
		},
	}
}

func defaultState() state {
	return state{
		Broker: brokerState{
			Queues: make(queueStateMap),
		},
	}
}

func autoSaveDisabled() bool {
	return activeState.DisableAutoSave
}

func Init(config Configuration, l *log.Logger) {
	SetLogger(l)
	SetRunningConfig(config)
}

func SetLogger(l *log.Logger) {
	logger = l.NewFrom("CONFIG")
}

func SetRunningConfig(config Configuration) {
	logger.Debug("Set running config to:\n%s", config.String())
	activeState = config.state()
	autoSave()
}

func DisableAutoSave(disable bool) {
	logger.Debug("Auto save: %s", util.BoolLabels(!disable, "enabled", "disabled"))
	activeState.DisableAutoSave = disable
	autoSave()
}

func CreateQueue(queueName string, cfg QueueConfig) {
	mx.Lock()
	defer mx.Unlock()

	activeState.Broker.Queues[queueName] = cfg.state()
	logger.Debug("Queue `%s` created", queueName)
	autoSave()
}

func DeleteQueue(queueName string) {
	mx.Lock()
	mx.Unlock()
	logger.Debug("Queue `%s` deleted", queueName)
	delete(activeState.Broker.Queues, queueName)
	autoSave()
}

func SetDefaultMinLen(value int64) {
	atomic.StoreInt64(&activeState.Broker.DefaultMinLength, value)
	logger.Debug("default min len set to %d", value)
	autoSave()
}

func SetDefaultMaxLen(value int64) {
	atomic.StoreInt64(&activeState.Broker.DefaultMaxLength, value)
	logger.Debug("default max len set to %d", value)
	autoSave()
}

func SetDefaultMaxSizeBytes(value int64) {
	atomic.StoreInt64(&activeState.Broker.DefaultMaxSizeBytes, value)
	logger.Debug("default max size bytes set to %d", value)
	autoSave()
}

func SetDefaultTTL(value int64) {
	activeState.Broker.DefaultTTL = util.Seconds(atomic.LoadInt64(&value))
	logger.Debug("default ttl set to %ds", value)
	autoSave()
}

func autoSave() {
	if autoSaveDisabled() {
		return
	}
	if err := CopyRunningConfigToStartupConfig(); err != nil {
		logger.Error("failed to auto save config:", err)
	}
}

func GetRunningConfig() Configuration {
	return activeState.asConfig()
}

func CopyRunningConfigToStartupConfig() error {
	f, err := os.OpenFile(configFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open state log file %s: %v", configFilePath, err)
	}
	defer f.Close()
	cfg := activeState.asConfig()
	return json.NewEncoder(f).Encode(&cfg)
}
