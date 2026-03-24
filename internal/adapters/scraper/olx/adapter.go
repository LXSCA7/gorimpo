package olx

import (
	"errors"
	"log/slog"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
	"github.com/LXSCA7/gorimpo/internal/core/ports"
)

type Adapter struct {
	isHeadless     bool
	config         ports.ConfigProvider
	identityGen    ports.IdentityGenerator
	proxy          ports.ProxyProvider
	currentProxy   string
	lastScreenshot []byte
	proxySessions  map[string]domain.UserAgent
}

func NewAdapter(isHeadless bool, cfg ports.ConfigProvider, idGen ports.IdentityGenerator, proxy ports.ProxyProvider) *Adapter {
	return &Adapter{
		isHeadless:    isHeadless,
		config:        cfg,
		identityGen:   idGen,
		proxy:         proxy,
		proxySessions: make(map[string]domain.UserAgent),
		currentProxy:  "",
	}
}

var _ ports.VisualScraper = (*Adapter)(nil)

func (o *Adapter) Search(term string) ([]domain.Offer, error) {
	scraperCfg := o.config.Get().Scraper
	o.applyJitter(scraperCfg)

	page, cleanup, err := o.accessOLX(term)
	if err != nil {
		if errors.Is(err, domain.ErrProxyFailure) {
			slog.Warn("⚠️ proxy error")
			o.currentProxy = ""
			return nil, nil
		} else {
			slog.Error("error when accessing olx.", "err", err)
			return nil, err
		}
	}
	defer cleanup()

	if err = o.waitForContent(page); err != nil {
		o.saveLastScreenshot(page)
		slog.Warn("⏳ OLX timed out", "term", term, "proxy", o.currentProxy)

		o.proxy.MarkInvalid(o.currentProxy)
		o.currentProxy = ""

		return nil, nil
	}

	rawOffers, err := o.evaluatePage(page)
	if err != nil {
		return nil, err
	}

	return o.mapToDomain(rawOffers), nil
}

func (o *Adapter) GetLastScreenshot() []byte {
	return o.lastScreenshot
}
