package integration

import (
	"testing"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/leorolland/sortir.in/pkg/application/applicationtest"
	"github.com/leorolland/sortir.in/pkg/infrastructure/server"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/stretchr/testify/require"
)

func TestDeleteExpiredEvents(t *testing.T) {
	app := setupTestPocketBase(t)

	now := time.Now()
	events := applicationtest.MustValidateEvents(t, []application.Event{
		{
			Name:  "Event already terminated",
			Begin: now.Add(-2 * time.Hour),
			End:   now.Add(1 * time.Hour),
			Loc:   application.EventLocation{Lat: 48.8566, Lon: 2.3522},
			Kind:  application.KindConcert,
		},
		{
			Name:  "Event running right now",
			Begin: now.Add(-1 * time.Hour),
			End:   now.Add(1 * time.Hour),
			Loc:   application.EventLocation{Lat: 48.8566, Lon: 2.3522},
			Kind:  application.KindConcert,
		},
		{
			Name:  "Event starting later today",
			Begin: now.Add(30 * time.Minute),
			End:   now.Add(23 * time.Hour),
			Loc:   application.EventLocation{Lat: 48.8566, Lon: 2.3522},
			Kind:  application.KindConcert,
		},
	})

	// Bypass IsValid (which rejects terminated events) by rewriting the first
	// event's end date in the database once it is stored.
	resp, err := putEvents(t, events)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	records, err := app.FindAllRecords("events")
	require.NoError(t, err)
	require.Len(t, records, 3)

	for _, record := range records {
		if record.GetString("name") != "Event already terminated" {
			continue
		}
		record.Set("end", types.NowDateTime().Add(-time.Hour))
		require.NoError(t, app.Save(record))
	}

	require.NoError(t, server.DeleteExpiredEvents(app))

	records, err = app.FindAllRecords("events")
	require.NoError(t, err)
	require.Len(t, records, 2)
	for _, record := range records {
		require.NotEqual(t, "Event already terminated", record.GetString("name"))
	}
}
