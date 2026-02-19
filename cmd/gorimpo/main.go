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
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

var Version = "dev"

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
		AddSource:  true,
	}))
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		logger.Warn("Ficheiro .env não encontrado, a usar variáveis de sistema")
	}

	token := os.Getenv("TELEGRAM_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if token == "" || chatID == "" {
		logger.Error("TELEGRAM_TOKEN ou TELEGRAM_CHAT_ID em falta!")
		os.Exit(1)
	}

	telegram := notifier.NewTelegram(token, chatID)

	logger.Info("gorimpo started!", slog.String("version", Version))
	bootMsg := fmt.Sprintf("🟢 <b>GOrimpo %s</b> iniciado e pronto para garimpar!", Version)
	if err := telegram.SendText(bootMsg); err != nil {
		logger.Error("Erro ao enviar mensagem de boot", "erro", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info("booting")

	go func() {
		logger.Info("starting search loop")

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logger.Info("search loop stopped by system")
				return
			case t := <-ticker.C:
				logger.Debug("searching...", slog.Time("tick", t))
			}
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	sig := <-stopChan
	logger.Warn("graceful shutdown...", slog.String("signal", sig.String()))
	shutdownMsg := fmt.Sprintf("🔴 <b>GOrimpo %s</b> fechando!", Version)
	if err := telegram.SendText(shutdownMsg); err != nil {
		logger.Error("Erro ao enviar mensagem de boot", "erro", err)
	}

	cancel()
	time.Sleep(2 * time.Second)

	logger.Info("👋 gorimpo stopped succesfuly. bye!")
}
