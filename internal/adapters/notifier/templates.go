package notifier

import "github.com/LXSCA7/gorimpo/internal/core/domain"

func firstTemplateConfig(templates []domain.NotificationTemplates) domain.NotificationTemplates {
	if len(templates) == 0 {
		return domain.NotificationTemplates{}
	}

	return templates[0]
}
