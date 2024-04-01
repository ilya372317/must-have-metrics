package v1

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowHandler(t *testing.T) {
	type testAlert struct {
		Type       string
		Name       string
		FloatValue float64
		IntValue   int64
	}
	type want struct {
		response string
		code     int
	}
	type args struct {
		typ  string
		name string
	}

	tests := []struct {
		fields map[string]testAlert
		name   string
		args   args
		want   want
	}{
		{
			name: "success simple test",
			want: want{
				response: "1",
				code:     http.StatusOK,
			},
			fields: map[string]testAlert{
				"alert": {
					Type:     "counter",
					Name:     "alert",
					IntValue: int64(1),
				},
			},
			args: args{
				typ:  "counter",
				name: "alert",
			},
		},
		{
			name: "not found test",
			want: want{
				response: "alert not found\n",
				code:     http.StatusNotFound,
			},
			fields: map[string]testAlert{},
			args: args{
				typ:  "counter",
				name: "alert",
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
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", tt.args.typ)
			rctx.URLParams.Add("name", tt.args.name)
			request, err := http.NewRequest(http.MethodGet, "localhost:8080/value/{type}/{name}", nil)
			require.NoError(t, err)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
			writer := httptest.NewRecorder()
			handler := ShowHandler(strg)
			handler.ServeHTTP(writer, request)

			res := writer.Result()
			defer res.Body.Close() //nolint //conflicts with practicum static tests
			responseBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.response, string(responseBody))
		})
	}
}
