package main

import (
	"fmt"

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

	b := broker.NewMessageBroker()
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
	if err = server.ServeHTTP(8080); err != nil {
		fmt.Println("http server error: ", err)
	}
	//testingBroker(b)

}
