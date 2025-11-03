package grpcinterceptor

import (
    "context"

    serverpb "github.com/Soliard/go-tpl-metrics/internal/proto/server"
    "github.com/Soliard/go-tpl-metrics/internal/compressor"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/codes"
)

func DecompressGzipInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        batch, ok := req.(*serverpb.BatchBytes)
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


