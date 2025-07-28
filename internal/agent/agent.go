package agent

import (
	"strings"
	"time"

	"github.com/Soliard/go-tpl-metrics/cmd/agent/config"
	"github.com/Soliard/go-tpl-metrics/internal/signer"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type Agent struct {
	serverHostURL    string
	httpClient       *resty.Client
	Logger           *zap.Logger
	pollInterval     time.Duration
	reportInterval   time.Duration
	signKey          []byte
	requestRateLimit int
}

func New(config *config.Config, logger *zap.Logger) *Agent {
	client := resty.New().
		SetRetryCount(3).
		SetRetryMaxWaitTime(2)

	return &Agent{
		serverHostURL:    normalizeServerURL(config.ServerHost),
		httpClient:       client,
		Logger:           logger,
		pollInterval:     time.Second * time.Duration(config.PollIntervalSeconds),
		reportInterval:   time.Second * time.Duration(config.ReportIntervalSeconds),
		signKey:          []byte(config.SignKey),
		requestRateLimit: config.RequestsLimit,
	}
}

func normalizeServerURL(url string) string {
	if strings.HasPrefix(url, "http") {
		return url
	}
	return "http://" + url
}

func (a *Agent) hasSignKey() bool {
	return signer.SignKeyExists(a.signKey)
}
