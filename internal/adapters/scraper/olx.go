package scraper

import (
	"encoding/json"
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
	identityGen    ports.IdentityGenerator
	proxy          ports.ProxyProvider
	lastScreenshot []byte
	proxySessions  map[string]domain.UserAgent
}

func NewOLX(isHeadless bool, cfg ports.ConfigProvider, idGen ports.IdentityGenerator, proxy ports.ProxyProvider) *OLXAdapter {
	return &OLXAdapter{
		isHeadless:    isHeadless,
		config:        cfg,
		identityGen:   idGen,
		proxy:         proxy,
		proxySessions: make(map[string]domain.UserAgent),
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

/*
	func (o *OLXAdapter) _Search(term string) ([]domain.Offer, error) {
		scraperCfg := o.config.Get().Scraper
		o.applyJitter(scraperCfg)

		page, cleanup, err := o.accessOLX(term)
		if err != nil {
			return nil, err
		}
		defer cleanup()

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

		result, err := page.Locator("section.olx-adcard").EvaluateAll()
		if err != nil {
			return nil, fmt.Errorf("erro ao extrair dados via JS: %v", err)
		}

		type jsOffer struct {
			Link       string   `json:"link"`
			Title      string   `json:"title"`
			Price      string   `json:"price"`
			Image      string   `json:"image"`
			Tags       []string `json:"tags"`
			IsFeatured bool     `json:"isFeatured"`
			PostDate   string   `json:"postDate"`
		}

		bytes, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("error on marshal olx results: %v", err)
		}

		var tempItems []jsOffer
		if err := json.Unmarshal(bytes, &tempItems); err != nil {
			return nil, fmt.Errorf("erro ao unmarshal de resultados: %v", err)
		}

		var ofertas []domain.Offer
		for _, item := range tempItems {
			postDate := parseOLXDate(item.PostDate)
			if item.Link != "" && strings.Contains(item.Link, "olx.com.br") {
				ofertas = append(ofertas, domain.Offer{
					Title:      item.Title,
					Price:      parsePrice(item.Price),
					Link:       item.Link,
					Source:     "OLX",
					ImageURL:   item.Image,
					Tags:       item.Tags,
					IsFeatured: item.IsFeatured,
					PostDate:   postDate,
				})
			}
		}

		return ofertas, nil
	}
*/
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

func (o *OLXAdapter) applyJitter(scraperCfg domain.ScraperSettings) {
	if scraperCfg.MaxJitter > 0 {
		jitter := rand.IntN(scraperCfg.MaxJitter-scraperCfg.MinJitter+1) + scraperCfg.MinJitter
		slog.Debug("⏱️  Aplicando Jitter", "segundos", jitter)
		time.Sleep(time.Duration(jitter) * time.Second)
	}

}

func (o *OLXAdapter) accessOLX(term string) (playwright.Page, func(), error) {
	max := 50
	if o.proxy == nil {
		max = 1
	}

	for range max {
		var proxyURL string
		if o.proxy != nil {
			proxyURL, _ = o.proxy.GetProxy()
		}

		userAgent := o.getStickyIdentity(proxyURL)

		page, cleanup, err := o.setupBrowser(userAgent, proxyURL)
		if err != nil {
			return nil, nil, err
		}

		buscaStr := url.QueryEscape(term)
		targetURL := fmt.Sprintf("https://www.olx.com.br/brasil?q=%s&sf=1", buscaStr)

		slog.Info(fmt.Sprintf("🕵️  Acessando a OLX: %s\n", targetURL))

		if _, err = page.Goto(targetURL, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			Timeout:   playwright.Float(30000),
		}); err != nil {
			page.Close()
			o.proxy.MarkInvalid(proxyURL)
			cleanup()
			continue
		}

		return page, cleanup, nil
	}
	return nil, nil, fmt.Errorf("max attempts.")
}

func (o *OLXAdapter) evaluatePage(page playwright.Page) ([]jsOffer, error) {
	result, err := page.Locator("section.olx-adcard").EvaluateAll(olxScraperScript)
	if err != nil {
		return nil, fmt.Errorf("erro JS: %v", err)
	}

	var items []jsOffer
	bytes, _ := json.Marshal(result)
	if err := json.Unmarshal(bytes, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (o *OLXAdapter) waitForContent(page playwright.Page) error {
	slog.Info("⏳ Esperando renderização...")
	time.Sleep(2 * time.Second)

	return page.Locator("section.olx-adcard").First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(10000),
	})
}

func (o *OLXAdapter) mapToDomain(items []jsOffer) []domain.Offer {
	var offers []domain.Offer
	for _, item := range items {
		if item.Link == "" || !strings.Contains(item.Link, "olx.com.br") {
			continue
		}

		offers = append(offers, domain.Offer{
			Title:      item.Title,
			Price:      parsePrice(item.Price),
			Link:       item.Link,
			Source:     "OLX",
			ImageURL:   item.Image,
			Tags:       item.Tags,
			IsFeatured: item.IsFeatured,
			PostDate:   parseOLXDate(item.PostDate),
		})
	}
	return offers
}

func (o *OLXAdapter) getStickyIdentity(proxyURL string) domain.UserAgent {
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

func (o *OLXAdapter) setupBrowser(userAgent domain.UserAgent, proxyURL string) (playwright.Page, func(), error) {
	pw, err := playwright.Run(&playwright.RunOptions{})

	if err != nil {
		return nil, nil, fmt.Errorf("não foi possível iniciar o playwright: %v", err)
	}

	var browser playwright.Browser
	launchOptions := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(o.isHeadless),
	}

	if proxyURL != "" {
		slog.Debug("🌐 Configurando proxy no navegador", "proxy", proxyURL)
		launchOptions.Proxy = &playwright.Proxy{
			Server: proxyURL,
		}
	}

	switch userAgent.Browser {
	case "chromium":
		launchOptions.Args = []string{"--disable-blink-features=AutomationControlled"}
		browser, err = pw.Chromium.Launch(launchOptions)
		slog.Info("🌐  UserAgent selecionado", "user_agent", userAgent.UserAgent)
	case "firefox":
		browser, err = pw.Firefox.Launch(launchOptions)
		slog.Info("🦊  UserAgent selecionado", "user_agent", userAgent.UserAgent)
	default:
		browser, err = pw.WebKit.Launch(launchOptions)
		slog.Info("🧭  UserAgent selecionado", "user_agent", userAgent.UserAgent)
	}

	if err != nil {
		pw.Stop()
		return nil, nil, fmt.Errorf("não foi possível lançar o browser: %v", err)
	}

	browserContext, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(userAgent.UserAgent),
		ExtraHttpHeaders: map[string]string{
			"Accept-Language": "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7",
			"Connection":      "keep-alive",
		},
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

func parseOLXDate(dateStr string) time.Time {
	now := time.Now()
	cleanStr := strings.ReplaceAll(strings.ToLower(dateStr), ",", "")

	parts := strings.Split(strings.ToLower(dateStr), ", ")
	timePart := "00:00"
	if len(parts) > 1 {
		timePart = parts[1]
	}

	t, _ := time.Parse("15:04", timePart)

	if strings.Contains(dateStr, "hoje") {
		return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
	}

	if strings.Contains(dateStr, "ontem") {
		yesterday := now.AddDate(0, 0, -1)
		return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
	}

	months := map[string]time.Month{
		"jan": time.January, "fev": time.February, "mar": time.March,
		"abr": time.April, "mai": time.May, "jun": time.June,
		"jul": time.July, "ago": time.August, "set": time.September,
		"out": time.October, "nov": time.November, "dez": time.December,
	}

	fields := strings.Fields(cleanStr)
	if len(fields) >= 3 {
		dia, _ := strconv.Atoi(fields[0])
		mesStr := fields[2]
		if mes, ok := months[mesStr]; ok {
			return time.Date(now.Year(), mes, dia, t.Hour(), t.Minute(), 0, 0, now.Location())
		}
	}

	return now
}

const olxScraperScript = `elements => {
   return elements.map(el => {
      const linkEl = el.querySelector('a[data-testid="adcard-link"]');
      const titleEl = el.querySelector('.olx-adcard__title');
      const priceEl = el.querySelector('h3');
      const imgEl = el.querySelector('img');
      const dateEl = el.querySelector('.olx-adcard__date');
      
      const badgeElements = Array.from(el.querySelectorAll('.olx-adcard__badges .olx-core-badge'));
      const tags = badgeElements.map(badge => badge.innerText.trim());

      const featuredBadge = el.querySelector('.olx-adcard__primary-badge');
      const isFeatured = featuredBadge ? featuredBadge.innerText.includes("Destaque") : false;

      return {
         link: linkEl ? linkEl.href : "",
         title: titleEl ? titleEl.innerText.trim() : "",
         price: priceEl ? priceEl.innerText.trim() : "",
         image: imgEl ? (imgEl.src || imgEl.getAttribute('data-src') || "") : "",
         postDate: dateEl ? dateEl.innerText.trim() : "",
         tags: tags,
         isFeatured: isFeatured
      };
   }).filter(item => item.price !== "" && item.title !== "");
}`

type jsOffer struct {
	Link       string   `json:"link"`
	Title      string   `json:"title"`
	Price      string   `json:"price"`
	Image      string   `json:"image"`
	Tags       []string `json:"tags"`
	IsFeatured bool     `json:"isFeatured"`
	PostDate   string   `json:"postDate"`
}
