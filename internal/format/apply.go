package format

import (
	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/format/creole"
	"github.com/akikareha/himewiki/internal/format/markdown"
	"github.com/akikareha/himewiki/internal/format/nomark"
)

// Detect detects format mode from text and config.
func Detect(cfg *config.Config, text string) string {
	if len(text) > 0 {
		// detect markup by very first character of input text.
		// * '=' : Creole
		// * '#' : Markdown
		// * '!' : Nomark
		head := text[0]
		if head == '=' {
			return "creole"
		} else if head == '#' {
			return "markdown"
		} else if head == '!' {
			return "nomark"
		}
	}

	// no headers
	// Chosen by config
	conf := cfg.Wiki.Format
	if conf == "creole" || conf == "markdown" || conf == "nomark" {
		return conf
	}

	// invalid config
	// Nomark as fallback
	return "nomark"
}

// Apply applies wiki formatting on input text
// and returns title, wiki text, plain text, HTML.
func Apply(cfg *config.Config, title string, text string) (string, string, string, string) {
	mode := Detect(cfg, text)
	if mode == "creole" {
		fc := creole.ToFormatConfig(cfg)
		return creole.Apply(fc, title, text)
	} else if mode == "markdown" {
		fc := markdown.ToFormatConfig(cfg)
		return markdown.Apply(fc, title, text)
	} else {
		fc := nomark.ToFormatConfig(cfg)
		return nomark.Apply(fc, title, text)
	}
}
