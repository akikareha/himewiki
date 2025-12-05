package filter

import (
	"fmt"

	"golang.org/x/text/unicode/norm"

	"github.com/akikareha/himewiki/internal/config"
)

func Apply(cfg *config.Config, title string, content string) (string, error) {
	normTitle := norm.NFC.String(title)
	normContent := norm.NFC.String(content)

	if cfg.Filter.Agent == "openai" {
		filtered, err := withOpenAI(cfg, normTitle, normContent)
		return norm.NFC.String(filtered), err
	} else if cfg.Filter.Agent == "nil" {
		return normContent, nil
	} else {
		return "", fmt.Errorf("Invalid filter agent. If you want to disable filter, set it to \"nil\".")
	}
}

func ImageApply(cfg *config.Config, title string, data []byte) ([]byte, error) {
	normTitle := norm.NFC.String(title)

	if cfg.ImageFilter.Agent == "openai" {
		return imageWithOpenAI(cfg, normTitle, data)
	} else if cfg.ImageFilter.Agent == "nil" {
		return data, nil
	} else {
		return nil, fmt.Errorf("Invalid image filter agent. If you want to disable filter, set it to \"nil\".")
	}
}

func GnomeApply(cfg *config.Config, title string, content string) (string, error) {
	normTitle := norm.NFC.String(title)
	normContent := norm.NFC.String(content)

	if cfg.Gnome.Agent == "openai" {
		filtered, err := gnomeWithOpenAI(cfg, normTitle, normContent)
		return norm.NFC.String(filtered), err
	} else if cfg.Gnome.Agent == "nil" {
		return normContent, nil
	} else {
		return "", fmt.Errorf("Invalid gnome filter agent. If you want to disable filter, set it to \"nil\".")
	}
}
