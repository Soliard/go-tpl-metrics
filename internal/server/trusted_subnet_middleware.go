package server

import (
	"net"
	"net/http"

	"go.uber.org/zap"
)

func TrustedSubnetMiddleware(cidr string, logger *zap.Logger) func(http.Handler) http.Handler {
	// Предварительно парсим CIDR один раз
	var ipnet *net.IPNet
	if cidr != "" {
		if _, n, err := net.ParseCIDR(cidr); err == nil {
			ipnet = n
		} else {
			logger.Warn("invalid trusted_subnet CIDR, middleware will be bypassed", zap.String("cidr", cidr), zap.Error(err))
		}
	}

	return func(next http.Handler) http.Handler {
		// Если подсеть не настроена — пропускаем без проверок
		if cidr == "" || ipnet == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				http.Error(w, "forbidden: missing X-Real-IP", http.StatusForbidden)
				return
			}
			ip := net.ParseIP(realIP)
			if ip == nil {
				http.Error(w, "forbidden: invalid X-Real-IP", http.StatusForbidden)
				return
			}
			if !ipnet.Contains(ip) {
				http.Error(w, "forbidden: ip not in trusted subnet", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
