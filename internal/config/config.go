package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App struct {
		Mode string `yaml:"mode"`
		Addr string `yaml:"addr"`
	} `yaml:"app"`

	Database struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
		User string `yaml:"user"`
		Password string `yaml:"password"`
		Name string `yaml:"name"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`

	Site struct {
		Name string `yaml:"name"`
		Front string `yaml:"front"`
	} `yaml:"site"`

	Image struct {
		Domains []string `yaml:"domains"`
		Extensions []string `yaml:"extensions"`
	} `yaml:"image"`

	Filter struct {
		Agent string `yaml:"agent"`
		Key string `yaml:"key"`
		System string `yaml:"system"`
		Prompt string `yaml:"prompt"`
	}
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

	return &cfg
}
