package middleware

type IncorrectPath struct{}

func (e *IncorrectPath) Error() string {
	return "Incorrect url"
}
