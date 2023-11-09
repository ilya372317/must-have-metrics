package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/server/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateHandler(t *testing.T) {
	type query struct {
		typ   string
		name  string
		value string
	}
	type want struct {
		alert entity.Alert
		code  int
	}

	tests := []struct {
		name   string
		query  query
		fields map[string]entity.Alert
		want   want
	}{
		{
			name: "success gauge with empty storage case",
			query: query{
				typ:   "gauge",
				name:  "alert",
				value: "1.1",
			},
			want: want{
				alert: entity.Alert{
					Type:  "gauge",
					Name:  "alert",
					Value: 1.1,
				},
				code: http.StatusOK,
			},
			fields: map[string]entity.Alert{},
		},
		{
			name: "success gauge with not empty storage case",
			query: query{
				typ:   "gauge",
				name:  "alert",
				value: "1.2",
			},
			want: want{
				alert: entity.Alert{
					Type:  "gauge",
					Name:  "alert",
					Value: 1.2,
				},
				code: http.StatusOK,
			},
			fields: map[string]entity.Alert{
				"alert": {
					Type:  "gauge",
					Name:  "alert",
					Value: 1.1,
				},
			},
		},
		{
			name: "success counter with empty storage case",
			query: query{
				typ:   "counter",
				name:  "alert",
				value: "10",
			},
			want: want{
				alert: entity.Alert{
					Type:  "counter",
					Name:  "alert",
					Value: int64(10),
				},
				code: http.StatusOK,
			},
			fields: map[string]entity.Alert{},
		},
		{
			name: "success counter with not empty storage case",
			query: query{
				typ:   "counter",
				name:  "alert",
				value: "10",
			},
			want: want{
				alert: entity.Alert{
					Type:  "counter",
					Name:  "alert",
					Value: int64(20),
				},
				code: http.StatusOK,
			},
			fields: map[string]entity.Alert{
				"alert": {
					Type:  "counter",
					Name:  "alert",
					Value: int64(10),
				},
			},
		},
		{
			name: "negative gauge invalid value case",
			query: query{
				typ:   "gauge",
				name:  "alert",
				value: "invalid value",
			},
			want: want{
				alert: entity.Alert{},
				code:  http.StatusBadRequest,
			},
			fields: map[string]entity.Alert{},
		},
		{
			name: "negative counter invalid case value",
			query: query{
				typ:   "counter",
				name:  "alert",
				value: "invalid value",
			},
			want: want{
				alert: entity.Alert{},
				code:  http.StatusBadRequest,
			},
			fields: map[string]entity.Alert{},
		},
		{
			name: "replace gauge type to counter",
			query: query{
				typ:   "counter",
				name:  "alert",
				value: "1",
			},
			fields: map[string]entity.Alert{
				"alert": {
					Type:  "gauge",
					Name:  "alert",
					Value: 1.32,
				},
			},
			want: want{
				alert: entity.Alert{
					Type:  "counter",
					Name:  "alert",
					Value: int64(1),
				},
				code: http.StatusOK,
			},
		},
		{
			name: "replace from counter to gauge type",
			query: query{
				typ:   "gauge",
				name:  "alert",
				value: "1.15",
			},
			fields: map[string]entity.Alert{
				"alert": {
					Type:  "counter",
					Name:  "alert",
					Value: 1,
				},
			},
			want: want{
				alert: entity.Alert{
					Type:  "gauge",
					Name:  "alert",
					Value: 1.15,
				},
				code: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", tt.query.typ)
			rctx.URLParams.Add("name", tt.query.name)
			rctx.URLParams.Add("value", tt.query.value)
			request, err := http.NewRequest(
				"POST",
				"http://localhost:8080/update/{type}/{name}/{value}",
				nil,
			)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
			require.NoError(t, err)
			writer := httptest.NewRecorder()
			repo := storage.MakeInMemoryStorage()
			for name, alert := range tt.fields {
				repo.Save(name, alert)
			}

			handler := UpdateHandler(repo)
			handler(writer, request)
			res := writer.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
			if res.StatusCode >= 400 {
				return
			}

			addedAlert, err := repo.Get(tt.query.name)
			assert.NoError(t, err)

			assert.Equal(t, addedAlert, tt.want.alert)
		})
	}
}

