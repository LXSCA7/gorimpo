package services

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
	"github.com/LXSCA7/gorimpo/internal/core/ports"
)

type GorimpoService struct {
	scraper   ports.Scraper
	offerRepo ports.OfferRepository
	notifier  ports.Notifier
	metrics   ports.Metrics
	config    ports.ConfigProvider

	consecutiveErrors     int
	circuitOpenUntil      time.Time
	cycleCount            int
	circuitBreakerCounter int
}

func NewGorimpoService(
	s ports.Scraper,
	or ports.OfferRepository,
	n ports.Notifier,
	m ports.Metrics,
	c ports.ConfigProvider,
) *GorimpoService {
	return &GorimpoService{
		scraper:   s,
		offerRepo: or,
		notifier:  n,
		metrics:   m,
		config:    c,
	}
}

func (g *GorimpoService) Start(version string) {
	if version == "dev" {
		version = "vDEV"
	}
	err := g.notifier.SendText(fmt.Sprintf("🟢 <b>GOrimpo %s</b> iniciado e pronto para garimpar!", version), "system")
	if err != nil {
		panic(fmt.Sprintf("Error sending Notification message : %v", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		slog.Info("Starting search routine...")
		g.runCycle()

		for {
			minutes := 2 + rand.IntN(4)
			seconds := rand.IntN(60)
			waitTime := time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second

			slog.Info("Aguardando próximo ciclo", "tempo_total", waitTime.String())

			select {
			case <-ctx.Done():
				return
			case <-time.After(waitTime):
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

	config := g.config.Get()
	for _, search := range config.Searches {
		if time.Now().Before(g.circuitOpenUntil) {
			slog.Warn("🚧 Circuit Breaker ATIVADO. Pulando ciclo para evitar ban...",
				"volto_em", time.Until(g.circuitOpenUntil).Round(time.Second))
			return
		}
		g.processSearch(search)
		time.Sleep(2 * time.Second)
	}

	g.cycleCount++
	if g.cycleCount >= 50 {
		minutes := 5 + rand.IntN(4)
		slog.Info("♻️  Reciclagem atingida.", "descanso_minutos", minutes)

		g.cycleCount = 0
		time.Sleep(time.Duration(minutes) * time.Minute)
	}
}

func (g *GorimpoService) processSearch(search domain.Search) {
	slog.Debug("🔎 Searching...", "term", search.Term)

	rawOffers, err := g.scraper.Search(search.Term)
	if err != nil {
		g.consecutiveErrors++
		slog.Error("Error scraping", "term", search.Term, "error", err)

		if visual, ok := g.scraper.(ports.VisualScraper); ok {
			if img := visual.GetLastScreenshot(); img != nil {
				g.notifier.SendPhoto(img, "📸 Erro ao buscar: "+search.Term, "system")
			}
		}

		if g.consecutiveErrors >= 3 {
			cooldown := time.Duration((15 + 10*g.circuitBreakerCounter)) * time.Minute

			g.circuitBreakerCounter++
			g.circuitOpenUntil = time.Now().Add(cooldown)
			g.consecutiveErrors = 0

			slog.Warn("🚧 Circuit Breaker ATIVADO. Aplicando descanso...",
				"nivel", g.circuitBreakerCounter,
				"circuit_breaker_cooldown", cooldown,
			)

			msg := fmt.Sprintf("🚧 <b>CIRCUIT BREAKER ATIVADO!</b>\n3 falhas seguidas. Entrando em cooldown de %v.", cooldown)
			g.notifier.SendText(msg, "system")
		}
		return
	}
	if g.circuitBreakerCounter > 0 {
		g.circuitBreakerCounter = 0
	}

	g.metrics.RecordScraped(search.Term, len(rawOffers))

	var validOffers []domain.Offer
	discardedByPrice := 0
	discardedByFilter := 0
	duplicated := 0
	irrelevantFeatured := 0

	for _, offer := range rawOffers {
		if offer.IsFeatured && !g.contains(offer.Title, search.Term) {
			irrelevantFeatured++
			slog.Debug("🚫 Ignorando destaque irrelevante", "title", offer.Title)
			continue
		}

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
		"irrelevant_featured", irrelevantFeatured,
	)

	newOffersCount := 0
	validDuplicated := 0

	for _, offer := range validOffers {
		exists, err := g.offerRepo.OfferExists(offer.Link)
		if err != nil || exists {
			validDuplicated++
			continue
		}

		if err := g.notifier.Send(offer, search.Category, search.Term, search.ShowSearchTerm); err != nil {
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
		sleep := 15 + rand.IntN(31)
		slog.Debug("🤷 No new offers. Waiting...",
			"term", search.Term,
			"valid_duplicated", validDuplicated,
			"seconds", sleep,
		)
		time.Sleep(time.Duration(sleep) * time.Second)
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

func (g *GorimpoService) contains(str, word string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(word))
}
