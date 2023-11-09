package rest

import (
	"context"
	"testing"
	"time"
	"yambol/config"
	"yambol/pkg/transport/httpx/rest"

	"github.com/stretchr/testify/assert"
)

func TestAPI(t *testing.T) {
	server, client, ctx, cancel := testInit(t)
	defer reset(t)
	defer cancel()
	run(t, server)
	testConfig(t, ctx, client)
	testQueueManagement(t, ctx, client)
	testQueueLogic(t, ctx, client)

}

func testConfig(t *testing.T, ctx context.Context, client *rest.Client) {

	testNoInitialStartupConfig(t, ctx, client)
	testInitialRunningConfig(t, ctx, client)
	testSaveRunCfgToStartCfg(t, ctx, client)
	testStartConfigExists(t, ctx, client)

}

func testQueueManagement(t *testing.T, ctx context.Context, client *rest.Client) {

	testNoInitialQueues(t, ctx, client)

	qOpts := config.QueueConfig{
		MinLength:    12,
		MaxLength:    1024,
		MaxSizeBytes: 42069,
		TTL:          defaultTimeoutSeconds,
	}

	testCreateQueue(t, ctx, client, qOpts)
	testRemoveQueue(t, ctx, client)
}

func testQueueLogic(t *testing.T, ctx context.Context, client *rest.Client) {

	testValue := "test_value"
	testCreateQueue(t, ctx, client, config.QueueConfig{
		MinLength:    12,
		MaxLength:    1024,
		MaxSizeBytes: 42069,
		TTL:          defaultTimeoutSeconds,
	})

	//_, err := client.ConsumeContext(ctx, defaultTestQueueName)
	//assert.Error(t, err, "was able to consume from an empty queue")

	err := client.PublishContext(ctx, defaultTestQueueName, testValue)
	assert.NoError(t, err, "failed to publish to queue")

	val, err := client.ConsumeContext(ctx, defaultTestQueueName)
	assert.NoError(t, err, "failed to consume from queue")
	assert.Equal(t, testValue, val, "failed to consume correct value from queue")

	err = client.PublishContextTimeout(ctx, defaultTestQueueName, testValue, time.Second)
	assert.NoError(t, err, "failed to publish to queue")
	time.Sleep(time.Second)
	v, err := client.Consume(defaultTestQueueName)
	assert.Equal(t, "", v, "expected empty value from empty queue")
	assert.NoError(t, err, "no error expected by consuming from empty queue")
}

func testNoInitialStartupConfig(t *testing.T, ctx context.Context, client *rest.Client) {
	cfg, err := client.GetStartupConfigContext(ctx)
	assert.Error(t, err, "expected no startup config")
	assert.Nil(t, cfg, "expected no startup config")
}

func testInitialRunningConfig(t *testing.T, ctx context.Context, client *rest.Client) {
	cfg, err := client.GetRunningConfigContext(ctx)
	assert.NoError(t, err, "expected existing running config")
	assert.Equal(t, defaultConfig, *cfg, "expected existing running config to be the default config")
}

func testSaveRunCfgToStartCfg(t *testing.T, ctx context.Context, client *rest.Client) {
	cfg, err := client.CopyRunCfgToStartCfgContext(ctx)
	assert.NoError(t, err, "failed to copy running config to startup config")
	assert.Equal(t, defaultConfig, *cfg, "failed to copy running config to startup config correctly")
}

func testStartConfigExists(t *testing.T, ctx context.Context, client *rest.Client) {
	cfg, err := client.GetStartupConfigContext(ctx)
	assert.NoError(t, err, "failed to get startup config")
	assert.Equal(t, defaultConfig, *cfg, "failed to get startup config correctly")
}

func testNoInitialQueues(t *testing.T, ctx context.Context, client *rest.Client) {
	queues, err := client.GetQueuesContext(ctx)
	assert.NoError(t, err, "failed to get queues")
	assert.Empty(t, queues, "there should be no queues available")
}

func testCreateQueue(t *testing.T, ctx context.Context, client *rest.Client, qOpts config.QueueConfig) {
	err := client.CreateQueueContext(ctx, defaultTestQueueName, qOpts)
	assert.NoError(t, err, "failed to create queue")

	queues, err := client.GetQueuesContext(ctx)
	assert.NoError(t, err, "failed to get queues")
	assert.Contains(t, *queues, defaultTestQueueName, "the queue was not created correctly")

	runCfg := config.GetRunningConfig()
	qInfo, ok := runCfg.Broker.Queues[defaultTestQueueName]
	assert.True(t, ok, "the queue cannot be found in the running config")
	assert.Equal(t,
		config.QueueConfig{
			MinLength:    qOpts.MinLength,
			MaxLength:    qOpts.MaxLength,
			MaxSizeBytes: qOpts.MaxSizeBytes,
			TTL:          qOpts.TTL,
		},
		qInfo,
		"mismatched queue state options",
	)
}

func testRemoveQueue(t *testing.T, ctx context.Context, client *rest.Client) {
	err := client.DeleteQueueContext(ctx, defaultTestQueueName)
	assert.NoError(t, err, "failed to delete queue")

	queues, err := client.GetQueuesContext(ctx)
	assert.NoError(t, err, "failed to get queues")
	assert.NotContains(t, *queues, defaultTestQueueName, "the queue was not deleted correctly")

	runCfg := config.GetRunningConfig()
	assert.NotContains(t, runCfg.Broker.Queues, defaultTestQueueName, "the queue is still in the running config")
}
