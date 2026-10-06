package collector

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
)

func loadOpenAgendaRecords(t *testing.T, filename string) []openAgendaRecord {
	t.Helper()

	raw, err := os.ReadFile(fmt.Sprintf("testdata/%s", filename))
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var response openAgendaResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatalf("failed to unmarshal fixture: %v", err)
	}

	return response.Results
}

func findOpenAgendaRecord(t *testing.T, records []openAgendaRecord, uid string) openAgendaRecord {
	t.Helper()
	for _, record := range records {
		if record.UID == uid {
			return record
		}
	}
	t.Fatalf("record %q not found in fixture", uid)
	return openAgendaRecord{}
}

func TestParseOpenAgendaRecords(t *testing.T) {
	records := loadOpenAgendaRecords(t, "openagenda_records.json")

	if len(records) != 6 {
		t.Fatalf("expected 6 records, got %d", len(records))
	}

	exposition := findOpenAgendaRecord(t, records, "96691677")
	if exposition.Title != "Exposition de Matthieu Ricard, Vol commercial - Galerie Françoise Livinec" {
		t.Errorf("unexpected exposition title: %q", exposition.Title)
	}
	if exposition.LocationCoords == nil || exposition.LocationCoords.Lat != 48.872728 || exposition.LocationCoords.Lon != 2.315075 {
		t.Errorf("unexpected exposition coordinates: %v", exposition.LocationCoords)
	}
	if exposition.CanonicalURL != "https://openagenda.com/bicentenaire-photographie-fr/events/exposition-de-matthieu-ricard-vol-commercial-galerie-francoise-livinec" {
		t.Errorf("unexpected exposition url: %q", exposition.CanonicalURL)
	}
}

func TestOpenAgendaRecordsToEvents(t *testing.T) {
	records := loadOpenAgendaRecords(t, "openagenda_records.json")
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

	events := openAgendaRecordsToEvents(records, now)

	// 29 upcoming timings across the 6 records; the two JEUDI CHANSON
	// records share a name but not a date, so none is deduplicated.
	if len(events) != 29 {
		t.Fatalf("expected 29 events, got %d", len(events))
	}

	var jeudi, exposition, cabaret, paradoxe, salon *application.Event
	expositionCount := 0
	for i := range events {
		switch {
		case events[i].Name == "JEUDI CHANSON" && events[i].Begin.Month() == time.October:
			jeudi = &events[i]
		case events[i].Name == "JEUDI CHANSON" && events[i].Begin.Year() == 2027:
			// Second JEUDI CHANSON record, covered by the count assertion.
		case events[i].Name == "Les Voisins - cabaret du Père-Lachaise":
			cabaret = &events[i]
		case events[i].Name == "PARADOXE":
			paradoxe = &events[i]
		case events[i].Name == "SALON REUSSIR SANS ATTENDRE":
			salon = &events[i]
		default:
			exposition = &events[i]
			expositionCount++
		}
	}

	if jeudi == nil {
		t.Fatalf("expected the October JEUDI CHANSON event")
	}
	if jeudi.Kind != application.KindConcert {
		t.Errorf("unexpected kind: got %q, want %q", jeudi.Kind, application.KindConcert)
	}
	expectedBegin := time.Date(2026, 10, 15, 20, 30, 0, 0, time.FixedZone("CEST", 2*3600))
	if !jeudi.Begin.Equal(expectedBegin) {
		t.Errorf("unexpected begin: got %v, want %v", jeudi.Begin, expectedBegin)
	}
	if !jeudi.End.Equal(expectedBegin.Add(3 * time.Hour)) {
		t.Errorf("unexpected end: got %v", jeudi.End)
	}
	if jeudi.Price != nil {
		// "Entrée libre : 0€" must not yield a price.
		t.Errorf("expected no price for a free event, got %v", *jeudi.Price)
	}
	if len(jeudi.Genres) != 2 || jeudi.Genres[0] != "Concerts" {
		t.Errorf("unexpected genres: %v", jeudi.Genres)
	}
	if jeudi.Source == "" || !strings.HasPrefix(jeudi.Source, "https://openagenda.com/") {
		t.Errorf("unexpected source: %q", jeudi.Source)
	}
	if jeudi.Place != "La Manufacture Chanson" {
		t.Errorf("unexpected place: %q", jeudi.Place)
	}

	if expositionCount != 23 {
		t.Errorf("expected 23 exposition occurrences, got %d", expositionCount)
	}
	if exposition.Kind != application.KindUnknown {
		t.Errorf("unexpected kind: got %q, want %q", exposition.Kind, application.KindUnknown)
	}

	if cabaret == nil {
		t.Fatalf("expected the cabaret event")
	}
	if cabaret.Price == nil || *cabaret.Price != 16 {
		t.Errorf("unexpected price: %v", cabaret.Price)
	}
	if cabaret.PriceCurrency == nil || *cabaret.PriceCurrency != "EUR" {
		t.Errorf("unexpected currency: %v", cabaret.PriceCurrency)
	}

	if paradoxe == nil {
		t.Fatalf("expected the PARADOXE events")
	}
	if paradoxe.Kind != application.KindTheater {
		t.Errorf("unexpected kind: got %q, want %q", paradoxe.Kind, application.KindTheater)
	}

	if salon == nil {
		t.Fatalf("expected the SALON event")
	}
	if salon.Kind != application.KindUnknown {
		t.Errorf("unexpected kind: got %q, want %q", salon.Kind, application.KindUnknown)
	}
	if len(salon.Genres) != 3 {
		t.Errorf("unexpected genres: %v", salon.Genres)
	}
}

