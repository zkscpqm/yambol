package broker

import (
	"fmt"
	"time"
	"yambol/pkg/queue"
)

var (
	defaultMinLen       = 100
	defaultMaxLen       = 1 << 31
	defaultMaxSizeBytes = 1024 * 1024 * 1024 // 1GB
	defaultTTL          = time.Minute
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
	setLTE0(value, 1<<31, &defaultMaxLen)
}

func SetDefaultMaxSizeBytes(value int) {
	setLTE0(value, 1024*1024*1024, &defaultMaxSizeBytes)
}

func SetDefaultTTL(value time.Duration) {
	if value < 0 {
		value = 0
	}
	defaultTTL = time.Minute * value
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
	Name         string
	MinLen       int
	MaxLen       int
	MaxSizeBytes int
}

type MessageBroker struct {
	queues map[string]*queue.Queue
	unsent map[string][]string
}

func NewMessageBroker() *MessageBroker {
	return &MessageBroker{
		queues: make(map[string]*queue.Queue),
		unsent: make(map[string][]string),
	}
}

func (mb *MessageBroker) AddDefaultQueue(queueName string) {
	mb.AddQueue(queueName, GetDefaultMinLen(), GetDefaultMaxLen(), GetDefaultMaxSizeBytes(), GetDefaultTTL())
}

func (mb *MessageBroker) AddQueue(queueName string, minLen, maxLen, maxSizeBytes int, ttl time.Duration) {
	mb.queues[queueName] = queue.New(
		determineMinLen(minLen),
		determineMaxLen(maxLen),
		determineMaxSizeBytes(maxSizeBytes),
		determineTTL(ttl),
	)
	mb.unsent[queueName] = make([]string, 0)
}

func (mb *MessageBroker) formatMultipleErrors(base string, errors map[string]error) error {
	if len(errors) > 0 {
		var queueName string
		err := fmt.Errorf(base)
		for queueName, err = range errors {
			err = fmt.Errorf("\n [%s] -> %s", queueName, err)
		}
		return err
	}
	return nil
}

func (mb *MessageBroker) Publish(message string, queueNames ...string) error {
	if len(queueNames) == 0 {
		return fmt.Errorf("no queue name provided")
	}
	errors := make(map[string]error)
	for _, queueName := range queueNames {
		if q, ok := mb.queues[queueName]; !ok {
			errors[queueName] = fmt.Errorf("queue '%s' not found", queueName)
		} else {
			if _, err := q.Push(message); err != nil {
				errors[queueName] = err
				mb.unsent[queueName] = append(mb.unsent[queueName], message)
			}
		}
	}
	return mb.formatMultipleErrors("one or more queues failed to send message:", errors)
}

func (mb *MessageBroker) Receive(queueName string) (string, error) {
	if q, ok := mb.queues[queueName]; !ok {
		return "", fmt.Errorf("queue '%s' not found", queueName)
	} else {
		return q.Pop()
	}
}
