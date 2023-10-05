package main

import (
	"fmt"
	"path/filepath"

	"yambol/config"
	"yambol/pkg/api/rest"
	"yambol/pkg/broker"
	"yambol/pkg/util"
)

func main() {
	fmt.Println("Running Yambol...")
	cfg, err := config.FromFile("config.json")
	if err != nil {
		fmt.Println("failed to load config file: ", err)
		return
	}

	broker.SetDefaultMinLen(cfg.Broker.DefaultMinLength)
	broker.SetDefaultMaxLen(cfg.Broker.DefaultMaxLength)
	broker.SetDefaultMaxSizeBytes(cfg.Broker.DefaultMaxSizeBytes)
	broker.SetDefaultTTL(util.Seconds(cfg.Broker.DefaultTTL))

	b := broker.New()
	for qName, qCfg := range cfg.Broker.Queues {
		if err = b.AddQueue(qName, broker.QueueOptions{
			MinLen:       qCfg.MinLength,
			MaxLen:       qCfg.MaxLength,
			MaxSizeBytes: qCfg.MaxSizeBytes,
			DefaultTTL:   util.Seconds(qCfg.TTL),
		}); err != nil {
			fmt.Println("failed to add queue: ", err)
			return
		}
	}
	server := rest.NewYambolHTTPServer(b, nil)
	certPath, err := filepath.Abs(cfg.API.Certificate)
	if err != nil {
		fmt.Println("failed to get TLS certificate path: ", err)
		return
	}
	keyPath, err := filepath.Abs(cfg.API.Key)
	if err != nil {
		fmt.Println("failed to get TLS key path: ", err)
		return
	}
	if cfg.API.HTTP.Port > 0 {
		if err = server.ListenAndServe(cfg.API.HTTP.Port, certPath, keyPath); err != nil {
			fmt.Println("failed to start http server: ", err)
			return
		}
	}
}
