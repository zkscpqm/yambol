package broker

import (
	"encoding/json"
	"fmt"
	"time"

	"yambol/config"
	"yambol/pkg/queue"
	"yambol/pkg/telemetry"
	"yambol/pkg/util/log"
)

type MessageBroker struct {
	queues    map[string]*queue.Queue
	unsent    map[string][]string
	stats     *telemetry.Collector
	ephemeral bool
	logger    *log.Logger
}

func New(logger *log.Logger) *MessageBroker {
	return &MessageBroker{
		queues: make(map[string]*queue.Queue),
		unsent: make(map[string][]string),
		stats:  telemetry.NewCollector(),
		logger: logger.NewFrom("BROKER"),
	}
}

func (mb *MessageBroker) AddDefaultQueue(queueName string) error {
	return mb.AddQueue(queueName,
		config.QueueConfig{
			MinLength:    GetDefaultMinLen(),
			MaxLength:    GetDefaultMaxLen(),
			MaxSizeBytes: GetDefaultMaxSizeBytes(),
			TTL:          GetDefaultTTL(),
		},
	)
}

func (mb *MessageBroker) AddQueue(queueName string, cfg config.QueueConfig) error {
	mb.logger.Info("Trying to add queue `%s`: %s", queueName, cfg)

	_, exists := mb.queues[queueName]
	if exists {
		mb.logger.Error("failed to add queue `%s` as it already exists", queueName)
		return fmt.Errorf("queue %s already exists", queueName)
	}
	queueStats := mb.stats.AddQueue(queueName)

	cfg.MinLength = determineMinLen(cfg.MinLength)
	cfg.MaxLength = determineMaxLen(cfg.MaxLength)
	cfg.MaxSizeBytes = determineMaxSizeBytes(cfg.MaxSizeBytes)
	cfg.TTL = determineTTL(cfg.TTL)

	mb.logger.Debug("Adding queue `%s` with determined: %s", queueName, cfg)
	mb.queues[queueName] = queue.New(cfg, queueStats)
	mb.unsent[queueName] = make([]string, 0)
	config.CreateQueue(queueName, cfg)
	mb.logger.Info("Queue `%s` created", queueName)
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

func (mb *MessageBroker) PublishWithTTL(message string, ttl *time.Duration, queueNames ...string) (err error) {
	if len(queueNames) == 0 {
		return fmt.Errorf("no queue name provided")
	}
	errors := make(map[string]error)
	for _, queueName := range queueNames {
		if q, ok := mb.queues[queueName]; !ok {
			errors[queueName] = fmt.Errorf("queue '%s' not found", queueName)
			mb.logger.Error(errors[queueName].Error())
		} else {
			if _, err = q.PushWithTTL(message, ttl); err != nil {
				errors[queueName] = err
				mb.logger.Error("failed to push message to queue `%s`: %v", queueName, err)
				mb.unsent[queueName] = append(mb.unsent[queueName], message)
			}
		}
	}
	err = mb.formatMultipleErrors("one or more queues failed to send message:", errors)
	if err != nil {
		mb.logger.Error(err.Error())
	}
	return
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
	stats := mb.stats.Stats()
	if mb.logger.GetLevel() <= log.LevelDebug {
		// slightly more complex debug logic to avoid marshaling each time
		b, err := json.MarshalIndent(stats, "", "    ")
		if err == nil {
			mb.logger.Debug("got stats with err `%v`:\n%s", err, string(b))
		}
	}
	return stats
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
	mb.logger.Info("Trying to remove queue `%s`", queueName)
	_, ok := mb.queues[queueName]
	if !ok {
		err := fmt.Errorf("queue '%s' not found", queueName)
		mb.logger.Error(err.Error())
		return err
	}
	delete(mb.queues, queueName)
	mb.stats.RemoveQueue(queueName)
	config.DeleteQueue(queueName)
	// TODO: Save the queue messages?
	mb.logger.Info("Queue `%s` removed", queueName)
	return nil

}

// Todo: Close() error -> dump queues to file or share to peers. Also, peers?
