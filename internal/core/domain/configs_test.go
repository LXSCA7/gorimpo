package domain

import (
	"strings"
	"testing"
)

func TestNotificationTemplatesTemplate(t *testing.T) {
	templates := NotificationTemplates{
		NewOffer:       "offer",
		CircuitBreaker: "circuit",
		Error:          "error",
	}

	tests := map[string]string{
		NotificationTemplateNewOffer:       "offer",
		NotificationTemplateCircuitBreaker: "circuit",
		NotificationTemplateError:          "error",
		"unknown":                          "",
	}

	for name, want := range tests {
		if got := templates.Template(name); got != want {
			t.Fatalf("Template(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestRenderNotificationTemplateUsesDefaultWhenCustomIsEmpty(t *testing.T) {
	got, err := RenderNotificationTemplate("", "cooling down for {{.Cooldown}}", NotificationTemplateData{
		Cooldown: "15m0s",
	})
	if err != nil {
		t.Fatalf("RenderNotificationTemplate returned error: %v", err)
	}

	if got != "cooling down for 15m0s" {
		t.Fatalf("rendered message = %q", got)
	}
}

func TestRenderNotificationTemplateUsesCustomTemplate(t *testing.T) {
	got, err := RenderNotificationTemplate("{{.Title}} costs {{.Price}} at {{.Link}}", "default", NotificationTemplateData{
		Title: "Nintendo 64",
		Price: "R$ 250.00",
		Link:  "https://example.com/offer",
	})
	if err != nil {
		t.Fatalf("RenderNotificationTemplate returned error: %v", err)
	}

	for _, want := range []string{"Nintendo 64", "R$ 250.00", "https://example.com/offer"} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered message %q does not contain %q", got, want)
		}
	}
}

func TestRenderNotificationTemplateReturnsParseErrors(t *testing.T) {
	if _, err := RenderNotificationTemplate("{{.Title", "default", NotificationTemplateData{}); err == nil {
		t.Fatal("expected parse error")
	}
}
