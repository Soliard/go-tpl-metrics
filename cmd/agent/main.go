package main

import (
	"fmt"

	"github.com/Soliard/go-tpl-metrics/cmd/agent/config"
	"github.com/Soliard/go-tpl-metrics/internal/agent"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
)

func main() {
	logger, err := logger.New(logger.ComponentServer)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Close()
	logger.Info("Starting agent...")

	config := config.New(logger)
	logger.Info("Agent config: ", config)

	agent := agent.New(config, logger)
	logger.Info("Agent works with service on ", config.ServerHost)
	agent.Run()
}
