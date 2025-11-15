package format

import (
	"strings"

	"github.com/akikareha/himewiki/internal/config"
)

// Apply applies wiki formatting on input text
// and returns title, wiki text, plain text, HTML.
func Apply(cfg *config.Config, title string, text string) (string, string, string, string) {
	if len(text) < 1 {
		return title, "", "", ""
	}
	head := text[0]

	fc := toFormatConfig(cfg)

	// Detect markup by very first character of input text.
	// * '=' : Creole
	// * '#' : Markdown
	// * Others : Nomark
	if head == '=' {
		return creole(fc, title, text)
	} else if head == '#' {
		return markdown(fc, title, text)
	} else {
		return nomark(fc, title, text)
	}
}

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
