package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/Soliard/go-tpl-metrics/internal/compressor"
	"github.com/Soliard/go-tpl-metrics/internal/crypto"
	serverpb "github.com/Soliard/go-tpl-metrics/internal/proto/server"
	"github.com/Soliard/go-tpl-metrics/internal/signer"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/Soliard/go-tpl-metrics/models"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// grpcServer реализует grpcapi.MetricsServer поверх MetricsService
type grpcServer struct {
	svc *MetricsService
	serverpb.UnimplementedMetricsServer
}

// Updates принимает зашифрованный/сжатый JSON пакета метрик, проверяет подпись и обновляет хранилище
func (g *grpcServer) Updates(ctx context.Context, req *serverpb.BatchBytes) (*emptypb.Empty, error) {
	logger := g.svc.Logger

	// trusted subnet check via metadata or peer
	if g.svc.trustedSubnet != "" {
		if ok := checkTrustedSubnet(ctx, g.svc.trustedSubnet, logger); !ok {
			return nil, fmt.Errorf("forbidden: ip not in trusted subnet")
		}
	}

	payload := req.Payload

	// Verify signature if key exists
	if signer.SignKeyExists(g.svc.signKey) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			vals := md.Get("HashSHA256")
			if len(vals) == 0 {
				return nil, fmt.Errorf("missing signature")
			}
			sig, err := signer.DecodeSign(vals[0])
			if err != nil {
				return nil, fmt.Errorf("invalid signature encoding: %w", err)
			}
			if !signer.Verify(payload, sig, g.svc.signKey) {
				return nil, fmt.Errorf("signature verification failed")
			}
		}
	}

	// Decrypt if private key present
	if g.svc.privateKey != nil {
		dec, err := crypto.DecryptHybrid(payload, g.svc.privateKey)
		if err != nil {
			return nil, fmt.Errorf("decrypt failed: %w", err)
		}
		payload = dec
	}

	// Decompress gzip
	buf, err := compressor.UncompressData(payload)
	if err != nil {
		return nil, fmt.Errorf("decompress failed: %w", err)
	}

	var metrics []*models.Metrics
	if err := json.Unmarshal(buf, &metrics); err != nil {
		logger.Warn("cant decode body to metric slice", zap.Error(err))
		return nil, fmt.Errorf("cant decode body to metrics: %w", err)
	}

	if err := g.svc.UpdateMetrics(ctx, metrics); err != nil {
		if errors.Is(err, store.ErrInvalidMetricReceived) {
			return nil, err
		}
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// NewGRPCServer создает gRPC сервер и регистрирует сервис
func NewGRPCServer(svc *MetricsService, opts ...grpc.ServerOption) *grpc.Server {
	gs := grpc.NewServer(opts...)
	serverpb.RegisterMetricsServer(gs, &grpcServer{svc: svc})
	return gs
}

// checkTrustedSubnet проверяет IP из metadata x-real-ip или из peer.Addr на вхождение в CIDR
func checkTrustedSubnet(ctx context.Context, cidr string, logger *zap.Logger) bool {
	if cidr == "" {
		return true
	}
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil || ipnet == nil {
		logger.Warn("invalid trusted_subnet CIDR", zap.String("cidr", cidr), zap.Error(err))
		return true // пропускаем, как и HTTP middleware
	}

	var ipStr string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		vals := md.Get("x-real-ip")
		if len(vals) > 0 {
			ipStr = vals[0]
		}
	}
	if ipStr == "" {
		if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
			host, _, _ := net.SplitHostPort(p.Addr.String())
			if host == "" {
				host = p.Addr.String()
			}
			ipStr = host
		}
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return ipnet.Contains(ip)
}
