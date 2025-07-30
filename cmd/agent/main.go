package main

import (
	"context"
	"log"
	"os"

	"github.com/Soliard/go-tpl-metrics/cmd/agent/config"
	"github.com/Soliard/go-tpl-metrics/internal/agent"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"go.uber.org/zap"
)

func main() {
	defer os.Stdout.Sync()
	config, err := config.New()
	if err != nil {
		log.Fatalf("cannot create config for agent %v", err)
	}

	logger, err := logger.New(config.LogLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	logger.Info("agent config: ", zap.Any("config", config))

	agent := agent.New(config, logger)
	logger.Info("agent works with service on", zap.String("serverhost", config.ServerHost))
	ctx := context.Background()
	agent.Run(ctx)
}
