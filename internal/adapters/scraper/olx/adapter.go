package olx

import (
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
		return nil, err
	}
	defer cleanup()

	if err = o.waitForContent(page); err != nil {
		return nil, err
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
