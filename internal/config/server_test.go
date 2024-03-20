package config

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestServerConfig_parseFromFile(t *testing.T) {
	const tempFileConfigPath = "/tmp/server-config.json"
	tests := []struct {
		name        string
		baseConfig  ServerConfig
		fileConfigs ServerConfig
		filePath    string
		wantErr     bool
		want        ServerConfig
	}{
		{
			name:        "invalid file path case",
			baseConfig:  ServerConfig{},
			fileConfigs: ServerConfig{},
			filePath:    "/tmp/invalid-file-path.json",
			wantErr:     true,
			want:        ServerConfig{},
		},
		{
			name: "success case with default base config",
			baseConfig: ServerConfig{
				Host:          defaultServerHostValue,
				FilePath:      defaultServerFilePathValue,
				DatabaseDSN:   defaultServerDatabaseDSNValue,
				SecretKey:     defaultServerSecretKeyValue,
				CryptoKey:     defaultServerCryptoKeyValue,
				StoreInterval: defaultServerStoreIntervalValue,
				Restore:       defaultServerRestoreValue,
			},
			fileConfigs: ServerConfig{
				Host:          ":8090",
				FilePath:      "/tmp/some-test.db",
				DatabaseDSN:   "test dsn",
				SecretKey:     "secret-key-test",
				CryptoKey:     "123",
				StoreInterval: 400,
				Restore:       false,
			},
			filePath: tempFileConfigPath,
			wantErr:  false,
			want: ServerConfig{
				Host:          ":8090",
				FilePath:      "/tmp/some-test.db",
				DatabaseDSN:   "test dsn",
				SecretKey:     "secret-key-test",
				ConfigPath:    tempFileConfigPath,
				CryptoKey:     "123",
				StoreInterval: 400,
				Restore:       false,
			},
		},
		{
			name: "no effect case",
			baseConfig: ServerConfig{
				Host:          ":8090",
				FilePath:      "test-123-file",
				DatabaseDSN:   "123-123-123",
				SecretKey:     "123",
				CryptoKey:     "321",
				StoreInterval: 500,
				Restore:       false,
			},
			fileConfigs: ServerConfig{
				Host:          ":8091",
				FilePath:      "test-321-file",
				DatabaseDSN:   "321-321-321",
				SecretKey:     "321",
				CryptoKey:     "123",
				StoreInterval: 600,
				Restore:       true,
			},
			filePath: tempFileConfigPath,
			wantErr:  false,
			want: ServerConfig{
				Host:          ":8090",
				FilePath:      "test-123-file",
				DatabaseDSN:   "123-123-123",
				ConfigPath:    tempFileConfigPath,
				SecretKey:     "123",
				CryptoKey:     "321",
				StoreInterval: 500,
				Restore:       false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configData, err := json.Marshal(&tt.fileConfigs)
			require.NoError(t, err)
			err = os.WriteFile(tempFileConfigPath, configData, 0650)
			require.NoError(t, err)

			tt.baseConfig.ConfigPath = tt.filePath
			err = tt.baseConfig.parseFromFile()
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, tt.baseConfig)

			err = os.RemoveAll(tempFileConfigPath)
			require.NoError(t, err)
		})
	}
}
