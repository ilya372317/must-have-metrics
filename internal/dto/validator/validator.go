package validator

import (
	"fmt"
	"sync"

	"github.com/asaskevich/govalidator"
)

var mu = sync.Mutex{}

// Validate make validation on given dataTransferObject.
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

// ValidateRequired validate given dataTransferObject in all fields required mode.
func ValidateRequired(dataTransferObject interface{}) (bool, error) {
	return Validate(dataTransferObject, true)
}
