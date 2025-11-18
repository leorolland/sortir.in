package pb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
	jsonData, err := json.Marshal(events)
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
