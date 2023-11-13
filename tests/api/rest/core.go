package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
	"yambol/pkg/util"

	"yambol/config"
	"yambol/pkg/broker"
	"yambol/pkg/transport/httpx/rest"
	"yambol/pkg/util/log"
)

const (
	restApiTestServerPort = 21519
	defaultTimeoutSeconds = 5
	defaultTestQueueName  = "_rest_api_test_queue"
)

var (
	defaultConfig = config.Configuration{
		DisableAutoSave: true,
		API: config.ApiConfig{
			REST: config.Server{
				Enabled:    true,
				Port:       restApiTestServerPort,
				TlsEnabled: false,
			},
		},
		Broker: config.BrokerConfig{
			DefaultMinLength:    10,
			DefaultMaxLength:    1000,
			DefaultMaxSizeBytes: 1000,
			DefaultTTLSeconds:   10,
			Queues:              config.QueueMap{},
		},
		Log: config.LogConfig{
			Level: "debug",
		},
	}
)

func testInit(t *testing.T) (*rest.Server, *rest.Client, context.Context, context.CancelFunc) {
	reset(t)

	logger := log.New("REST_API_TESTS", log.LevelDebug, log.NewDefaultStdioHandler())
	config.Init(defaultConfig, logger)

	server := rest.NewServer(
		broker.New(logger),
		nil,
		logger,
	)

	client := rest.NewClient(fmt.Sprintf("http://0.0.0.0:%d", restApiTestServerPort), http.DefaultClient, util.Seconds(defaultTimeoutSeconds))
	ctx, cancel := context.WithTimeout(context.Background(), util.Seconds(defaultTimeoutSeconds))
	return server, client, ctx, cancel
}

func run(t *testing.T, s *rest.Server) {
	go func() {
		if err := s.ListenAndServeInsecure(restApiTestServerPort); err != nil {
			t.Fatalf(">>>>>>REST API server FAILED: %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)
}

func removeConfigFile() (err error) {
	// Check if "config.json" exists in the current directory.
	configPath := "config.json"
	if _, err = os.Stat(configPath); err == nil {
		// If it exists, delete it.
		if err = os.Remove(configPath); err != nil {
			return fmt.Errorf("failed to remove config.json: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check for config.json: %v", err)
	}
	return nil
}

func reset(t *testing.T) {
	// Get the current working directory.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Check if the directory matches the pattern.
	matched, err := regexp.MatchString(expectedPathRegex(), cwd)
	if err != nil {
		t.Fatalf("Failed to execute regex: %v", err)
	}

	if !matched {
		t.Fatal("FATAL: Test running in a non-test directory")
	}
	if err = removeConfigFile(); err != nil {
		t.Fatalf("Failed to remove config.json file from REST API test dir: %v", err)
	}
}

func expectedPathRegex() string {
	sep := regexp.QuoteMeta(string(filepath.Separator))
	return fmt.Sprintf(`.*%stests%sapi%srest`, sep, sep, sep)
}

func logJson(t *testing.T, obj interface{}) {
	b, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	t.Logf("%s", b)
}
