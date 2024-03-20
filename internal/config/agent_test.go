package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentConfig_parseFromFile(t *testing.T) {
	const tempFileConfigPath = "/tmp/agent-config.json"
	tests := []struct {
		name        string
		baseConfig  AgentConfig
		fileConfigs AgentConfig
		filePath    string
		wantErr     bool
		want        AgentConfig
	}{
		{
			name:        "invalid file path case",
			baseConfig:  AgentConfig{},
			fileConfigs: AgentConfig{},
			filePath:    "/tmp/invalid-file-path.json",
			wantErr:     true,
			want:        AgentConfig{},
		},
		{
			name: "success case with default base config",
			baseConfig: AgentConfig{
				Host:           defaultAgentAddressValue,
				SecretKey:      defaultAgentSecretKeyValue,
				CryptoKey:      defaultAgentCryptoKeyValue,
				PollInterval:   defaultAgentPollIntervalValue,
				ReportInterval: defaultAgentReportIntervalValue,
				RateLimit:      defaultAgentRateLimitValue,
				ConfigPath:     defaultAgentConfigValue,
			},
			fileConfigs: AgentConfig{
				Host:           "localhost:9090",
				SecretKey:      "123",
				CryptoKey:      "321",
				PollInterval:   5,
				ReportInterval: 15,
				RateLimit:      20,
			},
			filePath: tempFileConfigPath,
			wantErr:  false,
			want: AgentConfig{
				Host:           "localhost:9090",
				SecretKey:      "123",
				CryptoKey:      "321",
				PollInterval:   5,
				ReportInterval: 15,
				RateLimit:      20,
				ConfigPath:     tempFileConfigPath,
			},
		},
		{
			name: "no effect case",
			baseConfig: AgentConfig{
				Host:           "localhost:8090",
				SecretKey:      "123",
				CryptoKey:      "123",
				PollInterval:   4,
				ReportInterval: 5,
				RateLimit:      6,
			},
			fileConfigs: AgentConfig{
				Host:           "localhost:8091",
				SecretKey:      "321",
				CryptoKey:      "321",
				PollInterval:   5,
				ReportInterval: 6,
				RateLimit:      7,
			},
			filePath: tempFileConfigPath,
			wantErr:  false,
			want: AgentConfig{
				Host:           "localhost:8090",
				SecretKey:      "123",
				CryptoKey:      "123",
				PollInterval:   4,
				ReportInterval: 5,
				RateLimit:      6,
				ConfigPath:     tempFileConfigPath,
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
