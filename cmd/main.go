package main

import (
	"fmt"
	"yambol/config"
	"yambol/pkg/broker"
	"yambol/pkg/transport/httpx/rest"
	"yambol/pkg/util"
)

func main() {
	fmt.Println("Running Yambol...")
	cfg, err := config.FromFile()
	if err != nil {
		fmt.Println("failed to load config file: ", err)
		return
	}
	config.Init(*cfg)

	broker.SetDefaultMinLen(cfg.Broker.DefaultMinLength)
	broker.SetDefaultMaxLen(cfg.Broker.DefaultMaxLength)
	broker.SetDefaultMaxSizeBytes(cfg.Broker.DefaultMaxSizeBytes)
	broker.SetDefaultTTL(util.Seconds(cfg.Broker.DefaultTTL))

	b, err := broker.New()
	if err != nil {
		fmt.Println("failed to create broker:", err)
		return
	}
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

	//s := grpcx.NewYambolGRPCServer(b)
	//go func() {
	//	err = s.Serve("localhost", cfg.API.GRPC.Port)
	//	if err != nil {
	//		fmt.Println("grpc server failed: ", err)
	//	}
	//}()

	//c, err := grpcx.NewDefaultInsecureClient("localhost", 8081)
	//if err != nil {
	//	fmt.Println("failed to create client: ", err)
	//	return
	//}
	//m, err := c.Home(context.Background())
	//if err != nil {
	//	fmt.Println("failed to get home: ", err)
	//	return
	//}
	//fmt.Println(m.Uptime, m.Version)

	server := rest.NewYambolHTTPServer(b, nil)
	//certPath, err := filepath.Abs(cfg.API.Certificate)
	//if err != nil {
	//	fmt.Println("failed to get TLS certificate path: ", err)
	//	return
	//}
	//keyPath, err := filepath.Abs(cfg.API.Key)
	//if err != nil {
	//	fmt.Println("failed to get TLS key path: ", err)
	//	return
	//}
	if cfg.API.HTTP.Port > 0 {
		//if err = server.ListenAndServe(cfg.API.HTTP.Port, certPath, keyPath); err != nil {
		//	fmt.Println("failed to start http server: ", err)
		//	return
		//}
		if err = server.ListenAndServeInsecure(cfg.API.HTTP.Port); err != nil {
			fmt.Println("failed to start http server: ", err)
			return
		}
	}
}
