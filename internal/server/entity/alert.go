package entity

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

// Alert entity representing metrics.
type Alert struct {
	IntValue   *int64
	FloatValue *float64
	Type       string
	Name       string
}

// MakeGaugeAlert constructor from create gauge metric.
func MakeGaugeAlert(name string, data float64) Alert {
	return Alert{
		Type:       TypeGauge,
		Name:       name,
		FloatValue: &data,
	}
}

// MakeCounterAlert constructor for create counter metric.
func MakeCounterAlert(name string, data int64) Alert {
	return Alert{
		Type:     TypeCounter,
		Name:     name,
		IntValue: &data,
	}
}

// GetValue return pointer to metric value.
func (a *Alert) GetValue() interface{} {
	if a.FloatValue != nil {
		return *a.FloatValue
	}
	if a.IntValue != nil {
		return *a.IntValue
	}
	return nil
}
