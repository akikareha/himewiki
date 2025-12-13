package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type Link struct {
	Key string `yaml:"key"`
	URL string `yaml:"url"`
}

type Config struct {
	App struct {
		Mode string `yaml:"mode"`
		Addr string `yaml:"addr"`
	} `yaml:"app"`

	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`

	Vacuum struct {
		CheckEvery     int   `yaml:"check-every"`
		Threshold      int64 `yaml:"threshold"`
		ImageThreshold int64 `yaml:"image-threshold"`
	} `yaml:"vacuum"`

	Site struct {
		Base   string `yaml:"base"`
		Name   string `yaml:"name"`
		Static string `yaml:"static"`
		Card   string `yaml:"card"`
	} `yaml:"site"`

	Wiki struct {
		Front  string `yaml:"front"`
		Format string `yaml:"format"`
	} `yaml:"wiki"`

	Image struct {
		Domains    []string `yaml:"domains"`
		Extensions []string `yaml:"extensions"`
	} `yaml:"image"`

	Filter struct {
		Agent       string  `yaml:"agent"`
		Key         string  `yaml:"key"`
		Temperature float64 `yaml:"temperature"`
		TopP        float64 `yaml:"top_p"`
	} `yaml:"filter"`

	ImageFilter struct {
		Agent     string `yaml:"agent"`
		Key       string `yaml:"key"`
		MaxLength int    `yaml:"max-length"`
		MaxSize   int    `yaml:"max-size"`
	} `yaml:"image-filter"`

	Gnome struct {
		Agent       string  `yaml:"agent"`
		Key         string  `yaml:"key"`
		Temperature float64 `yaml:"temperature"`
		TopP        float64 `yaml:"top_p"`
		Ratio       int     `yaml:"ratio"`
		Recent      int     `yaml:"recent"`
	} `yaml:"gnome"`

	PromptsPath string `yaml:"prompts-path"`

	Prompts *Prompts

	Links []Link `yaml:"links"`
}

func Load(path string) *Config {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	cfg.Prompts = loadPrompts(cfg.PromptsPath)

	return &cfg
}
