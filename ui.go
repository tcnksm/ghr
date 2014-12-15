package main

import (
	"fmt"

	"github.com/mitchellh/colorstring"
)

// Colorize color msg.
func Colorize(msg string, color string) (out string) {
	// If color is blank return plain text
	if color == "" {
		return msg
	}

	return colorstring.Color(fmt.Sprintf("[%s]%s[reset]", color, msg))
}

// ColoredError returned red colored msg.
// This is used display error.
func ColoredError(msg string) string {
	return Colorize(msg, "red")
}
