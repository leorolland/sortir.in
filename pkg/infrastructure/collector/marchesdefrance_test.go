package collector

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
)

func loadMarchesDeFranceCSV(t *testing.T) []marcheDeFranceRow {
	t.Helper()

	raw, err := os.ReadFile("testdata/marches_de_france.csv")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	rows, err := parseMarchesDeFranceCSV(raw)
	if err != nil {
		t.Fatalf("failed to parse fixture: %v", err)
	}
	return rows
}

func findMarketEvents(t *testing.T, events []application.Event, name string) []application.Event {
	t.Helper()
	matches := []application.Event{}
	for _, event := range events {
		if event.Name == name {
			matches = append(matches, event)
		}
	}
	return matches
}

func TestParseMarchesDeFranceCSV(t *testing.T) {
	rows := loadMarchesDeFranceCSV(t)

	// Only the market without coordinates is dropped at parse time; the one
	// without days is kept but cannot be expanded.
	if len(rows) != 7 {
		t.Fatalf("expected 7 rows, got %d", len(rows))
	}
	first := rows[0]
	if first.nom != "Marché d'Ambérieux" || first.commune != "Ambérieux" {
		t.Errorf("unexpected first row: %+v", first)
	}
	if first.lat != 45.99861 || first.lon != 4.90178 {
		t.Errorf("unexpected coordinates: %v, %v", first.lat, first.lon)
	}
	if first.adresse != "Parking de la salle polyvalente" {
		t.Errorf("unexpected address: %q", first.adresse)
	}
	if rows[5].saison != "Nov-Mar" {
		t.Errorf("unexpected season: %q", rows[5].saison)
	}
}

func TestDisambiguatedNames(t *testing.T) {
	rows := loadMarchesDeFranceCSV(t)
	names := disambiguatedNames(rows)

	expected := []string{
		"Marché d'Ambérieux",
		"Marché hebdomadaire (Ville-A) · Place A",
		"Marché hebdomadaire (Ville-B)",
		"Marché hebdomadaire (Ville-A) · Place A, coin Est",
		"Marché nocturne",
		"Marché d'hiver",
	}
	for i, want := range expected {
		if names[i] != want {
			t.Errorf("name %d: got %q, want %q", i, names[i], want)
		}
	}
}

func TestParseMarketDays(t *testing.T) {
	tests := []struct {
		input    string
		expected []time.Weekday
	}{
		{input: "", expected: nil},
		{input: "   ", expected: nil},
		{input: "Sa", expected: []time.Weekday{time.Saturday}},
		{input: "We,Sa", expected: []time.Weekday{time.Wednesday, time.Saturday}},
		{input: "Tu-Su", expected: []time.Weekday{time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday}},
		{input: "Mo-Sa", expected: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday}},
		{input: "Tu,Th-Sa", expected: []time.Weekday{time.Tuesday, time.Thursday, time.Friday, time.Saturday}},
		{input: "xx,Sa", expected: []time.Weekday{time.Saturday}},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			days := parseMarketDays(test.input)
			if len(days) != len(test.expected) {
				t.Fatalf("expected %d days, got %v", len(test.expected), days)
			}
			for _, day := range test.expected {
				if !days[day] {
					t.Errorf("expected %v in %v", day, days)
				}
			}
		})
	}
}

func TestParseMarketHours(t *testing.T) {
	tests := []struct {
		input     string
		wantBegin int
		wantEnd   int
	}{
		{input: "", wantBegin: 480, wantEnd: 780},
		{input: "08:00-13:00", wantBegin: 480, wantEnd: 780},
		{input: "07:30-12:30", wantBegin: 450, wantEnd: 750},
		{input: "le matin", wantBegin: 480, wantEnd: 780},
		{input: "l'après-midi", wantBegin: 840, wantEnd: 1080},
		{input: "en soirée", wantBegin: 1080, wantEnd: 1320},
		{input: "07:00-14:30 / 07:00-15:00", wantBegin: 420, wantEnd: 900},
		{input: "16:00-20:00 / 08:00-13:00", wantBegin: 480, wantEnd: 1200},
		{input: "20:00-00:00", wantBegin: 1200, wantEnd: 1440},
		{input: "n'importe quoi", wantBegin: 480, wantEnd: 780},
		{input: "08:00 / 09:00", wantBegin: 480, wantEnd: 780},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			begin, end := parseMarketHours(test.input)
			if begin != test.wantBegin || end != test.wantEnd {
				t.Errorf("got %d-%d, want %d-%d", begin, end, test.wantBegin, test.wantEnd)
			}
		})
	}
}