func TestOpenAgendaRecordsToEventsUpcomingOnly(t *testing.T) {
	records := loadOpenAgendaRecords(t, "openagenda_records.json")
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

	events := openAgendaRecordsToEvents(records, now)

	// The October one-off events, the exposition occurrences before
	// November 6th and PARADOXE's first show are all gone.
	if len(events) != 20 {
		t.Fatalf("expected 20 events, got %d", len(events))
	}

	expositionCount := 0
	for _, event := range events {
		if event.Name == "Exposition de Matthieu Ricard, Vol commercial - Galerie Françoise Livinec" {
			expositionCount++
		}
		if event.End.Before(now) {
			t.Errorf("event %q already ended at now: %v", event.Name, event.End)
		}
	}
	if expositionCount != 17 {
		t.Errorf("expected 17 upcoming exposition occurrences, got %d", expositionCount)
	}
}

func TestOpenAgendaRecordToEventsEdgeCases(t *testing.T) {
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	coords := &openAgendaCoordinates{Lat: 47.2184, Lon: -1.5536}

	t.Run("no coordinates is skipped", func(t *testing.T) {
		record := openAgendaRecord{UID: "1", Title: "No coords", FirstDateBegin: "2026-10-15T20:00:00+02:00", FirstDateEnd: "2026-10-15T22:00:00+02:00"}
		if events := openAgendaRecordToEvents(record, now); len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})

	t.Run("no title is skipped", func(t *testing.T) {
		record := openAgendaRecord{UID: "2", LocationCoords: coords, FirstDateBegin: "2026-10-15T20:00:00+02:00", FirstDateEnd: "2026-10-15T22:00:00+02:00"}
		if events := openAgendaRecordToEvents(record, now); len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})

	t.Run("no timings falls back to the date range", func(t *testing.T) {
		record := openAgendaRecord{
			UID:            "3",
			Title:          "Fallback event",
			LocationCoords: coords,
			FirstDateBegin: "2026-10-15T20:00:00+02:00",
			FirstDateEnd:   "2026-10-17T22:00:00+02:00",
		}
		events := openAgendaRecordToEvents(record, now)
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if !events[0].End.Equal(time.Date(2026, 10, 17, 22, 0, 0, 0, time.FixedZone("CEST", 2*3600))) {
			t.Errorf("unexpected end: %v", events[0].End)
		}
	})

	t.Run("date range spanning more than 15 days is skipped", func(t *testing.T) {
		record := openAgendaRecord{
			UID:            "4",
			Title:          "Too long",
			LocationCoords: coords,
			FirstDateBegin: "2026-10-15T20:00:00+02:00",
			FirstDateEnd:   "2027-01-31T22:00:00+01:00",
		}
		if events := openAgendaRecordToEvents(record, now); len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})

	t.Run("unparseable timings falls back to the date range", func(t *testing.T) {
		brokenTimings := "not json"
		record := openAgendaRecord{
			UID:            "5",
			Title:          "Broken timings",
			LocationCoords: coords,
			FirstDateBegin: "2026-10-15T20:00:00+02:00",
			FirstDateEnd:   "2026-10-15T22:00:00+02:00",
			Timings:        &brokenTimings,
		}
		events := openAgendaRecordToEvents(record, now)
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0].Begin.Month() != time.October {
			t.Errorf("unexpected begin: %v", events[0].Begin)
		}
	})

	t.Run("timing without end falls back to begin", func(t *testing.T) {
		timings := `[{"begin": "2026-10-15T20:00:00+02:00"}]`
		record := openAgendaRecord{
			UID:            "6",
			Title:          "No end",
			LocationCoords: coords,
			FirstDateBegin: "2026-10-15T20:00:00+02:00",
			FirstDateEnd:   "2026-10-15T22:00:00+02:00",
			Timings:        &timings,
		}
		events := openAgendaRecordToEvents(record, now)
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if !events[0].End.Equal(events[0].Begin) {
			t.Errorf("expected end to fall back to begin, got %v", events[0].End)
		}
	})

	t.Run("excluded origin agenda is skipped", func(t *testing.T) {
		timings := `[{"begin": "2026-10-15T20:00:00+02:00", "end": "2026-10-15T22:00:00+02:00"}]`
		record := openAgendaRecord{
			UID:             "8",
			Title:           "Job-search assistance",
			LocationCoords:  coords,
			OriginAgendaUID: "38495884",
			FirstDateBegin:  "2026-10-15T20:00:00+02:00",
			FirstDateEnd:    "2026-10-15T22:00:00+02:00",
			Timings:         &timings,
		}
		if events := openAgendaRecordToEvents(record, now); len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})

	t.Run("address is completed with postal code and city", func(t *testing.T) {
		timings := `[{"begin": "2026-10-15T20:00:00+02:00", "end": "2026-10-15T22:00:00+02:00"}]`
		record := openAgendaRecord{
			UID:                "7",
			Title:              "Address event",
			LocationCoords:     coords,
			LocationName:       "Le Lieu unique",
			LocationAddress:    "Quai Ferdinand Favre",
			LocationPostalCode: "44000",
			LocationCity:       "Nantes",
			FirstDateBegin:     "2026-10-15T20:00:00+02:00",
			FirstDateEnd:       "2026-10-15T22:00:00+02:00",
			Timings:            &timings,
		}
		events := openAgendaRecordToEvents(record, now)
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0].Address != "Quai Ferdinand Favre 44000, Nantes" {
			t.Errorf("unexpected address: %q", events[0].Address)
		}
	})
}

