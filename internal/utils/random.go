package utils

import (
	"math/rand"
	"time"
)

func GetRandomValue(min, max int) int {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomValue := random.Intn(max-min+1) + min
	return randomValue
}
