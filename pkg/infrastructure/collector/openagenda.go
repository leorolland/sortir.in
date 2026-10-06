package collector

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
)

// OpenAgenda events via the national public aggregate published on
// Opendatasoft (dataset "evenements-publics-openagenda"). It merges the
// events of the public OpenAgenda agendas used by French collectivités
// (~39k upcoming events country-wide), under an open license.
//
// The official OpenAgenda API (https://developers.openagenda.com, free
// oa_pk_ key) is a per-agenda API; its transverse /v2/events search is still
// experimental, so the aggregate is the only national discovery surface.
const (
	openAgendaBaseURL  = "https://public.opendatasoft.com/api/explore/v2.1/catalog/datasets/evenements-publics-openagenda/records"
	openAgendaPageSize = 100 // ODS records endpoint maximum
	// A big city area can hold several thousand upcoming records; the cap
	// keeps a pathological query from running forever.
	openAgendaMaxPages = 50

	openAgendaSelectFields = "uid,title_fr,conditions_fr,keywords_fr,image," +
		"firstdate_begin,firstdate_end,timings,canonicalurl," +
		"location_name,location_address,location_postalcode,location_city," +
		"location_coordinates,originagenda_uid"
)

// Public agendas that are not leisure events and dominate the national feed,
// excluded by origin agenda UID.
var openAgendaExcludedAgendas = map[string]string{
	"38495884": "Mes événements France Travail", // job-search assistance, ~30% of the feed
}

type openAgendaCollector struct {
	client *http.Client
}

