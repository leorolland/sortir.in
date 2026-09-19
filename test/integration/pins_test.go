package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/leorolland/sortir.in/pkg/application/applicationtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPinsGetSuccess(t *testing.T) {
	testCases := map[string]struct {
		events   []application.Event
		bounds   application.Bounds
		minDate  time.Time
		maxDate  time.Time
		expected []application.Pin
	}{
		"without events": {
			events: []application.Event{},
			bounds: application.Bounds{
				North: 45.0000,
				South: 41.0000,
				East:  5.0000,
				West:  1.0000,
			},
			minDate:  time.Now(),
			maxDate:  time.Now().Add(time.Hour * 24),
			expected: []application.Pin{},
		},
		"with 1 event outside bounds and 1 event inside bounds but entirely before the window": {
			events: applicationtest.MustValidateEvents(t, []application.Event{
				{
					Name:  "Event outside bounds",
					Loc:   application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind:  application.KindBusiness,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
				{
					Name:  "Event inside bounds",
					Loc:   application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:  application.KindConcert,
					Begin: time.Now(),
					End:   time.Now().Add(time.Hour * 24),
				},
			}),
			bounds: application.Bounds{
				North: 45.0000,
				South: 41.0000,
				East:  5.0000,
				West:  1.0000,
			},
			minDate:  time.Now().Add(time.Hour * 48),
			maxDate:  time.Now().Add(time.Hour * 96),
			expected: []application.Pin{},
		},
		"with 1 event outside bounds and 1 event inside bounds but beginning after the window": {
			events: applicationtest.MustValidateEvents(t, []application.Event{
				{
					Name:  "Event outside bounds",
					Loc:   application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind:  application.KindBusiness,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
				{
					Name:  "Event inside bounds",
					Loc:   application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:  application.KindConcert,
					Begin: time.Now().Add(time.Hour * 48),
					End:   time.Now().Add(time.Hour * 72),
				},
			}),
			bounds: application.Bounds{
				North: 45.0000,
				South: 41.0000,
				East:  5.0000,
				West:  1.0000,
			},
			minDate:  time.Now(),
			maxDate:  time.Now().Add(time.Hour * 24),
			expected: []application.Pin{},
		},
		"with 1 event outside bounds and 1 ongoing event inside bounds spanning the window": {
			events: applicationtest.MustValidateEvents(t, []application.Event{
				{
					Name:  "Event outside bounds",
					Loc:   application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind:  application.KindBusiness,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
				{
					Name:  "Event inside bounds",
					Loc:   application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:  application.KindConcert,
					Begin: time.Now().Add(-time.Hour * 24),
					End:   time.Now().Add(time.Hour * 48),
				},
			}),
			bounds: application.Bounds{
				North: 45.0000,
				South: 41.0000,
				East:  5.0000,
				West:  1.0000,
			},
			minDate: time.Now(),
			maxDate: time.Now().Add(time.Hour * 24),
			expected: []application.Pin{
				{
					Loc:    application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:   application.KindConcert,
					Amount: 1,
				},
			},
		},
		"with 1 event outside bounds and 1 event inside bounds ending after the window (night activities)": {
			events: applicationtest.MustValidateEvents(t, []application.Event{
				{
					Name:  "Event outside bounds",
					Loc:   application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind:  application.KindBusiness,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
				{
					Name:  "Event inside bounds",
					Loc:   application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:  application.KindConcert,
					Begin: time.Now().Add(time.Hour * 2),
					End:   time.Now().Add(time.Hour * 20),
				},
			}),
			bounds: application.Bounds{
				North: 45.0000,
				South: 41.0000,
				East:  5.0000,
				West:  1.0000,
			},
			minDate: time.Now(),
			maxDate: time.Now().Add(time.Hour * 12),
			expected: []application.Pin{
				{
					Loc:    application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:   application.KindConcert,
					Amount: 1,
				},
			},
		},
		"with 1 event outside bounds and 1 event inside bounds and max date": {
			events: applicationtest.MustValidateEvents(t, []application.Event{
				{
					Name:  "Event outside bounds",
					Loc:   application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind:  application.KindBusiness,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
				{
					Name:  "Event inside bounds",
					Loc:   application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:  application.KindConcert,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
			}),
			bounds: application.Bounds{
				North: 45.0000,
				South: 41.0000,
				East:  5.0000,
				West:  1.0000,
			},
			minDate: time.Now(),
			maxDate: time.Now().Add(time.Hour * 24 * 4),
			expected: []application.Pin{
				{
					Loc:    application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:   application.KindConcert,
					Amount: 1,
				},
			},
		},
		"with 1 event outside bounds and 2 different events inside bounds": {
			events: applicationtest.MustValidateEvents(t, []application.Event{
				{
					Name:  "Event outside bounds",
					Loc:   application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind:  application.KindBusiness,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
				{
					Name:  "Event inside bounds 1",
					Loc:   application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:  application.KindConcert,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
				{
					Name:  "Event inside bounds 2",
					Loc:   application.EventLocation{Lat: 42.9, Lon: 1.3},
					Kind:  application.KindConcert,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
			}),
			bounds: application.Bounds{
				North: 45.0000,
				South: 41.0000,
				East:  5.0000,
				West:  1.0000,
			},
			minDate: time.Now(),
			maxDate: time.Now().Add(time.Hour * 24 * 4),
			expected: []application.Pin{
				{
					Loc:    application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:   application.KindConcert,
					Amount: 1,
				},
				{
					Loc:    application.EventLocation{Lat: 42.9, Lon: 1.3},
					Kind:   application.KindConcert,
					Amount: 1,
				},
			},
		},
		"with 1 event outside bounds and 3 events inside, 2 are at the same place and same kind": {
			events: applicationtest.MustValidateEvents(t, []application.Event{
				{
					Name:  "Event outside bounds",
					Loc:   application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind:  application.KindBusiness,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
				{
					Name:  "Event inside bounds 1",
					Loc:   application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:  application.KindConcert,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
				{
					Name:  "Event inside bounds 2",
					Loc:   application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:  application.KindConcert,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
				{
					Name:  "Event inside bounds 3",
					Loc:   application.EventLocation{Lat: 42.9, Lon: 1.3},
					Kind:  application.KindConcert,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
			}),
			bounds: application.Bounds{
				North: 45.0000,
				South: 41.0000,
				East:  5.0000,
				West:  1.0000,
			},
			minDate: time.Now(),
			maxDate: time.Now().Add(time.Hour * 24 * 4),
			expected: []application.Pin{
				{
					Loc:    application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:   application.KindConcert,
					Amount: 2,
				},
				{
					Loc:    application.EventLocation{Lat: 42.9, Lon: 1.3},
					Kind:   application.KindConcert,
					Amount: 1,
				},
			},
		},
		"with 1 event outside bounds and 3 events inside, 2 are at the same place and different kind": {
			events: applicationtest.MustValidateEvents(t, []application.Event{
				{
					Name:  "Event outside bounds",
					Loc:   application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind:  application.KindBusiness,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
				{
					Name:  "Event inside bounds 1",
					Loc:   application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:  application.KindConcert,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
				{
					Name:  "Event inside bounds 2",
					Loc:   application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:  application.KindBusiness,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
				{
					Name:  "Event inside bounds 3",
					Loc:   application.EventLocation{Lat: 42.9, Lon: 1.3},
					Kind:  application.KindConcert,
					Begin: time.Now().Add(time.Hour * 24),
					End:   time.Now().Add(time.Hour * 24 * 2),
				},
			}),
			bounds: application.Bounds{
				North: 45.0000,
				South: 41.0000,
				East:  5.0000,
				West:  1.0000,
			},
			minDate: time.Now(),
			maxDate: time.Now().Add(time.Hour * 24 * 4),
			expected: []application.Pin{
				{
					Loc:    application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:   application.KindBusiness,
					Amount: 1,
				},
				{
					Loc:    application.EventLocation{Lat: 42.8, Lon: 1.3},
					Kind:   application.KindConcert,
					Amount: 1,
				},
				{
					Loc:    application.EventLocation{Lat: 42.9, Lon: 1.3},
					Kind:   application.KindConcert,
					Amount: 1,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			_ = setupTestPocketBase(t)

			resp, err := putEvents(t, testCase.events)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode)

			resp, err = getPins(t, testCase.bounds, testCase.minDate, testCase.maxDate)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode)

			var pins []application.Pin
			err = json.NewDecoder(resp.Body).Decode(&pins)
			require.NoError(t, err)

			assert.Equal(t, testCase.expected, pins)
		})
	}
}

func getPins(t *testing.T, bounds application.Bounds, minDate, maxDate time.Time) (*http.Response, error) {
	url := fmt.Sprintf("http://127.0.0.1:%d/api/pins", PORT)
	req, err := http.NewRequest("GET", url, nil)

	query := req.URL.Query()
	query.Add("north", strconv.FormatFloat(bounds.North, 'f', -1, 64))
	query.Add("south", strconv.FormatFloat(bounds.South, 'f', -1, 64))
	query.Add("east", strconv.FormatFloat(bounds.East, 'f', -1, 64))
	query.Add("west", strconv.FormatFloat(bounds.West, 'f', -1, 64))
	query.Add("max_time", maxDate.Format(time.RFC3339))
	if !minDate.IsZero() {
		query.Add("min_time", minDate.Format(time.RFC3339))
	}

	req.URL.RawQuery = query.Encode()
	require.NoError(t, err)

	client := &http.Client{}
	return client.Do(req)
}
