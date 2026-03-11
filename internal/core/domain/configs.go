package domain

type AppSettings struct {
	DefaultNotifier string `yaml:"default_notifier"`
	UseTopics       *bool  `yaml:"use_topics"`
}

type Search struct {
	Term           string   `yaml:"term"`
	MinPrice       float64  `yaml:"min_price"`
	MaxPrice       float64  `yaml:"max_price"`
	Category       string   `yaml:"category"`
	Exclude        []string `yaml:"exclude"`
	ShowSearchTerm bool     `yaml:"show_search_term"`
}

type ScraperSettings struct {
	MinJitter      int `yaml:"min_jitter"`
	MaxJitter      int `yaml:"max_jitter"`
	UserAgentCount int `yaml:"user_agent_count"`
}

type Config struct {
	App        AppSettings     `yaml:"app"`
	Scraper    ScraperSettings `yaml:"scraper"`
	Categories []string        `yaml:"categories"`
	Searches   []Search        `yaml:"searches"`
}
