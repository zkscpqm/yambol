package broker

import "time"

var (
	defaultMinLen       = 100
	defaultMaxLen       = 1024 * 1024 * 1024
	defaultMaxSizeBytes = 1024 * 1024 * 1024 // 1GB
	defaultTTL          = time.Duration(0)
)

func setLTE0(value, default_ int, target *int) {
	if value <= 0 {
		value = default_
	}
	*target = value
}

func SetDefaultMinLen(value int) {
	setLTE0(value, 100, &defaultMinLen)
}

func SetDefaultMaxLen(value int) {
	setLTE0(value, 1024*1024*1024, &defaultMaxLen)
}

func SetDefaultMaxSizeBytes(value int) {
	setLTE0(value, 1024*1024*1024, &defaultMaxSizeBytes)
}

func SetDefaultTTL(value time.Duration) {
	if value < 0 {
		value = time.Duration(0)
	}
	defaultTTL = value
}

func GetDefaultMinLen() int {
	return defaultMinLen
}

func GetDefaultMaxLen() int {
	return defaultMaxLen
}

func GetDefaultMaxSizeBytes() int {
	return defaultMaxSizeBytes
}

func GetDefaultTTL() time.Duration {
	return defaultTTL
}

func determineMinLen(value int) int {
	if value >= 0 {
		return value
	}
	return defaultMinLen
}

func determineMaxLen(value int) int {
	if value > 0 {
		return value
	}
	return defaultMaxLen
}

func determineMaxSizeBytes(value int) int {
	if value > 0 {
		return value
	}
	return defaultMaxSizeBytes
}

func determineTTL(value time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	return defaultTTL
}

type QueueOptions struct {
	MinLen       int
	MaxLen       int
	MaxSizeBytes int
	DefaultTTL   time.Duration
}
