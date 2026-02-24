package ports

import "github.com/LXSCA7/gorimpo/internal/core/domain"

type ConfigProvider interface {
	Get() *domain.Config
}

type ConfigWatcher interface {
	Watch()
}

type ConfigManager interface {
	ConfigProvider
	ConfigWatcher
}
