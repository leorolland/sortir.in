package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
	_ "time/tzdata" // guarantee Europe/Paris is available in stripped deployments

	"github.com/leorolland/sortir.in/pkg/application"
)

// DATAtourisme events ("Fêtes et Manifestations") via the anonymous
// Elasticsearch endpoint that powers the public explorer
// (https://explore.datatourisme.fr). The national portal aggregates the
// tourist offices data of the whole French territory under an open license
// (~45k event POIs, 12% of its ~375k POIs).
//
// The official REST API (https://api.datatourisme.fr/v1/docs, free key after
// registration, 1000 req/hour) exposes the same index; the explorer endpoint
// was preferred because it needs no key and supports the same geo + type +
// date filters. If it ever breaks, switch to the official API.
const (
	dataTourismeSearchURL = "https://explore.datatourisme.fr/api/_search"
	// The index stores ontology URIs; term-filtering on the full URI matches
	// every event sub-type (concert, festival, market, ...).
	dataTourismeEventType = "https://www.datatourisme.fr/ontology/core#EntertainmentAndEvent"
	dataTourismePageSize  = 100
	// Explorer limit: from+size must stay under the 10000 result window.
	dataTourismeMaxPages = 50

	dataTourismeParisLocation = "Europe/Paris"
)

type dataTourismeCollector struct {
	client *http.Client
}

func NewDataTourismeCollector() application.Collector {
	return &dataTourismeCollector{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// flexString accepts a JSON string, a list of strings or null and keeps the
// first non-empty value.
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*f = flexString(flexJoin(v, false))
	return nil
}

// addressLines accepts a JSON string or a list of strings, joined with ", ".
type addressLines []string

func (a *addressLines) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	if line := flexJoin(v, false); line != "" {
		*a = addressLines{line}
		return nil
	}
	if list, ok := v.([]any); ok {
		for _, item := range list {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				*a = append(*a, strings.TrimSpace(s))
			}
		}
	}
	return nil
}

// flexJoin renders a JSON value (string or list of strings) as a string,
// either keeping the first non-empty item or joining them all.
func flexJoin(v any, joinAll bool) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case []any:
		parts := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				parts = append(parts, strings.TrimSpace(s))
				if !joinAll {
					break
				}
			}
		}
		return strings.Join(parts, ", ")
	}
	return ""
}

// dataTourismeLabel is a multilingual object {"@fr": "...", "@en": ...}; the
// values can be strings or lists of strings.
type dataTourismeLabel map[string]flexString

func (l *dataTourismeLabel) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*l = dataTourismeLabel{}
	if obj, ok := v.(map[string]any); ok {
		for key, value := range obj {
			(*l)[key] = flexString(flexJoin(value, false))
		}
	} else if s := flexJoin(v, false); s != "" {
		(*l)[""] = flexString(s)
	}
	return nil
}

// French returns the French label, falling back to any non-empty label.
func (l dataTourismeLabel) French() string {
	if fr, ok := l["@fr"]; ok && fr != "" {
		return string(fr)
	}
	for _, value := range l {
		if value != "" {
			return string(value)
		}
	}
	return ""
}

type dataTourismeGeo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type dataTourismeAddress struct {
	Street addressLines `json:"streetAddress"`
	Postal flexString   `json:"postalCode"`
	City   *struct {
		Label dataTourismeLabel `json:"label"`
	} `json:"hasAddressCity"`
	Locality flexString `json:"addressLocality"`
}

type dataTourismeLocation struct {
	Geo     *dataTourismeGeo     `json:"geo"`
	Address *dataTourismeAddress `json:"address"`
}

// dataTourismeOccurrence mirrors takesPlaceAt entries; the times are local
// France times written as "-T14:30:00" (the date part is elided) and can be
// absent for all-day events.
type dataTourismeOccurrence struct {
	StartDate flexString `json:"startDate"`
	StartTime flexString `json:"startTime"`
	EndDate   flexString `json:"endDate"`
	EndTime   flexString `json:"endTime"`
}

type dataTourismeRepresentation struct {
	RelatedResource *struct {
		Locator flexString `json:"locator"`
	} `json:"hasRelatedResource"`
}

// dataTourismeContact mirrors hasContact entries; the homepage carries the
// original page of the event on the publisher's website.
type dataTourismeContact struct {
	Homepage flexString `json:"homepage"`
}

type dataTourismeEvent struct {
	URI                   string                      `json:"uri"`
	Label                 dataTourismeLabel           `json:"label"`
	Type                  []string                    `json:"type"`
	TakesPlaceAt          json.RawMessage             `json:"takesPlaceAt"`
	IsLocatedAt           *dataTourismeLocation       `json:"isLocatedAt"`
	HasMainRepresentation *dataTourismeRepresentation `json:"hasMainRepresentation"`
	HasRepresentation     json.RawMessage             `json:"hasRepresentation"`
	HasContact            json.RawMessage             `json:"hasContact"`
}

type dataTourismeResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source dataTourismeEvent `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (c *dataTourismeCollector) Collect(location application.CollectLocation) ([]application.Event, error) {
	paris, err := time.LoadLocation(dataTourismeParisLocation)
	if err != nil {
		return nil, fmt.Errorf("error loading %s timezone: %w", dataTourismeParisLocation, err)
	}
	now := time.Now().In(paris)
	today := now.Format("2006-01-02")

	docs := []dataTourismeEvent{}
	total := -1
	from := 0
	for page := 0; page < dataTourismeMaxPages; page++ {
		response, err := c.fetchPage(location, from, today)
		if err != nil {
			return nil, err
		}
		if page == 0 {
			total = response.Hits.Total.Value
			slog.Info("Requesting DATAtourisme events", "city", location.City, "total", total)
		}

		for _, hit := range response.Hits.Hits {
			docs = append(docs, hit.Source)
		}
		from += len(response.Hits.Hits)
		if len(response.Hits.Hits) == 0 || from >= total {
			break
		}
	}
	if total > from {
		slog.Warn("DATAtourisme collection truncated", "city", location.City, "collected", from, "total", total)
	}

	events := dataTourismeEventsToEvents(docs, now, paris)
	slog.Info("Collected DATAtourisme events", "city", location.City, "count", len(events))
	return dedupeEvents(events), nil
}

func dataTourismeEventsToEvents(docs []dataTourismeEvent, now time.Time, paris *time.Location) []application.Event {
	events := []application.Event{}
	for _, doc := range docs {
		events = append(events, dataTourismeEventToEvents(doc, now, paris)...)
	}
	return events
}

func (c *dataTourismeCollector) fetchPage(location application.CollectLocation, from int, today string) (dataTourismeResponse, error) {
	body, err := json.Marshal(dataTourismeQuery(location, from, today))
	if err != nil {
		return dataTourismeResponse{}, fmt.Errorf("error encoding query: %w", err)
	}

	resp, err := c.client.Post(dataTourismeSearchURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return dataTourismeResponse{}, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return dataTourismeResponse{}, fmt.Errorf("error reading response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return dataTourismeResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response dataTourismeResponse
	if err := json.Unmarshal(payload, &response); err != nil {
		return dataTourismeResponse{}, fmt.Errorf("error decoding response: %w", err)
	}
	return response, nil
}

// dataTourismeQuery builds the Elasticsearch search body: event POIs around
// the location whose last occurrence has not ended yet.
func dataTourismeQuery(location application.CollectLocation, from int, today string) map[string]any {
	return map[string]any{
		"size": dataTourismePageSize,
		"from": from,
		"sort": []map[string]any{
			{"lastUpdate": map[string]string{"order": "desc"}},
		},
		"_source": map[string]any{
			"includes": []string{
				"uri",
				"label",
				"type",
				"takesPlaceAt",
				"isLocatedAt.geo",
				"isLocatedAt.address.streetAddress",
				"isLocatedAt.address.addressLocality",
				"isLocatedAt.address.postalCode",
				"isLocatedAt.address.hasAddressCity.label",
				"hasMainRepresentation.hasRelatedResource.locator",
				"hasRepresentation.hasRelatedResource.locator",
				"hasContact",
			},
		},
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []map[string]any{
					{"term": map[string]string{"type": dataTourismeEventType}},
					{"geo_distance": map[string]any{
						"distance":             fmt.Sprintf("%gkm", location.Radius),
						"isLocatedAt.geoPoint": map[string]float64{"lat": location.Lat, "lon": location.Lon},
					}},
					{"range": map[string]any{
						"takesPlaceAt.endDate": map[string]string{"gte": today},
					}},
				},
			},
		},
	}
}

func dataTourismeEventToEvents(doc dataTourismeEvent, now time.Time, paris *time.Location) []application.Event {
	name := doc.Label.French()
	if name == "" || doc.IsLocatedAt == nil || doc.IsLocatedAt.Geo == nil {
		return nil
	}

	kind := dataTourismeKind(doc.Type)
	genres := dataTourismeGenres(doc.Type)
	image := dataTourismeImage(doc)
	address := dataTourismeAddressLine(doc.IsLocatedAt.Address)
	// The data.datatourisme.fr URI is a last resort: the homepage of the
	// contact is the original page of the event on the publisher's site.
	source := doc.URI
	if homepage := dataTourismeHomepage(doc.HasContact); homepage != "" {
		source = homepage
	}

	events := []application.Event{}
	for _, occurrence := range dataTourismeOccurrences(doc.TakesPlaceAt) {
		begin, end, ok := dataTourismeOccurrencePeriod(occurrence, paris)
		if !ok {
			continue
		}
		// Skip events that already ended or span too long (the server
		// rejects them); seasonal ranges are not real events anyway.
		if end.Before(now) || end.Sub(begin) > time.Hour*24*15 {
			continue
		}
		events = append(events, application.Event{
			Name:    name,
			Kind:    kind,
			Genres:  genres,
			Begin:   begin,
			End:     end,
			Loc:     application.EventLocation{Lat: doc.IsLocatedAt.Geo.Latitude, Lon: doc.IsLocatedAt.Geo.Longitude},
			Address: address,
			Source:  source,
			Img:     image,
		})
	}
	return events
}

// dataTourismeOccurrences normalizes takesPlaceAt, which can be a single
// occurrence object or a list of occurrences.
func dataTourismeOccurrences(raw json.RawMessage) []dataTourismeOccurrence {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var single dataTourismeOccurrence
	if err := json.Unmarshal(raw, &single); err == nil && single.StartDate != "" {
		return []dataTourismeOccurrence{single}
	}
	var list []dataTourismeOccurrence
	if err := json.Unmarshal(raw, &list); err != nil {
		slog.Warn("error parsing DATAtourisme occurrences", "error", err)
		return nil
	}
	return list
}

// dataTourismeOccurrencePeriod builds the local (Europe/Paris) begin and end
// times of an occurrence. An absent end time falls back to the begin time,
// while multi-day and all-day occurrences cover their whole last day.
func dataTourismeOccurrencePeriod(occurrence dataTourismeOccurrence, paris *time.Location) (time.Time, time.Time, bool) {
	startDate := string(occurrence.StartDate)
	startTime := dataTourismeNormalizedTime(string(occurrence.StartTime))
	endDate := string(occurrence.EndDate)
	if endDate == "" {
		endDate = startDate
	}
	endTime := dataTourismeNormalizedTime(string(occurrence.EndTime))

	begin, ok := dataTourismeDatetime(startDate, startTime, false, paris)
	if !ok {
		return time.Time{}, time.Time{}, false
	}

	var end time.Time
	switch {
	case endTime != "":
		end, ok = dataTourismeDatetime(endDate, endTime, false, paris)
	case endDate != startDate || startTime == "":
		end, ok = dataTourismeDatetime(endDate, "", true, paris)
	default:
		end, ok = begin, true
	}
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	if end.Before(begin) {
		return time.Time{}, time.Time{}, false
	}
	return begin, end, true
}

// dataTourismeNormalizedTime validates a local "-T14:30:00" time (the date
// part is elided by DATAtourisme) and returns an empty string for absent or
// malformed values.
func dataTourismeNormalizedTime(dashTime string) string {
	timePart := strings.TrimPrefix(strings.TrimSpace(dashTime), "-")
	if timePart == "" {
		return ""
	}
	if _, err := time.ParseInLocation("T15:04:05", timePart, time.UTC); err != nil {
		slog.Warn("error parsing DATAtourisme time, ignoring it", "time", dashTime, "error", err)
		return ""
	}
	return timePart
}

// dataTourismeDatetime parses a "2006-01-02" date and an optional validated
// "T14:30:00" local time; all-day occurrences start at midnight and, for
// ends, cover the whole day.
func dataTourismeDatetime(date, timePart string, endOfDay bool, paris *time.Location) (time.Time, bool) {
	if date == "" {
		return time.Time{}, false
	}
	day, err := time.ParseInLocation("2006-01-02", date, paris)
	if err != nil {
		slog.Warn("error parsing DATAtourisme date", "date", date, "error", err)
		return time.Time{}, false
	}

	if timePart == "" {
		if endOfDay {
			return day.Add(23*time.Hour + 59*time.Minute + 59*time.Second), true
		}
		return day, true
	}

	hour, err := time.ParseInLocation("T15:04:05", timePart, paris)
	if err != nil {
		slog.Warn("error parsing DATAtourisme time, falling back to day bounds", "time", timePart, "error", err)
		return dataTourismeDatetime(date, "", endOfDay, paris)
	}
	y, m, d := day.Date()
	h, mi, s := hour.Clock()
	return time.Date(y, m, d, h, mi, s, 0, paris), true
}

// dataTourismeTypeKinds maps the DATAtourisme event sub-types to application
// kinds, ordered from the most specific to the most generic: the index
// returns the whole type hierarchy and the first match wins.
var dataTourismeTypeKinds = []struct {
	typeName string
	kind     application.Kind
}{
	{"concert", application.KindConcert},
	{"recital", application.KindConcert},
	{"musicevent", application.KindConcert},
	{"theaterevent", application.KindTheater},
	{"cirque", application.KindCircus},
	{"screeningevent", application.KindMovie},
	{"cinema", application.KindMovie},
	{"exhibitionevent", application.KindExhibitions},
	{"exhibition", application.KindExhibitions},
	{"visualartsevent", application.KindExhibitions},
	{"conference", application.KindExhibitions},
	{"fairorshow", application.KindExhibitions},
	{"festival", application.KindFestival},
	{"market", application.KindMarket},
	{"bricabrac", application.KindFleaMarket},
	{"garagesale", application.KindFleaMarket},
	{"saleevent", application.KindFleaMarket},
	{"childrensevent", application.KindWorkshop},
	{"traineeship", application.KindWorkshop},
	{"trainingworkshop", application.KindWorkshop},
	{"reading", application.KindWorkshop},
	{"visit", application.KindWorkshop},
	{"rambling", application.KindWorkshop},
	{"businessevent", application.KindBusiness},
	{"sportscompetition", application.KindSports},
	{"sportsevent", application.KindSports},
}

// dataTourismeGenericTypes are hierarchy or catch-all types too vague to be
// a kind or a genre.
var dataTourismeGenericTypes = map[string]bool{
	"event":                  true,
	"culturalevent":          true,
	"socialevent":            true,
	"showevent":              true,
	"localanimation":         true,
	"traditionalcelebration": true,
	"religiousevent":         true,
	"parade":                 true,
	"openday":                true,
	"game":                   true,
	"product":                true,
	"pointofinterest":        true,
	"placeofinterest":        true,
	"entertainmentandevent":  true,
}

func dataTourismeShortTypes(types []string) []string {
	short := make([]string, 0, len(types))
	for _, t := range types {
		name := strings.ToLower(t)
		if i := strings.LastIndexAny(name, "#/"); i >= 0 {
			name = name[i+1:]
		}
		if strings.Contains(name, "schema.org") {
			continue
		}
		short = append(short, name)
	}
	return short
}

func dataTourismeKind(types []string) application.Kind {
	short := dataTourismeShortTypes(types)
	lookup := make(map[string]bool, len(short))
	for _, name := range short {
		lookup[name] = true
	}
	for _, mapping := range dataTourismeTypeKinds {
		if lookup[mapping.typeName] {
			return mapping.kind
		}
	}
	return application.KindUnknown
}

func dataTourismeGenres(types []string) []string {
	genres := []string{}
	for _, name := range dataTourismeShortTypes(types) {
		if !dataTourismeGenericTypes[name] {
			genres = append(genres, name)
		}
	}
	return genres
}

func dataTourismeImage(doc dataTourismeEvent) string {
	if doc.HasMainRepresentation != nil && doc.HasMainRepresentation.RelatedResource != nil {
		return string(doc.HasMainRepresentation.RelatedResource.Locator)
	}
	for _, representation := range dataTourismeRepresentations(doc.HasRepresentation) {
		if representation.RelatedResource != nil && representation.RelatedResource.Locator != "" {
			return string(representation.RelatedResource.Locator)
		}
	}
	return ""
}

// dataTourismeRepresentations normalizes hasRepresentation, which can be an
// object or a list.
func dataTourismeRepresentations(raw json.RawMessage) []dataTourismeRepresentation {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var single dataTourismeRepresentation
	if err := json.Unmarshal(raw, &single); err == nil {
		return []dataTourismeRepresentation{single}
	}
	var list []dataTourismeRepresentation
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil
	}
	return list
}

// dataTourismeHomepage extracts the original page of the event from
// hasContact, which can be a single contact or a list; the first homepage
// wins.
func dataTourismeHomepage(raw json.RawMessage) string {
	contacts := []dataTourismeContact{}
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var single dataTourismeContact
	if err := json.Unmarshal(raw, &single); err == nil {
		contacts = append(contacts, single)
	} else if err := json.Unmarshal(raw, &contacts); err != nil {
		slog.Warn("error parsing DATAtourisme contacts", "error", err)
		return ""
	}
	for _, contact := range contacts {
		if homepage := string(contact.Homepage); homepage != "" {
			return homepage
		}
	}
	return ""
}

func dataTourismeAddressLine(address *dataTourismeAddress) string {
	if address == nil {
		return ""
	}

	parts := []string{}
	parts = append(parts, []string(address.Street)...)
	if postal := string(address.Postal); postal != "" {
		parts = append(parts, postal)
	}
	if address.City != nil {
		if city := address.City.Label.French(); city != "" {
			parts = append(parts, city)
			return strings.Join(parts, ", ")
		}
	}
	if locality := string(address.Locality); locality != "" {
		parts = append(parts, locality)
	}
	return strings.Join(parts, ", ")
}
