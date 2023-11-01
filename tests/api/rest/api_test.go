package rest

import (
	"context"
	"testing"

	"yambol/pkg/broker"
	"yambol/pkg/transport/httpx/rest"

	"github.com/stretchr/testify/assert"
)

func TestAPI(t *testing.T) {
	server, client, ctx, cancel := testInit(t)
	defer cancel()
	run(t, server)
	testConfig(t, ctx, client)

}

func testConfig(t *testing.T, ctx context.Context, client *rest.Client) {
	cfg, err := client.GetStartupConfigContext(ctx)
	assert.Error(t, err, "expected no startup config")
	assert.Nil(t, cfg, "expected no startup config")

	cfg, err = client.GetRunningConfigContext(ctx)
	assert.NoError(t, err, "expected existing running config")
	assert.Equal(t, defaultConfig, *cfg, "expected existing running config to be the default config")

	startCfg, err := client.CopyRunCfgToStartCfgContext(ctx)
	assert.NoError(t, err, "failed to copy running config to startup config")
	assert.Equal(t, defaultConfig, *startCfg, "failed to copy running config to startup config correctly")

	cfg, err = client.GetStartupConfigContext(ctx)
	assert.NoError(t, err, "failed to get startup config")
	assert.Equal(t, defaultConfig, *cfg, "failed to get startup config correctly")

	assert.NoErrorf(t, removeConfigFile(t), "failed to remove local config file after testing")
}

func testQueueManagement(t *testing.T, ctx context.Context, client *rest.Client) {
	// TODO: This
	err := client.CreateQueueContext(ctx, "test-queue", broker.QueueOptions{})
	assert.NoError(t, err, "failed to create queue")

	_, err = client.GetQueueContext(ctx, "test-queue")
	assert.NoError(t, err, "failed to get queue")

	_, err = client.DeleteQueueContext(ctx, "test-queue")
	assert.NoError(t, err, "failed to delete queue")
}
