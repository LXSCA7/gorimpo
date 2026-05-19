package notifier

import (
	"strings"
	"testing"
	"time"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
)

func testOffer() domain.Offer {
	return domain.Offer{
		Title:    "Nintendo 64",
		Price:    250,
		Link:     "https://example.com/offer",
		Source:   "OLX",
		Tags:     []string{"console", "retro"},
		PostDate: time.Date(2025, time.January, 2, 3, 4, 0, 0, time.UTC),
	}
}

func TestTelegramFormatOfferMessageDefaultTemplate(t *testing.T) {
	adapter := NewTelegram("token", "chat")

	msg, err := adapter.formatOfferMessage(testOffer(), "n64", true)
	if err != nil {
		t.Fatalf("formatOfferMessage returned error: %v", err)
	}

	for _, want := range []string{
		"Search: n64",
		"NEW FIND ON OLX",
		"Nintendo 64",
		"R$ 250.00",
		"console | retro",
		"02/01/2025 at 03:04",
		"https://example.com/offer",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("message %q does not contain %q", msg, want)
		}
	}
}

func TestTelegramFormatOfferMessageCustomTemplate(t *testing.T) {
	adapter := NewTelegram("token", "chat", domain.NotificationTemplates{
		NewOffer: "{{.Title}}|{{.Price}}|{{.Link}}|{{.Date}}",
	})

	msg, err := adapter.formatOfferMessage(testOffer(), "", false)
	if err != nil {
		t.Fatalf("formatOfferMessage returned error: %v", err)
	}

	want := "Nintendo 64|R$ 250.00|https://example.com/offer|02/01/2025 at 03:04"
	if msg != want {
		t.Fatalf("message = %q, want %q", msg, want)
	}
}

func TestGotifyFormatOfferMessageDefaultTemplate(t *testing.T) {
	adapter := NewGotify("https://gotify.example", "token")

	msg, err := adapter.formatOfferMessage(testOffer(), "n64", true)
	if err != nil {
		t.Fatalf("formatOfferMessage returned error: %v", err)
	}

	for _, want := range []string{
		"Search: n64",
		"New find on OLX",
		"Nintendo 64",
		"R$ 250.00",
		"02/01/2025 at 03:04",
		"https://example.com/offer",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("message %q does not contain %q", msg, want)
		}
	}
}

func TestGotifyFormatOfferMessageCustomTemplate(t *testing.T) {
	adapter := NewGotify("https://gotify.example", "token", domain.NotificationTemplates{
		NewOffer: "{{.Title}}: {{.Price}}",
	})

	msg, err := adapter.formatOfferMessage(testOffer(), "", false)
	if err != nil {
		t.Fatalf("formatOfferMessage returned error: %v", err)
	}

	if msg != "Nintendo 64: R$ 250.00" {
		t.Fatalf("message = %q", msg)
	}
}
