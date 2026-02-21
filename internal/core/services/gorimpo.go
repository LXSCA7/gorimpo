package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LXSCA7/gorimpo/internal/config"
	"github.com/LXSCA7/gorimpo/internal/core/domain"
	"github.com/LXSCA7/gorimpo/internal/core/ports"
)

type GorimpoService struct {
	scraper   ports.Scraper
	offerRepo ports.OfferRepository
	notifier  ports.Notifier
	config    *config.Config
}

func NewGorimpoService(
	s ports.Scraper,
	or ports.OfferRepository,
	n ports.Notifier,
	c *config.Config,
) *GorimpoService {
	return &GorimpoService{
		scraper:   s,
		offerRepo: or,
		notifier:  n,
		config:    c,
	}
}

func (g *GorimpoService) Start(version string) {
	if version == "dev" {
		version = "vDEV"
	}
	err := g.notifier.SendText(fmt.Sprintf("🟢 <b>GOrimpo %s</b> iniciado e pronto para garimpar!", version), "system")
	if err != nil {
		panic(fmt.Sprintf("erro ao enviar mensagem ao telegram: %v", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		slog.Info("Iniciando rotina de busca...")
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
	slog.Warn("Graceful shutdown iniciado...")
	g.notifier.SendText("🔴 <b>GOrimpo</b> desligando. Fui!", "system")

	cancel()
	time.Sleep(2 * time.Second)
}

func (g *GorimpoService) runCycle() {
	slog.Info("⛏️ Iniciando ciclo de garimpo do YAML...")

	for _, search := range g.config.Searches {
		g.processSearch(search)
		time.Sleep(2 * time.Second)
	}
}

func (g *GorimpoService) processSearch(search config.Search) {
	slog.Debug("🔎 Buscando...", "termo", search.Term)

	ofertasBrutas, err := g.scraper.Search(search.Term)
	if err != nil {
		slog.Error("Erro ao garimpar", "termo", search.Term, "erro", err)
		return
	}

	var ofertasValidas []domain.Offer
	descartadasPreco := 0

	for _, item := range ofertasBrutas {
		if item.Price < search.MinPrice || item.Price > search.MaxPrice {
			descartadasPreco++
			continue
		}
		ofertasValidas = append(ofertasValidas, item)
	}

	slog.Info("📊 Resumo", "termo", search.Term, "validas", len(ofertasValidas), "descartadas", descartadasPreco)

	novasOfertas := 0
	for _, item := range ofertasValidas {
		existe, err := g.offerRepo.OfferExists(item.Link)
		if err != nil || existe {
			continue
		}

		if err := g.notifier.Send(item, search.Category); err != nil {
			slog.Error("Erro ao enviar pro Telegram", "erro", err)
			time.Sleep(3 * time.Second)
			continue
		}

		_ = g.offerRepo.SaveOffer(item)
		novasOfertas++
		time.Sleep(3 * time.Second)
	}

	if novasOfertas > 0 {
		slog.Info("💎 Ofertas enviadas!", "termo", search.Term, "qtd", novasOfertas)
	} else {
		slog.Debug("🤷 Nenhuma oferta nova.", "termo", search.Term)
	}
}
