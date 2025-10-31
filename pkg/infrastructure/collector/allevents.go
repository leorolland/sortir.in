package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type allEventsRequest struct {
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

type allEventsResponse struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
	Data    []struct {
		Eventname string `json:"eventname"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Venue     struct {
			Latitude  string `json:"latitude"`
			Longitude string `json:"longitude"`
		} `json:"venue"`
	} `json:"data"`
}

func (c *allEventsCollector) Collect(location application.CollectLocation) ([]application.Event, error) {
	reqBody := allEventsRequest{
		City:          location.City,
		Page:          0,
		Rows:          500,
		Radius:        int(location.Radius),
		ExcludeCities: []string{location.City},
		Category:      "music",
		IsTimeFilter:  false,
		StartDate:     strconv.FormatInt(time.Now().Unix(), 10),
		EndDate:       strconv.FormatInt(time.Now().AddDate(0, 1, 0).Unix(), 10),
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
	req.Header.Set("Referer", "https://allevents.in/"+location.City+"/music")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var allEventsResp allEventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&allEventsResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if allEventsResp.Error != 0 {
		return nil, fmt.Errorf("API error: %s", allEventsResp.Message)
	}

	var events []application.Event
	for _, eventData := range allEventsResp.Data {
		startTimeUnix, err := strconv.ParseInt(eventData.StartTime, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing start time %q: %w", eventData.StartTime, err)
		}

		startTime := time.Unix(startTimeUnix, 0)

		var endTime time.Time
		endTimeUnix, err := strconv.ParseInt(eventData.EndTime, 10, 64)
		if err == nil {
			endTime = time.Unix(endTimeUnix, 0)
		}

		lat, err := strconv.ParseFloat(eventData.Venue.Latitude, 64)
		if err != nil {
			lat = 0
		}
		lon, err := strconv.ParseFloat(eventData.Venue.Longitude, 64)
		if err != nil {
			lon = 0
		}

		event := application.Event{
			Name:  eventData.Eventname,
			Lat:   lat,
			Lon:   lon,
			Begin: startTime,
			End:   endTime,
		}
		events = append(events, event)
	}

	return events, nil
}
