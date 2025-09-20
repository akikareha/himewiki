package format

import (
	"github.com/akikareha/himewiki/internal/config"
)

func Creole(cfg *config.Config, title string, text string) (string, string, string, string) {
	return Nomark(cfg, title, text) // TODO
}
