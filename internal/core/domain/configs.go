package domain

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type AppSettings struct {
	DefaultNotifier string `yaml:"default_notifier"`
	UseTopics       *bool  `yaml:"use_topics"`
}

const (
	NotificationTemplateNewOffer       = "new_offer"
	NotificationTemplateCircuitBreaker = "circuit_breaker"
	NotificationTemplateError          = "error"
)

type NotifierSettings struct {
	Templates NotificationTemplates `yaml:"templates"`
}

type NotificationTemplates struct {
	NewOffer       string `yaml:"new_offer"`
	CircuitBreaker string `yaml:"circuit_breaker"`
	Error          string `yaml:"error"`
}

func (t NotificationTemplates) Template(name string) string {
	switch name {
	case NotificationTemplateNewOffer:
		return t.NewOffer
	case NotificationTemplateCircuitBreaker:
		return t.CircuitBreaker
	case NotificationTemplateError:
		return t.Error
	default:
		return ""
	}
}

type NotificationTemplateData struct {
	Title      string
	Price      string
	Link       string
	Date       string
	Cooldown   string
	Error      string
	Source     string
	SearchTerm string
	Tags       string
}

func RenderNotificationTemplate(customTemplate, defaultTemplate string, data NotificationTemplateData) (string, error) {
	templateText := defaultTemplate
	if strings.TrimSpace(customTemplate) != "" {
		templateText = customTemplate
	}

	tmpl, err := template.New("notification").Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("parse notification template: %w", err)
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("execute notification template: %w", err)
	}

	return out.String(), nil
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
	App        AppSettings      `yaml:"app"`
	Notifier   NotifierSettings `yaml:"notifier"`
	Scraper    ScraperSettings  `yaml:"scraper"`
	Proxy      Proxy            `yaml:"proxy"`
	Categories []string         `yaml:"categories"`
	Searches   []Search         `yaml:"searches"`
}
