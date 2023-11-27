package entity

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

type Alert struct {
	Value interface{}
	Type  string
	Name  string
}

func (a *Alert) UnmarshalJSON(bytes []byte) error {
	type alertAlias Alert

	alertValue := &struct {
		*alertAlias
	}{
		alertAlias: (*alertAlias)(a),
	}

	if err := json.Unmarshal(bytes, alertValue); err != nil {
		return fmt.Errorf("failed unmarshal alert entity: %w", err)
	}

	switch alertValue.Type {
	case TypeCounter:
		_, ok := alertValue.Value.(int64)
		if !ok {
			alertValue.Value = int64(alertValue.Value.(float64))
		}
	case TypeGauge:
		_, ok := alertValue.Value.(float64)
		if !ok {
			alertValue.Value = float64(alertValue.Value.(int64))
		}
	default:
		return errors.New("invalid value in alert")
	}

	return nil
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
