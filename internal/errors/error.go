package errors

type AlertNotFound struct{}

func (e *AlertNotFound) Error() string {
	return "alert not found in storage"
}
