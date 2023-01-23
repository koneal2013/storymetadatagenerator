package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func LogRequest(next http.Handler) http.Handler {
	logger := zap.L().Named("http_server")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Sugar().Info("incoming request", r)
		next.ServeHTTP(w, r)
	})
}
