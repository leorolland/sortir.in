package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/leorolland/sortir.in/pkg/application/applicationtest"
	"github.com/stretchr/testify/require"
)

func TestEventsPutSuccess(t *testing.T) {
	t.Run("insert 2 events and then update one of them", func(t *testing.T) {
		app := setupTestPocketBase(t)

		price := 10.0
		priceCurrency := "EUR"
		now := time.Now()
		begin := now.Add(24 * time.Hour)
		end := now.Add(25 * time.Hour)

		events := applicationtest.MustValidateEvents(t, []application.Event{
			{
				Name:  "Test Event",
				Begin: begin,
				End:   end,
				Loc: application.EventLocation{
					Lat: 48.8566,
					Lon: 2.3522,
				},
				Place:         "Test Place",
				Address:       "Test Address",
				Price:         &price,
				PriceCurrency: &priceCurrency,
				Source:        "https://example.com",
				Img:           "https://example.com/image.jpg",
				Genres:        []string{"Test Genre"},
				Kind:          "movie",
			},
			{
				Name:  "Test Event 2",
				Begin: begin.Add(24 * time.Hour),
				End:   end.Add(25 * time.Hour),
				Loc: application.EventLocation{
					Lat: 48.8566,
					Lon: 2.3522,
				},
				Place:         "Test Place 2",
				Address:       "Test Address 2",
				Price:         &price,
				PriceCurrency: &priceCurrency,
				Source:        "https://example.com",
				Img:           "https://example.com/image.jpg",
				Genres:        []string{"Test Genre 2"},
				Kind:          "movie",
			},
		})

		resp, err := putEvents(t, events)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		records, err := app.FindAllRecords("events")
		require.NoError(t, err)

		assertEqualEvents(t, events, records)

		// Update the first event
		updatedEvent := application.Event{
			Name:  events[0].Name,
			Begin: events[0].Begin,
			End:   events[0].End,
			Loc: application.EventLocation{
				Lat: events[0].Loc.Lat + 0.0100,
				Lon: events[0].Loc.Lon + 0.0100,
			},
			Place:         "updated place",
			Address:       "updated address",
			Price:         nil,
			PriceCurrency: nil,
			Source:        "updated source",
			Img:           "updated img",
			Genres:        []string{"updated genre"},
			Kind:          "updated kind",
		}

		resp, err = putEvents(t, []application.Event{updatedEvent})
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		records, err = app.FindAllRecords("events")
		require.NoError(t, err)

		assertEqualEvents(t, []application.Event{updatedEvent, events[1]}, records)
	})
}

func putEvents(t *testing.T, events []application.Event) (*http.Response, error) {
	eventsJSON, err := json.Marshal(events)
	require.NoError(t, err)

	url := fmt.Sprintf("http://127.0.0.1:%d/api/events", PORT)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(eventsJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(req)
}
