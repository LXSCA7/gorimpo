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
	host   string
	token  string
	apiURL string
	client *http.Client
	routes map[string]string
}

func NewGotify(host, token string) *GotifyAdapter {
	normalizedHost := strings.TrimRight(strings.TrimSpace(host), "/")

	return &GotifyAdapter{
		host:   normalizedHost,
		token:  strings.TrimSpace(token),
		apiURL: fmt.Sprintf("%s/message?token=%s", normalizedHost, strings.TrimSpace(token)),
		client: &http.Client{Timeout: 10 * time.Second},
		routes: make(map[string]string),
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
	var msg strings.Builder

	if showSearchTerm {
		fmt.Fprintf(&msg, "🔎 Busca: %s\n", searchTerm)
	}

	fmt.Fprintf(&msg, "🚨 Novo achado no %s\n", offer.Source)
	fmt.Fprintf(&msg, "🎮 %s\n", offer.Title)
	fmt.Fprintf(&msg, "💰 Preço: R$ %.2f\n", offer.Price)
	if !offer.PostDate.IsZero() {
		fmt.Fprintf(&msg, "🕗 Postado em: %s\n", formatDate(offer.PostDate))
	}
	fmt.Fprintf(&msg, "🔗 %s", offer.Link)

	return g.SendText(msg.String(), category)
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
