package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type Prompts struct {
	Filter   string `yaml:"filter"`
	Common   string `yaml:"common"`
	Nomark   string `yaml:"nomark"`
	Creole   string `yaml:"creole"`
	Markdown string `yaml:"markdown"`
	Gnome    string `yaml:"gnome"`
}

func loadPrompts(path string) *Prompts {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to load prompts: %v", err)
	}

	var prompts Prompts
	if err := yaml.Unmarshal(data, &prompts); err != nil {
		log.Fatalf("failed to parse prompts: %v", err)
	}

	return &prompts
}
