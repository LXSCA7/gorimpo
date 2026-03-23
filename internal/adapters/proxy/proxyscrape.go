package proxy

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/LXSCA7/gorimpo/internal/core/ports"
)

type ProxyscrapeAdapter struct {
	apiURL      string
	pool        []*ProxyStatus
	initialSize int
	mu          sync.Mutex
}

type ProxyStatus struct {
	URL     string
	IsValid bool
}

var _ ports.ProxyProvider = (*ProxyscrapeAdapter)(nil)

func NewProxyscrapeProvider(url string) ports.ProxyProvider {
	return &ProxyscrapeAdapter{apiURL: url, initialSize: 0}
}

func (h *ProxyscrapeAdapter) GetProxy() (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.pool) == 0 {
		if err := h.harvest(); err != nil {
			return "", err
		}
	}

	for len(h.pool) > 0 {
		p := h.pool[0]
		h.pool = h.pool[1:]

		if p.IsValid {
			slog.Info("proxy selected", "proxy url", p.URL)
			return p.URL, nil
		}
	}
	return "", fmt.Errorf("tanque vazio")
}

func (h *ProxyscrapeAdapter) MarkInvalid(proxyURL string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	invalidCount := 0
	for _, p := range h.pool {
		if p.URL == proxyURL {
			p.IsValid = false
		}
		if !p.IsValid {
			invalidCount++
		}
	}

	if invalidCount >= h.initialSize/2 {
		h.pool = nil
	}
}

func (h *ProxyscrapeAdapter) harvest() error {
	resp, err := http.Get(h.apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	list := strings.Split(string(body), "\r\n")

	var cleanList []*ProxyStatus
	for _, p := range list {
		if p != "" {
			cleanList = append(cleanList, &ProxyStatus{URL: p, IsValid: true})
		}
	}

	h.pool = cleanList
	h.initialSize = len(cleanList)
	return nil
}
