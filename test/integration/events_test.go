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
	t.Run("insert 1 event successfully", func(t *testing.T) {
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
		})

		resp, err := putEvents(t, events)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		records, err := app.FindAllRecords("events")
		require.Equal(t, 1, len(records))
		require.Equal(t, "Test Event", records[0].GetString("name"))
		require.Equal(t, begin.Unix(), records[0].GetDateTime("begin").Time().Unix())
		require.Equal(t, end.Unix(), records[0].GetDateTime("end").Time().Unix())
		require.Equal(t, 48.8566, records[0].GetGeoPoint("loc").Lat)
		require.Equal(t, 2.3522, records[0].GetGeoPoint("loc").Lon)
		require.Equal(t, "Test Place", records[0].GetString("place"))
		require.Equal(t, "Test Address", records[0].GetString("address"))
		require.Equal(t, price, records[0].GetFloat("price"))
		require.Equal(t, priceCurrency, records[0].GetString("price_currency"))
		require.Equal(t, "https://example.com", records[0].GetString("source"))
		require.Equal(t, "https://example.com/image.jpg", records[0].GetString("img"))
		require.Equal(t, []string{"Test Genre"}, records[0].GetStringSlice("genres"))
		require.Equal(t, "movie", records[0].GetString("kind"))
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
