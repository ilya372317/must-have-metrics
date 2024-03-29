package router

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/ilya372317/must-have-metrics/internal/utils/compress"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cnfg = &config.ServerConfig{
	Host:          "localhost:8080",
	FilePath:      "/tmp/metrics.json",
	Restore:       true,
	StoreInterval: 300,
}

func TestAlertRouter(t *testing.T) {
	err := logger.Init()
	require.NoError(t, err)
	strg := storage.NewInMemoryStorage()
	ts := httptest.NewServer(AlertRouter(strg, cnfg))
	defer ts.Close()

	type testAlert struct {
		Type       string
		Name       string
		FloatValue float64
		IntValue   int64
	}

	type want struct {
		body   string
		status int
	}

	var testTable = []struct {
		fields      map[string]testAlert
		name        string
		url         string
		method      string
		requestBody string
		want        want
	}{
		{
			name: "index success case",
			fields: map[string]testAlert{
				"alert": {
					Type:     "counter",
					Name:     "alert",
					IntValue: int64(1),
				},
			},
			url: "/",
			want: want{
				status: http.StatusOK,
				body:   "",
			},
			method: http.MethodGet,
		},
		{
			name:   "show success case",
			url:    "/value/counter/alert",
			method: http.MethodGet,
			fields: map[string]testAlert{
				"alert": {
					Type:     "counter",
					Name:     "alert",
					IntValue: int64(1),
				},
			},
			want: want{
				status: http.StatusOK,
			},
		},
		{
			name:   "negative show case",
			url:    "/value/counter/alert1",
			method: http.MethodGet,
			fields: map[string]testAlert{},
			want: want{
				status: http.StatusNotFound,
			},
		},
		{
			name:   "success update case",
			url:    "/update/counter/alert/1",
			method: http.MethodPost,
			fields: nil,
			want: want{
				status: http.StatusOK,
			},
		},
		{
			name:   "update type is invalid case",
			url:    "/update/invalidType/alert/1",
			method: http.MethodPost,
			fields: nil,

			want: want{
				status: http.StatusBadRequest,
			},
		},
		{
			name:   "update value is invalid case",
			url:    "/update/gauge/name/invalidValue",
			method: http.MethodPost,
			fields: nil,
			want: want{
				status: http.StatusBadRequest,
			},
		},
		{
			name:   "update json gauge success case",
			url:    "/update",
			method: http.MethodPost,
			fields: nil,
			want: want{
				status: http.StatusOK,
				body:   "{\"id\":\"alert\",\"type\":\"gauge\",\"value\":363334.99574712414}",
			},
			requestBody: "{\"id\":\"alert\",\"type\":\"gauge\",\"value\":363334.99574712414}",
		},
		{
			name:   "update json counter success case",
			url:    "/update",
			method: http.MethodPost,
			fields: map[string]testAlert{
				"alert": {
					IntValue: int64(1),
					Type:     "counter",
					Name:     "alert",
				},
			},
			want: want{
				status: http.StatusOK,
				body:   "{\"id\":\"alert\",\"type\":\"counter\",\"delta\":2}",
			},
			requestBody: "{\"id\":\"alert\",\"type\":\"counter\",\"delta\":1}",
		},
		{
			name:   "update json negative case give delta in gauge type",
			url:    "/update",
			method: http.MethodPost,
			fields: nil,
			want: want{
				status: http.StatusBadRequest,
				body:   "",
			},
			requestBody: "{\"id\":\"alert\",\"type\":\"gauge\",\"delta\":1}",
		},
		{
			name:   "update json negative case give value on counter type",
			url:    "/update",
			method: http.MethodPost,
			fields: nil,
			want: want{
				status: http.StatusBadRequest,
				body:   "",
			},
			requestBody: "{\"id\":\"alert\",\"type\":\"counter\",\"value\":1}",
		},
		{
			name:   "update json negative case give empty body",
			url:    "/update",
			method: http.MethodPost,
			fields: nil,
			want: want{
				status: http.StatusBadRequest,
				body:   "",
			},
			requestBody: "",
		},
		{
			name:   "update json negative case missing type in request body",
			url:    "/update",
			method: http.MethodPost,
			fields: nil,
			want: want{
				status: http.StatusBadRequest,
				body:   "",
			},
			requestBody: "{\"id\":\"alert\",\"delta\":1}",
		},
		{
			name:   "update json negative case missing value in request body",
			url:    "/update",
			method: http.MethodPost,
			fields: nil,
			want: want{
				status: http.StatusBadRequest,
				body:   "",
			},
			requestBody: "{\"id\":\"alert\",\"type\":\"counter\"}",
		},
		{
			name:   "update json negative case missing name in request body",
			url:    "/update",
			method: http.MethodPost,
			fields: nil,
			want: want{
				status: http.StatusBadRequest,
				body:   "",
			},
			requestBody: "\"type\":\"counter\",\"value\":1}",
		},
		{
			name:   "show json success counter case",
			url:    "/value",
			method: http.MethodPost,
			fields: map[string]testAlert{
				"alert": {
					IntValue: int64(1),
					Type:     "counter",
					Name:     "alert",
				},
			},
			want: want{
				status: http.StatusOK,
				body:   "{\"id\":\"alert\",\"type\":\"counter\",\"delta\":1}",
			},
			requestBody: "{\"id\":\"alert\",\"type\":\"counter\"}",
		},
		{
			name:   "show json success gauge case",
			url:    "/value",
			method: http.MethodPost,
			fields: map[string]testAlert{
				"alert": {
					FloatValue: 1.1,
					Type:       "gauge",
					Name:       "alert",
				},
			},
			want: want{
				status: http.StatusOK,
				body:   "{\"id\":\"alert\",\"type\":\"gauge\",\"value\":1.1}",
			},
			requestBody: "{\"id\":\"alert\",\"type\":\"counter\"}",
		},
		{
			name:   "negative show json case",
			url:    "/value",
			method: http.MethodPost,
			fields: nil,
			want: want{
				status: http.StatusNotFound,
				body:   "",
			},
			requestBody: "{\"id\":\"alert\",\"type\":\"counter\"}",
		},
		{
			name:   "negative show json without type case",
			url:    "/value",
			method: http.MethodPost,
			fields: nil,
			want: want{
				status: http.StatusBadRequest,
				body:   "",
			},
			requestBody: "{\"id\":\"alert\"}",
		},
		{
			name:   "negative show json without id case",
			url:    "/value",
			method: http.MethodPost,
			fields: nil,
			want: want{
				status: http.StatusBadRequest,
				body:   "",
			},
			requestBody: "{\"type\":\"counter\"}",
		},
		{
			name:   "negative show json empty body case",
			url:    "/value",
			method: http.MethodPost,
			fields: nil,
			want: want{
				status: http.StatusBadRequest,
				body:   "",
			},
			requestBody: "",
		},
		{
			name:   "updates mass",
			url:    "/updates",
			method: http.MethodPost,
			fields: map[string]testAlert{
				"Some2": {
					Type:       "counter",
					Name:       "Some2",
					FloatValue: 0,
					IntValue:   2,
				},
			},
			want: want{
				status: http.StatusOK,
				body:   `[{"id":"Some","type":"gauge","value":1.234234},{"id":"Some2","type":"counter","delta":4}]`,
			},
			requestBody: `[{"id":"Some","type":"gauge","value":1.234234},{"id":"Some2","type":"counter","delta":2}]`,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
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
			resp, responseBody := testRequest(t, ts, tt.method, tt.url, tt.requestBody)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.Println(err)
				}
			}()
			assert.Equal(t, tt.want.status, resp.StatusCode)
			if tt.want.body != "" {
				assert.JSONEq(t, tt.want.body, responseBody)
			}
			strg.Reset()
		})
	}
}

func testRequest(
	t *testing.T,
	ts *httptest.Server,
	method,
	path,
	body string,
) (*http.Response, string) {
	t.Helper()
	compressedBody, err := compress.Do([]byte(body))
	require.NoError(t, err)
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewReader(compressedBody))
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	gzipReader, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)

	respBody, err := io.ReadAll(gzipReader)
	require.NoError(t, err)

	return resp, string(respBody)
}
