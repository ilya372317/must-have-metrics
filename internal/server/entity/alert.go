package entity

import "github.com/ilya372317/must-have-metrics/internal/constant"

type Alert struct {
	Type  string
	Name  string
	Value interface{}
}

func MakeGaugeAlert(name string, data float64) Alert {
	return Alert{
		Type:  constant.TypeGauge,
		Name:  name,
		Value: data,
	}
}

func MakeCounterAlert(name string, data int64) Alert {
	return Alert{
		Type:  constant.TypeCounter,
		Name:  name,
		Value: data,
	}
}
