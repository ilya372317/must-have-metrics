package router

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

func TestAlertRouter(t *testing.T) {
	strg := storage.NewInMemoryStorage()
	ts := httptest.NewServer(AlertRouter(strg))
	defer ts.Close()

	var testTable = []struct {
		name   string
		url    string
		method string
		fields map[string]entity.Alert
		status int
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
			url:    "/",
			status: http.StatusOK,
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
			status: http.StatusOK,
		},
		{
			name:   "negative show case",
			url:    "/value/counter/alert1",
			method: http.MethodGet,
			fields: map[string]entity.Alert{},
			status: http.StatusNotFound,
		},
		{
			name:   "success update case",
			url:    "/update/counter/alert/1",
			method: http.MethodPost,
			fields: nil,
			status: http.StatusOK,
		},
		{
			name:   "update type is invalid case",
			url:    "/update/invalidType/alert/1",
			method: http.MethodPost,
			fields: nil,
			status: http.StatusBadRequest,
		},
		{
			name:   "update value is invalid case",
			url:    "/update/gauge/name/invalidValue",
			method: http.MethodPost,
			fields: nil,
			status: http.StatusBadRequest,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			for name, alert := range tt.fields {
				strg.Save(name, alert)
			}
			resp, _ := testRequest(t, ts, tt.method, tt.url)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.Println(err)
				}
			}()
			assert.Equal(t, tt.status, resp.StatusCode)
		})
	}
}

func testRequest(
	t *testing.T,
	ts *httptest.Server,
	method,
	path string,
) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(method, ts.URL+path, nil)
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
