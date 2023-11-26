package telemetry

import (
	"encoding/json"
	"sync/atomic"
	"time"

	"yambol/pkg/util/atomicx"
)

type QueueStats struct {
	Processed        int64 `json:"processed"`
	Dropped          int64 `json:"dropped"`
	TotalTimeInQueue int64 `json:"total_time_in_queue_ms"`
	MaxTimeInQueue   int64 `json:"max_time_in_queue_ms"`
}

func (qs *QueueStats) allMessages() int64 {
	return qs.Processed + qs.Dropped
}

func (qs *QueueStats) Process(timeInQueue time.Duration) {
	atomic.AddInt64(&qs.Processed, 1)
	qs.update(timeInQueue)
}

func (qs *QueueStats) Drop(timeInQueue time.Duration) {
	atomic.AddInt64(&qs.Dropped, 1)
	qs.update(timeInQueue)
}

func (qs *QueueStats) update(timeInQueue time.Duration) {
	tiq := timeInQueue.Milliseconds()
	atomicx.MaxSwap64(&qs.MaxTimeInQueue, tiq)
	atomic.AddInt64(&qs.TotalTimeInQueue, tiq)
}

func (qs *QueueStats) averageTimeInQueue() int64 {
	if qs.allMessages() == 0 {
		return 0
	}
	return qs.TotalTimeInQueue / qs.allMessages()
}

func (qs *QueueStats) MarshalJSON() ([]byte, error) {
	// We need to prevent infinite recursion by aliasing the type,
	// so that our Render won't be called again.
	type Alias QueueStats

	// Define a new struct to hold the averageTimeInQueue value
	aux := struct {
		Alias
		AverageTimeInQueue int64 `json:"average_time_in_queue_ms"`
	}{
		Alias:              (Alias)(*qs),
		AverageTimeInQueue: qs.averageTimeInQueue(),
	}

	// Marshal the aux struct into JSON
	return json.Marshal(aux)
}

type Collector struct {
	qStats map[string]*QueueStats
}

func NewCollector(queues ...string) *Collector {
	statMap := make(map[string]*QueueStats)
	for _, queue := range queues {
		statMap[queue] = &QueueStats{}
	}
	return &Collector{
		qStats: statMap,
	}
}

func (c *Collector) AddQueue(queue string) *QueueStats {
	qs, ok := c.qStats[queue]
	if !ok {
		qs = new(QueueStats)
		c.qStats[queue] = qs
	}
	return qs
}

func (c *Collector) RemoveQueue(queue string) {
	if _, ok := c.qStats[queue]; ok {
		delete(c.qStats, queue)
	}
}

func (c *Collector) Stats() map[string]QueueStats {
	s := make(map[string]QueueStats)
	for queueName, q := range c.qStats {
		s[queueName] = *q
	}
	return s
}
