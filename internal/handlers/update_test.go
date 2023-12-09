package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var serverConfig = &config.ServerConfig{
	Host:          "localhost:8080",
	FilePath:      "/tmp/metrics.json",
	Restore:       true,
	StoreInterval: 300,
}

func TestUpdateHandler(t *testing.T) {
	type testAlert struct {
		Type       string
		Name       string
		FloatValue float64
		IntValue   int64
	}
	type query struct {
		typ   string
		name  string
		value string
	}
	type want struct {
		alert testAlert
		code  int
	}

	tests := []struct {
		name   string
		query  query
		fields map[string]testAlert
		want   want
	}{
		{
			name: "success gauge with empty storage case",
			query: query{
				typ:   "gauge",
				name:  "alert",
				value: "1.1",
			},
			want: want{
				alert: testAlert{
					Type:       "gauge",
					Name:       "alert",
					FloatValue: 1.1,
				},
				code: http.StatusOK,
			},
			fields: map[string]testAlert{},
		},
		{
			name: "success gauge with not empty storage case",
			query: query{
				typ:   "gauge",
				name:  "alert",
				value: "1.2",
			},
			want: want{
				alert: testAlert{
					Type:       "gauge",
					Name:       "alert",
					FloatValue: 1.2,
				},
				code: http.StatusOK,
			},
			fields: map[string]testAlert{
				"alert": {
					Type:       "gauge",
					Name:       "alert",
					FloatValue: 1.1,
				},
			},
		},
		{
			name: "success counter with empty storage case",
			query: query{
				typ:   "counter",
				name:  "alert",
				value: "10",
			},
			want: want{
				alert: testAlert{
					Type:     "counter",
					Name:     "alert",
					IntValue: int64(10),
				},
				code: http.StatusOK,
			},
			fields: map[string]testAlert{},
		},
		{
			name: "success counter with not empty storage case",
			query: query{
				typ:   "counter",
				name:  "alert",
				value: "10",
			},
			want: want{
				alert: testAlert{
					Type:     "counter",
					Name:     "alert",
					IntValue: int64(20),
				},
				code: http.StatusOK,
			},
			fields: map[string]testAlert{
				"alert": {
					Type:     "counter",
					Name:     "alert",
					IntValue: int64(10),
				},
			},
		},
		{
			name: "negative gauge invalid value case",
			query: query{
				typ:   "gauge",
				name:  "alert",
				value: "invalid value",
			},
			want: want{
				alert: testAlert{},
				code:  http.StatusBadRequest,
			},
			fields: map[string]testAlert{},
		},
		{
			name: "negative counter invalid case value",
			query: query{
				typ:   "counter",
				name:  "alert",
				value: "invalid value",
			},
			want: want{
				alert: testAlert{},
				code:  http.StatusBadRequest,
			},
			fields: map[string]testAlert{},
		},
		{
			name: "replace gauge type to counter",
			query: query{
				typ:   "counter",
				name:  "alert",
				value: "1",
			},
			fields: map[string]testAlert{
				"alert": {
					Type:       "gauge",
					Name:       "alert",
					FloatValue: 1.32,
				},
			},
			want: want{
				alert: testAlert{
					Type:     "counter",
					Name:     "alert",
					IntValue: int64(1),
				},
				code: http.StatusOK,
			},
		},
		{
			name: "replace from counter to gauge type",
			query: query{
				typ:   "gauge",
				name:  "alert",
				value: "1.15",
			},
			fields: map[string]testAlert{
				"alert": {
					Type:     "counter",
					Name:     "alert",
					IntValue: int64(1),
				},
			},
			want: want{
				alert: testAlert{
					Type:       "gauge",
					Name:       "alert",
					FloatValue: 1.15,
				},
				code: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", tt.query.typ)
			rctx.URLParams.Add("name", tt.query.name)
			rctx.URLParams.Add("value", tt.query.value)
			request, err := http.NewRequest(
				http.MethodPost,
				"http://localhost:8080/update/{type}/{name}/{value}",
				nil,
			)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
			require.NoError(t, err)
			writer := httptest.NewRecorder()
			repo := storage.NewInMemoryStorage()
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
				err = repo.Save(context.Background(), name, alert)
				require.NoError(t, err)
			}

			handler := UpdateHandler(repo, serverConfig)
			handler(writer, request)
			res := writer.Result()
			defer res.Body.Close() //nolint //conflicts with practicum static tests

			assert.Equal(t, tt.want.code, res.StatusCode)
			if res.StatusCode >= 400 {
				return
			}

			addedAlert, err := repo.Get(context.Background(), tt.query.name)
			assert.NoError(t, err)

			wantAlert := entity.Alert{
				Type: tt.want.alert.Type,
				Name: tt.want.alert.Name,
			}
			if tt.want.alert.FloatValue != 0 {
				floatValue := tt.want.alert.FloatValue
				wantAlert.FloatValue = &floatValue
			}
			if tt.want.alert.IntValue != 0 {
				intValue := tt.want.alert.IntValue
				wantAlert.IntValue = &intValue
			}
			assert.Equal(t, addedAlert, wantAlert)
		})
	}
}
