package format

import (
	"github.com/akikareha/himewiki/internal/config"
)

// TODO
func creole(cfg *config.Config, title string, text string) (string, string, string, string) {
	return nomark(cfg, title, text) // fallback
}
