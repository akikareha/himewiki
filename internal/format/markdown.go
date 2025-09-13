package format

import (
	"github.com/akikareha/himewiki/internal/config"
)

func Markdown(cfg *config.Config, text string) (string, string) {
	return Nomark(cfg, text) // TODO
}
