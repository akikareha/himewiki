package filter

import (
	"fmt"

	"github.com/akikareha/himewiki/internal/config"
)

func Apply(cfg *config.Config, text string) (string, error) {
	if cfg.Filter.Agent == "ChatGPT" {
		return withChatGPT(cfg, text)
	} else if cfg.Filter.Agent == "nil" {
		return text, nil
	} else {
		return "", fmt.Errorf("Invalid filter agent. If you want to disable filter, set it to \"nil\".")
	}
}
