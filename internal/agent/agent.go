// Package agent предоставляет клиент для сбора и отправки метрик на сервер.
// Собирает системные метрики (память, CPU) и отправляет их на сервер с настраиваемой периодичностью.
package agent

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/Soliard/go-tpl-metrics/internal/compressor"
	"github.com/Soliard/go-tpl-metrics/internal/config"
	"github.com/Soliard/go-tpl-metrics/internal/crypto"
	metricspb "github.com/Soliard/go-tpl-metrics/internal/proto"
	"github.com/Soliard/go-tpl-metrics/internal/signer"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Agent представляет клиент для сбора и отправки метрик.
// Собирает метрики с заданными интервалами и отправляет их на сервер.
type Agent struct {
	serverHostURL    string
	grpcServerHost   string
	httpClient       *resty.Client
	Logger           *zap.Logger
	pollInterval     time.Duration
	reportInterval   time.Duration
	signKey          []byte
	requestRateLimit int
	publicKey        *rsa.PublicKey
	agentIP          string
	// gRPC
	grpcConn   *grpc.ClientConn
	grpcClient metricspb.MetricsClient
	grpcOnce   sync.Once
}

// New создает новый экземпляр агента с указанной конфигурацией.
// Настраивает HTTP клиент с повторными попытками и нормализует URL сервера.
func New(config *config.AgentConfig, logger *zap.Logger) *Agent {
	// Инициализируем HTTP клиент только если HTTP адрес задан
	var client *resty.Client
	if config.ServerHost != "" {
		client = resty.New().
			SetRetryCount(3).
			SetRetryMaxWaitTime(2)
	}

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
		grpcServerHost:   config.GRPCServerHost,
		httpClient:       client,
		Logger:           logger,
		pollInterval:     time.Second * time.Duration(config.PollIntervalSeconds),
		reportInterval:   time.Second * time.Duration(config.ReportIntervalSeconds),
		signKey:          []byte(config.SignKey),
		requestRateLimit: config.RequestsLimit,
		publicKey:        publicKey,
		agentIP:          detectOutboundIP(),
	}
}

func (a *Agent) ensureGRPCConn(ctx context.Context) error {
	var err error
	a.grpcOnce.Do(func() {
		if a.grpcServerHost == "" {
			return
		}

		// Новый API с опциями
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}

		conn, connErr := grpc.NewClient(a.grpcServerHost, opts...)
		if connErr != nil {
			err = fmt.Errorf("failed to create gRPC client: %w", connErr)
			return
		}

		a.grpcConn = conn
		a.grpcClient = metricspb.NewMetricsClient(conn)
	})
	return err
}

func (a *Agent) closeGRPCConn() {
	if a.grpcConn != nil {
		_ = a.grpcConn.Close()
		a.grpcConn = nil
		a.grpcClient = nil
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

// detectOutboundIP определяет исходящий IP-адрес
// через UDP подключение к публичному серверу
func detectOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	localAddr := conn.LocalAddr()
	udpAddr, ok := localAddr.(*net.UDPAddr)
	if !ok || udpAddr.IP == nil {
		return ""
	}
	return udpAddr.IP.String()
}

// prepareJSONPayload подготавливает тело запроса:
// json.Marshal -> EncryptHybrid (если ключ есть) -> gzip.
// Возвращает сжатый буфер и подпись (по сжатому буферу) в base64, если есть signKey.
func (a *Agent) prepareJSONPayload(v any) (compressed []byte, signatureB64 string, err error) {
	buf, err := json.Marshal(v)
	if err != nil {
		return nil, "", fmt.Errorf("cant marshal payload: %v", err)
	}
	if a.hasCryptoKey() {
		enc, err := crypto.EncryptHybrid(buf, a.publicKey)
		if err != nil {
			return nil, "", fmt.Errorf("cant encrypt data: %v", err)
		}
		buf = enc
		a.Logger.Info("payload encrypted successfully")
	}
	comp, err := compressor.CompressData(buf)
	if err != nil {
		return nil, "", fmt.Errorf("cant compress data: %v", err)
	}
	var signB64 string
	if a.hasSignKey() {
		sig := signer.Sign(comp, a.signKey)
		signB64 = signer.EncodeSign(sig)
	}
	return comp, signB64, nil
}
