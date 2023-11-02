package middleware

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func ValueValidator() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value := chi.URLParam(r, "value")
			if !valueIsValid(value) {
				http.Error(w, "value is invalid", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func valueIsValid(value string) bool {
	_, intErr := strconv.Atoi(value)
	_, floatErr := strconv.ParseFloat(value, 64)
	return intErr == nil || floatErr == nil
}
