package olx

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
	"github.com/playwright-community/playwright-go"
)

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

func (o *Adapter) accessOLX(term string) (playwright.Page, func(), error) {
	max := 10
	if o.proxy == nil {
		max = 1
	}

	for i := range max {
		var proxyURL string
		if o.proxy != nil && o.currentProxy == "" {
			proxyURL, _ = o.proxy.GetProxy()
		}

		if o.currentProxy != "" {
			slog.Info("using current proxy", "current_proxy", o.currentProxy)
			proxyURL = o.currentProxy
		}

		userAgent := o.getStickyIdentity(proxyURL)

		page, cleanup, err := o.setupBrowser(userAgent, proxyURL)
		if err != nil {
			return nil, nil, err
		}

		buscaStr := url.QueryEscape(term)
		targetURL := fmt.Sprintf("https://www.olx.com.br/brasil?q=%s&sf=1", buscaStr)

		slog.Info(fmt.Sprintf("🕵️  Acessando a OLX: %s", targetURL))

		if _, err = page.Goto(targetURL, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			Timeout:   playwright.Float(30000),
		}); err != nil {
			o.currentProxy = ""
			page.Close()
			o.proxy.MarkInvalid(proxyURL)
			cleanup()
			o.applyJitter(o.config.Get().Scraper)
			slog.Warn("⚠️ Invalid proxy, trying again...", "current_attempt", i, "max_attempts", max)
			continue
		}

		o.currentProxy = proxyURL
		return page, cleanup, nil
	}
	return nil, nil, domain.ErrProxyFailure
}

func (o *Adapter) evaluatePage(page playwright.Page) ([]jsOffer, error) {
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

func (o *Adapter) waitForContent(page playwright.Page) error {
	slog.Info("⏳ Esperando renderização...")
	time.Sleep(2 * time.Second)

	err := page.Locator("section.olx-adcard").First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(10000),
	})

	if err == nil {
		return nil
	}
	return err
}

func (o *Adapter) setupBrowser(userAgent domain.UserAgent, proxyURL string) (playwright.Page, func(), error) {
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

func (o *Adapter) saveLastScreenshot(page playwright.Page) {
	img, _ := page.Screenshot(playwright.PageScreenshotOptions{
		Type:    playwright.ScreenshotTypeJpeg,
		Quality: playwright.Int(80),
	})
	o.lastScreenshot = img
}
