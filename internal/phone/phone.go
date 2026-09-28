// Package phone normalizes phone numbers to a canonical E.164-style form so
// that "(555) 123-4567" and "555-123-4567" map to the same unique customer id.
package phone

import (
	"fmt"
	"regexp"
	"strings"
)

var nonDigit = regexp.MustCompile(`\D`)

// Normalize strips formatting and returns an E.164-style string (e.g. +15551234567).
// A bare 10-digit number is assumed to be US (+1). An 11-digit number starting
// with 1 is treated as US as well. Other lengths are accepted as international.
func Normalize(raw string) (string, error) {
	digits := nonDigit.ReplaceAllString(raw, "")
	switch {
	case len(digits) == 10:
		return "+1" + digits, nil
	case len(digits) == 11 && strings.HasPrefix(digits, "1"):
		return "+" + digits, nil
	case len(digits) >= 8 && len(digits) <= 15:
		return "+" + digits, nil
	default:
		return "", fmt.Errorf("invalid phone number: %q", raw)
	}
}
