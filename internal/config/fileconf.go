package config

import (
	"fmt"
	"os"
)

func getConfigFileContent(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return []byte(""), fmt.Errorf("failed read config file content on path %s: %w", filePath, err)
	}

	return data, nil
}
