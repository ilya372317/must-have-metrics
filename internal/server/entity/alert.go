package entity

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

type Alert struct {
	Type       string
	Name       string
	FloatValue *float64
	IntValue   *int64
}

func MakeGaugeAlert(name string, data float64) Alert {
	return Alert{
		Type:       TypeGauge,
		Name:       name,
		FloatValue: &data,
	}
}

func MakeCounterAlert(name string, data int64) Alert {
	return Alert{
		Type:     TypeCounter,
		Name:     name,
		IntValue: &data,
	}
}

func (a *Alert) GetValue() interface{} {
	if a.FloatValue != nil {
		return *a.FloatValue
	}
	if a.IntValue != nil {
		return *a.IntValue
	}
	return nil
}
