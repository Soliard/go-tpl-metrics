package grpcinterceptor

import (
    "context"
    "crypto/rsa"

    serverpb "github.com/Soliard/go-tpl-metrics/internal/proto/server"
    "github.com/Soliard/go-tpl-metrics/internal/crypto"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/codes"
)

func DecryptInterceptor(privateKey *rsa.PrivateKey, logger *zap.Logger) grpc.UnaryServerInterceptor {
    if privateKey == nil {
        return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
            return handler(ctx, req)
        }
    }
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        batch, ok := req.(*serverpb.BatchBytes)
        if !ok {
            return handler(ctx, req)
        }
        dec, err := crypto.DecryptHybrid(batch.Payload, privateKey)
        if err != nil {
            return nil, status.Error(codes.InvalidArgument, "decrypt failed")
        }
        batch.Payload = dec
        return handler(ctx, req)
    }
}


