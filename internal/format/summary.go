package format

import (
	"unicode"
	"unicode/utf8"
)

// TrimForSummary trims text to specified rune length with ".." ellipsis.
// It also compresses all whitespace (including tabs/newlines)
// into single spaces.
func TrimForSummary(text string, limit int) string {
	if limit < 0 {
		panic("program error")
	}
	if limit < 1 {
		return ""
	}

	runeCount := utf8.RuneCountInString(text)

	if limit < 2 {
		if runeCount < 2 {
			return text
		}
		return "."
	}

	var (
		buf       = make([]rune, 0, limit)
		prevSpace = false
		count     = 0
	)

	for _, r := range text {
		if unicode.IsSpace(r) {
			if prevSpace {
				continue
			}
			r = ' '
			prevSpace = true
		} else {
			prevSpace = false
		}

		if count >= limit-2 && count+2 < runeCount {
			return string(buf) + ".."
		}

		buf = append(buf, r)
		count++
	}

	return string(buf)
}
