package middleware

import (
	"net/http"
	"strings"
)

const partsCountWithoutType = 1
const partsCountWithoutName = 2
const partsCountWithoutValue = 3
const partsCountIsValid = 4

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
		return nil, http.StatusOK
	default:
		return &IncorrectPath{}, http.StatusBadRequest
	}
}
