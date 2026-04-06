package olx

import (
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
)

func (o *Adapter) applyJitter(scraperCfg domain.ScraperSettings) {
	if scraperCfg.MaxJitter > 0 {
		jitter := rand.IntN(scraperCfg.MaxJitter-scraperCfg.MinJitter+1) + scraperCfg.MinJitter
		slog.Debug("⏱️  Applying Jitter", "seconds", jitter)
		time.Sleep(time.Duration(jitter) * time.Second)
	}

}

func (o *Adapter) getStickyIdentity(proxyURL string) domain.UserAgent {
	if proxyURL == "" {
		return o.identityGen.GetRandom()
	}

	if ua, ok := o.proxySessions[proxyURL]; ok {
		return ua
	}

	newUA := o.identityGen.GetRandom()
	o.proxySessions[proxyURL] = newUA
	return newUA
}
