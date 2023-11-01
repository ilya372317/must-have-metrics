package middleware

import (
	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/must-have-metrics/internal/constant"
	"net/http"
)

func TypeValidator() Middleware {
	return func(f http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			typ := chi.URLParam(r, "type")
			if !typeIsValid(typ) {
				http.Error(w, "invalid type parameter", http.StatusBadRequest)
			}
			f.ServeHTTP(w, r)
		})
	}
}

func typeIsValid(typ string) bool {
	return typ == constant.TypeGauge || typ == constant.TypeCounter
}
