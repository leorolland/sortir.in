package pb

import (
	"bytes"
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

func (c *pbClient) SaveEvents(events []application.Event) error {
	for _, event := range events {
		if err := c.saveEvent(event); err != nil {
			return err
		}
	}
	return nil
}

type saveEventRequest struct {
	Name   string    `json:"name"`
	Kind   string    `json:"kind"`
	Genres []string  `json:"genres"`
	Begin  time.Time `json:"begin"`
	End    time.Time `json:"end"`
	Loc    struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"loc"`
	Place         string   `json:"place"`
	Address       string   `json:"address"`
	Price         *float64 `json:"price,omitempty"`
	PriceCurrency *string  `json:"price_currency,omitempty"`
	Source        string   `json:"source"`
	Img           string   `json:"img"`
}

func (c *pbClient) saveEvent(event application.Event) error {
	req := saveEventRequest{
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
	}
	req.Loc.Lon = event.Loc.Lon
	req.Loc.Lat = event.Loc.Lat

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/collections/events/records", c.baseURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode == 400 {
			var errorResp struct {
				Data struct {
					Name struct {
						Code string `json:"code"`
					} `json:"name"`
				} `json:"data"`
			}

			if err := json.Unmarshal(body, &errorResp); err == nil &&
				errorResp.Data.Name.Code == "validation_not_unique" {
				// Skip this error - duplicate event name
				return nil
			}
		}

		return fmt.Errorf("request failed with status code: %d and body: %s", resp.StatusCode, string(body))
	}

	return nil
}
