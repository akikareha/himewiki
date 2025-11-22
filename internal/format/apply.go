package format

import (
	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/format/creole"
	"github.com/akikareha/himewiki/internal/format/markdown"
	"github.com/akikareha/himewiki/internal/format/nomark"
)

// Apply applies wiki formatting on input text
// and returns title, wiki text, plain text, HTML.
func Apply(cfg *config.Config, title string, text string) (string, string, string, string) {
	// output is empty if input is empty
	if len(text) < 1 {
		return title, "", "", ""
	}

	// detect markup by very first character of input text.
	// * '=' : Creole
	// * '#' : Markdown
	// * '!' : Nomark
	head := text[0]
	if head == '=' {
		fc := creole.ToFormatConfig(cfg)
		return creole.Apply(fc, title, text)
	} else if head == '#' {
		fc := markdown.ToFormatConfig(cfg)
		return markdown.Apply(fc, title, text)
	} else if head == '!' {
		fc := nomark.ToFormatConfig(cfg)
		return nomark.Apply(fc, title, text)
	}

	// no headers
	// Chosen by config
	conf := cfg.Wiki.Format
	if conf == "creole" {
		fc := creole.ToFormatConfig(cfg)
		return creole.Apply(fc, title, text)
	} else if conf == "markdown" {
		fc := markdown.ToFormatConfig(cfg)
		return markdown.Apply(fc, title, text)
	} else if conf == "nomark" {
		fc := nomark.ToFormatConfig(cfg)
		return nomark.Apply(fc, title, text)
	}

	// invalid config
	// Nomark as fallback
	fc := nomark.ToFormatConfig(cfg)
	return nomark.Apply(fc, title, text)
}
