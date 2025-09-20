package format

import (
	"strings"

	"github.com/akikareha/himewiki/internal/config"
)

// Returns title, wiki text, plain text, HTML
func Apply(cfg *config.Config, title string, text string) (string, string, string, string) {
	if len(text) < 1 {
		return title, text, text, ""
	}
	head := text[0]

	// Detect markup by very first character of input text.
	// * '=' : Creole
	// * '#' : Markdown
	// * Others : Nomark
	if head == '=' {
		return creole(cfg, title, text)
	} else if head == '#' {
		return markdown(cfg, title, text)
	} else {
		return nomark(cfg, title, text)
	}
}

func Summarize(s string, n int) string {
	if n < 2 {
		return "."
	}

	var b strings.Builder
	long := false
	space := false
	i := 0
	for _, r := range(s) {
		if i >= n - 2 {
			long = true
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
		b.WriteRune(r)
		i++
	}

	if long {
		return b.String() + ".."
	} else {
		return b.String()
	}
}
