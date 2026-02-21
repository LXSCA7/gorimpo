package ports

import "github.com/LXSCA7/gorimpo/internal/core/domain"

type Notifier interface {
	SetRoutes(routes map[string]string)
	Send(offer domain.Offer, category string) error
	SendText(message, category string) error
}
