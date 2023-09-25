package main

import (
	"fmt"
	"yambol/config"
	"yambol/pkg/broker"
)

func testingBroker(b *broker.MessageBroker) {
	fmt.Println("Testing broker...")
	err := b.Publish(`{'message': 'Hello, World!'}`, "default", "noexistent")
	if err != nil {
		fmt.Println("broker publish error", err)
	}
	msg, err := b.Receive("default")
	if err != nil {
		fmt.Println("broker receive error", err)
	}
	fmt.Println("message received:", msg)
}

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

	b := broker.NewMessageBroker()
	for qName, qCfg := range cfg.Broker.Queues {
		b.AddQueue(qName, qCfg.MinLength, qCfg.MaxLength, qCfg.MaxSizeBytes)
	}
	testingBroker(b)

}
