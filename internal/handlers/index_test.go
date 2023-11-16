package handlers

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexHandler(t *testing.T) {
	type want struct {
		response string
		code     int
	}
	tests := []struct {
		name   string
		want   want
		fields map[string]entity.Alert
	}{
		{
			name: "success test case",
			want: want{
				response: "<!DOCTYPE html>\n<html lang=\"ru\">\n<head>\n    <meta charset=\"UTF-8\">\n" +
					"    <title>Some awesome metrics</title>\n</head>\n<section>\n    <ul>\n        \n    </ul>\n</section>\n</html>",
				code: http.StatusOK,
			},
			fields: map[string]entity.Alert{},
		},
		{
			name: "success test case with fields",
			want: want{
				response: "<!DOCTYPE html>\n<html lang=\"ru\">\n<head>\n    <meta charset=\"UTF-8\">\n" +
					"    <title>Some awesome metrics</title>\n</head>\n<section>\n    <ul>\n        \n " +
					"       <li>alert1: 100</li>\n        \n        <li>alert2: 2.33434</li>\n " +
					"       \n    </ul>\n</section>\n</html>",
				code: http.StatusOK,
			},
			fields: map[string]entity.Alert{
				"alert1": {
					Type:  "counter",
					Name:  "alert1",
					Value: int64(100),
				},
				"alert2": {
					Type:  "gauge",
					Name:  "alert2",
					Value: 2.33434,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strg := storage.NewInMemoryStorage()
			for name, alert := range tt.fields {
				strg.Save(name, alert)
			}

			request, err := http.NewRequest(http.MethodGet, "localhost:8080/", nil)
			require.NoError(t, err)
			writer := httptest.NewRecorder()
			handlerToTest := IndexHandler(strg, "../../static")
			handlerToTest.ServeHTTP(writer, request)

			res := writer.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Println(err)
				}
			}()
			responseBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.response, string(responseBody))
		})
	}
}
