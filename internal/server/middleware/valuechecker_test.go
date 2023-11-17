package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueValidator(t *testing.T) {
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
				value: "123",
			},
		},
		{
			name: "correct gauge type test",
			want: want{
				code:     http.StatusOK,
				response: "",
			},
			args: args{
				value: "1.23445",
			},
		},
		{
			name: "incorrect type test",
			want: want{
				code:     http.StatusBadRequest,
				response: "value is invalid\n",
			},
			args: args{
				value: "invalid value",
			},
		},
		{
			name: "empty parameter test",
			want: want{
				code:     http.StatusBadRequest,
				response: "value is invalid\n",
			},
			args: args{value: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			})
			request, err := http.NewRequest(http.MethodPost, "localhost:8080/{value}", nil)
			require.NoError(t, err)
			writer := httptest.NewRecorder()
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("value", tt.args.value)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
			middlewareToTest := ValueValidator()
			funcForHandle := middlewareToTest(nextHandler)
			funcForHandle.ServeHTTP(writer, request)

			res := writer.Result()
			defer res.Body.Close()
			responseBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.response, string(responseBody))
		})
	}
}
