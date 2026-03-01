package wizard

import "fmt"

// FormatMenuLabel renders a two-column menu label: [tag] + aligned text.
// width controls the fixed width for the [tag] column; values <= 0 default to 12.
func FormatMenuLabel(tag, text string, width int) string {
	if width <= 0 {
		width = 12
	}
	left := ""
	if tag != "" {
		left = "[" + tag + "]"
	}
	return fmt.Sprintf("%-*s %s", width, left, text)
}

// Colorize wraps a string with an ANSI color code and resets formatting.
// Pass empty color to keep the text unchanged.
func Colorize(text, color string) string {
	if color == "" {
		return text
	}
	return color + text + "\033[0m"
}
