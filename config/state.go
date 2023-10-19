package config

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	_config          = Empty()
	autoSaveDisabled bool
	mx               = &sync.RWMutex{}
)

func Init(config Configuration) {
	SetRunningConfig(config)
}

func SetRunningConfig(config Configuration) {
	_config = config
	autoSaveDisabled = config.DisableAutoSave
	autoSave()
}

func DisableAutoSave(b bool) {
	_config.DisableAutoSave = b
	autoSaveDisabled = b
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
	autoSave()
}

func DeleteQueue(queueName string) {
	mx.Lock()
	mx.Unlock()
	delete(_config.Broker.Queues, queueName)
	autoSave()
}

func SetDefaultMinLen(value int64) {
	atomic.StoreInt64(&_config.Broker.DefaultMinLength, value)
	autoSave()
}

func SetDefaultMaxLen(value int64) {
	atomic.StoreInt64(&_config.Broker.DefaultMaxLength, value)
	autoSave()
}

func SetDefaultMaxSizeBytes(value int64) {
	atomic.StoreInt64(&_config.Broker.DefaultMaxSizeBytes, value)
	autoSave()
}

func SetDefaultTTL(value int64) {
	atomic.StoreInt64(&_config.Broker.DefaultTTL, value)
	autoSave()
}

func autoSave() {
	if autoSaveDisabled {
		return
	}
	if err := CopyRunningConfigToStartupConfig(); err != nil {
		fmt.Println("failed to auto save config:", err)
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
