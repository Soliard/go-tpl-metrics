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
	ServerHost            string
	PollIntervalSeconds   int
	ReportIntervalSeconds int
}

func NewAgent(config Config) *Agent {
	return &Agent{
		serverHostURL:  normalizeServerURL(config.ServerHost),
		collector:      NewStatsCollector(),
		httpClient:     resty.New(),
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