func TestOpenAgendaWhere(t *testing.T) {
	location := application.CollectLocation{City: "Paris", Lat: 48.8566, Lon: 2.3522, Radius: 10.0}
	now := time.Date(2026, 10, 6, 10, 0, 0, 0, time.UTC)

	where := openAgendaWhere(location, now)

	expected := "within_distance(location_coordinates, geom'POINT(2.3522 48.8566)', 10km) " +
		"AND location_countrycode = \"FR\" " +
		"AND firstdate_end >= \"2026-10-06T10:00:00Z\""
	if where != expected {
		t.Errorf("unexpected where clause:\ngot  %q\nwant %q", where, expected)
	}
}

func TestOpenAgendaEventsDedupe(t *testing.T) {
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	record := openAgendaRecord{
		UID:            "42",
		Title:          "Same event",
		LocationCoords: &openAgendaCoordinates{Lat: 47.2184, Lon: -1.5536},
		FirstDateBegin: "2026-10-15T20:00:00+02:00",
		FirstDateEnd:   "2026-10-15T22:00:00+02:00",
	}
	timings := `[{"begin": "2026-10-15T20:00:00+02:00", "end": "2026-10-15T22:00:00+02:00"}]`
	record.Timings = &timings

	events := dedupeEvents(openAgendaRecordsToEvents([]openAgendaRecord{record, record}, now))
	if len(events) != 1 {
		t.Fatalf("expected 1 deduplicated event, got %d", len(events))
	}
}
