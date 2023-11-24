package params

type FlagFunc func(name string, value interface{}, usage string) *interface{}

type Parameter struct {
	Flag     *string
	Argument string
	Value    string
}

func (h *Parameter) GetFlag() *string {
	return h.Flag
}

func (h *Parameter) SetFlag(str *string) {
	h.Flag = str
}

func (h *Parameter) SetValue(s string) {
	h.Value = s
}

func (h *Parameter) GetValue() string {
	return h.Value
}
