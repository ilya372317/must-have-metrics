package entity

const GaugeType = "gauge"
const CounterType = "counter"

type Alert struct {
	Type  string
	Name  string
	Value AlertValue
}

type AlertValue interface {
	SetFloatValue(value float64)
	SetIntValue(value int64)
	GetFloatValue() float64
	GetIntValue() int64
	Add(newValue AlertValue) AlertValue
}

type AlertIntValue struct {
	Data int64
}

func (alert *AlertIntValue) Add(newValue AlertValue) AlertValue {
	alert.Data += newValue.GetIntValue()
	return alert
}

func (alert *AlertIntValue) SetFloatValue(value float64) {
	alert.Data = int64(value)
}

func (alert *AlertIntValue) SetIntValue(value int64) {
	alert.Data = value
}

func (alert *AlertIntValue) GetFloatValue() float64 {
	return float64(alert.Data)
}

func (alert *AlertIntValue) GetIntValue() int64 {
	return alert.Data
}

type AlertFloatValue struct {
	Data float64
}

func (alert *AlertFloatValue) SetFloatValue(value float64) {
	alert.Data = value
}

func (alert *AlertFloatValue) SetIntValue(value int64) {
	alert.Data = float64(value)
}

func (alert *AlertFloatValue) GetFloatValue() float64 {
	return alert.Data
}

func (alert *AlertFloatValue) GetIntValue() int64 {
	return int64(alert.Data)
}

func MakeGaugeAlert(name string, data float64) Alert {
	return Alert{
		Type:  GaugeType,
		Name:  name,
		Value: &AlertFloatValue{Data: data},
	}
}

func MakeCounterAlert(name string, data int64) Alert {
	return Alert{
		Type:  CounterType,
		Name:  name,
		Value: &AlertIntValue{Data: data},
	}
}

func (alert *AlertFloatValue) Add(newValue AlertValue) AlertValue {
	alert.Data += newValue.GetFloatValue()
	return alert
}