func TestParseMarketSeason(t *testing.T) {
	tests := []struct {
		input string
		dates map[string]bool
	}{
		{input: "", dates: nil},
		{input: "n'importe quoi", dates: nil},
		{input: "Apr 15", dates: nil},
		{input: "Jul-Aug", dates: map[string]bool{
			"2026-07-01": true, "2026-08-16": true, "2026-08-31": true,
			"2026-06-30": false, "2026-10-07": false,
		}},
		{input: "Jun 15-Sep 15", dates: map[string]bool{
			"2026-06-15": true, "2026-09-15": true,
			"2026-06-14": false, "2026-09-16": false,
		}},
		{input: "Nov-Mar", dates: map[string]bool{
			"2026-11-01": true, "2026-12-25": true, "2026-03-31": true,
			"2026-10-07": false, "2026-04-01": false,
		}},
		{input: "Jan,Feb,Mar,Jul,Aug", dates: map[string]bool{
			"2026-01-10": true, "2026-08-20": true,
			"2026-04-10": false, "2026-09-10": false,
		}},
		{input: "Jan-Jun, Sep-Dec", dates: map[string]bool{
			"2026-03-10": true, "2026-10-07": true,
			"2026-07-10": false, "2026-08-10": false,
		}},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			season := parseMarketSeason(test.input)
			for date, want := range test.dates {
				day, err := time.Parse("2006-01-02", date)
				if err != nil {
					t.Fatalf("bad test date: %v", err)
				}
				if got := season == nil || season.allows(day); got != want {
					t.Errorf("%s: got %v, want %v", date, got, want)
				}
			}
		})
	}
}

