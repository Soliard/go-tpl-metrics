package agent

import (
	"strings"
	"time"

	"github.com/Soliard/go-tpl-metrics/cmd/agent/config"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/go-resty/resty/v2"
)

type Agent struct {
	serverHostURL  string
	collector      *StatsCollector
	httpClient     *resty.Client
	Logger         *logger.Logger
	pollInterval   time.Duration
	reportInterval time.Duration
}

func New(config *config.Config, logger *logger.Logger) *Agent {
	return &Agent{
		serverHostURL:  normalizeServerURL(config.ServerHost),
		collector:      NewStatsCollector(),
		httpClient:     resty.New(),
		Logger:         logger,
		pollInterval:   time.Second * time.Duration(config.PollIntervalSeconds),
		reportInterval: time.Second * time.Duration(config.ReportIntervalSeconds),
	}
}

func normalizeServerURL(url string) string {
	if strings.HasPrefix(url, "http") {
		return url
	}
	return "http://" + url
}
