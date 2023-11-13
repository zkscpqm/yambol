package broker

import (
	"yambol/config"
)

var (
	defaultMinLen       = int64(100)
	defaultMaxLen       = int64(1024 * 1024 * 1024)
	defaultMaxSizeBytes = int64(1024 * 1024 * 1024) // 1GB
	defaultTTLSeconds   = int64(0)
)

func setLTE0(value, default_ int64, target *int64) int64 {
	if value <= 0 {
		value = default_
	}
	*target = value
	return value
}

func SetDefaultMinLen(value int64) {
	config.SetDefaultMinLen(setLTE0(value, 100, &defaultMinLen))
}

func SetDefaultMaxLen(value int64) {
	config.SetDefaultMaxLen(setLTE0(value, 1024*1024*1024, &defaultMaxLen))
}

func SetDefaultMaxSizeBytes(value int64) {
	config.SetDefaultMaxSizeBytes(setLTE0(value, 1024*1024*1024, &defaultMaxSizeBytes))
}

func SetDefaultTTL(value int64) {
	if value < 0 {
		value = 0
	}
	defaultTTLSeconds = value
	config.SetDefaultTTL(value)
}

func GetDefaultMinLen() int64 {
	return defaultMinLen
}

func GetDefaultMaxLen() int64 {
	return defaultMaxLen
}

func GetDefaultMaxSizeBytes() int64 {
	return defaultMaxSizeBytes
}

func GetDefaultTTL() int64 {
	return defaultTTLSeconds
}

func determineMinLen(value int64) int64 {
	if value >= 0 {
		return value
	}
	return defaultMinLen
}

func determineMaxLen(value int64) int64 {
	if value > 0 {
		return value
	}
	return defaultMaxLen
}

func determineMaxSizeBytes(value int64) int64 {
	if value > 0 {
		return value
	}
	return defaultMaxSizeBytes
}

func determineTTL(value int64) int64 {
	if value > 0 {
		return value
	}
	return defaultTTLSeconds
}
