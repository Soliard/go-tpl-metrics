package netutil

import (
    "context"
    "net"
    "net/http"

    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/peer"
)

// ExtractIPFromHTTPRequest извлекает IP из заголовка X-Real-IP
func ExtractIPFromHTTPRequest(r *http.Request) net.IP {
    realIP := r.Header.Get("X-Real-IP")
    if realIP == "" {
        return nil
    }
    return net.ParseIP(realIP)
}

// ExtractIPFromGRPCContext извлекает IP из metadata x-real-ip или из peer.Addr
func ExtractIPFromGRPCContext(ctx context.Context) net.IP {
    if md, ok := metadata.FromIncomingContext(ctx); ok {
        vals := md.Get("x-real-ip")
        if len(vals) > 0 {
            if ip := net.ParseIP(vals[0]); ip != nil {
                return ip
            }
        }
    }
    if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
        host, _, _ := net.SplitHostPort(p.Addr.String())
        if host == "" {
            host = p.Addr.String()
        }
        if ip := net.ParseIP(host); ip != nil {
            return ip
        }
    }
    return nil
}

// ParseCIDR безопасно парсит CIDR и возвращает *net.IPNet (или nil)
func ParseCIDR(cidr string) *net.IPNet {
    if cidr == "" {
        return nil
    }
    _, ipnet, err := net.ParseCIDR(cidr)
    if err != nil {
        return nil
    }
    return ipnet
}


