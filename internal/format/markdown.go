package format

import (
	"github.com/akikareha/himewiki/internal/config"
)

func Markdown(cfg *config.Config, title string, text string) (string, string, string, string) {
	return Nomark(cfg, title, text) // TODO
}
