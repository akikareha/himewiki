package config

type Public struct {
	Site struct {
		Base string
		Name string
		Card string
	}

	Wiki struct {
		Front  string
		Format string
	}

	Image struct {
		Domains    []string
		Extensions []string
	}

	Filter struct {
		Agent       string
		Temperature float64
		TopP        float64
	}

	ImageFilter struct {
		Agent     string
		MaxLength int
		MaxSize   int
	}

	Gnome struct {
		Agent       string
		Temperature float64
		TopP        float64
		Ratio       int
		Recent      int
	}

	Prompts Prompts
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
			Front  string
			Format string
		}{
			Front:  cfg.Wiki.Front,
			Format: cfg.Wiki.Format,
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
			Temperature float64
			TopP        float64
		}{
			Agent:       cfg.Filter.Agent,
			Temperature: cfg.Filter.Temperature,
			TopP:        cfg.Filter.TopP,
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
			Temperature float64
			TopP        float64
			Ratio       int
			Recent      int
		}{
			Agent:       cfg.Gnome.Agent,
			Temperature: cfg.Gnome.Temperature,
			TopP:        cfg.Gnome.TopP,
			Ratio:       cfg.Gnome.Ratio,
			Recent:      cfg.Gnome.Recent,
		},

		Prompts: *cfg.Prompts,
	}
}
