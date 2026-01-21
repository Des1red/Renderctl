package logger

import (
	"fmt"
	"os"
)

// ANSI color codes
const (
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[38;5;45m"
	purple = "\033[35m"
	cyan   = "\033[36m"

	gray  = "\033[90m"
	white = "\033[97m"

	reset = "\033[0m"
)

var verbose bool

func SetVerbose(v bool) {
	verbose = v
}

// User interaction / prompt
func Prompt(format string, a ...any) {
	fmt.Printf(white+"[PROMPT] "+format+reset, a...)
}

// Runtime state / banner
func Status(format string, a ...any) {
	fmt.Printf(gray+"[STATUS] "+format+reset+"\n", a...)
}

// Fatal error (red)
func Error(format string, a ...any) {
	fmt.Fprintf(os.Stderr, red+"[ERROR] "+format+reset+"\n", a...)
	os.Exit(0)
}

// Task finished (neutral)
func Done(format string, a ...any) {
	fmt.Printf(cyan+"[DONE] "+format+reset+"\n", a...)
}

// Success (blue neon)
func Success(format string, a ...any) {
	fmt.Printf(blue+"[SUCCESS] "+format+reset+"\n", a...)
}

// In progress / running (green)
func Info(format string, a ...any) {
	if !verbose {
		return
	}
	fmt.Printf(green+"[INFO] "+format+reset+"\n", a...)
}

// Final result / summary (purple)
func Result(format string, a ...any) {
	fmt.Printf(purple+"[RESULT] "+format+reset+"\n", a...)
}

// Notification / warning (yellow)
func Notify(format string, a ...any) {
	fmt.Printf(yellow+"[NOTICE] "+format+reset+"\n", a...)
}
