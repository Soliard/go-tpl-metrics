package main

import (
	"fmt"

	"github.com/Soliard/go-tpl-metrics/cmd/agent/config"
	"github.com/Soliard/go-tpl-metrics/internal/agent"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
)

func main() {
	if err := logger.InitLogger("agent"); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.LogInfo("agent", "Starting agent...")

	config := config.New()
	logger.LogConfig("agent", config)

	agent := agent.New(config)
	logger.LogInfo("agent", fmt.Sprintf("Agent works with service on %s", config.ServerHost))
	agent.Run()
}
