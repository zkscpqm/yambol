package broker

import (
	"fmt"
	"time"
	"yambol/config"

	"yambol/pkg/queue"
	"yambol/pkg/telemetry"
)

type MessageBroker struct {
	queues    map[string]*queue.Queue
	unsent    map[string][]string
	stats     *telemetry.Collector
	ephemeral bool
}

func New() (*MessageBroker, error) {
	return &MessageBroker{
		queues: make(map[string]*queue.Queue),
		unsent: make(map[string][]string),
		stats:  telemetry.NewCollector(),
	}, nil
}

func (mb *MessageBroker) AddDefaultQueue(queueName string) error {
	return mb.AddQueue(queueName,
		QueueOptions{
			GetDefaultMinLen(),
			GetDefaultMaxLen(),
			GetDefaultMaxSizeBytes(),
			GetDefaultTTL(),
		},
	)
}

func (mb *MessageBroker) AddQueue(queueName string, opts QueueOptions) error {
	_, exists := mb.queues[queueName]
	if exists {
		return fmt.Errorf("queue %s already exists", queueName)
	}
	queueStats := mb.stats.AddQueue(queueName)

	minLen := determineMinLen(opts.MinLen)
	maxLen := determineMaxLen(opts.MaxLen)
	maxSizeBytes := determineMaxSizeBytes(opts.MaxSizeBytes)
	ttl := determineTTL(opts.DefaultTTL)

	mb.queues[queueName] = queue.New(
		determineMinLen(opts.MinLen),
		determineMaxLen(opts.MaxLen),
		determineMaxSizeBytes(opts.MaxSizeBytes),
		determineTTL(opts.DefaultTTL),
		queueStats,
	)
	mb.unsent[queueName] = make([]string, 0)
	config.CreateQueue(queueName, minLen, maxLen, maxSizeBytes, ttl)
	return nil
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

func (mb *MessageBroker) PublishWithTTL(message string, ttl *time.Duration, queueNames ...string) error {
	if len(queueNames) == 0 {
		return fmt.Errorf("no queue name provided")
	}
	errors := make(map[string]error)
	for _, queueName := range queueNames {
		if q, ok := mb.queues[queueName]; !ok {
			errors[queueName] = fmt.Errorf("queue '%s' not found", queueName)
		} else {
			if _, err := q.PushWithTTL(message, ttl); err != nil {
				errors[queueName] = err
				mb.unsent[queueName] = append(mb.unsent[queueName], message)
			}
		}
	}
	return mb.formatMultipleErrors("one or more queues failed to send message:", errors)
}

func (mb *MessageBroker) Publish(message string, queueNames ...string) error {
	return mb.PublishWithTTL(message, nil, queueNames...)
}

func (mb *MessageBroker) Broadcast(message string) error {
	return mb.Publish(message, mb.Queues()...)
}

func (mb *MessageBroker) BroadcastWithTTL(message string, ttl *time.Duration) error {
	return mb.PublishWithTTL(message, ttl, mb.Queues()...)
}

func (mb *MessageBroker) Consume(queueName string) (string, error) {
	if q, ok := mb.queues[queueName]; !ok {
		return "", fmt.Errorf("queue '%s' not found", queueName)
	} else {
		return q.Pop()
	}
}

func (mb *MessageBroker) Stats() map[string]telemetry.QueueStats {
	return mb.stats.Stats()
}

func (mb *MessageBroker) QueueExists(queueName string) bool {
	_, ok := mb.queues[queueName]
	return ok
}

func (mb *MessageBroker) Queues() (queueNames []string) {
	for queueName := range mb.queues {
		queueNames = append(queueNames, queueName)
	}
	return
}

func (mb *MessageBroker) RemoveQueue(queueName string) error {
	_, ok := mb.queues[queueName]
	if !ok {
		return fmt.Errorf("queue '%s' not found", queueName)
	}
	delete(mb.queues, queueName)
	config.DeleteQueue(queueName)
	// TODO: Save the queue messages?
	return nil

}

// Todo: Close() error -> dump queues to file or share to peers. Also, peers?
