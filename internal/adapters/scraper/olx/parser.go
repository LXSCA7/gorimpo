package olx

import (
	"strconv"
	"strings"
	"time"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
)

type jsOffer struct {
	Link       string   `json:"link"`
	Title      string   `json:"title"`
	Price      string   `json:"price"`
	Image      string   `json:"image"`
	Tags       []string `json:"tags"`
	IsFeatured bool     `json:"isFeatured"`
	PostDate   string   `json:"postDate"`
}

func parsePrice(p string) float64 {
	p = strings.ReplaceAll(p, "R$", "")
	p = strings.ReplaceAll(p, ".", "")
	p = strings.ReplaceAll(p, ",", ".")
	p = strings.TrimSpace(p)
	val, _ := strconv.ParseFloat(p, 64)
	return val
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

func (o *Adapter) mapToDomain(items []jsOffer) []domain.Offer {
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
