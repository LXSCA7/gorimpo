package olx

import (
	"regexp"
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

var (
	olxPostTimeRE = regexp.MustCompile(`\b([01]?\d|2[0-3]):([0-5]\d)\b`)
	olxLocation   = loadOLXLocation()
)

func parsePrice(p string) float64 {
	p = strings.ReplaceAll(p, "R$", "")
	p = strings.ReplaceAll(p, ".", "")
	p = strings.ReplaceAll(p, ",", ".")
	p = strings.TrimSpace(p)
	val, _ := strconv.ParseFloat(p, 64)
	return val
}

func parseOLXDate(dateStr string) time.Time {
	return parseOLXDateAt(dateStr, time.Now())
}

func parseOLXDateAt(dateStr string, now time.Time) time.Time {
	if strings.TrimSpace(dateStr) == "" {
		return now.In(olxLocation)
	}

	now = now.In(olxLocation)
	cleanStr := normalizeOLXDate(dateStr)
	hour, minute := parseOLXTime(cleanStr)

	if strings.Contains(cleanStr, "hoje") {
		return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, olxLocation)
	}

	if strings.Contains(cleanStr, "ontem") {
		yesterday := now.AddDate(0, 0, -1)
		return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), hour, minute, 0, 0, olxLocation)
	}

	months := map[string]time.Month{
		"jan": time.January, "janeiro": time.January,
		"fev": time.February, "fevereiro": time.February,
		"mar": time.March, "marco": time.March, "março": time.March,
		"abr": time.April, "abril": time.April,
		"mai": time.May, "maio": time.May,
		"jun": time.June, "junho": time.June,
		"jul": time.July, "julho": time.July,
		"ago": time.August, "agosto": time.August,
		"set": time.September, "setembro": time.September,
		"out": time.October, "outubro": time.October,
		"nov": time.November, "novembro": time.November,
		"dez": time.December, "dezembro": time.December,
	}

	fields := strings.Fields(cleanStr)
	for i, field := range fields {
		day, err := strconv.Atoi(field)
		if err != nil || day < 1 || day > 31 || i+1 >= len(fields) {
			continue
		}
		month, ok := months[fields[i+1]]
		if !ok {
			continue
		}
		year := now.Year()
		if i+2 < len(fields) {
			if parsedYear, err := strconv.Atoi(fields[i+2]); err == nil && parsedYear >= 2000 {
				year = parsedYear
			}
		}
		parsed := time.Date(year, month, day, hour, minute, 0, 0, olxLocation)
		if parsed.After(now.Add(24 * time.Hour)) {
			parsed = parsed.AddDate(-1, 0, 0)
		}
		return parsed
	}

	return now
}

func normalizeOLXDate(dateStr string) string {
	replacer := strings.NewReplacer(
		",", " ",
		".", " ",
		"postado em", " ",
		"publicado em", " ",
		"às", " ",
		" as ", " ",
		" de ", " ",
	)
	clean := strings.ToLower(strings.TrimSpace(dateStr))
	clean = replacer.Replace(clean)
	return strings.Join(strings.Fields(clean), " ")
}

func parseOLXTime(cleanStr string) (int, int) {
	match := olxPostTimeRE.FindStringSubmatch(cleanStr)
	if len(match) != 3 {
		return 0, 0
	}
	hour, _ := strconv.Atoi(match[1])
	minute, _ := strconv.Atoi(match[2])
	return hour, minute
}

func loadOLXLocation() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("America/Sao_Paulo", -3*60*60)
	}
	return loc
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
