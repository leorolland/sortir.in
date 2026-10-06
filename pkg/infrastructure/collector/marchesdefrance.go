package collector

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // guarantee Europe/Paris is available in stripped deployments

	"github.com/leorolland/sortir.in/pkg/application"
)

// Les marchés de France (https://www.marches-de-france.fr) is a merged
// national dataset of ~5.7k markets, itself built from OpenStreetMap
// (amenity=marketplace + opening_hours, ODbL) and DATAtourisme (Licence
// Ouverte). The CSV download URL serves the full base; attribution is
// required: "Les marchés de France, © contributeurs OpenStreetMap (ODbL) &
// DATAtourisme (LO 2.0)".
//
// Markets are weekly recurrences, not dated events: rows carrying structured
// days and hours are expanded into concrete Event instances over the next two
// weeks (beyond "Cette semaine", which ends the following Monday 3am, users
// never look further). Rows without days (~37%, DATAtourisme-only entries)
// cannot be expanded and are skipped. Re-runs upsert on (name, begin, end),
// so instances stay stable across runs.
const (
	marchesDeFranceCSVURL      = "https://www.marches-de-france.fr/donnees-licences.html?csv=1"
	marchesDeFranceSource      = "https://www.marches-de-france.fr/"
	marchesDeFranceLocation    = "Europe/Paris"
	marchesDeFranceHorizonDays = 14
	marchesDeFranceMaxCSVSize  = 10 << 20

	// Default hours for markets without structured ones ("le matin", empty).
	marketDefaultBeginMinutes = 8 * 60
	marketDefaultEndMinutes   = 13 * 60
)

type marchesDeFranceCollector struct {
	client *http.Client
	csvURL string
}

func NewMarchesDeFranceCollector() application.GlobalCollector {
	return &marchesDeFranceCollector{
		client: &http.Client{Timeout: 30 * time.Second},
		csvURL: marchesDeFranceCSVURL,
	}
}

type marcheDeFranceRow struct {
	slug     string
	nom      string
	commune  string
	lat      float64
	lon      float64
	adresse  string
	jours    string
	horaires string
	saison   string
}

// marcheDeFranceURL returns the market page on the site (one page per slug)
// and falls back to the homepage when a row has no slug.
func marcheDeFranceURL(slug string) string {
	if slug == "" {
		return marchesDeFranceSource
	}
	return marchesDeFranceSource + "marche/" + url.PathEscape(slug) + ".html"
}

func (c *marchesDeFranceCollector) CollectAll() ([]application.Event, error) {
	data, err := c.fetchCSV()
	if err != nil {
		return nil, err
	}

	rows, err := parseMarchesDeFranceCSV(data)
	if err != nil {
		return nil, err
	}

	paris, err := time.LoadLocation(marchesDeFranceLocation)
	if err != nil {
		return nil, fmt.Errorf("error loading %s timezone: %w", marchesDeFranceLocation, err)
	}

	events := marchesDeFranceRowsToEvents(rows, time.Now().In(paris), paris)
	slog.Info("Collected markets", "count", len(events))
	return dedupeEvents(events), nil
}

func (c *marchesDeFranceCollector) fetchCSV() ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.csvURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	}
	req.Header.Set("User-Agent", "sortir.in populate (+https://sortir.in)")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, marchesDeFranceMaxCSVSize))
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}
	return data, nil
}

func parseMarchesDeFranceCSV(data []byte) ([]marcheDeFranceRow, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error parsing CSV: %w", err)
	}
	if len(records) < 1 {
		return nil, fmt.Errorf("empty CSV")
	}

	columns := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		columns[strings.TrimSpace(name)] = i
	}

	rows := []marcheDeFranceRow{}
	for _, record := range records[1:] {
		row := marcheDeFranceRow{
			slug:     csvColumn(record, columns, "slug"),
			nom:      csvColumn(record, columns, "nom"),
			commune:  csvColumn(record, columns, "commune"),
			adresse:  csvColumn(record, columns, "adresse"),
			jours:    csvColumn(record, columns, "jours"),
			horaires: csvColumn(record, columns, "horaires"),
			saison:   csvColumn(record, columns, "saison"),
		}
		lat, latErr := strconv.ParseFloat(strings.TrimSpace(csvColumn(record, columns, "latitude")), 64)
		lon, lonErr := strconv.ParseFloat(strings.TrimSpace(csvColumn(record, columns, "longitude")), 64)
		if row.nom == "" || latErr != nil || lonErr != nil {
			continue
		}
		row.lat, row.lon = lat, lon
		rows = append(rows, row)
	}
	return rows, nil
}

