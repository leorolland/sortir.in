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

type parisEventsCollector struct {
	client *http.Client
}

func NewParisEventsCollector() application.Collector {
	return &parisEventsCollector{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ParisEventsResponse represents the top-level response from the Paris Events API
type parisEventsResponse struct {
	Records []struct {
		RecordID string          `json:"recordid"`
		Fields   json.RawMessage `json:"fields"`
	} `json:"records"`
	NHits int `json:"nhits"`
}

// ParisEventsFields contains the actual event data
type parisEventsFields struct {
	ID             string  `json:"id"`
	EventID        int     `json:"event_id"`
	URL            string  `json:"url"`
	Title          string  `json:"title"`
	LeadText       string  `json:"lead_text"`
	Description    string  `json:"description"`
	DateStart      string  `json:"date_start"`
	DateEnd        string  `json:"date_end"`
	Occurrences    *string `json:"occurrences"`
	CoverURL       string  `json:"cover_url"`
	CoverAlt       *string `json:"cover_alt"`
	CoverCredit    *string `json:"cover_credit"`
	AddressName    string  `json:"address_name"`
	AddressStreet  string  `json:"address_street"`
	AddressZipcode string  `json:"address_zipcode"`
	AddressCity    string  `json:"address_city"`
	PriceType      string  `json:"price_type"`
	PriceDetail    *string `json:"price_detail"`
	QfapTags       string  `json:"qfap_tags"`
	// We'll handle lat_lon separately since it can be in different formats
}

func (c *parisEventsCollector) Collect(location application.CollectLocation) ([]application.Event, error) {
	// Only collect events for Paris
	if location.City != "Paris" {
		return []application.Event{}, nil
	}

	// Cannot add >= 2 filters because in this case the API limits the number of results to 100
	// So better get the first 1000 ordered by date_start

	// Build the URL with query parameters
	baseURL := "https://opendata.paris.fr/api/records/1.0/search/"
	params := url.Values{}
	params.Add("dataset", "que-faire-a-paris-")
	params.Add("rows", "1000")
	params.Add("sort", "date_start")

	requestURL := baseURL + "?" + params.Encode()

	slog.Info("Requesting Paris events", "url", requestURL)

	// Create and execute the request
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var parisResp parisEventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parisResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	slog.Info("Received Paris events", "count", parisResp.NHits)

	// Convert the response to events
	events, err := toParisEvents(parisResp)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func toParisEvents(parisResp parisEventsResponse) ([]application.Event, error) {
	events := []application.Event{}

	for _, record := range parisResp.Records {
		var eventFields parisEventsFields
		if err := json.Unmarshal(record.Fields, &eventFields); err != nil {
			slog.Warn("error unmarshaling event fields", "error", err)
			continue
		}

		// Log the raw occurrences field for debugging
		if eventFields.Occurrences != nil {
			slog.Info("Raw occurrences field", "event", eventFields.Title, "occurrences", *eventFields.Occurrences)
		}

		// Extract lat/lon from raw fields
		var lat, lon float64
		var latLonMap map[string]interface{}
		if err := json.Unmarshal(record.Fields, &latLonMap); err == nil {
			if latLonVal, ok := latLonMap["lat_lon"]; ok {
				switch v := latLonVal.(type) {
				case string:
					// Handle string format "lat,lon"
					parts := strings.Split(v, ",")
					if len(parts) == 2 {
						lat, _ = strconv.ParseFloat(parts[0], 64)
						lon, _ = strconv.ParseFloat(parts[1], 64)
					}
				case map[string]interface{}:
					// Handle object format {"lat": x, "lon": y}
					if latVal, ok := v["lat"].(float64); ok {
						lat = latVal
					}
					if lonVal, ok := v["lon"].(float64); ok {
						lon = lonVal
					}
				case []interface{}:
					// Handle array format [lat, lon]
					if len(v) >= 2 {
						if latVal, ok := v[0].(float64); ok {
							lat = latVal
						}
						if lonVal, ok := v[1].(float64); ok {
							lon = lonVal
						}
					}
				}
			}
		}

		// If occurrences field exists and is not empty, create an event for each occurrence
		if eventFields.Occurrences != nil && *eventFields.Occurrences != "" {
			occurrenceEvents, err := createEventsFromOccurrences(eventFields, lat, lon)
			if err != nil {
				slog.Warn("error parsing occurrences, skipping event", "event", eventFields.Title, "error", err)
				continue
			}
			slog.Info("Created events from occurrences", "event", eventFields.Title, "count", len(occurrenceEvents))
			events = append(events, occurrenceEvents...)
		} else {
			// Otherwise, create a single event from the start and end dates
			event, err := createEventFromDates(eventFields, lat, lon)
			if err != nil {
				slog.Warn("error parsing dates, skipping event", "event", eventFields.Title, "error", err)
				continue
			}
			events = append(events, event)
		}
	}

	return events, nil
}

func createEventsFromOccurrences(eventFields parisEventsFields, lat, lon float64) ([]application.Event, error) {
	events := []application.Event{}

	// Occurrences format: "2025-01-01T17:30:00+02:00_2025-01-01T18:30:00+02:00;2025-01-08T17:30:00+02:00_2025-01-08T18:30:00+02:00;..."
	occurrences := strings.Split(*eventFields.Occurrences, ";")

	slog.Info("Processing occurrences", "event", eventFields.Title, "count", len(occurrences))

	for i, occurrence := range occurrences {
		// Split each occurrence into start and end times
		times := strings.Split(occurrence, "_")
		if len(times) != 2 {
			slog.Warn("invalid occurrence format", "occurrence", occurrence)
			continue
		}

		startTime, err := time.Parse(time.RFC3339, times[0])
		if err != nil {
			slog.Warn("error parsing occurrence start time", "time", times[0], "error", err)
			continue
		}

		endTime, err := time.Parse(time.RFC3339, times[1])
		if err != nil {
			slog.Warn("error parsing occurrence end time", "time", times[1], "error", err)
			continue
		}

		// Skip past events
		if endTime.Before(time.Now()) {
			slog.Info("Skipping past event", "event", eventFields.Title, "start", startTime, "end", endTime)
			continue
		}

		slog.Info("Creating occurrence event", "event", eventFields.Title, "index", i, "start", startTime, "end", endTime)
		event := createEvent(eventFields, startTime, endTime, lat, lon)
		events = append(events, event)
	}

	return events, nil
}

func createEventFromDates(eventFields parisEventsFields, lat, lon float64) (application.Event, error) {
	startTime, err := time.Parse(time.RFC3339, eventFields.DateStart)
	if err != nil {
		return application.Event{}, fmt.Errorf("error parsing start time: %w", err)
	}

	endTime, err := time.Parse(time.RFC3339, eventFields.DateEnd)
	if err != nil {
		return application.Event{}, fmt.Errorf("error parsing end time: %w", err)
	}

	// Skip past events
	if endTime.Before(time.Now()) {
		return application.Event{}, fmt.Errorf("event already ended")
	}

	return createEvent(eventFields, startTime, endTime, lat, lon), nil
}

func createEvent(eventFields parisEventsFields, startTime, endTime time.Time, lat, lon float64) application.Event {
	// Determine price if available
	var price *float64
	var priceCurrency *string

	if eventFields.PriceType != "gratuit" && eventFields.PriceDetail != nil {
		// Extract the first number from the price detail
		priceStr := *eventFields.PriceDetail

		// Use regex to extract the first number in the string
		re := regexp.MustCompile(`\d+([,.]\d+)?`)
		matches := re.FindStringSubmatch(priceStr)

		if len(matches) > 0 {
			// Clean up the extracted number
			extractedPrice := matches[0]
			extractedPrice = strings.ReplaceAll(extractedPrice, ",", ".")

			if priceVal, err := strconv.ParseFloat(extractedPrice, 64); err == nil && priceVal > 0 {
				price = &priceVal
				currency := "EUR"
				priceCurrency = &currency
			}
		}
	}

	// Extract tags for genres
	var genres []string
	if eventFields.QfapTags != "" {
		genres = strings.Split(eventFields.QfapTags, ";")
	}

	// Determine kind based on tags
	kind := application.FirstKindMatch(genres)

	return application.Event{
		Name:   eventFields.Title,
		Kind:   kind,
		Genres: genres,
		Begin:  startTime,
		End:    endTime,
		Loc: application.EventLocation{
			Lat: lat,
			Lon: lon,
		},
		Place:         eventFields.AddressName,
		Address:       fmt.Sprintf("%s, %s %s", eventFields.AddressStreet, eventFields.AddressZipcode, eventFields.AddressCity),
		Price:         price,
		PriceCurrency: priceCurrency,
		Source:        eventFields.URL,
		Img:           eventFields.CoverURL,
	}
}
