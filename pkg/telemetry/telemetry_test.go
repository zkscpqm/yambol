package telemetry

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	defaultTIQ = time.Millisecond * 4
)

func TestQueueStatsAPI(t *testing.T) {

	qs := QueueStats{}

	qs.Process(defaultTIQ)
	assert.Equal(t, int64(1), qs.Processed, "should have processed 1 item")
	assert.Equal(t, defaultTIQ.Milliseconds(), qs.TotalTimeInQueue)
	assert.Equal(t, defaultTIQ.Milliseconds(), qs.MaxTimeInQueue)

	qs.Drop(defaultTIQ * 3)
	assert.Equal(t, int64(1), qs.Dropped, "should have dropped 1 item")
	assert.Equal(t, defaultTIQ.Milliseconds()*4, qs.TotalTimeInQueue) // 1 dtiq before, 3 now
	assert.Equal(t, defaultTIQ.Milliseconds()*3, qs.MaxTimeInQueue)   // new dtiq is this 3

}

func TestQueueStatsEnrichment(t *testing.T) {
	qs := QueueStats{}
	assert.Zero(t, qs.averageTimeInQueue(), "initial stats should have 0 average")
	qs.Process(time.Second)
	assert.Equal(t, time.Second.Milliseconds(), qs.averageTimeInQueue(), "should have processed 1 item so avg == item tiq")
	qs.Process(time.Second * 3)
	assert.Equal(t, (time.Second * 2).Milliseconds(), qs.averageTimeInQueue(), "should have processed 1+3 item so avg == 4/2=2")
}

func TestQueueStatsRender(t *testing.T) {

	// TODO: Make this test less shit

	qs := QueueStats{}
	qs.Process(defaultTIQ * 3)
	qs.Drop(defaultTIQ)
	expectedJsonMap := fmt.Sprintf(
		`{"processed": 1, "dropped": 1, "total_time_in_queue_ms": %d, "max_time_in_queue_ms": %d, "average_time_in_queue_ms": %d}`,
		defaultTIQ.Milliseconds()*4,
		defaultTIQ.Milliseconds()*3,
		defaultTIQ.Milliseconds()*2,
	)

	// because on map[string]interface{} JSON unmarshals numbers as float64, not uint64.
	// Might be annoying when the stats get more complicated (nested)
	actualB, err := qs.MarshalJSON()
	assert.NoError(t, err, "failed to marshal actual stats JSON")
	//err = json.Unmarshal(actualB, &actualJsonMap)
	//assert.NoError(t, err, "failed to unmarshal actual data to map")

	assert.JSONEq(t, expectedJsonMap, string(actualB), "discrepancies in stats JSON representation")
}

func TestCollector(t *testing.T) {
	initialQueues := []string{"test1", "test2", "test3"}
	c := NewCollector(initialQueues...)
	assert.Len(t, c.qStats, len(initialQueues), "wrong initial queueStats length")
	for _, qName := range initialQueues {
		assert.Contains(t, c.qStats, qName, "queue name not found:", qName)
	}
	c.AddQueue("test4")
	assert.Contains(t, c.qStats, "test4", "failed to add new queue")
	// TODO: collector stats reporting?
}
