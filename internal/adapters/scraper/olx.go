package scraper

import (
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
	"github.com/LXSCA7/gorimpo/internal/core/ports"
	"github.com/playwright-community/playwright-go"
)

type OLXAdapter struct {
	isHeadless bool
}

func NewOLX(isHeadless bool) *OLXAdapter {
	return &OLXAdapter{
		isHeadless: isHeadless,
	}
}

var _ ports.Scraper = (*OLXAdapter)(nil)

func parsePrice(p string) float64 {
	p = strings.ReplaceAll(p, "R$", "")
	p = strings.ReplaceAll(p, ".", "")
	p = strings.ReplaceAll(p, ",", ".")
	p = strings.TrimSpace(p)
	val, _ := strconv.ParseFloat(p, 64)
	return val
}

func (o *OLXAdapter) Search(term string) ([]domain.Offer, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("não foi possível iniciar o playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(o.isHeadless),
	})
	if err != nil {
		return nil, fmt.Errorf("não foi possível lançar o browser: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		return nil, err
	}

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
