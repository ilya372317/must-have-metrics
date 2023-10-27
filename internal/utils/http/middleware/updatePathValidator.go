package middleware

import (
	"github.com/ilya372317/must-have-metrics/internal/entity"
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

			if err, statusCode := checkUrl(urlPath); err != nil {
				http.Error(w, err.Error(), statusCode)
				return
			}

			f(w, r)
		}
	}
}

func checkUrl(path string) (error, int) {
	pathParts := strings.Split(path, "/")
	pathPartsWithoutEmpty := make([]string, 0, cap(pathParts))
	for _, part := range pathParts {
		if strings.TrimSpace(part) != "" {
			pathPartsWithoutEmpty = append(pathPartsWithoutEmpty, part)
		}
	}

	switch len(pathPartsWithoutEmpty) {
	case partsCountWithoutType:
		return &IncorrectPath{}, http.StatusBadRequest
	case partsCountWithoutName:
		return &IncorrectPath{}, http.StatusNotFound
	case partsCountWithoutValue:
		return &IncorrectPath{}, http.StatusBadRequest
	case partsCountIsValid:
		return validateParts(pathPartsWithoutEmpty)
	default:
		return &IncorrectPath{}, http.StatusBadRequest
	}
}

func validateParts(pathParts []string) (error, int) {
	if !typeIsValid(pathParts[typePartPosition]) {
		return &IncorrectPath{}, http.StatusBadRequest
	}

	if !valueIsValid(pathParts[valuePartPosition]) {
		return &IncorrectPath{}, http.StatusBadRequest
	}

	return nil, http.StatusOK
}

func valueIsValid(value string) bool {
	_, intErr := strconv.Atoi(value)
	_, floatErr := strconv.ParseFloat(value, 64)
	return intErr == nil || floatErr == nil
}

func typeIsValid(typ string) bool {
	return typ == entity.GaugeType || typ == entity.CounterType
}
