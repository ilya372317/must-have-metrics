package middleware

import (
	"net"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/config"
)

func WithTrustedSubnet(serverConfig *config.ServerConfig) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.Header.Get("X-Real-IP")
			_, trustepIPNet, err := net.ParseCIDR(serverConfig.TrustedSubnet)
			if err != nil {
				http.Error(w, "invalid server trusted subnet configuration", http.StatusInternalServerError)
				return
			}
			ip := net.ParseIP(clientIP)
			if ip == nil {
				http.Error(w, "invalid X-Real-IP header given", http.StatusForbidden)
				return
			}
			isNotTrusted := !trustepIPNet.Contains(ip)
			if isNotTrusted {
				http.Error(w, "client ip address not in trusted subnet", http.StatusForbidden)
				return
			}
			handler.ServeHTTP(w, r)
		})
	}
}
