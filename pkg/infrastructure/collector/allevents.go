package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
)

type allEventsCollector struct {
	client *http.Client
}

func NewAllEventsCollector() application.Collector {
	return &allEventsCollector{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *allEventsCollector) Collect(location application.CollectLocation) ([]application.Event, error) {
	// First collect events from the mobile API
	var mobileQueryEvents []application.Event
	mEvents, err := c.mobileQueryCollect(location)
	if err != nil {
		slog.Warn("error collecting events from mobile API", "error", err)
	} else {
		slog.Info("Collecting allevents from mobile API", "city", location.City, "found", len(mEvents))
		mobileQueryEvents = mEvents
	}

	// Then collect events from the category API
	categories := []string{"music", "parties", "entertainment", "art", "food-drinks", "business", "sports", "exhibitions", "health-wellness", "workshops", "lgbt-pride", "theatre"}
	var categoryQueryEvents []application.Event

	for _, category := range categories {
		events, err := c.categoryQueryCollect(location, category)
		if err != nil {
			slog.Warn("error collecting events for category", "category", category, "error", err)
			continue
		}
		slog.Info("Collecting allevents by category", "city", location.City, "category", category, "found", len(events))
		categoryQueryEvents = append(categoryQueryEvents, events...)
	}

	// Merge results, with category events taking precedence (acting as upsert)
	allEvents := mergeMobileAndCategoryQueryEvents(mobileQueryEvents, categoryQueryEvents)

	return removeDuplicateEvents(allEvents), nil
}

// removeDuplicateEvents removes duplicate events based on event name
func removeDuplicateEvents(events []application.Event) []application.Event {
	uniqueEvents := make([]application.Event, 0, len(events))
	seen := make(map[string]bool)

	for _, event := range events {
		if _, exists := seen[event.Name]; !exists {
			seen[event.Name] = true
			uniqueEvents = append(uniqueEvents, event)
		}
	}

	return uniqueEvents
}

// mergeMobileAndCategoryQueryEvents merges events from mobile and category APIs
// with category events taking precedence when there are duplicates
func mergeMobileAndCategoryQueryEvents(mobileQueryEvents, categoryQueryEvents []application.Event) []application.Event {
	// Create a map of event names to events from the mobile API
	mobileQueryEventMap := make(map[string]application.Event)
	for _, event := range mobileQueryEvents {
		mobileQueryEventMap[event.Name] = event
	}

	// Add all category events to the result
	result := make([]application.Event, len(categoryQueryEvents))
	copy(result, categoryQueryEvents)

	// Add mobile events that don't exist in category events
	categoryQueryEventNames := make(map[string]bool)
	for _, event := range categoryQueryEvents {
		categoryQueryEventNames[event.Name] = true
	}

	for name, event := range mobileQueryEventMap {
		if !categoryQueryEventNames[name] {
			result = append(result, event)
		}
	}

	return result
}

type allEventsCategoryQueryRequest struct {
	City          string   `json:"city"`
	Page          int      `json:"page"`
	Rows          int      `json:"rows"`
	Radius        int      `json:"radius"`
	ExcludeCities []string `json:"exclude_cities"`
	Category      string   `json:"category"`
	IsTimeFilter  bool     `json:"is_time_filter"`
	StartDate     string   `json:"start_date"`
	EndDate       string   `json:"end_date"`
}

type allEventsMobileQueryRequest struct {
	Latitude        string `json:"latitude"`
	Longitude       string `json:"longitude"`
	City            string `json:"city"`
	StartDate       string `json:"start_date"`
	SearchScope     string `json:"search_scope"`
	Page            int    `json:"page"`
	Rows            int    `json:"rows"`
	ShowLongDateFmt bool   `json:"show_long_date_format"`
	Distance        int    `json:"distance"`
	UserLat         string `json:"user_lat"`
	UserLong        string `json:"user_long"`
}

type allEventsCategoryQueryResult struct {
	Eventname  string   `json:"eventname"` // Name
	ThumbURL   string   `json:"thumb_url"` // Img
	StartTime  string   `json:"start_time"`
	EndTime    string   `json:"end_time"`
	Location   *string  `json:"location"`   // Place
	Categories []string `json:"categories"` // Kind candidates
	Venue      struct {
		Street    string `json:"street"` // Address
		Latitude  string `json:"latitude"`
		Longitude string `json:"longitude"`
	} `json:"venue"`
	ShareURL string `json:"share_url"` // Source
	Tickets  struct {
		TicketCurrency *string     `json:"ticket_currency,omitempty"`
		MinTicketPrice interface{} `json:"min_ticket_price,omitempty"`
	} `json:"tickets"`
	CustomParams struct {
		HighConfidenceMergedLookup []string `json:"high_confidence_merged_lookup"`
	} `json:"custom_params"`
}

type allEventsCategoryQueryResponse struct {
	Data []allEventsCategoryQueryResult `json:"data"`
}

type allEventsMobileQueryResult struct {
	Eventname string  `json:"eventname"` // Name
	ThumbURL  string  `json:"thumb_url"` // Img
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	Location  *string `json:"location"` // Place
	Venue     struct {
		Street    string `json:"street"` // Address
		Latitude  string `json:"latitude"`
		Longitude string `json:"longitude"`
	} `json:"venue"`
	ShareURL string `json:"share_url"` // Source
	Tickets  struct {
		TicketCurrency *string     `json:"ticket_currency,omitempty"`
		MinTicketPrice interface{} `json:"min_ticket_price,omitempty"`
	} `json:"tickets"`
}

type allEventsMobileQueryResponse struct {
	SearchResults []allEventsMobileQueryResult `json:"search_result"`
	Page          int                          `json:"page"`
	Rows          int                          `json:"rows"`
	Error         int                          `json:"error"`
}

func (c *allEventsCollector) categoryQueryCollect(location application.CollectLocation, category string) ([]application.Event, error) {
	reqBody := allEventsCategoryQueryRequest{
		City:          location.City,
		Page:          0,
		Rows:          1000,
		Radius:        100000,
		ExcludeCities: []string{"online"},
		Category:      category,
		IsTimeFilter:  true,
		StartDate:     strconv.FormatInt(time.Now().AddDate(0, 0, -7).Unix(), 10),
		EndDate:       strconv.FormatInt(time.Now().AddDate(0, 0, 15).Unix(), 10),
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://allevents.in/api/index.php/events/find-events-from-nearby-cities", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", "https://allevents.in")
	req.Header.Set("Referer", "https://allevents.in/"+location.City)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var allEventsResp allEventsCategoryQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&allEventsResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return categoryQueryToEvents(allEventsResp)
}

func categoryQueryToEvents(allEventsResp allEventsCategoryQueryResponse) ([]application.Event, error) {
	events := []application.Event{}
	for _, eventData := range allEventsResp.Data {
		startTimeUnix, err := strconv.ParseInt(eventData.StartTime, 10, 64)
		if err != nil {
			slog.Warn("error parsing start time, skipping event", "event", eventData.Eventname, "start_time", eventData.StartTime, "error", err)
			continue
		}

		startTime := time.Unix(startTimeUnix, 0)

		var endTime time.Time
		if eventData.EndTime != "" {
			endTimeUnix, err := strconv.ParseInt(eventData.EndTime, 10, 64)
			if err == nil {
				endTime = time.Unix(endTimeUnix, 0)
			}
		}

		lat, err := strconv.ParseFloat(eventData.Venue.Latitude, 64)
		if err != nil {
			lat = 0
		}
		lon, err := strconv.ParseFloat(eventData.Venue.Longitude, 64)
		if err != nil {
			lon = 0
		}

		// Handle ticket prices
		var price *float64
		var priceCurrency *string

		if eventData.Tickets.MinTicketPrice != nil {
			switch v := eventData.Tickets.MinTicketPrice.(type) {
			case string:
				priceStr := v
				priceVal, err := strconv.ParseFloat(priceStr, 64)
				if err == nil && priceVal != 0 { // Set price if it's not 0, allevents put a lot of 0 prices even whitout knowing
					price = &priceVal
				}
			case float64:
				priceVal := v
				if priceVal != 0 {
					price = &priceVal
				}
			}
			priceCurrency = eventData.Tickets.TicketCurrency
		}

		var place string
		if eventData.Location != nil {
			place = *eventData.Location
		}

		event := application.Event{
			Name:   eventData.Eventname,
			Kind:   application.FirstKindMatch(append(eventData.Categories, eventData.CustomParams.HighConfidenceMergedLookup...)),
			Genres: eventData.CustomParams.HighConfidenceMergedLookup,
			Begin:  startTime,
			End:    endTime,
			Loc: application.EventLocation{
				Lat: lat,
				Lon: lon,
			},
			Place:         place,
			Address:       eventData.Venue.Street,
			Price:         price,
			PriceCurrency: priceCurrency,
			Source:        eventData.ShareURL,
			Img:           eventData.ThumbURL,
		}
		events = append(events, event)
	}
	return events, nil
}

func (c *allEventsCollector) mobileQueryCollect(location application.CollectLocation) ([]application.Event, error) {
	lat := strconv.FormatFloat(location.Lat, 'f', 10, 64)
	lon := strconv.FormatFloat(location.Lon, 'f', 10, 64)

	reqBody := allEventsMobileQueryRequest{
		Latitude:        lat,
		Longitude:       lon,
		City:            location.City,
		StartDate:       time.Now().Format("2006-01-02"),
		SearchScope:     "city",
		Page:            0,
		Rows:            3000,
		ShowLongDateFmt: false,
		Distance:        50,
		UserLat:         lat,
		UserLong:        lon,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://allevents.in/api/index.php/mobile_apps/v2/qs/search_with_filters_v2", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", "https://allevents.in")
	req.Header.Set("Referer", "https://allevents.in/"+location.City)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var allEventsResp allEventsMobileQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&allEventsResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return mobileQueryToEvents(allEventsResp)
}

func mobileQueryToEvents(allEventsResp allEventsMobileQueryResponse) ([]application.Event, error) {
	events := []application.Event{}
	for _, eventData := range allEventsResp.SearchResults {
		startTimeUnix, err := strconv.ParseInt(eventData.StartTime, 10, 64)
		if err != nil {
			slog.Warn("error parsing start time, skipping event", "event", eventData.Eventname, "start_time", eventData.StartTime, "error", err)
			continue
		}

		startTime := time.Unix(startTimeUnix, 0)

		var endTime time.Time
		if eventData.EndTime != "" {
			endTimeUnix, err := strconv.ParseInt(eventData.EndTime, 10, 64)
			if err == nil {
				endTime = time.Unix(endTimeUnix, 0)
			}
		}

		lat, err := strconv.ParseFloat(eventData.Venue.Latitude, 64)
		if err != nil {
			lat = 0
		}
		lon, err := strconv.ParseFloat(eventData.Venue.Longitude, 64)
		if err != nil {
			lon = 0
		}

		// Handle ticket prices
		var price *float64
		var priceCurrency *string

		if eventData.Tickets.MinTicketPrice != nil {
			switch v := eventData.Tickets.MinTicketPrice.(type) {
			case string:
				priceStr := v
				priceVal, err := strconv.ParseFloat(priceStr, 64)
				if err == nil && priceVal != 0 { // Set price if it's not 0, allevents put a lot of 0 prices even whitout knowing
					price = &priceVal
				}
			case float64:
				priceVal := v
				if priceVal != 0 {
					price = &priceVal
				}
			}
			priceCurrency = eventData.Tickets.TicketCurrency
		}

		var place string
		if eventData.Location != nil {
			place = *eventData.Location
		}

		event := application.Event{
			Name:   eventData.Eventname,
			Kind:   application.KindUnknown, // Default kind since mobile API doesn't provide categories
			Genres: []string{},              // No genres in the mobile API response
			Begin:  startTime,
			End:    endTime,
			Loc: application.EventLocation{
				Lat: lat,
				Lon: lon,
			},
			Place:         place,
			Address:       eventData.Venue.Street,
			Price:         price,
			PriceCurrency: priceCurrency,
			Source:        eventData.ShareURL,
			Img:           eventData.ThumbURL,
		}
		events = append(events, event)
	}
	return events, nil
}