func TestExpandMarket(t *testing.T) {
	paris, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Fatalf("failed to load Paris timezone: %v", err)
	}

	t.Run("weekly saturday market over two weeks", func(t *testing.T) {
		now := time.Date(2026, 10, 7, 12, 0, 0, 0, paris) // Wednesday
		row := marcheDeFranceRow{slug: "ville-a", nom: "Marché du samedi", jours: "Sa", horaires: "08:00-13:00", lat: 48.1, lon: -1.6}
		events := expandMarket(row, "Marché du samedi", now, paris)

		if len(events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(events))
		}
		if !events[0].Begin.Equal(time.Date(2026, 10, 10, 8, 0, 0, 0, paris)) {
			t.Errorf("unexpected begin: %v", events[0].Begin)
		}
		if !events[0].End.Equal(time.Date(2026, 10, 10, 13, 0, 0, 0, paris)) {
			t.Errorf("unexpected end: %v", events[0].End)
		}
		if !events[1].Begin.Equal(time.Date(2026, 10, 17, 8, 0, 0, 0, paris)) {
			t.Errorf("unexpected second begin: %v", events[1].Begin)
		}
		if events[0].Kind != application.KindMarket {
			t.Errorf("unexpected kind: %v", events[0].Kind)
		}
		if events[0].Source != "https://www.marches-de-france.fr/marche/ville-a.html" {
			t.Errorf("unexpected source URL: %q", events[0].Source)
		}
		if events[0].Place != "" || events[0].Address != "" {
			t.Errorf("unexpected place/address: %q / %q", events[0].Place, events[0].Address)
		}
	})

	t.Run("today's market still running is kept, finished one is dropped", func(t *testing.T) {
		now := time.Date(2026, 10, 10, 14, 0, 0, 0, paris) // Saturday 14h
		row := marcheDeFranceRow{nom: "Marché", jours: "Sa", horaires: "08:00-13:00"}
		if events := expandMarket(row, "Marché", now, paris); len(events) != 1 {
			t.Fatalf("expected 1 remaining event, got %d", len(events))
		}

		now = time.Date(2026, 10, 10, 10, 0, 0, 0, paris) // Saturday 10h
		if events := expandMarket(row, "Marché", now, paris); len(events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(events))
		}
	})

	t.Run("overnight market ends at midnight of the next day", func(t *testing.T) {
		now := time.Date(2026, 7, 8, 12, 0, 0, 0, paris)
		row := marcheDeFranceRow{nom: "Nocturne", jours: "Sa", horaires: "20:00-00:00"}
		events := expandMarket(row, "Nocturne", now, paris)

		if len(events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(events))
		}
		if !events[0].End.Equal(time.Date(2026, 7, 12, 0, 0, 0, 0, paris)) {
			t.Errorf("unexpected end: %v", events[0].End)
		}
	})

	t.Run("out-of-season market is skipped", func(t *testing.T) {
		row := marcheDeFranceRow{nom: "Estival", jours: "Sa", horaires: "08:00-13:00", saison: "Jul-Aug"}

		july := time.Date(2026, 7, 8, 12, 0, 0, 0, paris)
		if events := expandMarket(row, "Estival", july, paris); len(events) != 2 {
			t.Errorf("expected 2 July events, got %d", len(events))
		}

		october := time.Date(2026, 10, 7, 12, 0, 0, 0, paris)
		if events := expandMarket(row, "Estival", october, paris); len(events) != 0 {
			t.Errorf("expected 0 October events, got %d", len(events))
		}
	})

	t.Run("no days is skipped", func(t *testing.T) {
		now := time.Date(2026, 10, 7, 12, 0, 0, 0, paris)
		row := marcheDeFranceRow{nom: "Sans jours", lat: 48.1, lon: -1.6}
		if events := expandMarket(row, "Sans jours", now, paris); len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})
}

func TestMarchesDeFranceRowsToEvents(t *testing.T) {
	paris, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Fatalf("failed to load Paris timezone: %v", err)
	}
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, paris)

	rows := loadMarchesDeFranceCSV(t)
	events := marchesDeFranceRowsToEvents(rows, now, paris)

	// Ambérieux (2 Wednesdays) + 3 Saturday markets (2 each);
	// the nocturne (Jul-Aug) and the winter market (Nov-Mar) are out of season.
	if len(events) != 8 {
		t.Fatalf("expected 8 events, got %d", len(events))
	}

	samediVilleB := findMarketEvents(t, events, "Marché hebdomadaire (Ville-B)")
	if len(samediVilleB) != 2 {
		t.Fatalf("expected 2 Ville-B events, got %d", len(samediVilleB))
	}
	if !samediVilleB[0].Begin.Equal(time.Date(2026, 10, 10, 8, 0, 0, 0, paris)) {
		t.Errorf("empty hours must default to the morning window, got %v", samediVilleB[0].Begin)
	}
	if samediVilleB[0].Address != "Ville-B" {
		t.Errorf("unexpected address: %q", samediVilleB[0].Address)
	}

	for _, name := range []string{"Marché mystère", "Marché perdu", "Marché nocturne", "Marché d'hiver"} {
		if found := findMarketEvents(t, events, name); len(found) != 0 {
			t.Errorf("expected no event for %q, got %d", name, len(found))
		}
	}
}

func TestMarchesDeFranceCollectAll(t *testing.T) {
	raw, err := os.ReadFile("testdata/marches_de_france.csv")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write(raw)
	}))
	defer server.Close()

	collector := &marchesDeFranceCollector{client: server.Client(), csvURL: server.URL}
	events, err := collector.CollectAll()
	if err != nil {
		t.Fatalf("CollectAll failed: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("expected events from fixture")
	}

	failed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failed.Close()

	collector = &marchesDeFranceCollector{client: failed.Client(), csvURL: failed.URL}
	if _, err := collector.CollectAll(); err == nil {
		t.Error("expected an error for a failing source")
	}
}
