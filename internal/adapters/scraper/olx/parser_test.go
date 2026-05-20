package olx

import (
	"testing"
	"time"
)

func TestParseOLXDateUsesSaoPauloForRelativeDates(t *testing.T) {
	now := time.Date(2026, time.May, 20, 2, 30, 0, 0, time.UTC)

	got := parseOLXDateAt("Postado em Hoje, 00:12", now)
	want := time.Date(2026, time.May, 19, 0, 12, 0, 0, olxLocation)
	if !got.Equal(want) || got.Location() != olxLocation {
		t.Fatalf("today date = %s (%s), want %s (%s)", got, got.Location(), want, want.Location())
	}

	got = parseOLXDateAt("Postado em Ontem, 23:58", now)
	want = time.Date(2026, time.May, 18, 23, 58, 0, 0, olxLocation)
	if !got.Equal(want) {
		t.Fatalf("yesterday date = %s, want %s", got, want)
	}
}

func TestParseOLXDateHandlesAbsoluteMetadata(t *testing.T) {
	now := time.Date(2026, time.May, 20, 12, 0, 0, 0, olxLocation)

	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{
			name:  "full month with connector",
			input: "Postado em 18 de maio às 09:45",
			want:  time.Date(2026, time.May, 18, 9, 45, 0, 0, olxLocation),
		},
		{
			name:  "abbreviated month",
			input: "18 mai, 09:45",
			want:  time.Date(2026, time.May, 18, 9, 45, 0, 0, olxLocation),
		},
		{
			name:  "explicit year",
			input: "31 dez 2025, 21:10",
			want:  time.Date(2025, time.December, 31, 21, 10, 0, 0, olxLocation),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOLXDateAt(tt.input, now)
			if !got.Equal(tt.want) {
				t.Fatalf("date = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestParseOLXDateRollsFutureMonthBackOneYear(t *testing.T) {
	now := time.Date(2026, time.January, 2, 12, 0, 0, 0, olxLocation)

	got := parseOLXDateAt("Postado em 31 dez, 22:00", now)
	want := time.Date(2025, time.December, 31, 22, 0, 0, 0, olxLocation)
	if !got.Equal(want) {
		t.Fatalf("date = %s, want %s", got, want)
	}
}
