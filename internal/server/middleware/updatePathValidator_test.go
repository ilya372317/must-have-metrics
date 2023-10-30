package middleware

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidUpdate(t *testing.T) {
	type want struct {
		code int
	}

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "success gauge test",
			url:  "http://localhost:8080/update/gauge/someName/1.15",
			want: want{code: http.StatusOK},
		},
		{
			name: "success counter test",
			url:  "http://localhost:8080/update/counter/someName/1",
			want: want{code: http.StatusOK},
		},
		{
			name: "invalid value test",
			url:  "http://localhost:8080/update/counter/someName/invalidValue",
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "invalid type request",
			url:  "http://localhost:8080/update/invalidType/someName/1.1",
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "without value test",
			url:  "http://localhost:8080/update/invalidType/someName",
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "without name test",
			url:  "http://localhost:8080/update/invalidType",
			want: want{code: http.StatusNotFound},
		},
		{
			name: "without type test",
			url:  "http://localhost:8080/update",
			want: want{code: http.StatusBadRequest},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := func(w http.ResponseWriter, r *http.Request) {}

			request, err := http.NewRequest(http.MethodPost, tt.url, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			middlewareToTest := ValidUpdate()
			handlerToTest := middlewareToTest(nextHandler)
			handlerToTest.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}

func Test_checkURL(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "success valid url",
			args:    args{path: "update/gauge/someName/1.1"},
			want:    http.StatusOK,
			wantErr: false,
		},
		{
			name:    "invalid path without update",
			args:    args{path: "gauge/someName/1.1"},
			want:    http.StatusBadRequest,
			wantErr: true,
		},
		{
			name:    "invalid path with invalid type",
			args:    args{path: "update/invalid_type/someName/1.1"},
			want:    http.StatusBadRequest,
			wantErr: true,
		},
		{
			name:    "invalid path without name",
			args:    args{path: "update/counter//1.1"},
			want:    http.StatusBadRequest,
			wantErr: true,
		},
		{
			name:    "invalid path without parameters",
			args:    args{path: "update"},
			want:    http.StatusBadRequest,
			wantErr: true,
		},
		{
			name:    "invalid path with only type",
			args:    args{path: "update/gauge"},
			want:    http.StatusNotFound,
			wantErr: true,
		},
		{
			name:    "invalid path without value",
			args:    args{path: "update/gauge/someName/"},
			want:    http.StatusBadRequest,
			wantErr: true,
		},
		{
			name:    "invalid path with incorrect value",
			args:    args{path: "update/gauge/someName/incorrect"},
			want:    http.StatusBadRequest,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkURL(tt.args.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equalf(t, tt.want, got, "checkURL(%v)", tt.args.path)
		})
	}
}

func Test_typeIsValid(t *testing.T) {
	type args struct {
		typ string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "success gauge test",
			args: args{typ: "gauge"},
			want: true,
		},
		{
			name: "success counter test",
			args: args{typ: "counter"},
			want: true,
		},
		{
			name: "negative test",
			args: args{typ: "type_not_present"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, typeIsValid(tt.args.typ), "typeIsValid(%v)", tt.args.typ)
		})
	}
}

func Test_validateParts(t *testing.T) {
	type args struct {
		pathParts []string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "all correct part test",
			args:    args{pathParts: []string{"update", "gauge", "test_name", "1.1"}},
			want:    200,
			wantErr: false,
		},
		{
			name:    "value of parts is invalid",
			args:    args{pathParts: []string{"update", "counter", "test_name", "invalid value"}},
			want:    http.StatusBadRequest,
			wantErr: true,
		},
		{
			name:    "type is invalid",
			args:    args{pathParts: []string{"update", "invalid_type", "test_name", "1"}},
			wantErr: true,
			want:    http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateParts(tt.args.pathParts)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equalf(t, tt.want, got, "validateParts(%v)", tt.args.pathParts)
		})
	}
}

func Test_valueIsValid(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "success float value test",
			args: args{value: "1.23"},
			want: true,
		},
		{
			name: "success integer value test",
			args: args{value: "34"},
			want: true,
		},
		{
			name: "negative string value test",
			args: args{value: "it is not pass"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, valueIsValid(tt.args.value), "valueIsValid(%v)", tt.args.value)
		})
	}
}
