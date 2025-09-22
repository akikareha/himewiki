package filter

import (
	"fmt"

	"github.com/akikareha/himewiki/internal/config"
)

func Apply(cfg *config.Config, title string, content string) (string, error) {
	if cfg.Filter.Agent == "ChatGPT" {
		return withChatGPT(cfg, title, content)
	} else if cfg.Filter.Agent == "nil" {
		return content, nil
	} else {
		return "", fmt.Errorf("Invalid filter agent. If you want to disable filter, set it to \"nil\".")
	}
}

func ImageApply(cfg *config.Config, title string, data []byte) ([]byte, error) {
	if cfg.Filter.Agent == "ChatGPT" {
		return imageWithChatGPT(cfg, title, data)
	} else if cfg.Filter.Agent == "nil" {
		return data, nil
	} else {
		return nil, fmt.Errorf("Invalid image filter agent. If you want to disable filter, set it to \"nil\".")
	}
}
