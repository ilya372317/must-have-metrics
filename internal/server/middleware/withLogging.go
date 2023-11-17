package middleware

import (
	"net/http"
	"time"

	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
)

var sLogger = logger.Get()

type (
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}

	responseData struct {
		status int
		size   int
	}
)

func WithLogging() Middleware {
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}
			h.ServeHTTP(&lw, r)

			duration := time.Since(start)

			sLogger.Infoln(
				"uri", r.RequestURI,
				"method", r.Method,
				"duration", duration,
				"status", responseData.status,
				"size", responseData.size,
			)
		}
		return http.HandlerFunc(logFn)
	}
}
