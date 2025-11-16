package format

import "strings"

// TrimForSummary trims text to specified length with ellipsis.
// Spaces are compressed.
func TrimForSummary(text string, length int) string {
	if length < 0 {
		panic("program error")
	}
	if length < 1 {
		return ""
	}
	if length < 2 {
		return "."
	}

	var buf strings.Builder
	overflowed := false
	space := false
	i := 0
	for _, r := range text {
		if i >= length-2 {
			overflowed = true
			break
		}
		if r == '\r' || r == '\n' || r == '\t' {
			if space {
				i++
				continue
			}
			r = ' '
			space = true
		} else {
			space = false
		}
		buf.WriteRune(r)
		i++
	}

	if overflowed {
		return buf.String() + ".."
	} else {
		return buf.String()
	}
}
