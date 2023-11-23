package middleware

import (
	"net/http"
	"strings"

	"github.com/ilya372317/must-have-metrics/internal/utils/compress"
	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
)

var compressLogger = logger.Get()

func Compressed() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			acceptGzip := strings.Contains(acceptEncoding, "gzip")
			if acceptGzip {
				cw := compress.NewWriter(w)
				ow = cw
				w.Header().Set("Content-Encoding", "gzip")
				defer cw.Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			contentCompressed := strings.Contains(contentEncoding, "gzip")
			if contentCompressed {

				cr, err := compress.NewReader(r.Body)
				if err != nil {
					http.Error(w, "failed create gzip compressor", http.StatusInternalServerError)
					compressLogger.Errorf("something went wrong with gzip compressor: %v", err)
					return
				}
				r.Body = cr
				defer cr.Close()
			}

			h.ServeHTTP(ow, r)
		})
	}
}
