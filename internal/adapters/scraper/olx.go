package scraper

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
	"github.com/LXSCA7/gorimpo/internal/core/ports"
	"github.com/playwright-community/playwright-go"
)

type OLXAdapter struct {
	isHeadless     bool
	config         ports.ConfigProvider
	lastScreenshot []byte
}

func NewOLX(isHeadless bool, cfg ports.ConfigProvider) *OLXAdapter {
	return &OLXAdapter{
		isHeadless: isHeadless,
		config:     cfg,
	}
}

var _ ports.VisualScraper = (*OLXAdapter)(nil)

func parsePrice(p string) float64 {
	p = strings.ReplaceAll(p, "R$", "")
	p = strings.ReplaceAll(p, ".", "")
	p = strings.ReplaceAll(p, ",", ".")
	p = strings.TrimSpace(p)
	val, _ := strconv.ParseFloat(p, 64)
	return val
}

func (o *OLXAdapter) Search(term string) ([]domain.Offer, error) {
	scraperCfg := o.config.Get().Scraper
	o.applyJitter(scraperCfg)
	userAgent := o.getUserAgent(scraperCfg)
	page, cleanup, err := o.setupBrowser(userAgent)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	buscaStr := url.QueryEscape(term)
	targetURL := fmt.Sprintf("https://www.olx.com.br/brasil?q=%s&sf=1", buscaStr)

	slog.Info(fmt.Sprintf("🕵️  Acessando a OLX: %s\n", targetURL))

	if _, err = page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(30000),
	}); err != nil {
		return nil, err
	}

	slog.Info("⏳ Esperando a OLX renderizar as ofertas...")
	time.Sleep(2 * time.Second)

	err = page.Locator("section.olx-adcard").First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		o.saveLastScreenshot(page)
		return nil, fmt.Errorf("anúncios reais não renderizaram: %v", err)
	}

	result, err := page.Locator("section.olx-adcard").EvaluateAll(`elements => {
		return elements.map(el => {
			// Caça os elementos de dentro do card principal
			const linkEl = el.querySelector('a');
			const titleEl = el.querySelector('h2');
			const priceEl = el.querySelector('h3');
			const imgEl = el.querySelector('img');
			
			return {
				link: linkEl ? linkEl.href : "",
				title: titleEl ? titleEl.innerText.trim() : "",
				price: priceEl ? priceEl.innerText.trim() : "",
				image: imgEl ? (imgEl.src || imgEl.getAttribute('data-src') || "") : ""
			};
		}).filter(item => item.price !== "" && item.title !== ""); // Só devolve se for um anúncio válido
	}`)
	if err != nil {
		return nil, fmt.Errorf("erro ao extrair dados via JS: %v", err)
	}

	var ofertas []domain.Offer
	items := result.([]interface{})

	for _, item := range items {
		data := item.(map[string]interface{})

		link := data["link"].(string)
		title := data["title"].(string)
		priceStr := data["price"].(string)
		image := data["image"].(string)

		if link != "" && strings.Contains(link, "olx.com.br") {
			ofertas = append(ofertas, domain.Offer{
				Title:    title,
				Price:    parsePrice(priceStr),
				Link:     link,
				Source:   "OLX",
				ImageURL: image,
			})
		}
	}

	return ofertas, nil
}

func (o *OLXAdapter) GetLastScreenshot() []byte {
	return o.lastScreenshot
}

func (o *OLXAdapter) saveLastScreenshot(page playwright.Page) {
	img, _ := page.Screenshot(playwright.PageScreenshotOptions{
		Type:    playwright.ScreenshotTypeJpeg,
		Quality: playwright.Int(80),
	})
	o.lastScreenshot = img
}

func (o *OLXAdapter) getUserAgent(scraperCfg domain.ScraperSettings) string {
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36" // Fallback seguro
	if len(scraperCfg.UserAgents) > 0 {
		userAgent = scraperCfg.UserAgents[rand.IntN(len(scraperCfg.UserAgents))]
	}
	slog.Debug("🕵️‍♂️ User-Agent", "ua", userAgent)
	return userAgent
}

func (o *OLXAdapter) applyJitter(scraperCfg domain.ScraperSettings) {
	if scraperCfg.MaxJitter > 0 {
		jitter := rand.IntN(scraperCfg.MaxJitter-scraperCfg.MinJitter+1) + scraperCfg.MinJitter
		slog.Debug("⏱️  Aplicando Jitter", "segundos", jitter)
		time.Sleep(time.Duration(jitter) * time.Second)
	}

}

func (o *OLXAdapter) setupBrowser(userAgent string) (playwright.Page, func(), error) {
	pw, err := playwright.Run(&playwright.RunOptions{})

	if err != nil {
		return nil, nil, fmt.Errorf("não foi possível iniciar o playwright: %v", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(o.isHeadless),
	})
	if err != nil {
		pw.Stop()
		return nil, nil, fmt.Errorf("não foi possível lançar o browser: %v", err)
	}

	browserContext, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(userAgent),
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
	})
	if err != nil {
		pw.Stop()
		browser.Close()
		return nil, nil, fmt.Errorf("erro ao criar contexto do browser: %v", err)
	}

	page, err := browserContext.NewPage()
	if err != nil {
		browserContext.Close()
		browser.Close()
		pw.Stop()
		return nil, nil, err
	}

	close := func() {
		browserContext.Close()
		browser.Close()
		pw.Stop()
	}

	return page, close, nil
}
