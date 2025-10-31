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
	Name  string    `json:"name"`
	Begin time.Time `json:"begin"`
	End   time.Time `json:"end"`
	Loc   struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"loc"`
}

func (c *pbClient) saveEvent(event application.Event) error {
	req := saveEventRequest{
		Name:  event.Name,
		Begin: event.Begin,
		End:   event.End,
	}
	req.Loc.Lat = event.Lat
	req.Loc.Lon = event.Lon

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
		return fmt.Errorf("request failed with status code: %d and body: %s", resp.StatusCode, string(body))
	}

	return nil
}
