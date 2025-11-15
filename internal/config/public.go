package config

type Public struct {
	Site struct {
		Base string
		Name string
		Card string
	}

	Wiki struct {
		Front string
	}

	Image struct {
		Domains    []string
		Extensions []string
	}

	Filter struct {
		Agent       string
		System      string
		Prompt      string
		Temperature float64
		Common      string
	}

	ImageFilter struct {
		Agent     string
		MaxLength int
		MaxSize   int
	}

	Gnome struct {
		Agent       string
		System      string
		Prompt      string
		Temperature float64
		Ratio       int
		Recent      int
	}
}

func Publish(cfg *Config) Public {
	return Public{

		Site: struct {
			Base string
			Name string
			Card string
		}{
			Base: cfg.Site.Base,
			Name: cfg.Site.Name,
			Card: cfg.Site.Card,
		},

		Wiki: struct {
			Front string
		}{
			Front: cfg.Wiki.Front,
		},

		Image: struct {
			Domains    []string
			Extensions []string
		}{
			Domains:    cfg.Image.Domains,
			Extensions: cfg.Image.Extensions,
		},

		Filter: struct {
			Agent       string
			System      string
			Prompt      string
			Temperature float64
			Common      string
		}{
			Agent:       cfg.Filter.Agent,
			System:      cfg.Filter.System,
			Prompt:      cfg.Filter.Prompt,
			Temperature: cfg.Filter.Temperature,
			Common:      cfg.Filter.Common,
		},

		ImageFilter: struct {
			Agent     string
			MaxLength int
			MaxSize   int
		}{
			Agent:     cfg.ImageFilter.Agent,
			MaxLength: cfg.ImageFilter.MaxLength,
			MaxSize:   cfg.ImageFilter.MaxSize,
		},

		Gnome: struct {
			Agent       string
			System      string
			Prompt      string
			Temperature float64
			Ratio       int
			Recent      int
		}{
			Agent:       cfg.Gnome.Agent,
			System:      cfg.Gnome.System,
			Prompt:      cfg.Gnome.Prompt,
			Temperature: cfg.Gnome.Temperature,
			Ratio:       cfg.Gnome.Ratio,
			Recent:      cfg.Gnome.Recent,
		},
	}
}
