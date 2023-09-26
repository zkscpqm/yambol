package queue

import (
	"strconv"
	"testing"
	"time"

	"yambol/pkg/telemetry"
	"yambol/pkg/util"

	"github.com/stretchr/testify/assert"
)

// TODO: Move QueueStats testing away from here. Maybe with a mock interface? Works for now I guess...

const (
	TestMinLen  = 10
	TestMaxLen  = 10000
	TestMaxSize = 1024 * 10 // 10KB
	TestTTL     = time.Second * 1
)

func queueSetUp() (*Queue, *telemetry.QueueStats) {
	qs := &telemetry.QueueStats{}
	return New(TestMinLen, TestMaxLen, TestMaxSize, TestTTL, qs), qs
}

func stringRange(n int) (rv []string) {
	rv = make([]string, n)
	for i := 0; i < n; i++ {
		rv[i] = strconv.Itoa(i)
	}
	return
}

func TestQueueBasicOps(t *testing.T) {
	q, qs := queueSetUp()
	var err error
	for i := 0; i < TestMaxLen; i++ {
		_, err = q.Push(strconv.Itoa(i))
		assert.NoError(t, err, "failed to push", i)
	}
	assert.Equal(t, TestMaxLen, q.Len(), "mismatched number of items")
	_, err = q.Push("oob")
	assert.Error(t, err, "managed to push after max size has been reached")
	var val string
	for i := 0; i < TestMaxLen; i++ {
		val, err = q.Pop()
		assert.NoError(t, err, "failed to pop", i)
		assert.NotNil(t, val, "empty value popped", i)
		assert.Equal(t, strconv.Itoa(i), val, "mismatched value popped", i, val)
	}
	assert.Empty(t, q.items, "queue not empty")
	assert.Equal(t, uint64(TestMaxLen), qs.Processed, "mismatched number of processed items")
	_, err = q.Pop()
	assert.Error(t, err, "popped from empty queue")
}

func TestQueueBulkOps(t *testing.T) {
	q, qs := queueSetUp()
	var err error
	_, err = q.PushBatch(stringRange(TestMaxLen - 9)...)
	assert.NoError(t, err, "failed to push batch")
	assert.Equal(t, TestMaxLen-9, q.Len(), "mismatched number of items")
	_, err = q.PushBatch(stringRange(10)...)
	assert.Error(t, err, "managed to push over max size")
	vals := q.Drain()
	assert.Equal(t, TestMaxLen-9, len(vals), "mismatched number of drained items")
	assert.Equal(t, uint64(TestMaxLen-9), qs.Processed, "mismatched number of processed items")
}

func TestQueueExpiration(t *testing.T) {
	q, qs := queueSetUp()
	_, err := q.Push("test")
	assert.NoError(t, err, "failed to push")
	ptr := q.peek()
	assert.NotNil(t, ptr, "peek returned nil")
	time.Sleep(util.LittleLongerThan(TestTTL))
	_, err = q.Pop()
	assert.Error(t, err, "popped from empty queue")
	assert.Equal(t, uint64(0), qs.Processed, "mismatched number of processed items")
	assert.Equal(t, uint64(1), qs.Dropped, "mismatched number of dropped items")
}
