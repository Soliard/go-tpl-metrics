package main

import (
	"fmt"

	"github.com/Soliard/go-tpl-metrics/cmd/agent/config"
	"github.com/Soliard/go-tpl-metrics/internal/agent"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"go.uber.org/zap"
)

func main() {

	config, err := config.New()
	if err != nil {
		panic(fmt.Errorf("cannot create config for agent %w", err))
	}

	logger, err := logger.New(logger.ComponentServer, config.LogLevel)

	if err != nil {
		panic(fmt.Errorf("failed to initialize logger: %w", err))
	}
	logger.Log.Info("Agent config: ", zap.Any("config", config))

	agent := agent.New(config, logger)
	logger.Log.Info("Agent works with service on", zap.String("ServerHost", config.ServerHost))
	agent.Run()
}
