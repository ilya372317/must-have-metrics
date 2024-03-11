package utils

import (
	"math/rand"
	"time"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

// GetRandomValue generate random integer between min and max value.
func GetRandomValue(min, max int) int {
	randomValue := random.Intn(max-min+1) + min
	return randomValue
}
