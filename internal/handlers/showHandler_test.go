package handlers

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShowHandler(t *testing.T) {
	type want struct {
		response string
		code     int
	}
	type args struct {
		typ  string
		name string
	}

	tests := []struct {
		name   string
		want   want
		fields map[string]entity.Alert
		args   args
	}{
		{
			name: "success simple test",
			want: want{
				response: "name: alert, type: counter, value: 1",
				code:     http.StatusOK,
			},
			fields: map[string]entity.Alert{
				"alert": {
					Type:  "counter",
					Name:  "alert",
					Value: 1,
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
			fields: map[string]entity.Alert{},
			args: args{
				typ:  "counter",
				name: "alert",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			strg := storage.MakeInMemoryStorage()
			for name, alert := range tt.fields {
				strg.SaveAlert(name, alert)
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
			defer res.Body.Close()
			responseBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.response, string(responseBody))
		})
	}
}
