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

type Proxy struct {
	Enabled    bool            `yaml:"enabled"`
	Provider   string          `yaml:"provider"`
	Strategies ProxyStrategies `yaml:"strategies"`
}

type ProxyStrategies struct {
	Proxyscrape Proxyscrape `yaml:"proxyscrape"`
}

type Proxyscrape struct {
	URL              string  `yaml:"url"`
	Timeout          int     `yaml:"timeout"`
	RefreshThreshold float64 `yaml:"refresh_threshold"`
}

type FixedProxy struct {
	URL string `yaml:"url"`
}

type Config struct {
	App        AppSettings     `yaml:"app"`
	Scraper    ScraperSettings `yaml:"scraper"`
	Proxy      Proxy           `yaml:"proxy"`
	Categories []string        `yaml:"categories"`
	Searches   []Search        `yaml:"searches"`
}
