package ports

import "github.com/LXSCA7/gorimpo/internal/core/domain"

type IdentityGenerator interface {
	GetRandom() domain.UserAgent
}
