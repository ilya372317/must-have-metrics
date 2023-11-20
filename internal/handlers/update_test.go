package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateHandler(t *testing.T) {
	type query struct {
		typ   string
		name  string
		value string
	}
	type want struct {
		alert entity.Alert
		code  int
	}

	tests := []struct {
		name   string
		query  query
		fields map[string]entity.Alert
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
				alert: entity.Alert{
					Type:  "gauge",
					Name:  "alert",
					Value: 1.1,
				},
				code: http.StatusOK,
			},
			fields: map[string]entity.Alert{},
		},
		{
			name: "success gauge with not empty storage case",
			query: query{
				typ:   "gauge",
				name:  "alert",
				value: "1.2",
			},
			want: want{
				alert: entity.Alert{
					Type:  "gauge",
					Name:  "alert",
					Value: 1.2,
				},
				code: http.StatusOK,
			},
			fields: map[string]entity.Alert{
				"alert": {
					Type:  "gauge",
					Name:  "alert",
					Value: 1.1,
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
				alert: entity.Alert{
					Type:  "counter",
					Name:  "alert",
					Value: int64(10),
				},
				code: http.StatusOK,
			},
			fields: map[string]entity.Alert{},
		},
		{
			name: "success counter with not empty storage case",
			query: query{
				typ:   "counter",
				name:  "alert",
				value: "10",
			},
			want: want{
				alert: entity.Alert{
					Type:  "counter",
					Name:  "alert",
					Value: int64(20),
				},
				code: http.StatusOK,
			},
			fields: map[string]entity.Alert{
				"alert": {
					Type:  "counter",
					Name:  "alert",
					Value: int64(10),
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
				alert: entity.Alert{},
				code:  http.StatusBadRequest,
			},
			fields: map[string]entity.Alert{},
		},
		{
			name: "negative counter invalid case value",
			query: query{
				typ:   "counter",
				name:  "alert",
				value: "invalid value",
			},
			want: want{
				alert: entity.Alert{},
				code:  http.StatusBadRequest,
			},
			fields: map[string]entity.Alert{},
		},
		{
			name: "replace gauge type to counter",
			query: query{
				typ:   "counter",
				name:  "alert",
				value: "1",
			},
			fields: map[string]entity.Alert{
				"alert": {
					Type:  "gauge",
					Name:  "alert",
					Value: 1.32,
				},
			},
			want: want{
				alert: entity.Alert{
					Type:  "counter",
					Name:  "alert",
					Value: int64(1),
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
			fields: map[string]entity.Alert{
				"alert": {
					Type:  "counter",
					Name:  "alert",
					Value: 1,
				},
			},
			want: want{
				alert: entity.Alert{
					Type:  "gauge",
					Name:  "alert",
					Value: 1.15,
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
			for name, alert := range tt.fields {
				repo.Save(name, alert)
			}

			handler := UpdateHandler(repo)
			handler(writer, request)
			res := writer.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
			if res.StatusCode >= 400 {
				return
			}

			addedAlert, err := repo.Get(tt.query.name)
			assert.NoError(t, err)

			assert.Equal(t, addedAlert, tt.want.alert)
		})
	}
}
