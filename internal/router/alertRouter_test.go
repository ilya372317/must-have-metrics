package router

import (
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAlertRouter(t *testing.T) {
	strg := storage.MakeInMemoryStorage()
	ts := httptest.NewServer(AlertRouter(strg, "../../static"))
	defer ts.Close()

	var testTable = []struct {
		name   string
		url    string
		method string
		want   string
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
			want:   "<!DOCTYPE html>\n<html>\n<head>\n    <meta charset=\"UTF-8\">\n    <title>Some awesome metrics</title>\n</head>\n<section>\n    <ul>\n        \n        <li>alert: 1</li>\n        \n    </ul>\n</section>\n</html>",
			status: http.StatusOK,
			method: http.MethodGet,
		},
		{
			name:   "show success case",
			url:    "/value/counter/alert",
			method: http.MethodGet,
			want:   "1",
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
			want:   "alert not found\n",
			fields: map[string]entity.Alert{},
			status: http.StatusNotFound,
		},
		{
			name:   "success update case",
			url:    "/update/counter/alert/1",
			method: http.MethodPost,
			want:   "",
			fields: nil,
			status: http.StatusOK,
		},
		{
			name:   "update type is invalid case",
			url:    "/update/invalidType/alert/1",
			method: http.MethodPost,
			want:   "invalid type parameter\n",
			fields: nil,
			status: http.StatusBadRequest,
		},
		{
			name:   "update name is invalid case",
			url:    "/update/gauge//1.1",
			method: http.MethodPost,
			want:   "given name is invalid\n",
			fields: nil,
			status: http.StatusNotFound,
		},
		{
			name:   "update value is invalid case",
			url:    "/update/gauge/name/invalidValue",
			method: http.MethodPost,
			want:   "value is invalid\n",
			fields: nil,
			status: http.StatusBadRequest,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			for name, alert := range tt.fields {
				strg.SaveAlert(name, alert)
			}
			resp, body := testRequest(t, ts, tt.method, tt.url)
			defer resp.Body.Close()
			assert.Equal(t, tt.status, resp.StatusCode)
			assert.Equal(t, tt.want, body)
		})
	}
}

func testRequest(
	t *testing.T,
	ts *httptest.Server,
	method,
	path string,
) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
