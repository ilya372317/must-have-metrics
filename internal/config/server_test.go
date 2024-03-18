package config

import (
	"reflect"
	"testing"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		want    *ServerConfig
		name    string
		wantErr bool
	}{
		{
			name: "success default case",
			want: &ServerConfig{
				Host:          ":8080",
				FilePath:      "/tmp/metrics-db.json",
				DatabaseDSN:   "",
				SecretKey:     "",
				StoreInterval: 300,
				Restore:       true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewServer()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServer() got = %v, want %v", got, tt.want)
			}
		})
	}
}
