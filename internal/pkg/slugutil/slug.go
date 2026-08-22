package slugutil

import (
	"regexp"
	"strings"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a name to a URL-safe slug:
// "Bakso Pedas Spesial!" -> "bakso-pedas-spesial".
func Slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	return strings.Trim(slugRe.ReplaceAllString(s, "-"), "-")
}
