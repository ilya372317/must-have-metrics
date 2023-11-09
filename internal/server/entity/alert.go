package entity

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

type Alert struct {
	Value interface{}
	Type  string
	Name  string
}

func MakeGaugeAlert(name string, data float64) Alert {
	return Alert{
		Type:  TypeGauge,
		Name:  name,
		Value: data,
	}
}

func MakeCounterAlert(name string, data int64) Alert {
	return Alert{
		Type:  TypeCounter,
		Name:  name,
		Value: data,
	}
}
