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
	reset  = "\033[0m"
)

// Fatal error (red)
func Fatal(format string, a ...any) {
	fmt.Fprintf(os.Stderr, red+"[ERROR] "+format+reset+"\n", a...)
	os.Exit(1)
}

// Success (blue neon)
func Success(format string, a ...any) {
	fmt.Printf(blue+"[SUCCESS] "+format+reset+"\n", a...)
}

// In progress / running (green)
func Info(format string, a ...any) {
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
