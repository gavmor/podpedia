package main

import "strings"

// slug normalizes a string for use in filenames by keeping only alphanumeric characters.
func slug(s string) string {
	if s == "" {
		return "unknown"
	}
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, s)
}
