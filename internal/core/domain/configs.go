package domain

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
