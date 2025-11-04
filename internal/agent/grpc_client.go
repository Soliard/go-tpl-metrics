package agent

import (
	"context"
	"fmt"
	"time"

	metricspb "github.com/Soliard/go-tpl-metrics/internal/proto"
	"github.com/Soliard/go-tpl-metrics/models"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func (a *Agent) reportMetricsBatchGRPC(metrics []*models.Metrics) error {
	if a.grpcServerHost == "" {
		return fmt.Errorf("grpc address not configured")
	}

	// подготовка полезной нагрузки (общая)
	comp, signB64, err := a.prepareJSONPayload(metrics)
	if err != nil {
		return err
	}

	// ensure shared connection and client
	if err := a.ensureGRPCConn(context.Background()); err != nil {
		return err
	}
	client := a.grpcClient

	// metadata: signature and x-real-ip
	md := metadata.New(nil)
	if a.agentIP != "" {
		md.Set("x-real-ip", a.agentIP)
	}
	if signB64 != "" {
		md.Set("HashSHA256", signB64)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err = client.Updates(ctx, &metricspb.BatchBytes{Payload: comp})
	if err != nil {
		a.Logger.Error("grpc Updates failed", zap.Error(err))
		return err
	}
	return nil
}
