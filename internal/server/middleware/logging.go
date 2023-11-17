package middleware

import (
	"net/http"
	"time"

	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
)

var sLogger = logger.Get()

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.responseData.size = size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.responseData.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func WithLogging() Middleware {
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   rData,
			}
			h.ServeHTTP(&lw, r)

			duration := time.Since(start)

			sLogger.Infoln(
				"uri", r.RequestURI,
				"method", r.Method,
				"duration", duration,
				"status", rData.status,
				"size", rData.size,
			)
		}
		return http.HandlerFunc(logFn)
	}
}