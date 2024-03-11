package utils

import (
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
	// Root contain root absolute path of this project.
	Root = filepath.Join(filepath.Dir(b), "../..")
)
