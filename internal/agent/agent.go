package agent

import (
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type Agent struct {
	serverHostURL  string
	collector      *StatsCollector
	httpClient     *resty.Client
	pollInterval   time.Duration
	reportInterval time.Duration
}

type Config struct {
	ServerHost            string `env:"ADDRESS"`
	PollIntervalSeconds   int    `env:"POLL_INTERVAL"`
	ReportIntervalSeconds int    `env:"REPORT_INTERVAL"`
}

func NewAgent(config Config) *Agent {
	return &Agent{
		serverHostURL:  normalizeServerURL(config.ServerHost),
		collector:      NewStatsCollector(),
		httpClient:     resty.New(),
		pollInterval:   time.Duration(config.PollIntervalSeconds),
		reportInterval: time.Duration(config.ReportIntervalSeconds),
	}
}

func normalizeServerURL(url string) string {
	if strings.HasPrefix(url, "http") {
		return url
	}
	return "http://" + url
}
