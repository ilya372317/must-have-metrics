package middleware

import (
	"github.com/ilya372317/must-have-metrics/internal/server/entity"
	"net/http"
	"strconv"
	"strings"
)

const partsCountWithoutType = 1
const partsCountWithoutName = 2
const partsCountWithoutValue = 3
const partsCountIsValid = 4

const typePartPosition = 1
const valuePartPosition = 3

func ValidUpdate() Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			urlPath := r.URL.Path

			if statusCode, err := checkURL(urlPath); err != nil {
				http.Error(w, err.Error(), statusCode)
				return
			}

			f(w, r)
		}
	}
}

func checkURL(path string) (int, error) {
	pathParts := strings.Split(path, "/")
	pathPartsWithoutEmpty := make([]string, 0, cap(pathParts))
	for _, part := range pathParts {
		if strings.TrimSpace(part) != "" {
			pathPartsWithoutEmpty = append(pathPartsWithoutEmpty, part)
		}
	}

	switch len(pathPartsWithoutEmpty) {
	case partsCountWithoutType:
		return http.StatusBadRequest, &IncorrectPath{}
	case partsCountWithoutName:
		return http.StatusNotFound, &IncorrectPath{}
	case partsCountWithoutValue:
		return http.StatusBadRequest, &IncorrectPath{}
	case partsCountIsValid:
		return validateParts(pathPartsWithoutEmpty)
	default:
		return http.StatusBadRequest, &IncorrectPath{}
	}
}

func validateParts(pathParts []string) (int, error) {
	if !typeIsValid(pathParts[typePartPosition]) {
		return http.StatusBadRequest, &IncorrectPath{}
	}

	if !valueIsValid(pathParts[valuePartPosition]) {
		return http.StatusBadRequest, &IncorrectPath{}
	}

	return http.StatusOK, nil
}

func valueIsValid(value string) bool {
	_, intErr := strconv.Atoi(value)
	_, floatErr := strconv.ParseFloat(value, 64)
	return intErr == nil || floatErr == nil
}

func typeIsValid(typ string) bool {
	return typ == entity.GaugeType || typ == entity.CounterType
}
