package queue

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
	"yambol/config"
	"yambol/pkg/telemetry"
	"yambol/pkg/util"

	"github.com/stretchr/testify/assert"
)

// TODO: Move QueueStats testing away from here. Maybe with a mock interface? Works for now I guess...

const (
	testQueueDefaultMinLen  = 10
	testQueueDefaultMaxLen  = 10000
	testQueueDefaultMaxSize = 1024 * 10 // 10KB
	testQueueDefaultTTL     = 1
)

func queueSetUp() (*Queue, *telemetry.QueueStats) {
	qs := &telemetry.QueueStats{}
	return New(
		config.QueueConfig{
			MinLength:    testQueueDefaultMinLen,
			MaxLength:    testQueueDefaultMaxLen,
			MaxSizeBytes: testQueueDefaultMaxSize,
			TTL:          testQueueDefaultTTL,
		}, qs), qs
}

func stringRange(n int) (rv []string) {
	rv = make([]string, n)
	for i := 0; i < n; i++ {
		rv[i] = strconv.Itoa(i)
	}
	return
}

func TestQueueBasicDefaultOps(t *testing.T) {
	q, qs := queueSetUp()
	var err error
	for i := 0; i < testQueueDefaultMaxLen; i++ {
		_, err = q.Push(strconv.Itoa(i))
		assert.NoError(t, err, "failed to push", i)
	}
	assert.Equal(t, testQueueDefaultMaxLen, q.Len(), "mismatched number of items")
	_, err = q.Push("test_val")
	assert.Error(t, err, "managed to push after max size has been reached")
	var val string
	for i := 0; i < testQueueDefaultMaxLen; i++ {
		itm := q.peek()
		assert.NotNil(t, itm, "nothing in queue")
		assert.Equal(t, util.Seconds(testQueueDefaultTTL), itm.ttl)
		val, err = q.Pop()
		assert.NoError(t, err, "failed to pop", i)
		assert.Equal(t, strconv.Itoa(i), val, "mismatched value popped", i, val)
	}
	assert.Empty(t, q.items, "queue not empty")
	assert.Equal(t, int64(testQueueDefaultMaxLen), qs.Processed, "mismatched number of processed items")
	_, err = q.Pop()
	assert.Error(t, err, "popped from empty queue")

	_, err = q.PushWithTTL("test", nil)
	assert.NoError(t, err, "failed to push with nil ttl")

	itm := q.peek()
	assert.NotNil(t, itm, "peek returned nil")
	assert.Equal(t, util.Seconds(testQueueDefaultTTL), itm.ttl)
	_, err = q.Pop()
	assert.NoError(t, err, "failed to pop nil ttl value")

	customTTL := time.Duration(rand.Int63())
	_, err = q.PushWithTTL("test", &customTTL)
	assert.NoError(t, err, "failed to push with nil ttl")

	itm = q.peek()
	assert.NotNil(t, itm, "peek returned nil")
	assert.Equal(t, customTTL, itm.ttl)
	_, err = q.Pop()
	assert.NoError(t, err, "failed to pop nil ttl value")

}

func TestQueueBulkOps(t *testing.T) {
	q, qs := queueSetUp()
	var err error
	_, err = q.PushBatch(stringRange(testQueueDefaultMaxLen - 9)...)
	assert.NoError(t, err, "failed to push batch")
	assert.Equal(t, testQueueDefaultMaxLen-9, q.Len(), "mismatched number of items")
	_, err = q.PushBatch(stringRange(10)...)
	assert.Error(t, err, "managed to push over max size")
	vals := q.Drain()
	assert.Equal(t, testQueueDefaultMaxLen-9, len(vals), "mismatched number of drained items")
	assert.Equal(t, int64(testQueueDefaultMaxLen-9), qs.Processed, "mismatched number of processed items")
}

func TestQueueExpiration(t *testing.T) {
	q, qs := queueSetUp()
	_, err := q.Push("test")
	assert.NoError(t, err, "failed to push")
	ptr := q.peek()
	assert.NotNil(t, ptr, "peek returned nil")
	time.Sleep(util.LittleLongerThan(util.Seconds(testQueueDefaultTTL)))
	_, err = q.Pop()
	assert.Error(t, err, "popped from empty queue")
	assert.Equal(t, int64(0), qs.Processed, "mismatched number of processed items")
	assert.Equal(t, int64(1), qs.Dropped, "mismatched number of dropped items")
}
