// Package output provides utilities for formatted terminal output,
// including colored text and styled messages.
package output

import "fmt"

// ANSI color codes for terminal text styling.
// These codes work on most modern terminals (Linux, macOS, Windows 10+).
const (
	ColorReset  = "\033[0m"  // Reset to default color
	ColorRed    = "\033[31m" // Red text (errors, failures)
	ColorGreen  = "\033[32m" // Green text (success, fast responses)
	ColorYellow = "\033[33m" // Yellow text (warnings, slow responses)
	ColorBlue   = "\033[34m" // Blue text (informational)
	ColorCyan   = "\033[36m" // Cyan text (exceptional performance)
)

// Green wraps the given text in ANSI green color codes.
func Green(text string) string {
	return colorize(text, ColorGreen)
}

// Red wraps the given text in ANSI red color codes.
func Red(text string) string {
	return colorize(text, ColorRed)
}

// Yellow wraps the given text in ANSI yellow color codes.
func Yellow(text string) string {
	return colorize(text, ColorYellow)
}

// Blue wraps the given text in ANSI blue color codes.
func Blue(text string) string {
	return colorize(text, ColorBlue)
}

// Cyan wraps the given text in ANSI cyan color codes.
// Use this for exceptional performance indicators.
func Cyan(text string) string {
	return colorize(text, ColorCyan)
}

// colorize is a helper function that wraps text with the specified
// color code and automatically resets the color at the end.
func colorize(text, color string) string {
	return fmt.Sprintf("%s%s%s", color, text, ColorReset)
}
