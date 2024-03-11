package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/config"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/storage"
)

var memStrg = storage.NewInMemoryStorage()

func TestMain(m *testing.M) {
	ctx := context.Background()
	metric := entity.Alert{
		Type:       "gauge",
		Name:       "metric",
		FloatValue: floatPointer(1.1),
	}
	_ = memStrg.Save(ctx, metric.Name, metric)
	m.Run()
}

func ExampleIndexHandler() {
	r := httptest.NewRequest(http.MethodGet, "/index", nil)
	w := httptest.NewRecorder()
	indexHandler := IndexHandler(memStrg)

	indexHandler.ServeHTTP(w, r)

	res := w.Result()
	defer func() {
		_ = res.Body.Close()
	}()
	body, _ := io.ReadAll(res.Body)

	fmt.Printf("Status code: %d\n", res.StatusCode)
	fmt.Printf("Body valid: %t", strings.Contains(string(body), "metric: 1.1"))
	// Output:
	// Status code: 200
	// Body valid: true
}

func ExamplePingHandler() {
	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	pingHandler := PingHandler(memStrg)
	pingHandler.ServeHTTP(w, r)
	res := w.Result()

	defer func() {
		_ = res.Body.Close()
	}()

	body, _ := io.ReadAll(res.Body)

	fmt.Printf("Status code: %d\n", res.StatusCode)
	fmt.Printf("Body: %s", body)
	// Output:
	// Status code: 200
	// Body: pong
}

func ExampleShowHandler() {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "metric")
	r := httptest.NewRequest(http.MethodGet, "/show/", nil)
	rContext := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
	r = r.WithContext(rContext)
	w := httptest.NewRecorder()

	showHandler := ShowHandler(memStrg)
	showHandler.ServeHTTP(w, r)

	res := w.Result()

	defer func() {
		_ = res.Body.Close()
	}()

	body, _ := io.ReadAll(res.Body)

	fmt.Printf("Status code: %d\n", res.StatusCode)
	fmt.Printf("Body: %s\n", body)

	// Output:
	// Status code: 200
	// Body: 1.1
}

func ExampleShowJSONHandler() {
	rMetric := struct {
		ID    string   `json:"id" valid:"type(string)"`
		MType string   `json:"type" valid:"in(gauge|counter)"`
		Value *float64 `json:"value,omitempty" valid:"optional"`
		Delta *int64   `json:"delta,omitempty" valid:"optional"`
	}{
		ID:    "metric",
		MType: "gauge",
		Value: floatPointer(1.1),
		Delta: nil,
	}
	rBodyString, _ := json.Marshal(rMetric)
	rBody := bytes.NewReader(rBodyString)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/show", rBody)

	showHandler := ShowJSONHandler(memStrg)
	showHandler.ServeHTTP(w, r)

	res := w.Result()

	defer func() {
		_ = res.Body.Close()
	}()

	body, _ := io.ReadAll(res.Body)

	fmt.Printf("Status code: %d\n", res.StatusCode)
	fmt.Printf("Body: %s", body)
	// Output:
	// Status code: 200
	// Body: {"value":1.1,"id":"metric","type":"gauge"}
}

func ExampleUpdateHandler() {
	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("name", "metric")
	chiContext.URLParams.Add("type", "gauge")
	chiContext.URLParams.Add("value", "2.1")
	r := httptest.NewRequest(http.MethodPost, "/update", nil)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiContext))
	w := httptest.NewRecorder()

	updateHandler := UpdateHandler(memStrg, &config.ServerConfig{FilePath: "/tmp/metrics.json"})
	updateHandler.ServeHTTP(w, r)

	res := w.Result()

	defer func() {
		_ = res.Body.Close()
	}()

	body, _ := io.ReadAll(res.Body)
	fmt.Printf("Status code: %d\n", res.StatusCode)
	fmt.Printf("Body: %s\n", body)
	// Output:
	// Status code: 200
	// Body:
}

func ExampleUpdateJSONHandler() {
	r := httptest.NewRequest(
		http.MethodPost,
		"/update",
		strings.NewReader(`{"id":"metric", "type": "gauge", "value":2.1})`),
	)
	w := httptest.NewRecorder()

	updateJSONHandler := UpdateJSONHandler(memStrg, &config.ServerConfig{FilePath: "/tmp/metrics.json"})
	updateJSONHandler.ServeHTTP(w, r)

	res := w.Result()

	defer func() {
		_ = res.Body.Close()
	}()

	body, _ := io.ReadAll(res.Body)
	fmt.Printf("Status code: %d\n", res.StatusCode)
	fmt.Printf("Body: %s\n", body)
	// Output:
	// Status code: 200
	// Body: {"value":2.1,"id":"metric","type":"gauge"}
}

func ExampleBulkUpdate() {
	r := httptest.NewRequest(
		http.MethodPost,
		"/updates",
		strings.NewReader(`[{"id":"metric", "type": "gauge", "value":2.1}]`),
	)
	w := httptest.NewRecorder()

	handler := BulkUpdate(memStrg)
	handler.ServeHTTP(w, r)

	res := w.Result()

	defer func() {
		_ = res.Body.Close()
	}()

	body, _ := io.ReadAll(res.Body)
	fmt.Printf("Status code: %d\n", res.StatusCode)
	fmt.Printf("Body: %s\n", body)
	// Output:
	// Status code: 200
	// Body: [{"value":2.1,"id":"metric","type":"gauge"}]
}

func floatPointer(val float64) *float64 {
	return &val
}
