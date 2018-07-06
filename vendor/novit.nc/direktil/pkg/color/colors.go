package color

import "io"

const (
	None  Color = ""
	Reset Color = "\033[0m"
	Bold  Color = "\033[1m"

	Red     Color = "\033[91m"
	Green   Color = "\033[92m"
	Yellow  Color = "\033[93m"
	Blue    Color = "\033[94m"
	Magenta Color = "\033[95m"
	Cyan    Color = "\033[96m"
	White   Color = "\033[97m"

	// Aligned versions (yes, I'm like that)

	Red____ Color = Red
	Green__ Color = Green
	Yellow_ Color = Yellow
	Blue___ Color = Blue
	Cyan___ Color = Cyan
	White__ Color = White

	// Extra colors
	DarkGreen Color = "\033[32m"
	DarkGray  Color = "\033[90m"
)

type Color string

func Write(out io.Writer, color Color, message string) {
	out.Write([]byte(string(color) + message + string(Reset)))
}
