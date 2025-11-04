package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	metricspb "github.com/Soliard/go-tpl-metrics/internal/proto"
	"github.com/Soliard/go-tpl-metrics/internal/server/grpcinterceptor"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/Soliard/go-tpl-metrics/models"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// grpcServer реализует grpcapi.MetricsServer поверх MetricsService
type grpcServer struct {
	svc *MetricsService
	metricspb.UnimplementedMetricsServer
}

// Updates принимает зашифрованный/сжатый JSON пакета метрик, проверяет подпись и обновляет хранилище
func (g *grpcServer) Updates(ctx context.Context, req *metricspb.BatchBytes) (*emptypb.Empty, error) {
	var metrics []*models.Metrics
	if err := json.Unmarshal(req.Payload, &metrics); err != nil {
		g.svc.Logger.Warn("cant decode body to metric slice", zap.Error(err))
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
	// chain: trusted subnet -> verify signature -> decrypt -> decompress
	chain := grpc.ChainUnaryInterceptor(
		grpcinterceptor.TrustedSubnetInterceptor(svc.trustedSubnet, svc.Logger),
		grpcinterceptor.VerifySignatureInterceptor(svc.signKey, svc.Logger),
		grpcinterceptor.DecryptInterceptor(svc.privateKey, svc.Logger),
		grpcinterceptor.DecompressGzipInterceptor(svc.Logger),
	)
	opts = append(opts, chain)
	gs := grpc.NewServer(opts...)
	metricspb.RegisterMetricsServer(gs, &grpcServer{svc: svc})
	return gs
}
