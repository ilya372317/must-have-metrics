package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressed(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "simple success case",
			body: "test123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer([]byte{})
			gzipWriter := gzip.NewWriter(buffer)
			_, err := gzipWriter.Write([]byte(tt.body))
			require.NoError(t, err)
			err = gzipWriter.Close()
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodGet, "/test", buffer)
			r.Header.Set("Accept-Encoding", "gzip")
			r.Header.Set("Content-Encoding", "gzip")
			w := httptest.NewRecorder()
			middleware := Compressed()
			middleware(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				body, err := io.ReadAll(request.Body)
				require.NoError(t, err)
				_, err = responseWriter.Write(body)
				require.NoError(t, err)
			})).ServeHTTP(w, r)

			res := w.Result()
			defer func() {
				_ = res.Body.Close()
			}()

			contentEncoding := res.Header.Get("Content-Encoding")
			assert.Equal(t, "gzip", contentEncoding)
			gzipReader, err := newReader(res.Body)
			require.NoError(t, err)
			response, err := io.ReadAll(gzipReader)
			require.NoError(t, err)
			assert.Equal(t, string(response), tt.body)
		})
	}
}
