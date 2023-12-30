package validator

import (
	"fmt"
	"sync"

	"github.com/asaskevich/govalidator"
)

var mu = sync.Mutex{}

func Validate(dataTransferObject interface{}, allFieldReq bool) (bool, error) {
	mu.Lock()
	govalidator.SetFieldsRequiredByDefault(allFieldReq)
	mu.Unlock()
	result, err := govalidator.ValidateStruct(dataTransferObject)
	if err != nil {
		err = fmt.Errorf("sturct is invalid: %w", err)
	}
	return result, err
}

func ValidateRequired(dataTransferObject interface{}) (bool, error) {
	return Validate(dataTransferObject, true)
}
