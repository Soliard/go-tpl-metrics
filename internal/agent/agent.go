// Package agent предоставляет клиент для сбора и отправки метрик на сервер.
// Собирает системные метрики (память, CPU) и отправляет их на сервер с настраиваемой периодичностью.
package agent

import (
	"crypto/rsa"
	"strings"
	"time"

	"github.com/Soliard/go-tpl-metrics/cmd/agent/config"
	"github.com/Soliard/go-tpl-metrics/internal/crypto"
	"github.com/Soliard/go-tpl-metrics/internal/signer"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// Agent представляет клиент для сбора и отправки метрик.
// Собирает метрики с заданными интервалами и отправляет их на сервер.
type Agent struct {
	serverHostURL    string
	httpClient       *resty.Client
	Logger           *zap.Logger
	pollInterval     time.Duration
	reportInterval   time.Duration
	signKey          []byte
	requestRateLimit int
	publicKey        *rsa.PublicKey
}

// New создает новый экземпляр агента с указанной конфигурацией.
// Настраивает HTTP клиент с повторными попытками и нормализует URL сервера.
func New(config *config.Config, logger *zap.Logger) *Agent {
	client := resty.New().
		SetRetryCount(3).
		SetRetryMaxWaitTime(2)

	// Загружаем публичный ключ для шифрования
	var publicKey *rsa.PublicKey
	if config.CryptoKey != "" {
		var err error
		publicKey, err = crypto.LoadPublicKey(config.CryptoKey)
		if err != nil {
			logger.Fatal("failed to load public key", zap.Error(err))
		}
		logger.Info("public key loaded successfully for encryption")
	}

	return &Agent{
		serverHostURL:    normalizeServerURL(config.ServerHost),
		httpClient:       client,
		Logger:           logger,
		pollInterval:     time.Second * time.Duration(config.PollIntervalSeconds),
		reportInterval:   time.Second * time.Duration(config.ReportIntervalSeconds),
		signKey:          []byte(config.SignKey),
		requestRateLimit: config.RequestsLimit,
		publicKey:        publicKey,
	}
}

// normalizeServerURL добавляет протокол http:// к URL если он не указан
func normalizeServerURL(url string) string {
	if strings.HasPrefix(url, "http") {
		return url
	}
	return "http://" + url
}

// hasSignKey проверяет, настроен ли ключ для подписи данных
func (a *Agent) hasSignKey() bool {
	return signer.SignKeyExists(a.signKey)
}

// hasCryptoKey проверяет, настроен ли ключ для шифрования данных
func (a *Agent) hasCryptoKey() bool {
	return a.publicKey != nil
}
