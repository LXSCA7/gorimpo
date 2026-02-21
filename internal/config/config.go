package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AppSettings struct {
	DefaultNotifier string `yaml:"default_notifier"`
	UseTopics       bool   `yaml:"use_topics"`
}

type Search struct {
	Term     string   `yaml:"term"`
	MinPrice float64  `yaml:"min_price"`
	MaxPrice float64  `yaml:"max_price"`
	Category string   `yaml:"category"`
	Exclude  []string `yaml:"exclude"`
}

type Config struct {
	App        AppSettings `yaml:"app"`
	Categories []string    `yaml:"categories"`
	Searches   []Search    `yaml:"searches"`
}

func Load(filepath string) (*Config, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o arquivo de configuracao: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("erro ao fazer o parse do yaml: %v", err)
	}

	return &cfg, nil
}
