package service

import (
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/dto"
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/ilya372317/must-have-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testAlert struct {
	Type       string
	Name       string
	FloatValue float64
	IntValue   int64
}

func Test_addAlert(t *testing.T) {
	type args struct {
		repo *storage.InMemoryStorage
		dto  dto.UpdateAlertDTO
	}
	tests := []struct {
		name    string
		args    args
		fields  map[string]testAlert
		wantErr bool
		want    testAlert
	}{
		{
			name: "success counter empty storage case",
			args: args{
				repo: storage.NewInMemoryStorage(),
				dto: dto.UpdateAlertDTO{
					Type: "counter",
					Name: "alert",
					Data: "10",
				},
			},
			fields:  map[string]testAlert{},
			wantErr: false,
			want: testAlert{
				Type:     "counter",
				Name:     "alert",
				IntValue: int64(10),
			},
		},
		{
			name: "success counter not empty storage case",
			args: args{
				repo: storage.NewInMemoryStorage(),
				dto: dto.UpdateAlertDTO{
					Type: "counter",
					Name: "alert",
					Data: "10",
				},
			},
			fields: map[string]testAlert{
				"alert": {
					Type:     "counter",
					Name:     "alert",
					IntValue: int64(20),
				},
			},
			wantErr: false,
			want: testAlert{
				Type:     "counter",
				Name:     "alert",
				IntValue: int64(30),
			},
		},
		{
			name: "success gauge empty storage case",
			args: args{
				repo: storage.NewInMemoryStorage(),
				dto: dto.UpdateAlertDTO{
					Type: "gauge",
					Name: "alert",
					Data: "1.1",
				},
			},
			fields:  map[string]testAlert{},
			wantErr: false,
			want: testAlert{
				Type:       "gauge",
				Name:       "alert",
				FloatValue: 1.1,
			},
		},
		{
			name: "success gauge not empty storage case",
			args: args{
				repo: storage.NewInMemoryStorage(),
				dto: dto.UpdateAlertDTO{
					Type: "gauge",
					Name: "alert",
					Data: "1.2",
				},
			},
			fields: map[string]testAlert{
				"alert": {
					Type:       "gauge",
					Name:       "alert",
					FloatValue: 1.1,
				},
			},
			wantErr: false,
			want: testAlert{
				Type:       "gauge",
				Name:       "alert",
				FloatValue: 1.2,
			},
		},
		{
			name: "negative gauge invalid value case",
			args: args{
				repo: storage.NewInMemoryStorage(),
				dto: dto.UpdateAlertDTO{
					Type: "gauge",
					Name: "alert",
					Data: "invalid value",
				},
			},
			fields:  map[string]testAlert{},
			wantErr: true,
			want:    testAlert{},
		},
		{
			name: "negative counter invalid value case",
			args: args{
				repo: storage.NewInMemoryStorage(),
				dto: dto.UpdateAlertDTO{
					Type: "counter",
					Name: "alert",
					Data: "invalid data",
				},
			},
			fields:  map[string]testAlert{},
			wantErr: true,
			want:    testAlert{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for name, alert := range tt.fields {
				tt.args.repo.Save(name, newAlertFromTestAlert(alert))
			}

			_, err := AddAlert(tt.args.repo, tt.args.dto)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			addedAlert, addAlertErr := tt.args.repo.Get(tt.args.dto.Name)
			require.NoError(t, addAlertErr)
			assert.Equal(t, addedAlert, newAlertFromTestAlert(tt.want))
		})
	}
}

func Test_updateCounterAlert(t *testing.T) {
	type args struct {
		dto  dto.UpdateAlertDTO
		repo *storage.InMemoryStorage
	}
	tests := []struct {
		name    string
		args    args
		fields  map[string]testAlert
		wantErr bool
		want    testAlert
	}{
		{
			name: "success case with empty storage",
			args: args{
				dto: dto.UpdateAlertDTO{
					Type: "counter",
					Name: "alert",
					Data: "10",
				},
				repo: storage.NewInMemoryStorage(),
			},
			fields:  map[string]testAlert{},
			wantErr: false,
			want: testAlert{
				Type:     "counter",
				Name:     "alert",
				IntValue: int64(10),
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
				repo: storage.NewInMemoryStorage(),
			},
			fields: map[string]testAlert{
				"alert": {
					Type:     "counter",
					Name:     "alert",
					IntValue: int64(10),
				},
			},
			wantErr: false,
			want: testAlert{
				Type:     "counter",
				Name:     "alert",
				IntValue: int64(20),
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
				repo: storage.NewInMemoryStorage(),
			},
			fields:  map[string]testAlert{},
			wantErr: true,
			want:    testAlert{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for name, alert := range tt.fields {
				tt.args.repo.Save(name, newAlertFromTestAlert(alert))
			}

			_, err := updateCounterAlert(tt.args.dto, tt.args.repo)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}
			updatedAlert, getAlertErr := tt.args.repo.Get(tt.args.dto.Name)
			require.NoError(t, getAlertErr)
			assert.Equal(t, updatedAlert, newAlertFromTestAlert(tt.want))
		})
	}
}

func Test_updateGaugeAlert(t *testing.T) {
	type args struct {
		dto        dto.UpdateAlertDTO
		repository *storage.InMemoryStorage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    testAlert
	}{
		{
			name: "positive test",
			args: args{
				dto: dto.UpdateAlertDTO{
					Type: "gauge",
					Name: "alert",
					Data: "1.12",
				},
				repository: storage.NewInMemoryStorage(),
			},
			wantErr: false,
			want: testAlert{
				Type:       "gauge",
				Name:       "alert",
				FloatValue: 1.12,
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
				repository: storage.NewInMemoryStorage(),
			},
			wantErr: false,
			want: testAlert{
				Type:       "gauge",
				Name:       "alert",
				FloatValue: 1.0,
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
				repository: storage.NewInMemoryStorage(),
			},
			wantErr: true,
			want:    testAlert{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := updateGaugeAlert(tt.args.dto, tt.args.repository)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			expectedAlert, getAlertError := tt.args.repository.Get(tt.args.dto.Name)
			require.NoError(t, getAlertError)
			assert.Equal(t, newAlertFromTestAlert(tt.want), expectedAlert)
		})
	}
}

func newAlertFromTestAlert(testAlert testAlert) entity.Alert {
	wantAlert := entity.Alert{
		Type: testAlert.Type,
		Name: testAlert.Name,
	}
	if testAlert.FloatValue != 0 {
		floatValue := testAlert.FloatValue
		wantAlert.FloatValue = &floatValue
	}
	if testAlert.IntValue != 0 {
		intValue := testAlert.IntValue
		wantAlert.IntValue = &intValue
	}

	return wantAlert
}
