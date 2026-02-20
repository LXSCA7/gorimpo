package ports

import "github.com/LXSCA7/gorimpo/internal/core/domain"

type OfferRepository interface {
	OfferExists(link string) (bool, error)
	SaveOffer(offer domain.Offer) error
}
