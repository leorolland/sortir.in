package pb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
)

type pbClient struct {
	baseURL string
}

func NewPBClient(baseURL string) *pbClient {
	return &pbClient{
		baseURL: baseURL,
	}
}

type saveEventLocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type saveEventRequestEntry struct {
	Name          string            `json:"name"`
	Kind          string            `json:"kind"`
	Genres        []string          `json:"genres"`
	Begin         time.Time         `json:"begin"`
	End           time.Time         `json:"end"`
	Loc           saveEventLocation `json:"loc"`
	Place         string            `json:"place"`
	Address       string            `json:"address"`
	Price         *float64          `json:"price,omitempty"`
	PriceCurrency *string           `json:"price_currency,omitempty"`
	Source        string            `json:"source"`
	Img           string            `json:"img"`
}

func (c *pbClient) SaveEvents(events []application.Event) error {
	eventsRequest := make([]saveEventRequestEntry, len(events))
	for i, event := range events {
		eventsRequest[i] = saveEventRequestEntry{
			Name:          event.Name,
			Kind:          event.Kind,
			Genres:        event.Genres,
			Begin:         event.Begin,
			End:           event.End,
			Place:         event.Place,
			Address:       event.Address,
			Price:         event.Price,
			PriceCurrency: event.PriceCurrency,
			Source:        event.Source,
			Img:           event.Img,
			Loc: saveEventLocation{
				Lat: event.Loc.Lat,
				Lon: event.Loc.Lon,
			},
		}
	}

	jsonData, err := json.Marshal(eventsRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, fmt.Sprintf("%s/api/events", c.baseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		return fmt.Errorf("request failed with status code: %d and body: %s", resp.StatusCode, string(body))
	}

	return nil
}
