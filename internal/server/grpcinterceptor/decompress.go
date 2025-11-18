package grpcinterceptor

import (
	"context"

	"github.com/Soliard/go-tpl-metrics/internal/compressor"
	metricspb "github.com/Soliard/go-tpl-metrics/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func DecompressGzipInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		batch, ok := req.(*metricspb.BatchBytes)
		if !ok {
			return handler(ctx, req)
		}
		buf, err := compressor.UncompressData(batch.Payload)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "decompress failed")
		}
		batch.Payload = buf
		return handler(ctx, req)
	}
}
