package collector

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
)

func loadInfomaniakTestCards(t *testing.T, filename string) []infomaniakCard {
	t.Helper()

	raw, err := os.ReadFile(fmt.Sprintf("testdata/%s", filename))
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var response struct {
		HTML string `json:"html"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatalf("failed to unmarshal fixture: %v", err)
	}

	return parseInfomaniakCards(response.HTML)
}

func TestParseInfomaniakCards(t *testing.T) {
	cards := loadInfomaniakTestCards(t, "infomaniak_cards.json")

	if len(cards) != 32 {
		t.Fatalf("expected 32 cards, got %d", len(cards))
	}

	first := cards[0]
	if first.Name != "Youness Hanifi - 7 vies" {
		t.Errorf("unexpected first card name: %q", first.Name)
	}
	if first.Type != "ComedyEvent" {
		t.Errorf("unexpected first card type: %q", first.Type)
	}
	expectedBegin := time.Date(2026, 9, 23, 21, 0, 0, 0, time.FixedZone("CEST", 2*3600))
	if !first.Begin.Equal(expectedBegin) {
		t.Errorf("unexpected begin: got %v, want %v", first.Begin, expectedBegin)
	}
	if !first.End.Equal(expectedBegin) {
		t.Errorf("unexpected end: got %v, want %v (end == begin for one-off events)", first.End, expectedBegin)
	}
	if first.URL == "" || first.Img == "" || first.Place == "" {
		t.Errorf("expected url, img and place to be set, got %q %q %q", first.URL, first.Img, first.Place)
	}
	if first.URL != "https://infomaniak.events/fr-ch/humour-et-comedie/youness-hanifi-7-vies/a35fee0a-3c20-4a4f-96f0-9476e23be16d/events/394118" {
		t.Errorf("unexpected first card url: %q", first.URL)
	}
}

func TestParseInfomaniakCardsEmpty(t *testing.T) {
	if cards := parseInfomaniakCards(""); cards != nil {
		t.Errorf("expected nil cards for empty html, got %v", cards)
	}
	if cards := parseInfomaniakCards("<div>not microdata</div>"); len(cards) != 0 {
		t.Errorf("expected no cards, got %v", cards)
	}
}

func TestInfomaniakCardsToEvents(t *testing.T) {
	cards := loadInfomaniakTestCards(t, "infomaniak_cards.json")

	loc := application.EventLocation{Lat: 46.520236381329596, Lon: 6.637344639748335}
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	events := infomaniakCardsToEvents(cards, loc, now)

	// 12 of the 32 cards are recurring-event series spanning more than 15 days
	// and must be dropped.
	if len(events) != 20 {
		t.Fatalf("expected 20 events, got %d", len(events))
	}

	for _, event := range events {
		if event.Loc != loc {
			t.Errorf("unexpected location: got %v, want %v", event.Loc, loc)
		}
		if event.Kind != application.KindTheater {
			t.Errorf("unexpected kind: got %q, want %q", event.Kind, application.KindTheater)
		}
		if len(event.Genres) != 1 {
			t.Errorf("unexpected genres: %v", event.Genres)
		}
		if event.Price != nil {
			t.Errorf("expected no price, got %v", *event.Price)
		}
	}
}

func TestInfomaniakCardsToEventsKinds(t *testing.T) {
	cards := loadInfomaniakTestCards(t, "infomaniak_cards_music.json")

	loc := application.EventLocation{Lat: 46.53332346584648, Lon: 6.545848120003939}
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	events := infomaniakCardsToEvents(cards, loc, now)

	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}

	expectedKinds := map[string]application.Kind{
		"Carrousel":                                              application.KindConcert,
		"Atelier cuisine: légumes de saison":                     application.KindWorkshop,
		"Danser pour le plaisir: Relaxation, danse libre guidée": application.KindParty,
		"Cyberrisques: comment éviter les pièges du numérique":   application.KindWorkshop,
	}
	for _, event := range events {
		expected, ok := expectedKinds[event.Name]
		if !ok {
			t.Errorf("unexpected event name: %q", event.Name)
			continue
		}
		if event.Kind != expected {
			t.Errorf("unexpected kind for %q: got %q, want %q", event.Name, event.Kind, expected)
		}
	}
}

func TestInfomaniakCardsToEventsFilters(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	loc := application.EventLocation{Lat: 46.5, Lon: 6.5}

	tz := time.FixedZone("CEST", 2*3600)
	cases := []struct {
		name    string
		begin   time.Time
		end     time.Time
		wantOK  bool
		wantEnd time.Time
	}{
		{
			name:    "upcoming one-off event",
			begin:   time.Date(2026, 10, 1, 20, 0, 0, 0, tz),
			end:     time.Date(2026, 10, 1, 22, 0, 0, 0, tz),
			wantOK:  true,
			wantEnd: time.Date(2026, 10, 1, 22, 0, 0, 0, tz),
		},
		{
			name:    "missing end falls back to begin",
			begin:   time.Date(2026, 10, 1, 20, 0, 0, 0, tz),
			end:     time.Time{},
			wantOK:  true,
			wantEnd: time.Date(2026, 10, 1, 20, 0, 0, 0, tz),
		},
		{
			name:   "already ended",
			begin:  time.Date(2026, 9, 18, 20, 0, 0, 0, tz),
			end:    time.Date(2026, 9, 18, 22, 0, 0, 0, tz),
			wantOK: false,
		},
		{
			name:   "series spanning more than 15 days",
			begin:  time.Date(2026, 10, 1, 20, 0, 0, 0, tz),
			end:    time.Date(2026, 10, 1, 20, 0, 0, 0, tz).Add(15*24*time.Hour + time.Hour),
			wantOK: false,
		},
		{
			name:    "series spanning exactly 15 days is kept",
			begin:   time.Date(2026, 10, 1, 20, 0, 0, 0, tz),
			end:     time.Date(2026, 10, 1, 20, 0, 0, 0, tz).Add(15 * 24 * time.Hour),
			wantOK:  true,
			wantEnd: time.Date(2026, 10, 1, 20, 0, 0, 0, tz).Add(15 * 24 * time.Hour),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			card := buildInfomaniakCard(t, tc.begin, tc.end)
			events := infomaniakCardsToEvents([]infomaniakCard{card}, loc, now)

			if tc.wantOK {
				if len(events) != 1 {
					t.Fatalf("expected 1 event, got %d", len(events))
				}
				if !events[0].End.Equal(tc.wantEnd) {
					t.Errorf("unexpected end: got %v, want %v", events[0].End, tc.wantEnd)
				}
			} else if len(events) != 0 {
				t.Fatalf("expected 0 events, got %d", len(events))
			}
		})
	}
}

func TestInfomaniakEventsDedupe(t *testing.T) {
	loc := application.EventLocation{Lat: 46.5, Lon: 6.5}
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)

	card := buildInfomaniakCard(t, time.Date(2026, 10, 1, 20, 0, 0, 0, time.UTC), time.Date(2026, 10, 1, 22, 0, 0, 0, time.UTC))
	events := infomaniakCardsToEvents([]infomaniakCard{card, card, card}, loc, now)

	if len(events) != 3 {
		t.Fatalf("expected 3 events before dedupe, got %d", len(events))
	}
}

// buildInfomaniakCard builds a test card with a stable name and URL.
func buildInfomaniakCard(t *testing.T, begin, end time.Time) infomaniakCard {
	t.Helper()

	beginStr := begin.Format(time.RFC3339)
	endStr := ""
	if !end.IsZero() {
		endStr = end.Format(time.RFC3339)
	}
	html := fmt.Sprintf(`<div itemscope itemtype="https://schema.org/ComedyEvent">
		<a itemprop="url" href="https://infomaniak.events/fr-ch/humour-et-comedie/test-event/8299e26e-718b-432f-97a9-a3128f5c75f9/events/398478">
			<h3 itemprop="name">Test Event</h3>
			<meta itemprop="startDate" content="%s">
			<meta itemprop="endDate" content="%s">
		</a>
	</div>`, beginStr, endStr)

	cards := parseInfomaniakCards(html)
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	return cards[0]
}
