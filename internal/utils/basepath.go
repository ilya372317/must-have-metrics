package utils

import (
	"path/filepath"
	"runtime"
)

func BasePath() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../..")
}
