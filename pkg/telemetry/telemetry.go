package telemetry

import (
	"encoding/json"
	"sync/atomic"
	"time"
	"yambol/pkg/util/atomicx"
)

type QueueStats struct {
	Processed        uint64 `json:"sent"`
	Dropped          uint64 `json:"dropped"`
	TotalTimeInQueue uint64 `json:"total_time_in_queue"`
	MaxTimeInQueue   uint64 `json:"max_time_in_queue"`
}

func (qs *QueueStats) allMessages() uint64 {
	return qs.Processed + qs.Dropped
}

func (qs *QueueStats) Process(timeInQueue time.Duration) {
	atomic.AddUint64(&qs.Processed, 1)
	qs.update(timeInQueue)
}

func (qs *QueueStats) Drop(timeInQueue time.Duration) {
	atomic.AddUint64(&qs.Dropped, 1)
	qs.update(timeInQueue)
}

func (qs *QueueStats) update(timeInQueue time.Duration) {
	tiq := uint64(timeInQueue.Milliseconds())
	atomicx.MaxSwap64(&qs.MaxTimeInQueue, tiq)
	atomic.AddUint64(&qs.TotalTimeInQueue, tiq)
}

func (qs *QueueStats) AverageTimeInQueue() uint64 {
	return qs.TotalTimeInQueue / qs.allMessages()
}

func (qs *QueueStats) MarshalJSON() ([]byte, error) {
	// We need to prevent infinite recursion by aliasing the type,
	// so that our MarshalJSON won't be called again.
	type Alias QueueStats

	// Define a new struct to hold the AverageTimeInQueue value
	aux := struct {
		Alias
		AverageTimeInQueue uint64 `json:"average_time_in_queue"`
	}{
		Alias:              (Alias)(*qs),
		AverageTimeInQueue: qs.AverageTimeInQueue(),
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
		qs = &QueueStats{}
		c.qStats[queue] = qs
	}
	return qs
}
