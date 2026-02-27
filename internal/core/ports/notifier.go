package ports

import "github.com/LXSCA7/gorimpo/internal/core/domain"

type Notifier interface {
	SetRoutes(routes map[string]string)
	Send(offer domain.Offer, category, searchTerm string, showSearchTerm bool) error
	SendText(message, category string) error
	SendPhoto(data []byte, caption string, category string) error
	CreateCategory(name string) (string, error)
}
