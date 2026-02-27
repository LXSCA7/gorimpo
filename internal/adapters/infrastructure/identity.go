package infrastructure

import (
	"fmt"
	"math/rand/v2"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
	"github.com/LXSCA7/gorimpo/internal/core/ports"
)

type RandomUAFactory struct {
	identities []domain.UserAgent
}

var _ ports.IdentityGenerator = (*RandomUAFactory)(nil)

func NewRandomUAFactory(count int) *RandomUAFactory {
	f := &RandomUAFactory{}
	if count == 0 {
		count = 3
	}
	f.identities = make([]domain.UserAgent, count)
	f.generate(count)
	return f
}

func (f *RandomUAFactory) GetRandom() domain.UserAgent {
	if len(f.identities) == 0 {
		return domain.UserAgent{
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
			Browser:   "chromium",
		}
	}
	return f.identities[rand.IntN(len(f.identities))]
}

func (f *RandomUAFactory) generate(count int) {
	osWindows := []string{"Windows NT 10.0; Win64; x64", "Windows NT 11.0; Win64; x64"}
	osMac := []string{"Macintosh; Intel Mac OS X 10_15_7", "Macintosh; Intel Mac OS X 14_2_1", "Macintosh; Intel Mac OS X 13_6"}
	osLinux := []string{"X11; Linux x86_64", "X11; Ubuntu; Linux x86_64"}

	if count == 0 {
		count = 3
	}

	for i := range count {
		engineWeight := rand.IntN(100)
		var engine, ua string

		if engineWeight < 70 {
			engine = "chromium"
			os := pickRandom(append(osWindows, append(osMac, osLinux...)...))

			major := 135 + rand.IntN(10)
			build := rand.IntN(6500)
			patch := rand.IntN(150)
			ua = fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%d.0.%d.%d Safari/537.36",
				os, major, build, patch)

		} else if engineWeight < 90 {
			engine = "firefox"
			os := pickRandom(osWindows)
			if rand.IntN(2) == 0 {
				os = pickRandom(osLinux)
			}

			ver := 140 + rand.IntN(10)
			ua = fmt.Sprintf("Mozilla/5.0 (%s; rv:%d.0) Gecko/20100101 Firefox/%d.0",
				os, ver, ver)

		} else {
			engine = "webkit"
			os := pickRandom(osMac)

			ver := 18 + rand.IntN(3)
			subVer := rand.IntN(5)
			ua = fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/%d.%d Safari/605.1.15",
				os, ver, subVer)
		}

		f.identities[i] = domain.UserAgent{
			UserAgent: ua,
			Browser:   engine,
		}
	}
}

func pickRandom(slice []string) string {
	return slice[rand.IntN(len(slice))]
}
