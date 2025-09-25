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
		Domains []string
		Extensions []string
	}

	Filter struct {
		Agent string
		System string
		Prompt string
	}

	ImageFilter struct {
		Agent string
		MaxLength string
		MaxSize string
	}
}

func Publish(cfg *Config) Public {
	return Public {

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
			Domains []string
			Extensions []string
		}{
			Domains: cfg.Image.Domains,
			Extensions: cfg.Image.Extensions,
		},

		Filter: struct {
			Agent string
			System string
			Prompt string
		}{
			Agent: cfg.Filter.Agent,
			System: cfg.Filter.System,
			Prompt: cfg.Filter.Prompt,
		},

		ImageFilter: struct {
			Agent string
			MaxLength string
			MaxSize string
		}{
			Agent: cfg.ImageFilter.Agent,
			MaxLength: cfg.ImageFilter.MaxLength,
			MaxSize: cfg.ImageFilter.MaxSize,
		},
	}
}
