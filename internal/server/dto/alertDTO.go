package dto

import (
	"net/http"
	"strings"
)

const typePartPosition = 1
const namePartPosition = 2
const valuePartPosition = 3

type UpdateAlertDTO struct {
	Type string
	Name string
	Data string
}

func CreateAlertDTOFromRequest(request *http.Request) UpdateAlertDTO {
	urlParts := strings.Split(request.URL.Path, "/")
	urlPartsWithoutEmptyValue := make([]string, 0, cap(urlParts))
	for _, part := range urlParts {
		if strings.TrimSpace(part) != "" {
			urlPartsWithoutEmptyValue = append(urlPartsWithoutEmptyValue, part)
		}
	}
	return UpdateAlertDTO{
		Type: urlPartsWithoutEmptyValue[typePartPosition],
		Name: urlPartsWithoutEmptyValue[namePartPosition],
		Data: urlPartsWithoutEmptyValue[valuePartPosition],
	}
}
