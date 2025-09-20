package format

import (
	"github.com/akikareha/himewiki/internal/config"
)

// TODO
func markdown(cfg *config.Config, title string, text string) (string, string, string, string) {
	return nomark(cfg, title, text) // fallback
}