func csvColumn(record []string, columns map[string]int, name string) string {
	index, ok := columns[name]
	if !ok || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

// disambiguatedNames returns one display name per row. Generic names are
// shared across communes ("Marché hebdomadaire" ×230) and the events table
// upserts on (name, begin, end): duplicated names get their commune appended,
// and markets still colliding within a commune also get their address.
func disambiguatedNames(rows []marcheDeFranceRow) []string {
	nameCount := make(map[string]int, len(rows))
	nameCommuneCount := make(map[string]int, len(rows))
	for _, row := range rows {
		nameCount[row.nom]++
		nameCommuneCount[row.nom+"\x00"+row.commune]++
	}

	names := make([]string, len(rows))
	for i, row := range rows {
		name := row.nom
		if nameCount[row.nom] > 1 {
			name = fmt.Sprintf("%s (%s)", name, row.commune)
			if nameCommuneCount[row.nom+"\x00"+row.commune] > 1 && row.adresse != "" {
				name = fmt.Sprintf("%s · %s", name, row.adresse)
			}
		}
		names[i] = name
	}
	return names
}

func marchesDeFranceRowsToEvents(rows []marcheDeFranceRow, now time.Time, paris *time.Location) []application.Event {
	names := disambiguatedNames(rows)
	events := []application.Event{}
	for i, row := range rows {
		events = append(events, expandMarket(row, names[i], now, paris)...)
	}
	return events
}

// expandMarket materializes one market recurrence into concrete instances over
// the next marchesDeFranceHorizonDays days.
func expandMarket(row marcheDeFranceRow, name string, now time.Time, paris *time.Location) []application.Event {
	days := parseMarketDays(row.jours)
	if len(days) == 0 {
		return nil
	}
	season := parseMarketSeason(row.saison)
	beginMinutes, endMinutes := parseMarketHours(row.horaires)

	events := []application.Event{}
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, paris)
	for offset := 0; offset < marchesDeFranceHorizonDays; offset++ {
		day := midnight.AddDate(0, 0, offset)
		if !days[day.Weekday()] {
			continue
		}
		if season != nil && !season.allows(day) {
			continue
		}
		begin := day.Add(time.Duration(beginMinutes) * time.Minute)
		end := day.Add(time.Duration(endMinutes) * time.Minute)
		if end.Before(now) {
			continue
		}
		events = append(events, application.Event{
			Name:    name,
			Kind:    application.KindMarket,
			Begin:   begin,
			End:     end,
			Loc:     application.EventLocation{Lat: row.lat, Lon: row.lon},
			Place:   row.adresse,
			Address: row.commune,
			Source:  marcheDeFranceURL(row.slug),
		})
	}
	return events
}

var marketDayCodes = map[string]time.Weekday{
	"mo": time.Monday,
	"tu": time.Tuesday,
	"we": time.Wednesday,
	"th": time.Thursday,
	"fr": time.Friday,
	"sa": time.Saturday,
	"su": time.Sunday,
}

// parseMarketDays parses the OSM-style day list: comma-separated tokens, each
// a two-letter day code ("We") or an inclusive range ("Tu-Su"). Unknown
// tokens are ignored; an empty result means the row cannot be expanded.
func parseMarketDays(s string) map[time.Weekday]bool {
	days := map[time.Weekday]bool{}
	for _, token := range strings.Split(s, ",") {
		token = strings.ToLower(strings.TrimSpace(token))
		if token == "" {
			continue
		}
		if bounds := strings.SplitN(token, "-", 2); len(bounds) == 2 {
			start, okStart := marketDayCodes[strings.TrimSpace(bounds[0])]
			end, okEnd := marketDayCodes[strings.TrimSpace(bounds[1])]
			if !okStart || !okEnd {
				continue
			}
			for day := start; ; day = (day + 1) % 7 {
				days[day] = true
				if day == end {
					break
				}
			}
			continue
		}
		if day, ok := marketDayCodes[token]; ok {
			days[day] = true
		}
	}
	return days
}

// parseMarketHours returns the begin and end of a market day as
// minutes-since-midnight; the end may exceed 24h for markets running past
// midnight ("20:00-00:00"). Multiple "HH:MM-HH:MM" variants separated by "/"
// are seasonal alternatives: their union covers every variant. Free-text
// values map to their obvious window, anything unparsable falls back to the
// default morning market.
func parseMarketHours(s string) (int, int) {
	switch {
	case s == "":
		return marketDefaultBeginMinutes, marketDefaultEndMinutes
	case strings.Contains(s, "matin"):
		return marketDefaultBeginMinutes, marketDefaultEndMinutes
	case strings.Contains(s, "après-midi"), strings.Contains(s, "apres-midi"):
		return 14 * 60, 18 * 60
	case strings.Contains(s, "midi"):
		return 11 * 60, 15 * 60
	case strings.Contains(s, "soir"):
		return 18 * 60, 22 * 60
	}

	begin, end := -1, -1
	for _, variant := range strings.Split(s, "/") {
		variantBegin, variantEnd, ok := parseMarketHoursRange(variant)
		if !ok {
			continue
		}
		if begin == -1 || variantBegin < begin {
			begin = variantBegin
		}
		if variantEnd > end {
			end = variantEnd
		}
	}
	if begin == -1 {
		return marketDefaultBeginMinutes, marketDefaultEndMinutes
	}
	return begin, end
}

