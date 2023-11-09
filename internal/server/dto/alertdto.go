package dto

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type UpdateAlertDTO struct {
	Type string
	Name string
	Data string
}

func CreateAlertDTOFromRequest(request *http.Request) UpdateAlertDTO {
	typ := chi.URLParam(request, "type")
	name := chi.URLParam(request, "name")
	value := chi.URLParam(request, "value")

	return UpdateAlertDTO{
		Type: typ,
		Name: name,
		Data: value,
	}
}
