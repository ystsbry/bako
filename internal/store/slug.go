package store

import (
	"strings"
	"unicode"
)

// Slugify converts an arbitrary string into a filesystem-safe slug consisting
// of lowercase ASCII alphanumerics separated by single hyphens. Runs of other
// characters (including spaces and non-ASCII text such as Japanese) collapse
// to a single hyphen. The result is trimmed of leading/trailing hyphens and
// may be empty when the input has no ASCII alphanumerics.
func Slugify(s string) string {
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		switch {
		case r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)):
			b.WriteRune(unicode.ToLower(r))
			prevHyphen = false
		default:
			if !prevHyphen && b.Len() > 0 {
				b.WriteByte('-')
				prevHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