func NewOpenAgendaCollector() application.Collector {
	return &openAgendaCollector{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type openAgendaCoordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type openAgendaRecord struct {
	UID        string   `json:"uid"`
	Title      string   `json:"title_fr"`
	Conditions *string  `json:"conditions_fr"`
	Keywords   []string `json:"keywords_fr"`
	Image      string   `json:"image"`
	// The aggregate collapses recurring events into a first and last
	// occurrence; the raw occurrences live in this JSON-encoded string.
	FirstDateBegin string  `json:"firstdate_begin"`
	FirstDateEnd   string  `json:"firstdate_end"`
	Timings        *string `json:"timings"`
	CanonicalURL   string  `json:"canonicalurl"`

	LocationName       string                 `json:"location_name"`
	LocationAddress    string                 `json:"location_address"`
	LocationPostalCode string                 `json:"location_postalcode"`
	LocationCity       string                 `json:"location_city"`
	LocationCoords     *openAgendaCoordinates `json:"location_coordinates"`

	OriginAgendaUID string `json:"originagenda_uid"`
}

type openAgendaTiming struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
}

type openAgendaResponse struct {
	TotalCount int                `json:"total_count"`
	Results    []openAgendaRecord `json:"results"`
}

func (c *openAgendaCollector) Collect(location application.CollectLocation) ([]application.Event, error) {
	now := time.Now()

	events := []application.Event{}
	total := -1
	offset := 0
	for page := 0; page < openAgendaMaxPages; page++ {
		response, err := c.fetchPage(location, offset, now)
		if err != nil {
			return nil, err
		}
		if page == 0 {
			total = response.TotalCount
			slog.Info("Requesting OpenAgenda events", "city", location.City, "total", total)
		}

		events = append(events, openAgendaRecordsToEvents(response.Results, now)...)

		offset += len(response.Results)
		if len(response.Results) == 0 || offset >= total {
			break
		}
	}
	if total > offset {
		slog.Warn("OpenAgenda collection truncated", "city", location.City, "collected", offset, "total", total)
	}

	slog.Info("Collected OpenAgenda events", "city", location.City, "count", len(events))
	return dedupeEvents(events), nil
}

func (c *openAgendaCollector) fetchPage(location application.CollectLocation, offset int, now time.Time) (openAgendaResponse, error) {
	params := url.Values{}
	params.Set("select", openAgendaSelectFields)
	params.Set("limit", strconv.Itoa(openAgendaPageSize))
	params.Set("offset", strconv.Itoa(offset))
	params.Set("where", openAgendaWhere(location, now))

	resp, err := c.client.Get(openAgendaBaseURL + "?" + params.Encode())
	if err != nil {
		return openAgendaResponse{}, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return openAgendaResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response openAgendaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return openAgendaResponse{}, fmt.Errorf("error decoding response: %w", err)
	}
	return response, nil
}

// openAgendaWhere builds the ODSQL filter: a radius around the location
// (kilometers, the unit-suffixed form — a bare number is silently
// misinterpreted by the API), restricted to France and upcoming events.
func openAgendaWhere(location application.CollectLocation, now time.Time) string {
	return fmt.Sprintf(
		"within_distance(location_coordinates, geom'POINT(%s %s)', %gkm) AND location_countrycode = \"FR\" AND firstdate_end >= \"%s\"",
		strconv.FormatFloat(location.Lon, 'f', -1, 64),
		strconv.FormatFloat(location.Lat, 'f', -1, 64),
		location.Radius,
		now.UTC().Format(time.RFC3339),
	)
}

func openAgendaRecordsToEvents(records []openAgendaRecord, now time.Time) []application.Event {
	events := []application.Event{}
	for _, record := range records {
		events = append(events, openAgendaRecordToEvents(record, now)...)
	}
	return events
}

// openAgendaRecordToEvents expands one aggregate record into one Event per
// upcoming timing occurrence, falling back to the first/last date range when
// the timings are missing or unreadable. It returns no event when the record
// is unusable (no coordinates, no title, nothing upcoming).
func openAgendaRecordToEvents(record openAgendaRecord, now time.Time) []application.Event {
	if _, excluded := openAgendaExcludedAgendas[record.OriginAgendaUID]; excluded {
		return nil
	}
	if record.Title == "" || record.LocationCoords == nil {
		return nil
	}

	kind := application.FirstKindMatch(record.Keywords)
	timings := openAgendaUpcomingTimings(record, now)

	events := []application.Event{}
	for _, timing := range timings {
		events = append(events, openAgendaRecordToEvent(record, kind, timing.begin, timing.end))
	}
	if len(events) > 0 {
		return events
	}

	// No usable timings: fall back to the collapsed date range, skipping
	// ranges that would be rejected server-side (> 15 days, e.g. month-long
	// expositions without per-day timings).
	begin, err := parseOpenAgendaTime(record.FirstDateBegin)
	if err != nil {
		return nil
	}
	end, err := parseOpenAgendaTime(record.FirstDateEnd)
	if err != nil || end.Before(now) || end.Sub(begin) > time.Hour*24*15 {
		return nil
	}
	return []application.Event{openAgendaRecordToEvent(record, kind, begin, end)}
}

type openAgendaPeriod struct {
	begin time.Time
	end   time.Time
}

// openAgendaUpcomingTimings parses the timings JSON string and keeps the
// occurrences that have not ended yet, each with a guaranteed end time.
func openAgendaUpcomingTimings(record openAgendaRecord, now time.Time) []openAgendaPeriod {
	if record.Timings == nil || *record.Timings == "" {
		return nil
	}

	var rawTimings []openAgendaTiming
	if err := json.Unmarshal([]byte(*record.Timings), &rawTimings); err != nil {
		slog.Warn("error parsing OpenAgenda timings, falling back to date range", "uid", record.UID, "error", err)
		return nil
	}

	periods := []openAgendaPeriod{}
	for _, raw := range rawTimings {
		begin, err := parseOpenAgendaTime(raw.Begin)
		if err != nil {
			slog.Warn("error parsing OpenAgenda timing begin, skipping occurrence", "uid", record.UID, "begin", raw.Begin, "error", err)
			continue
		}
		end, err := parseOpenAgendaTime(raw.End)
		if err != nil {
			end = begin
		}
		if end.Before(now) {
			continue
		}
		periods = append(periods, openAgendaPeriod{begin: begin, end: end})
	}
	return periods
}

func openAgendaRecordToEvent(record openAgendaRecord, kind application.Kind, begin, end time.Time) application.Event {
	price, currency := openAgendaPrice(record.Conditions)

	genres := record.Keywords
	if genres == nil {
		genres = []string{}
	}

	return application.Event{
		Name:          record.Title,
		Kind:          kind,
		Genres:        genres,
		Begin:         begin,
		End:           end,
		Loc:           application.EventLocation{Lat: record.LocationCoords.Lat, Lon: record.LocationCoords.Lon},
		Place:         record.LocationName,
		Address:       openAgendaAddress(record),
		Price:         price,
		PriceCurrency: currency,
		Source:        record.CanonicalURL,
		Img:           record.Image,
	}
}

func parseOpenAgendaTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// openAgendaAddress joins the location parts without repeating what the
// address string already contains (the aggregate often embeds "75008 Paris"
// in the address itself).
func openAgendaAddress(record openAgendaRecord) string {
	address := strings.TrimSpace(record.LocationAddress)
	if address == "" {
		return strings.TrimSpace(record.LocationCity)
	}

	postalCode := strings.TrimSpace(record.LocationPostalCode)
	city := strings.TrimSpace(record.LocationCity)
	lowerAddress := strings.ToLower(address)
	if postalCode != "" && !strings.Contains(lowerAddress, strings.ToLower(postalCode)) {
		address += " " + postalCode
	}
	if city != "" && !strings.Contains(lowerAddress, strings.ToLower(city)) {
		address += ", " + city
	}
	return address
}

var openAgendaPriceRegex = regexp.MustCompile(`\d+([,.]\d+)?`)

// openAgendaPrice extracts the first price mentioned in the free-text
// conditions ("Plein tarif : 12 euros", "Tarif réduit : 8,50 €", ...).
// "Entrée libre" and other conditions without a price yield no price.
func openAgendaPrice(conditions *string) (*float64, *string) {
	if conditions == nil {
		return nil, nil
	}
	match := openAgendaPriceRegex.FindStringSubmatch(*conditions)
	if match == nil {
		return nil, nil
	}
	price, err := strconv.ParseFloat(strings.ReplaceAll(match[0], ",", "."), 64)
	if err != nil || price <= 0 {
		return nil, nil
	}
	currency := "EUR"
	return &price, &currency
}
