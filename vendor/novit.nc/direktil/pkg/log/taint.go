package log

import (
	"novit.nc/direktil/pkg/color"
)

const (
	Normal Taint = iota
	Info
	Warning
	Error
	Fatal
	OK
)

type Taint byte

func (t Taint) Color() color.Color {
	switch t {
	case Info:
		return color.Blue
	case Warning:
		return color.Yellow
	case Error:
		return color.Red
	case Fatal:
		return color.Magenta
	case OK:
		return color.Green
	default:
		return color.None
	}
}
