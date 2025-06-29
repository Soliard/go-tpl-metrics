package main

import (
	"log"

	"github.com/Soliard/go-tpl-metrics/cmd/agent/config"
	"github.com/Soliard/go-tpl-metrics/internal/agent"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"go.uber.org/zap"
)

func main() {

	config, err := config.New()
	if err != nil {
		log.Fatalf("cannot create config for agent %v", err)
	}

	logger, err := logger.New(config.LogLevel)

	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	logger.Info("Agent config: ", zap.Any("config", config))

	agent := agent.New(config, logger)
	logger.Info("Agent works with service on", zap.String("ServerHost", config.ServerHost))
	agent.Run()
}
