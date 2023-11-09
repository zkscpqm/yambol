package rest

import (
	"context"
	"testing"
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
		TTL:          defaultTimeoutSeconds * 2,
	}

	testCreateQueue(t, ctx, client, qOpts)
	testRemoveQueue(t, ctx, client)
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
	err := client.CreateQueueContext(ctx, testingDefaultQueueName, qOpts)
	assert.NoError(t, err, "failed to create queue")

	queues, err := client.GetQueuesContext(ctx)
	assert.NoError(t, err, "failed to get queues")
	assert.Contains(t, *queues, testingDefaultQueueName, "the queue was not created correctly")

	runCfg := config.GetRunningConfig()
	qInfo, ok := runCfg.Broker.Queues[testingDefaultQueueName]
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
	err := client.DeleteQueueContext(ctx, testingDefaultQueueName)
	assert.NoError(t, err, "failed to delete queue")

	queues, err := client.GetQueuesContext(ctx)
	assert.NoError(t, err, "failed to get queues")
	assert.NotContains(t, *queues, testingDefaultQueueName, "the queue was not deleted correctly")

	runCfg := config.GetRunningConfig()
	assert.NotContains(t, runCfg.Broker.Queues, testingDefaultQueueName, "the queue is still in the running config")
}
