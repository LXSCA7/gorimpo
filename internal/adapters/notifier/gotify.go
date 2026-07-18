package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
	"github.com/LXSCA7/gorimpo/internal/core/ports"
)

var _ ports.Notifier = (*GotifyAdapter)(nil)

type GotifyAdapter struct {
	host      string
	token     string
	apiURL    string
	client    *http.Client
	routes    map[string]string
	templates domain.NotificationTemplates
}

const defaultGotifyNewOfferTemplate = `{{if .SearchTerm}}🔎 Search: {{.SearchTerm}}
{{end}}🚨 New find on {{.Source}}
🎮 {{.Title}}
💰 Price: {{.Price}}
{{if .Date}}🕗 Posted on: {{.Date}}
{{end}}🔗 {{.Link}}`

func NewGotify(host, token string, templates ...domain.NotificationTemplates) *GotifyAdapter {
	normalizedHost := strings.TrimRight(strings.TrimSpace(host), "/")

	return &GotifyAdapter{
		host:      normalizedHost,
		token:     strings.TrimSpace(token),
		apiURL:    fmt.Sprintf("%s/message?token=%s", normalizedHost, strings.TrimSpace(token)),
		client:    &http.Client{Timeout: 10 * time.Second},
		routes:    make(map[string]string),
		templates: firstTemplateConfig(templates),
	}
}

func (g *GotifyAdapter) SetRoutes(routes map[string]string) {
	g.routes = routes
}

func (g *GotifyAdapter) SendText(message, category string) error {
	title := "GOrimpo"
	if category != "" {
		title = fmt.Sprintf("GOrimpo • %s", category)
	}

	payload := map[string]any{
		"title":    title,
		"message":  message,
		"priority": 5,
	}

	return g.doRequest(payload)
}

func (g *GotifyAdapter) Send(offer domain.Offer, category, searchTerm string, showSearchTerm bool) error {
	msg, err := g.formatOfferMessage(offer, searchTerm, showSearchTerm)
	if err != nil {
		return err
	}

	return g.SendText(msg, category)
}

func (g *GotifyAdapter) formatOfferMessage(offer domain.Offer, searchTerm string, showSearchTerm bool) (string, error) {
	if !showSearchTerm {
		searchTerm = ""
	}

	date := ""
	if !offer.PostDate.IsZero() {
		date = formatDate(offer.PostDate)
	}

	return domain.RenderNotificationTemplate(
		g.templates.Template(domain.NotificationTemplateNewOffer),
		defaultGotifyNewOfferTemplate,
		domain.NotificationTemplateData{
			Title:      offer.Title,
			Price:      fmt.Sprintf("R$ %.2f", offer.Price),
			Link:       offer.Link,
			Date:       date,
			Source:     offer.Source,
			SearchTerm: searchTerm,
		},
	)
}

func (g *GotifyAdapter) SendPhoto(data []byte, caption string, category string) error {
	if len(data) == 0 {
		return g.SendText(caption, category)
	}

	message := fmt.Sprintf("%s\n\n📎 Screenshot size: %d bytes", caption, len(data))
	return g.SendText(message, category)
}

func (g *GotifyAdapter) CreateCategory(name string) (string, error) {
	return "0", nil
}

func (g *GotifyAdapter) doRequest(payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal gotify payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, g.apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create gotify request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("send gotify request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("gotify api error: status %d - %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
}
