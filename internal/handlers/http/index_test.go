package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/server/service"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestAlert struct {
	Type       string
	Name       string
	FloatValue float64
	IntValue   int64
}

func TestIndexHandler(t *testing.T) {
	type want struct {
		response string
		code     int
	}
	tests := []struct {
		fields map[string]TestAlert
		name   string
		want   want
	}{
		{
			name: "success test case",
			want: want{
				response: "<!DOCTYPE html>\n<html lang=\"ru\">\n<head>\n    <meta charset=\"UTF-8\">\n" +
					"    <title>Some awesome metrics</title>\n</head>\n<section>\n    <ul>\n        \n    </ul>\n</section>\n</html>",
				code: http.StatusOK,
			},
			fields: map[string]TestAlert{},
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
			fields: map[string]TestAlert{
				"alert1": {
					Type:     "counter",
					Name:     "alert1",
					IntValue: int64(100),
				},
				"alert2": {
					Type:       "gauge",
					Name:       "alert2",
					FloatValue: 2.33434,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strg := storage.NewInMemoryStorage()
			for name, tAlert := range tt.fields {
				alert := entity.Alert{
					Type: tAlert.Type,
					Name: tAlert.Name,
				}
				if tAlert.FloatValue != 0 {
					floatValue := tAlert.FloatValue
					alert.FloatValue = &floatValue
				}
				if tAlert.IntValue != 0 {
					intValue := tAlert.IntValue
					alert.IntValue = &intValue
				}
				err := strg.Save(context.Background(), name, alert)
				require.NoError(t, err)
			}

			request, err := http.NewRequest(http.MethodGet, "localhost:8080/", nil)
			require.NoError(t, err)
			writer := httptest.NewRecorder()
			serv := service.NewMetricsService(strg, serverConfig)
			handlerToTest := IndexHandler(serv)
			handlerToTest.ServeHTTP(writer, request)

			res := writer.Result()
			defer res.Body.Close() //nolint //conflicts with practicum static tests
			responseBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.response, string(responseBody))
		})
	}
}
