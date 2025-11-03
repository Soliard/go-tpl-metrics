package grpcinterceptor

import (
    "context"

    serverpb "github.com/Soliard/go-tpl-metrics/internal/proto/server"
    "github.com/Soliard/go-tpl-metrics/internal/signer"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/codes"
)

func VerifySignatureInterceptor(signKey []byte, logger *zap.Logger) grpc.UnaryServerInterceptor {
    if !signer.SignKeyExists(signKey) {
        return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
            return handler(ctx, req)
        }
    }
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        batch, ok := req.(*serverpb.BatchBytes)
        if !ok {
            return handler(ctx, req)
        }
        md, _ := metadata.FromIncomingContext(ctx)
        vals := md.Get("HashSHA256")
        if len(vals) == 0 {
            return nil, status.Error(codes.InvalidArgument, "missing signature")
        }
        sig, err := signer.DecodeSign(vals[0])
        if err != nil || !signer.Verify(batch.Payload, sig, signKey) {
            return nil, status.Error(codes.PermissionDenied, "signature verification failed")
        }
        return handler(ctx, req)
    }
}


