package grpcinterceptor

import (
	"context"
	"net"

	"github.com/Soliard/go-tpl-metrics/internal/netutil"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TrustedSubnetInterceptor(cidr string, logger *zap.Logger) grpc.UnaryServerInterceptor {
	if cidr == "" {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil || ipnet == nil {
		logger.Warn("invalid trusted_subnet CIDR, interceptor bypassed", zap.String("cidr", cidr), zap.Error(err))
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ip := netutil.ExtractIPFromGRPCContext(ctx)
		if ip == nil || !ipnet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "ip not in trusted subnet")
		}
		return handler(ctx, req)
	}
}
