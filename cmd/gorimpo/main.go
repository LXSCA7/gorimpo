package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/LXSCA7/gorimpo/internal/adapters/notifier"
	"github.com/LXSCA7/gorimpo/internal/adapters/repository"
	"github.com/LXSCA7/gorimpo/internal/adapters/scraper"
	"github.com/LXSCA7/gorimpo/internal/adapters/telemetry"
	"github.com/LXSCA7/gorimpo/internal/config"
	"github.com/LXSCA7/gorimpo/internal/core/services"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var Version = "dev"

func setupLogger() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	}))
	slog.SetDefault(logger)
}

func main() {
	setupLogger()
	_ = godotenv.Load()

	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		slog.Error("Erro ao criar pasta data", "erro", err)
		os.Exit(1)
	}

	cfg, err := config.Load("./config.yaml")
	if err != nil {
		slog.Error("Erro ao carregar configurações", "erro", err)
		os.Exit(1)
	}

	token, chatID := os.Getenv("TELEGRAM_TOKEN"), os.Getenv("TELEGRAM_CHAT_ID")
	if token == "" || chatID == "" {
		slog.Error("Variáveis TELEGRAM_TOKEN ou TELEGRAM_CHAT_ID ausentes")
		os.Exit(1)
	}

	telegram := notifier.NewTelegram(token, chatID)
	olxScraper := scraper.NewOLX(Version != "dev")

	repo, err := repository.NewSQLite("data/gorimpo.db")
	if err != nil {
		slog.Error("Erro ao iniciar o banco de dados", "erro", err)
		os.Exit(1)
	}

	systemSvc := services.NewSystemService(repo, telegram, cfg)
	_ = systemSvc.Setup(Version)

	http.Handle("/metrics", promhttp.Handler())
	metrics := telemetry.NewPrometheusMetrics()
	go func() {
		slog.Info("📈 Servidor de métricas rodando na porta :2112")
		http.ListenAndServe(":2112", nil)
	}()

	gorimpoSvc := services.NewGorimpoService(olxScraper, repo, telegram, metrics, cfg)
	slog.Info("🚀 GOrimpo starting...", slog.String("version", Version))
	gorimpoSvc.Start(Version)

	slog.Info("👋 Sistema encerrado.")
}
