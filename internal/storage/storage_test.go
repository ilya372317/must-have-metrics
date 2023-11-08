package storage

import (
	"reflect"
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryStorage_GetAlert(t *testing.T) {
	type fields struct {
		Records map[string]entity.Alert
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    entity.Alert
		wantErr bool
	}{
		{
			name: "positive test",
			fields: fields{Records: map[string]entity.Alert{
				"alert": {
					Type:  "gauge",
					Name:  "alert",
					Value: 1,
				},
			}},
			args: args{name: "alert"},
			want: entity.Alert{
				Type:  "gauge",
				Name:  "alert",
				Value: 1,
			},
			wantErr: false,
		},
		{
			name:    "negative case",
			fields:  fields{Records: map[string]entity.Alert{}},
			args:    args{name: "test"},
			want:    entity.Alert{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &InMemoryStorage{
				Records: tt.fields.Records,
			}
			got, err := storage.Get(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryStorage_HasAlert(t *testing.T) {
	type fields struct {
		Records map[string]entity.Alert
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "positive test",
			fields: fields{Records: map[string]entity.Alert{
				"test": {
					Type:  "gauge",
					Name:  "test",
					Value: 10,
				},
			}},
			args: args{name: "test"},
			want: true,
		},
		{
			name:   "negative test",
			fields: fields{Records: map[string]entity.Alert{}},
			args:   args{name: "test"},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &InMemoryStorage{
				Records: tt.fields.Records,
			}
			if got := storage.Has(tt.args.name); got != tt.want {
				t.Errorf("Has() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryStorage_SaveAlert(t *testing.T) {
	type args struct {
		name  string
		alert entity.Alert
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "simple success test case",
			args: args{
				name: "testValue",
				alert: entity.Alert{
					Type:  "gauge",
					Name:  "test",
					Value: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := MakeInMemoryStorage()
			storage.Save(tt.args.name, tt.args.alert)
			value, hasRecord := storage.Records[tt.args.name]
			assert.Equal(t, tt.args.alert.Value, value.Value)
			assert.True(t, hasRecord)
		})
	}
}

func TestInMemoryStorage_UpdateAlert(t *testing.T) {
	type fields struct {
		Records map[string]entity.Alert
	}
	type args struct {
		name     string
		newValue entity.Alert
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    entity.Alert
	}{
		{
			name: "positive case",
			fields: fields{Records: map[string]entity.Alert{
				"alert": {
					Type:  "gauge",
					Name:  "alert",
					Value: 1,
				},
			}},
			args: args{
				name: "alert",
				newValue: entity.Alert{
					Type:  "gauge",
					Name:  "alert",
					Value: 2,
				},
			},
			wantErr: false,
			want: entity.Alert{
				Type:  "gauge",
				Name:  "alert",
				Value: 2,
			},
		},
		{
			name: "negative case",
			fields: fields{Records: map[string]entity.Alert{
				"alert_test": {
					Type:  "gauge",
					Name:  "alert_test",
					Value: 1,
				},
			}},
			args: args{
				name: "alert",
				newValue: entity.Alert{
					Type:  "gauge",
					Name:  "alert",
					Value: 10,
				},
			},
			wantErr: true,
			want:    entity.Alert{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := MakeInMemoryStorage()
			storage.Records = tt.fields.Records
			err := storage.Update(tt.args.name, tt.args.newValue)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, storage.Records[tt.args.name], tt.want)
		})
	}
}
