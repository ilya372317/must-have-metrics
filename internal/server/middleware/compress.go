package middleware

import (
	"net/http"
	"strings"

	"github.com/ilya372317/must-have-metrics/internal/utils/compress"
	"github.com/ilya372317/must-have-metrics/internal/utils/logger"
)

const (
	contentEncodingHeader = "Content-Encoding"
	gzipEncoding          = "gzip"
)

var compressLogger = logger.Get()

func Compressed() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			acceptGzip := strings.Contains(acceptEncoding, gzipEncoding)
			if acceptGzip {
				cw := compress.NewWriter(w)
				ow = cw
				w.Header().Set(contentEncodingHeader, gzipEncoding)
				defer func() {
					_ = cw.Close()
				}()
			}

			contentEncoding := r.Header.Get(contentEncodingHeader)
			contentCompressed := strings.Contains(contentEncoding, gzipEncoding)
			if contentCompressed {
				cr, err := compress.NewReader(r.Body)
				if err != nil {
					http.Error(w, "failed create gzip compressor", http.StatusInternalServerError)
					compressLogger.Errorf("something went wrong with gzip compressor: %v", err)
					return
				}
				r.Body = cr
				defer func() {
					_ = cr.Close()
				}()
			}

			h.ServeHTTP(ow, r)
		})
	}
}
