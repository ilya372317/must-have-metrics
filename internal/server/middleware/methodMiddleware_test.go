package middleware

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMethod(t *testing.T) {
	type args struct {
		method string
	}
	type want struct {
		expectCode int
	}
	tests := []struct {
		name             string
		args             args
		middlewareMethod string
		want             want
	}{
		{
			name:             "Unexpected method",
			args:             args{method: http.MethodPost},
			middlewareMethod: http.MethodGet,
			want:             want{expectCode: http.StatusMethodNotAllowed},
		},
		{
			name:             "success test",
			args:             args{method: http.MethodGet},
			middlewareMethod: http.MethodGet,
			want:             want{expectCode: http.StatusOK},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := func(w http.ResponseWriter, r *http.Request) {}

			request, err := http.NewRequest(tt.args.method, "http://localhost:8080", nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			middlewareToTest := Method(tt.middlewareMethod)
			handlerToTest := middlewareToTest(nextHandler)
			handlerToTest.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.expectCode, res.StatusCode)
		})
	}
}
