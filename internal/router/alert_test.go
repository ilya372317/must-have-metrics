package router

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertRouter(t *testing.T) {
	strg := storage.NewInMemoryStorage()
	ts := httptest.NewServer(AlertRouter(strg))
	defer ts.Close()

	type want struct {
		status int
		body   string
	}

	var testTable = []struct {
		name        string
		url         string
		method      string
		fields      map[string]entity.Alert
		want        want
		requestBody string
	}{
		{
			name: "index success case",
			fields: map[string]entity.Alert{
				"alert": {
					Type:  "counter",
					Name:  "alert",
					Value: int64(1),
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
			fields: map[string]entity.Alert{
				"alert": {
					Type:  "counter",
					Name:  "alert",
					Value: int64(1),
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
			fields: map[string]entity.Alert{},
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
			fields: map[string]entity.Alert{
				"alert": {
					Value: int64(1),
					Type:  "counter",
					Name:  "alert",
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
			fields: map[string]entity.Alert{
				"alert": {
					Value: int64(1),
					Type:  "counter",
					Name:  "alert",
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
			fields: map[string]entity.Alert{
				"alert": {
					Value: 1.1,
					Type:  "gauge",
					Name:  "alert",
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
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			for name, alert := range tt.fields {
				strg.Save(name, alert)
			}
			resp, responseBody := testRequest(t, ts, tt.method, tt.url, tt.requestBody)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.Println(err)
				}
			}()
			assert.Equal(t, tt.want.status, resp.StatusCode)
			if tt.want.body != "" {
				assert.Equal(t, tt.want.body, responseBody)
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
	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
