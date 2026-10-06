package collector

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
)

func loadDataTourismeEvents(t *testing.T, filename string) []dataTourismeEvent {
	t.Helper()

	raw, err := os.ReadFile(fmt.Sprintf("testdata/%s", filename))
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var response dataTourismeResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatalf("failed to unmarshal fixture: %v", err)
	}

	events := make([]dataTourismeEvent, 0, len(response.Hits.Hits))
	for _, hit := range response.Hits.Hits {
		events = append(events, hit.Source)
	}
	return events
}

func findDataTourismeEvent(t *testing.T, docs []dataTourismeEvent, uuid string) dataTourismeEvent {
	t.Helper()
	for _, doc := range docs {
		if len(doc.URI) > len(uuid) && doc.URI[len(doc.URI)-len(uuid):] == uuid {
			return doc
		}
	}
	t.Fatalf("event %q not found in fixture", uuid)
	return dataTourismeEvent{}
}

// findEvent returns all events sharing a name.
func findEvent(t *testing.T, events []application.Event, name string) []application.Event {
	t.Helper()
	matches := []application.Event{}
	for _, event := range events {
		if event.Name == name {
			matches = append(matches, event)
		}
	}
	return matches
}

func TestDataTourismeQuery(t *testing.T) {
	location := application.CollectLocation{City: "Rennes", Lat: 48.1173, Lon: -1.6778, Radius: 10.0}

	raw, err := json.Marshal(dataTourismeQuery(location, 100, "2026-10-06"))
	if err != nil {
		t.Fatalf("failed to marshal query: %v", err)
	}
	var query map[string]any
	if err := json.Unmarshal(raw, &query); err != nil {
		t.Fatalf("failed to unmarshal query: %v", err)
	}

	if query["size"] != float64(dataTourismePageSize) || query["from"] != float64(100) {
		t.Errorf("unexpected pagination: size %v, from %v", query["size"], query["from"])
	}
	filters := query["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]any)
	if len(filters) != 3 {
		t.Fatalf("expected 3 filters, got %d", len(filters))
	}
	term := filters[0].(map[string]any)["term"].(map[string]any)
	if term["type"] != dataTourismeEventType {
		t.Errorf("unexpected type filter: %v", term["type"])
	}
	geo := filters[1].(map[string]any)["geo_distance"].(map[string]any)
	if geo["distance"] != "10km" {
		t.Errorf("unexpected geo distance: %v (a bare number is not a distance)", geo["distance"])
	}
	point := geo["isLocatedAt.geoPoint"].(map[string]any)
	if point["lat"] != 48.1173 || point["lon"] != -1.6778 {
		t.Errorf("unexpected geo point: %v", point)
	}
	endDate := filters[2].(map[string]any)["range"].(map[string]any)["takesPlaceAt.endDate"].(map[string]any)
	if endDate["gte"] != "2026-10-06" {
		t.Errorf("unexpected end date filter: %v", endDate["gte"])
	}
}

func TestDataTourismeEventsToEvents(t *testing.T) {
	docs := loadDataTourismeEvents(t, "datatourisme_events.json")
	if len(docs) != 7 {
		t.Fatalf("expected 7 events, got %d", len(docs))
	}

	paris, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Fatalf("failed to load Paris timezone: %v", err)
	}
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, paris)

	events := dataTourismeEventsToEvents(docs, now, paris)

	// 1 atelier + 1 telethon + 18 guided-tour occurrences + 1 concert
	// + 1 balade; the 31-day and 93-day seasonal ranges are dropped.
	if len(events) != 22 {
		t.Fatalf("expected 22 events, got %d", len(events))
	}

	atelier := findEvent(t, events, "Atelier Méditation")
	if len(atelier) != 1 {
		t.Fatalf("expected 1 Atelier Méditation event, got %d", len(atelier))
	}
	if atelier[0].Kind != application.KindWorkshop {
		t.Errorf("unexpected kind: got %q, want %q", atelier[0].Kind, application.KindWorkshop)
	}
	expectedBegin := time.Date(2026, 10, 6, 19, 15, 0, 0, paris)
	if !atelier[0].Begin.Equal(expectedBegin) {
		t.Errorf("unexpected begin: got %v, want %v", atelier[0].Begin, expectedBegin)
	}
	if !atelier[0].End.Equal(expectedBegin.Add(time.Hour)) {
		t.Errorf("unexpected end: got %v", atelier[0].End)
	}
	if atelier[0].Address != "1 Place de la Gare, 35160, Montfort-sur-Meu" {
		t.Errorf("unexpected address: %q", atelier[0].Address)
	}
	if atelier[0].Loc.Lat != 48.1388618 || atelier[0].Loc.Lon != -1.9489967 {
		t.Errorf("unexpected location: %v", atelier[0].Loc)
	}
	if atelier[0].Source != "https://montfort-sur-meu-tourisme.billetweb.fr/atelier-mediation" {
		t.Errorf("unexpected source: %q", atelier[0].Source)
	}
	if len(atelier[0].Genres) != 2 || atelier[0].Genres[0] != "traineeship" || atelier[0].Genres[1] != "practice" {
		t.Errorf("unexpected genres: %v", atelier[0].Genres)
	}

	telethon := findEvent(t, events, "Téléthon : rendez-vous à Landujan !")
	if len(telethon) != 1 {
		t.Fatalf("expected 1 telethon event, got %d", len(telethon))
	}
	// Date-only occurrences are all-day events.
	if !telethon[0].Begin.Equal(time.Date(2026, 10, 10, 0, 0, 0, 0, paris)) {
		t.Errorf("unexpected begin: %v", telethon[0].Begin)
	}
	if !telethon[0].End.Equal(time.Date(2026, 10, 11, 23, 59, 59, 0, paris)) {
		t.Errorf("unexpected end: %v", telethon[0].End)
	}
	if telethon[0].Kind != application.KindFestival {
		t.Errorf("unexpected kind: got %q, want %q", telethon[0].Kind, application.KindFestival)
	}

	visite := findEvent(t, events, "Visite de la coutellerie")
	if len(visite) != 18 {
		t.Fatalf("expected 18 upcoming tour occurrences, got %d", len(visite))
	}
	first := visite[0]
	if !first.Begin.Equal(time.Date(2026, 10, 30, 14, 0, 0, 0, paris)) {
		t.Errorf("unexpected begin: %v", first.Begin)
	}
	if !first.End.Equal(time.Date(2026, 10, 30, 14, 40, 0, 0, paris)) {
		t.Errorf("unexpected end: %v", first.End)
	}

	concert := findEvent(t, events, "Les Negresses Vertes - Concert")
	if len(concert) != 1 {
		t.Fatalf("expected 1 concert event, got %d", len(concert))
	}
	if concert[0].Kind != application.KindConcert {
		t.Errorf("unexpected kind: got %q, want %q", concert[0].Kind, application.KindConcert)
	}
	expectedConcertBegin := time.Date(2026, 11, 27, 20, 30, 0, 0, paris)
	if !concert[0].Begin.Equal(expectedConcertBegin) || !concert[0].End.Equal(expectedConcertBegin) {
		t.Errorf("unexpected times: %v -> %v", concert[0].Begin, concert[0].End)
	}
	if concert[0].Source != "https://www.billetterie-ferme-de-la-bintinais.fr/concert" {
		t.Errorf("unexpected source: %q (first contact without homepage must be skipped)", concert[0].Source)
	}

	if telethon[0].Source != "https://data.datatourisme.fr/38/10e1476a-14e6-372c-85f6-dd2707ba1fde" {
		t.Errorf("expected the data.datatourisme fallback for a contact-less event, got %q", telethon[0].Source)
	}

	balade := findEvent(t, events, "Balade nocturne : \"Les 5 sens d'Halloween\"")
	if len(balade) != 1 {
		t.Fatalf("expected 1 balade event, got %d", len(balade))
	}
	if !balade[0].Begin.Equal(time.Date(2026, 10, 30, 17, 30, 0, 0, paris)) {
		t.Errorf("unexpected begin: %v", balade[0].Begin)
	}
	// A multi-day occurrence without an end time covers its last day.
	if !balade[0].End.Equal(time.Date(2026, 10, 31, 23, 59, 59, 0, paris)) {
		t.Errorf("unexpected end: %v", balade[0].End)
	}
	if balade[0].Address != "35310, Paimpont" {
		t.Errorf("unexpected address without street: %q", balade[0].Address)
	}

	for _, name := range []string{"Collecte Octobre Rose", "Exposition \"Safari Graffiti\""} {
		if dropped := findEvent(t, events, name); len(dropped) != 0 {
			t.Errorf("expected %q seasonal range to be dropped, got %d events", name, len(dropped))
		}
	}
}

