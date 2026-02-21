package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LXSCA7/gorimpo/internal/adapters/notifier"
	"github.com/LXSCA7/gorimpo/internal/adapters/repository"
	"github.com/LXSCA7/gorimpo/internal/adapters/scraper"
	"github.com/LXSCA7/gorimpo/internal/config"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

var Version = "dev"

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	}))
	slog.SetDefault(logger)
	routes := make(map[string]string)

	cfg, err := config.Load("./config.yaml")
	if err != nil {
		panic(err)
	}

	for _, cat := range cfg.Categories {
		if !cfg.App.UseTopics {
			routes[cat] = "0"
		} else {
			if cat == "nintendo" {
				routes[cat] = "3"
			} else {
				routes[cat] = "0"
			}
		}
	}
	_ = godotenv.Load()

	token := os.Getenv("TELEGRAM_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if token == "" || chatID == "" {
		logger.Error("missing TELEGRAM_TOKEN or TELEGRAM_CHAT_ID")
		os.Exit(1)
	}

	telegram := notifier.NewTelegram(token, chatID)
	telegram.SetRoutes(routes)

	olxScraper := scraper.NewOLX()
	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		logger.Error("Erro ao criar pasta data", "erro", err)
		os.Exit(1)
	}

	repo, err := repository.NewSQLite("data/gorimpo.db")
	if err != nil {
		logger.Error("Erro ao iniciar o banco de dados", "erro", err)
		os.Exit(1)
	}

	logger.Info("🚀 GOrimpo starting...", slog.String("version", Version))
	err = telegram.SendText(fmt.Sprintf("🟢 <b>GOrimpo v%s</b> iniciado e pronto a garimpar!", Version), "nintendo")
	if err != nil {
		panic(fmt.Sprintf("erro ao enviar mensagem ao telegram: %v", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		logger.Info("Iniciando rotina de busca...")

		garimpar := func() {
			logger.Debug("⛏️ Abrindo o navegador e cavando na OLX...")
			ofertas, err := olxScraper.Search("Nintendo Switch")
			if err != nil {
				logger.Error("Erro ao garimpar", "erro", err)
				return
			}

			logger.Info("💎 Busca concluída!", "encontrados", len(ofertas))
			novasOfertas := 0

			for _, item := range ofertas {

				existe, err := repo.OfferExists(item.Link)
				if err != nil {
					logger.Error("Erro ao consultar o banco", "erro", err)
					continue
				}

				if existe {
					continue
				}

				if err := telegram.Send(item, "nintendo"); err != nil {
					logger.Error("Erro ao enviar pro Telegram", "erro", err)
					time.Sleep(3 * time.Second)
					continue
				}

				if err := repo.SaveOffer(item); err != nil {
					logger.Error("Erro ao salvar oferta no banco", "erro", err)
				}

				novasOfertas++
				time.Sleep(3 * time.Second)
			}

			if novasOfertas > 0 {
				logger.Info("💎 Novas ofertas enviadas com sucesso!", "quantidade", novasOfertas)
			} else {
				logger.Debug("🤷 Nenhuma oferta nova nessa rodada.")
			}
		}

		garimpar()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				garimpar()
			}
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	<-stopChan
	logger.Warn("Graceful shutdown iniciado...")
	telegram.SendText("🔴 <b>GOrimpo</b> desligando. Fui!", "nintendo")

	cancel()
	time.Sleep(2 * time.Second)
	logger.Info("👋 Sistema encerrado.")
}