func Test_addAlert(t *testing.T) {
	type args struct {
		repo storage.Storage
		dto  dto.UpdateAlertDTO
	}
	tests := []struct {
		name    string
		args    args
		fields  map[string]entity.Alert
		wantErr bool
		want    entity.Alert
	}{
		{
			name: "success counter empty storage case",
			args: args{
				repo: storage.MakeInMemoryStorage(),
				dto: dto.UpdateAlertDTO{
					Type: "counter",
					Name: "alert",
					Data: "10",
				},
			},
			fields:  map[string]entity.Alert{},
			wantErr: false,
			want: entity.Alert{
				Type:  "counter",
				Name:  "alert",
				Value: int64(10),
			},
		},
		{
			name: "success counter not empty storage case",
			args: args{
				repo: storage.MakeInMemoryStorage(),
				dto: dto.UpdateAlertDTO{
					Type: "counter",
					Name: "alert",
					Data: "10",
				},
			},
			fields: map[string]entity.Alert{
				"alert": {
					Type:  "counter",
					Name:  "alert",
					Value: int64(20),
				},
			},
			wantErr: false,
			want: entity.Alert{
				Type:  "counter",
				Name:  "alert",
				Value: int64(30),
			},
		},
		{
			name: "success gauge empty storage case",
			args: args{
				repo: storage.MakeInMemoryStorage(),
				dto: dto.UpdateAlertDTO{
					Type: "gauge",
					Name: "alert",
					Data: "1.1",
				},
			},
			fields:  map[string]entity.Alert{},
			wantErr: false,
			want: entity.Alert{
				Type:  "gauge",
				Name:  "alert",
				Value: 1.1,
			},
		},
		{
			name: "success gauge not empty storage case",
			args: args{
				repo: storage.MakeInMemoryStorage(),
				dto: dto.UpdateAlertDTO{
					Type: "gauge",
					Name: "alert",
					Data: "1.2",
				},
			},
			fields: map[string]entity.Alert{
				"alert": {
					Type:  "gauge",
					Name:  "alert",
					Value: 1.1,
				},
			},
			wantErr: false,
			want: entity.Alert{
				Type:  "gauge",
				Name:  "alert",
				Value: 1.2,
			},
		},
		{
			name: "negative gauge invalid value case",
			args: args{
				repo: storage.MakeInMemoryStorage(),
				dto: dto.UpdateAlertDTO{
					Type: "gauge",
					Name: "alert",
					Data: "invalid value",
				},
			},
			fields:  map[string]entity.Alert{},
			wantErr: true,
			want:    entity.Alert{},
		},
		{
			name: "negative counter invalid value case",
			args: args{
				repo: storage.MakeInMemoryStorage(),
				dto: dto.UpdateAlertDTO{
					Type: "counter",
					Name: "alert",
					Data: "invalid data",
				},
			},
			fields:  map[string]entity.Alert{},
			wantErr: true,
			want:    entity.Alert{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for name, alert := range tt.fields {
				tt.args.repo.Save(name, alert)
			}

			err := addAlert(tt.args.repo, tt.args.dto)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			addedAlert, addAlertErr := tt.args.repo.Get(tt.args.dto.Name)
			require.NoError(t, addAlertErr)
			assert.Equal(t, addedAlert, tt.want)
		})
	}
}

func Test_updateCounterAlert(t *testing.T) {
	type args struct {
		dto  dto.UpdateAlertDTO
		repo storage.Storage
	}
	tests := []struct {
		name    string
		args    args
		fields  map[string]entity.Alert
		wantErr bool
		want    entity.Alert
	}{
		{
			name: "success case with empty storage",
			args: args{
				dto: dto.UpdateAlertDTO{
					Type: "counter",
					Name: "alert",
					Data: "10",
				},
				repo: storage.MakeInMemoryStorage(),
			},
			fields:  map[string]entity.Alert{},
			wantErr: false,
			want: entity.Alert{
				Type:  "counter",
				Name:  "alert",
				Value: int64(10),
			},
		},
		{
			name: "success case with value in storage",
			args: args{
				dto: dto.UpdateAlertDTO{
					Type: "counter",
					Name: "alert",
					Data: "10",
				},
				repo: storage.MakeInMemoryStorage(),
			},
			fields: map[string]entity.Alert{
				"alert": {
					Type:  "counter",
					Name:  "alert",
					Value: int64(10),
				},
			},
			wantErr: false,
			want: entity.Alert{
				Type:  "counter",
				Name:  "alert",
				Value: int64(20),
			},
		},
		{
			name: "negative parse int case",
			args: args{
				dto: dto.UpdateAlertDTO{
					Type: "counter",
					Name: "alert",
					Data: "invalid data",
				},
				repo: storage.MakeInMemoryStorage(),
			},
			fields:  map[string]entity.Alert{},
			wantErr: true,
			want:    entity.Alert{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for name, alert := range tt.fields {
				tt.args.repo.Save(name, alert)
			}

			err := updateCounterAlert(tt.args.dto, tt.args.repo)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}
			updatedAlert, getAlertErr := tt.args.repo.Get(tt.args.dto.Name)
			require.NoError(t, getAlertErr)
			assert.Equal(t, updatedAlert, tt.want)
		})
	}
}

func Test_updateGaugeAlert(t *testing.T) {
	type args struct {
		dto        dto.UpdateAlertDTO
		repository storage.Storage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    entity.Alert
	}{
		{
			name: "positive test",
			args: args{
				dto: dto.UpdateAlertDTO{
					Type: "gauge",
					Name: "alert",
					Data: "1.12",
				},
				repository: storage.MakeInMemoryStorage(),
			},
			wantErr: false,
			want: entity.Alert{
				Type:  "gauge",
				Name:  "alert",
				Value: 1.12,
			},
		},
		{
			name: "parse integer value",
			args: args{
				dto: dto.UpdateAlertDTO{
					Type: "gauge",
					Name: "alert",
					Data: "1",
				},
				repository: storage.MakeInMemoryStorage(),
			},
			wantErr: false,
			want: entity.Alert{
				Type:  "gauge",
				Name:  "alert",
				Value: 1.0,
			},
		},
		{
			name: "negative parse float",
			args: args{
				dto: dto.UpdateAlertDTO{
					Type: "gauge",
					Name: "alert",
					Data: "invalid data",
				},
				repository: storage.MakeInMemoryStorage(),
			},
			wantErr: true,
			want:    entity.Alert{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := updateGaugeAlert(tt.args.dto, tt.args.repository)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			expectedAlert, getAlertError := tt.args.repository.Get(tt.args.dto.Name)
			require.NoError(t, getAlertError)
			assert.Equal(t, tt.want, expectedAlert)
		})
	}
}