func TestDataTourismeEventToEventsEdgeCases(t *testing.T) {
	paris, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Fatalf("failed to load Paris timezone: %v", err)
	}
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, paris)

	t.Run("no coordinates is skipped", func(t *testing.T) {
		doc := dataTourismeEvent{
			URI:          "https://data.datatourisme.fr/38/x",
			Label:        dataTourismeLabel{"@fr": "No coords"},
			TakesPlaceAt: json.RawMessage(`{"startDate": "2026-10-10", "endDate": "2026-10-10"}`),
		}
		if events := dataTourismeEventToEvents(doc, now, paris); len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})

	t.Run("no label is skipped", func(t *testing.T) {
		doc := dataTourismeEvent{
			URI:          "https://data.datatourisme.fr/38/x",
			IsLocatedAt:  &dataTourismeLocation{Geo: &dataTourismeGeo{Latitude: 48.1, Longitude: -1.6}},
			TakesPlaceAt: json.RawMessage(`{"startDate": "2026-10-10", "endDate": "2026-10-10"}`),
		}
		if events := dataTourismeEventToEvents(doc, now, paris); len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})

	t.Run("no occurrences is skipped", func(t *testing.T) {
		doc := dataTourismeEvent{
			URI:         "https://data.datatourisme.fr/38/x",
			Label:       dataTourismeLabel{"@fr": "No dates"},
			IsLocatedAt: &dataTourismeLocation{Geo: &dataTourismeGeo{Latitude: 48.1, Longitude: -1.6}},
		}
		if events := dataTourismeEventToEvents(doc, now, paris); len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})

	t.Run("end before begin is dropped", func(t *testing.T) {
		doc := dataTourismeEvent{
			URI:         "https://data.datatourisme.fr/38/x",
			Label:       dataTourismeLabel{"@fr": "Inverted"},
			IsLocatedAt: &dataTourismeLocation{Geo: &dataTourismeGeo{Latitude: 48.1, Longitude: -1.6}},
			TakesPlaceAt: json.RawMessage(`{
				"startDate": "2026-10-10", "startTime": "-T18:00:00",
				"endDate": "2026-10-10", "endTime": "-T17:00:00"
			}`),
		}
		if events := dataTourismeEventToEvents(doc, now, paris); len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})

	t.Run("malformed time falls back to day bounds", func(t *testing.T) {
		doc := dataTourismeEvent{
			URI:         "https://data.datatourisme.fr/38/x",
			Label:       dataTourismeLabel{"@fr": "Broken time"},
			IsLocatedAt: &dataTourismeLocation{Geo: &dataTourismeGeo{Latitude: 48.1, Longitude: -1.6}},
			TakesPlaceAt: json.RawMessage(`{
				"startDate": "2026-10-10", "startTime": "-Twhenever",
				"endDate": "2026-10-10"
			}`),
		}
		events := dataTourismeEventToEvents(doc, now, paris)
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if !events[0].Begin.Equal(time.Date(2026, 10, 10, 0, 0, 0, 0, paris)) {
			t.Errorf("unexpected begin: %v", events[0].Begin)
		}
		if !events[0].End.Equal(time.Date(2026, 10, 10, 23, 59, 59, 0, paris)) {
			t.Errorf("unexpected end: %v", events[0].End)
		}
	})

	t.Run("unknown sub-types fall back to kind unknown", func(t *testing.T) {
		doc := dataTourismeEvent{
			URI:   "https://data.datatourisme.fr/38/x",
			Label: dataTourismeLabel{"@fr": "Mystery"},
			Type:  []string{"https://www.datatourisme.fr/ontology/core#EntertainmentAndEvent"},
			IsLocatedAt: &dataTourismeLocation{
				Geo:     &dataTourismeGeo{Latitude: 48.1, Longitude: -1.6},
				Address: &dataTourismeAddress{Postal: "35000", Locality: "Rennes"},
			},
			TakesPlaceAt: json.RawMessage(`{"startDate": "2026-10-10", "endDate": "2026-10-10"}`),
		}
		events := dataTourismeEventToEvents(doc, now, paris)
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0].Kind != application.KindUnknown {
			t.Errorf("unexpected kind: got %q, want %q", events[0].Kind, application.KindUnknown)
		}
		if events[0].Address != "35000, Rennes" {
			t.Errorf("unexpected address fallback: %q", events[0].Address)
		}
	})

	t.Run("multilang label falls back to any language", func(t *testing.T) {
		doc := dataTourismeEvent{
			URI:          "https://data.datatourisme.fr/38/x",
			Label:        dataTourismeLabel{"@en": "English only"},
			IsLocatedAt:  &dataTourismeLocation{Geo: &dataTourismeGeo{Latitude: 48.1, Longitude: -1.6}},
			TakesPlaceAt: json.RawMessage(`{"startDate": "2026-10-10", "endDate": "2026-10-10"}`),
		}
		events := dataTourismeEventToEvents(doc, now, paris)
		if len(events) != 1 || events[0].Name != "English only" {
			t.Errorf("unexpected events: %v", events)
		}
	})
}
