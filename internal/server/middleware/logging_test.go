package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const testLogFilePath = "/tmp/test_log.txt"

func TestWithLogging(t *testing.T) {
	cnfg := zap.NewDevelopmentConfig()
	cnfg.OutputPaths = []string{testLogFilePath}
	log, err := cnfg.Build()
	require.NoError(t, err)
	logger.Log = log.Sugar()

	bodyReader := strings.NewReader("test-body")
	r := httptest.NewRequest(http.MethodGet, "/test", bodyReader)
	w := httptest.NewRecorder()
	middleware := WithLogging()
	middleware(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
	})).ServeHTTP(w, r)
	loggedContent, err := os.ReadFile(testLogFilePath)
	loggedString := string(loggedContent)
	fmt.Println(loggedString)
	require.NoError(t, err)
	assert.True(t, strings.Contains(loggedString, "INFO\tmiddleware/logging.go:54\turi /test method GET duration"))

	err = os.Remove(testLogFilePath)
	require.NoError(t, err)
}
