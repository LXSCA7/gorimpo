package ports

import "github.com/LXSCA7/gorimpo/internal/core/domain"

type Scraper interface {
	Search(term string) ([]domain.Offer, error)
}

type VisualScraper interface {
	Scraper
	GetLastScreenshot() []byte
}