func parseMarketHoursRange(s string) (int, int, bool) {
	parts := strings.SplitN(strings.TrimSpace(s), "-", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	begin, err := parseMarketClock(parts[0])
	if err != nil {
		return 0, 0, false
	}
	end, err := parseMarketClock(parts[1])
	if err != nil {
		return 0, 0, false
	}
	if end <= begin {
		end += 24 * 60
	}
	return begin, end, true
}

func parseMarketClock(s string) (int, error) {
	hour, minute, found := strings.Cut(strings.TrimSpace(s), ":")
	if !found {
		return 0, fmt.Errorf("invalid clock %q", s)
	}
	h, err := strconv.Atoi(hour)
	if err != nil || h < 0 || h > 23 {
		return 0, fmt.Errorf("invalid hour in %q", s)
	}
	m, err := strconv.Atoi(minute)
	if err != nil || m < 0 || m > 59 {
		return 0, fmt.Errorf("invalid minute in %q", s)
	}
	return h*60 + m, nil
}

var marketMonths = map[string]time.Month{
	"jan": time.January, "feb": time.February, "mar": time.March,
	"apr": time.April, "may": time.May, "jun": time.June,
	"jul": time.July, "aug": time.August, "sep": time.September,
	"oct": time.October, "nov": time.November, "dec": time.December,
}

// marketSeason narrows a market to part of the year: "Jul-Aug",
// "Jun 15-Sep 15", month lists ("Jan,Feb,Mar"), or several ranges
// ("Jan-Jun, Sep-Dec"). Ranges whose start follows their end wrap around the
// new year ("Sep-Jun"). A nil marketSeason allows every day.
type marketSeason struct {
	ranges []marketSeasonRange
}

type marketSeasonRange struct {
	start int
	end   int
}

func (s *marketSeason) allows(t time.Time) bool {
	key := int(t.Month())*100 + t.Day()
	for _, r := range s.ranges {
		if r.start <= r.end {
			if key >= r.start && key <= r.end {
				return true
			}
		} else if key >= r.start || key <= r.end {
			return true
		}
	}
	return false
}

// parseMarketSeason returns nil for empty or fully unparsable values (fail
// open: the market keeps its weekly schedule year-round). Single dated months
// ("Apr 15", a start date without end) are ignored: they would exclude the
// market from almost the whole year.
func parseMarketSeason(s string) *marketSeason {
	season := &marketSeason{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if bounds := strings.SplitN(part, "-", 2); len(bounds) == 2 {
			startMonth, startDay, okStart := parseMarketMonthDay(bounds[0])
			endMonth, endDay, okEnd := parseMarketMonthDay(bounds[1])
			if !okStart || !okEnd {
				continue
			}
			// A month without an explicit day starts on the 1st and
			// ends on the 31st ("Jul-Aug" covers all of August).
			if startDay == 0 {
				startDay = 1
			}
			if endDay == 0 {
				endDay = 31
			}
			season.ranges = append(season.ranges, marketSeasonRange{
				start: seasonKey(startMonth, startDay),
				end:   seasonKey(endMonth, endDay),
			})
			continue
		}
		// A bare month allows the whole month; a single dated month
		// ("Apr 15", a start date without end) is ignored: it would
		// exclude the market from almost the whole year.
		month, day, ok := parseMarketMonthDay(part)
		if !ok || day != 0 {
			continue
		}
		season.ranges = append(season.ranges, marketSeasonRange{
			start: seasonKey(month, 1),
			end:   seasonKey(month, 31),
		})
	}
	if len(season.ranges) == 0 {
		return nil
	}
	return season
}

func seasonKey(month time.Month, day int) int {
	return int(month)*100 + day
}

// parseMarketMonthDay parses "Jul" or "Jun 15"; a day of 0 means the month
// was given without an explicit day.
func parseMarketMonthDay(s string) (time.Month, int, bool) {
	fields := strings.Fields(strings.TrimSpace(s))
	if len(fields) == 0 || len(fields) > 2 {
		return 0, 0, false
	}
	month, ok := marketMonths[strings.ToLower(fields[0])]
	if !ok {
		return 0, 0, false
	}
	day := 0
	if len(fields) == 2 {
		parsed, err := strconv.Atoi(fields[1])
		if err != nil || parsed < 1 || parsed > 31 {
			return 0, 0, false
		}
		day = parsed
	}
	return month, day, true
}
