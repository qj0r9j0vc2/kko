package output

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Underline = "\033[4m"

	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Gray    = "\033[90m"
)

var colorEnabled = true

func SetColor(enabled bool) {
	colorEnabled = enabled
}

func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func ShouldColor() bool {
	if !colorEnabled {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return IsTerminal()
}

func Colorize(color, text string) string {
	if !ShouldColor() {
		return text
	}
	return color + text + Reset
}

func Header(text string) string  { return Colorize(Bold+White, text) }
func Label(text string) string   { return Colorize(Cyan, text) }
func Value(text string) string   { return text }
func Muted(text string) string   { return Colorize(Gray, text) }
func Success(text string) string { return Colorize(Green, text) }
func Warning(text string) string { return Colorize(Yellow, text) }
func Error(text string) string   { return Colorize(Red, text) }
func Link(text string) string    { return Colorize(Blue+Underline, text) }

func Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, Error("Error: ")+msg)
}
