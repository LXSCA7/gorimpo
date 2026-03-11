package services

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/LXSCA7/gorimpo/internal/core/ports"
)

type SystemService struct {
	repo          ports.SystemRepository
	notifier      ports.Notifier
	configManager ports.ConfigManager
}

func NewSystemService(r ports.SystemRepository, n ports.Notifier, c ports.ConfigManager) *SystemService {
	return &SystemService{
		repo:          r,
		notifier:      n,
		configManager: c,
	}
}

func (s *SystemService) Setup(currentVersion string) map[string]string {
	routes, newTopics := s.setupRoutes()
	s.notifier.SetRoutes(routes)
	s.checkVersion(currentVersion)
	if len(newTopics) > 0 {
		var msg strings.Builder
		msg.WriteString("<b>✨ Novos tópicos configurados:</b>\n")
		for _, cat := range newTopics {
			fmt.Fprintf(&msg, "• <code>%s</code>\n", cat)
		}
		_ = s.notifier.SendText(msg.String(), "system")
	}
	return routes
}

func (s *SystemService) checkVersion(currentVersion string) {
	lastVersion := s.repo.GetLastVersion()

	if lastVersion != "" && lastVersion != currentVersion {
		slog.Info("🎉 Atualização detectada!", "old", lastVersion, "new", currentVersion)

		changelogMsg := fmt.Sprintf(
			"🚀 <b>GOrimpo Atualizado com Sucesso!</b>\n\n"+
				"De: <code>%s</code>\nPara: <code>%s</code>\n\n"+
				"🔗 <a href=\"https://github.com/LXSCA7/gorimpo/releases\">Ver Changelog</a>",
			lastVersion, currentVersion,
		)
		_ = s.notifier.SendText(changelogMsg, "system")
	}

	_ = s.repo.SetCurrentVersion(currentVersion)
}

func (s *SystemService) setupRoutes() (map[string]string, []string) {
	routes := make(map[string]string)
	newTopics := []string{}
	config := s.configManager.Get()

	useTelegramTopics := (strings.EqualFold(config.App.DefaultNotifier, "telegram") && (config.App.UseTopics != nil && *config.App.UseTopics))

	slog.Info("🗺️ Configurando rotas de notificação por categoria...")

	categories := []string{"system"}
	categories = append(categories, config.Categories...)
	for _, category := range categories {
		if !useTelegramTopics {
			routes[category] = "0"
			continue
		}

		destID := s.repo.GetRoute(category)
		if destID == "" {
			slog.Info("✨ Criando novo tópico no Telegram...", "categoria", category)

			newID, err := s.notifier.CreateCategory(category)
			if err != nil {
				slog.Error("Erro ao criar tópico, jogando pro Geral", "erro", err)
				newID = "0"
			} else {
				_ = s.repo.SaveRoute(category, newID)
				newTopics = append(newTopics, category)
			}
			destID = newID
		}
		routes[category] = destID
	}

	return routes, newTopics
}
