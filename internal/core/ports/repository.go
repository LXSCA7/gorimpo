package ports

import "github.com/LXSCA7/gorimpo/internal/core/domain"

type OfferRepository interface {
	OfferExists(link string) (bool, error)
	SaveOffer(offer domain.Offer) error
	SaveDiscarded(offer domain.Offer, reason string) error
}

type SystemRepository interface {
	GetRoute(category string) string
	SaveRoute(category, destID string) error
	GetLastVersion() string
	SetCurrentVersion(version string) error
}
