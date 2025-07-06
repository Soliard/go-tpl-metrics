package main

import (
	"log"
	"os"

	"github.com/Soliard/go-tpl-metrics/cmd/agent/config"
	"github.com/Soliard/go-tpl-metrics/internal/agent"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"go.uber.org/zap"
)

func main() {

	config, err := config.New()
	if err != nil {
		log.Printf("cannot create config for agent %v", err)
		os.Stdout.Sync()
		os.Exit(1)
	}

	logger, err := logger.New(config.LogLevel)

	if err != nil {
		log.Printf("failed to initialize logger: %v", err)
		os.Stdout.Sync()
		os.Exit(1)
	}
	logger.Info("Agent config: ", zap.Any("config", config))

	agent := agent.New(config, logger)
	logger.Info("Agent works with service on", zap.String("ServerHost", config.ServerHost))
	os.Stdout.Sync()
	agent.Run()
}
