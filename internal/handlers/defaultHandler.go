package handlers

import "net/http"

func DefaultHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "incorrect route", http.StatusBadRequest)
	}
}
