package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Soliard/go-tpl-metrics/internal/compressor"
	"github.com/Soliard/go-tpl-metrics/internal/crypto"
	agentpb "github.com/Soliard/go-tpl-metrics/internal/proto/agent"
	"github.com/Soliard/go-tpl-metrics/internal/signer"
	"github.com/Soliard/go-tpl-metrics/models"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func (a *Agent) reportMetricsBatchGRPC(metrics []*models.Metrics) error {
	if a.grpcServerHost == "" {
		return fmt.Errorf("grpc address not configured")
	}

	// marshal -> encrypt (if any) -> gzip -> sign over compressed bytes
	buf, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("cant marshal metrics: %v", err)
	}
	if a.hasCryptoKey() {
		enc, err := crypto.EncryptHybrid(buf, a.publicKey)
		if err != nil {
			return fmt.Errorf("cant encrypt data: %v", err)
		}
		buf = enc
		a.Logger.Info("metrics encrypted successfully")
	}
	comp, err := compressor.CompressData(buf)
	if err != nil {
		return fmt.Errorf("cant compress data: %v", err)
	}

	// dial per call (простота). Можно оптимизировать позже до реюза соединения
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, a.grpcServerHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := newGRPCClient(conn)

	// metadata: signature and x-real-ip
	md := metadata.New(nil)
	if a.agentIP != "" {
		md.Set("x-real-ip", a.agentIP)
	}
	if a.hasSignKey() {
		sig := signer.Sign(comp, a.signKey)
		md.Set("HashSHA256", signer.EncodeSign(sig))
	}
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err = client.Updates(ctx, &agentpb.BatchBytes{Payload: comp})
	if err != nil {
		a.Logger.Error("grpc Updates failed", zap.Error(err))
		return err
	}
	return nil
}

type grpcClient interface {
	Updates(ctx context.Context, in *agentpb.BatchBytes, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

func newGRPCClient(cc grpc.ClientConnInterface) grpcClient {
	return &metricsClient{cc}
}

type metricsClient struct {
	cc grpc.ClientConnInterface
}

func (c *metricsClient) Updates(ctx context.Context, in *agentpb.BatchBytes, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/metrics.Metrics/Updates", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
