package nomark

import (
	"github.com/akikareha/himewiki/internal/config"
)

type imageConfig struct {
	domains    []string
	extensions []string
}

type formatConfig struct {
	image imageConfig
	links []config.Link
}

func ToFormatConfig(cfg *config.Config) formatConfig {
	fc := formatConfig{}
	fc.image.domains = cfg.Image.Domains
	fc.image.extensions = cfg.Image.Extensions
	fc.links = cfg.Links
	return fc
}
