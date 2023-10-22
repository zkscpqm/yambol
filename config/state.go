package config

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"yambol/pkg/util"
	"yambol/pkg/util/log"
)

var (
	_config          = Empty()
	autoSaveDisabled bool
	mx               = &sync.RWMutex{}
	logger           = log.New("CONFIG", log.LevelOff)
)

func Init(config Configuration, l *log.Logger) {
	logger = l.NewFrom("CONFIG")
	SetRunningConfig(config)
}

func SetRunningConfig(config Configuration) {
	logger.Debug("Set running config to:\n%s", config.String())
	_config = config
	autoSaveDisabled = config.DisableAutoSave
	autoSave()
}

func DisableAutoSave(disable bool) {
	logger.Debug("Auto save: %s", util.BoolLabels(!disable, "enabled", "disabled"))
	_config.DisableAutoSave = disable
	autoSaveDisabled = disable
	autoSave()
}

func CreateQueue(queueName string, minLength, maxLength, maxSizeBytes int64, ttl time.Duration) {
	mx.Lock()
	defer mx.Unlock()

	_config.Broker.Queues[queueName] = QueueState{
		MinLength:    minLength,
		MaxLength:    maxLength,
		MaxSizeBytes: maxSizeBytes,
		TTL:          ttl.Nanoseconds(),
	}
	logger.Debug("Queue `%s` created", queueName)
	autoSave()
}

func DeleteQueue(queueName string) {
	mx.Lock()
	mx.Unlock()
	logger.Debug("Queue `%s` deleted")
	delete(_config.Broker.Queues, queueName)
	autoSave()
}

func SetDefaultMinLen(value int64) {
	atomic.StoreInt64(&_config.Broker.DefaultMinLength, value)
	logger.Debug("default min len set to %d", value)
	autoSave()
}

func SetDefaultMaxLen(value int64) {
	atomic.StoreInt64(&_config.Broker.DefaultMaxLength, value)
	logger.Debug("default max len set to %d", value)
	autoSave()
}

func SetDefaultMaxSizeBytes(value int64) {
	atomic.StoreInt64(&_config.Broker.DefaultMaxSizeBytes, value)
	logger.Debug("default max size bytes set to %d", value)
	autoSave()
}

func SetDefaultTTL(value int64) {
	atomic.StoreInt64(&_config.Broker.DefaultTTL, value)
	logger.Debug("default ttl set to %d", value)
	autoSave()
}

func autoSave() {
	if autoSaveDisabled {
		logger.Debug("auto save disabled")
		return
	}
	if err := CopyRunningConfigToStartupConfig(); err != nil {
		logger.Error("failed to auto save config:", err)
	}
}

func GetRunningConfig() Configuration {
	return _config.Copy()
}

func GetStartupConfig() (Configuration, error) {
	cfg, err := FromFile()
	if err != nil {
		return Empty(), fmt.Errorf("failed to load startup config: %s", err)
	}
	return *cfg, nil
}
