package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/LXSCA7/gorimpo/internal/config"
	"github.com/LXSCA7/gorimpo/internal/core/domain"
	"github.com/LXSCA7/gorimpo/internal/core/ports"
)

type GorimpoService struct {
	scraper       ports.Scraper
	offerRepo     ports.OfferRepository
	notifier      ports.Notifier
	metrics       ports.Metrics
	configManager ports.ConfigManager
}

func NewGorimpoService(
	s ports.Scraper,
	or ports.OfferRepository,
	n ports.Notifier,
	m ports.Metrics,
	c ports.ConfigManager,
) *GorimpoService {
	return &GorimpoService{
		scraper:       s,
		offerRepo:     or,
		notifier:      n,
		metrics:       m,
		configManager: c,
	}
}

func (g *GorimpoService) Start(version string) {
	if version == "dev" {
		version = "vDEV"
	}
	err := g.notifier.SendText(fmt.Sprintf("🟢 <b>GOrimpo %s</b> iniciado e pronto para garimpar!", version), "system")
	if err != nil {
		panic(fmt.Sprintf("error sending message to telegram: %v", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		slog.Info("Starting search routine...")
		g.runCycle()

		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				g.runCycle()
			}
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	<-stopChan
	slog.Warn("Graceful shutdown initiated...")
	g.notifier.SendText("🔴 <b>GOrimpo</b> desligando. Fui!", "system")

	cancel()
	time.Sleep(2 * time.Second)
}

func (g *GorimpoService) runCycle() {
	slog.Info("⛏️ Starting YAML parsing cycle...")

	config := g.configManager.Get()
	for _, search := range config.Searches {
		g.processSearch(search)
		time.Sleep(2 * time.Second)
	}
}

func (g *GorimpoService) processSearch(search config.Search) {
	slog.Debug("🔎 Searching...", "term", search.Term)

	rawOffers, err := g.scraper.Search(search.Term)
	if err != nil {
		slog.Error("Error scraping", "term", search.Term, "error", err)
		return
	}
	g.metrics.RecordScraped(search.Term, len(rawOffers))

	var validOffers []domain.Offer
	discardedByPrice := 0
	discardedByFilter := 0
	duplicated := 0

	for _, offer := range rawOffers {
		if offer.Price < search.MinPrice || (search.MaxPrice > 0 && offer.Price > search.MaxPrice) {
			isNew, _ := g.offerRepo.SaveDiscarded(offer, "price")
			if isNew {
				discardedByPrice++
			} else {
				duplicated++
			}
			continue
		}

		if isExcluded(offer.Title, search.Exclude) {
			isNew, _ := g.offerRepo.SaveDiscarded(offer, "filter")
			if isNew {
				discardedByFilter++
			} else {
				duplicated++
			}
			continue
		}

		validOffers = append(validOffers, offer)
	}

	slog.Info("📊 Summary",
		"term", search.Term,
		"valid_total", len(validOffers),
		"discarded_price_new", discardedByPrice,
		"discarded_filter_new", discardedByFilter,
		"discarded_duplicated", duplicated,
	)

	newOffersCount := 0
	validDuplicated := 0

	for _, offer := range validOffers {
		exists, err := g.offerRepo.OfferExists(offer.Link)
		if err != nil || exists {
			validDuplicated++
			continue
		}

		if err := g.notifier.Send(offer, search.Category); err != nil {
			slog.Error("Error sending to Telegram", "error", err)
			time.Sleep(3 * time.Second)
			continue
		}

		_ = g.offerRepo.SaveOffer(offer)
		newOffersCount++
		time.Sleep(3 * time.Second)
	}

	g.metrics.RecordDiscarded(search.Term, "price", discardedByPrice)
	g.metrics.RecordDiscarded(search.Term, "filter", discardedByFilter)
	g.metrics.RecordValid(search.Term, len(validOffers))
	g.metrics.RecordSent(search.Term, newOffersCount)

	if newOffersCount > 0 {
		slog.Info("💎 Offers sent!", "term", search.Term, "count", newOffersCount)
	} else {
		slog.Debug("🤷 No new offers.", "term", search.Term, "valid_duplicated", validDuplicated)
	}
}

func isExcluded(title string, excludes []string) bool {
	if len(excludes) == 0 {
		return false
	}

	titleLower := strings.ToLower(title)

	for _, word := range excludes {
		if word == "" {
			continue
		}
		if strings.Contains(titleLower, strings.ToLower(word)) {
			return true
		}
	}
	return false
}
