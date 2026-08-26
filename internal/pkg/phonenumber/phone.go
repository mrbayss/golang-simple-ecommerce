package phonenumber

import (
	"fmt"
	"strings"
)

// Normalize converts Indonesian local phone formats to E.164 without '+':
// "0812-3456-789", "+62 812 3456 789", "628123456789" -> "628123456789"
func Normalize(input string) (string, error) {
	var b strings.Builder
	for _, r := range input {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	s := b.String()

	switch {
	case strings.HasPrefix(s, "62"):
		// already international format
	case strings.HasPrefix(s, "0"):
		s = "62" + s[1:]
	default:
		return "", fmt.Errorf("phone number must start with 08 or 62")
	}

	if len(s) < 10 || len(s) > 13 {
		return "", fmt.Errorf("invalid phone number length")
	}
	return s, nil
}
