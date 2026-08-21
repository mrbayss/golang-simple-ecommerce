package phonenumber

import (
	"fmt"
	"strings"
)

// Normalize converts Indonesian local formats to E.164 without '+':
// "0812-3456-789", "+62 812 3456 789", "628123456789" -> "628123456789"
func Normalize(input string) (string, error) {
	var b strings.Builder
	for _, r := range input {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
		// spasi, strip, kurung, '+' dibuang — hanya digit yang disimpan
	}
	s := b.String()

	switch {
	case strings.HasPrefix(s, "62"):
		// sudah format internasional
	case strings.HasPrefix(s, "0"):
		s = "62" + s[1:]
	default:
		return "", fmt.Errorf("nomor telepon harus diawali 08 atau 62")
	}

	if len(s) < 10 || len(s) > 13 {
		return "", fmt.Errorf("panjang nomor telepon tidak valid")
	}
	return s, nil
}
