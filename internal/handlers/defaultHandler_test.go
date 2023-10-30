package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultHandler(t *testing.T) {
	type want struct {
		code     int
		response string
	}
	type args struct {
		url string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "simple success test",
			args: args{
				url: "http://localhost:8080/test/1/3/45/",
			},
			want: want{code: http.StatusBadRequest, response: "incorrect route\n"},
		},
		{
			name: "another simple success test",
			args: args{
				url: "http://localhost:8080/test/1/3/45/434/34/34/34/",
			},
			want: want{code: http.StatusBadRequest, response: "incorrect route\n"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request, err := http.NewRequest(
				http.MethodPost,
				tt.args.url,
				nil,
			)
			require.NoError(t, err)

			writer := httptest.NewRecorder()

			handler := DefaultHandler()
			handler.ServeHTTP(writer, request)

			res := writer.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)
			defer res.Body.Close()
			responseBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)

			assert.Equal(t, tt.want.response, string(responseBody))
		})
	}
}
