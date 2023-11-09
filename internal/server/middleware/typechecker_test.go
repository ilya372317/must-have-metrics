package middleware

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeValidator(t *testing.T) {
	type want struct {
		code     int
		response string
	}
	type args struct {
		value string
	}
	tests := []struct {
		name string
		want want
		args args
	}{
		{
			name: "correct type counter test",
			want: want{
				code:     http.StatusOK,
				response: "",
			},
			args: args{
				value: "counter",
			},
		},
		{
			name: "correct gauge type test",
			want: want{
				code:     http.StatusOK,
				response: "",
			},
			args: args{
				value: "gauge",
			},
		},
		{
			name: "incorrect type test",
			want: want{
				code:     http.StatusBadRequest,
				response: "invalid type parameter\n",
			},
			args: args{
				value: "invalid value",
			},
		},
		{
			name: "empty parameter test",
			want: want{
				code:     http.StatusBadRequest,
				response: "invalid type parameter\n",
			},
			args: args{value: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			})
			request, err := http.NewRequest(http.MethodPost, "localhost:8080/{type}", nil)
			require.NoError(t, err)
			writer := httptest.NewRecorder()
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", tt.args.value)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
			middlewareToTest := TypeValidator()
			funcForHandle := middlewareToTest(nextHandler)
			funcForHandle.ServeHTTP(writer, request)

			res := writer.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Println(err)
				}
			}()
			responseBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.response, string(responseBody))
		})
	}
}
