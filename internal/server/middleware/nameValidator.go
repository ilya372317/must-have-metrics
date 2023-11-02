package middleware

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func NameValidator() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			name := chi.URLParam(request, "name")
			if !nameIsValid(name) {
				http.Error(writer, "given name is invalid", http.StatusNotFound)
				return
			}

			next.ServeHTTP(writer, request)
		})
	}
}

func nameIsValid(name string) bool {
	return len(name) > 0
}
