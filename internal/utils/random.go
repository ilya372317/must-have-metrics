package utils

import (
	"math/rand"
	"time"
)

// GetRandomValue generate random integer between min and max value.
func GetRandomValue(min, max int) int {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomValue := random.Intn(max-min+1) + min
	return randomValue
}
