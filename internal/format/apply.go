package format

import (
	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/format/nomark"
)

// Apply applies wiki formatting on input text
// and returns title, wiki text, plain text, HTML.
func Apply(cfg *config.Config, title string, text string) (string, string, string, string) {
	if len(text) < 1 {
		return title, "", "", ""
	}
	head := text[0]

	// Detect markup by very first character of input text.
	// * '=' : Creole
	// * '#' : Markdown
	// * Others : Nomark
	if head == '=' {
		// TODO implement Creole
		fc := nomark.ToFormatConfig(cfg)
		return nomark.Apply(fc, title, text)
	} else if head == '#' {
		// TODO implement Markdown
		fc := nomark.ToFormatConfig(cfg)
		return nomark.Apply(fc, title, text)
	} else {
		fc := nomark.ToFormatConfig(cfg)
		return nomark.Apply(fc, title, text)
	}
}
