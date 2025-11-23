package integration

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/leorolland/sortir.in/migrations"
	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/leorolland/sortir.in/pkg/infrastructure/server"
)

const PORT = 8035

func setupTestPocketBase(t *testing.T) *pocketbase.PocketBase {
	t.Helper()

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  t.TempDir(),
		HideStartBanner: true,
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/stop", func(e *core.RequestEvent) error {
			event := new(core.TerminateEvent)
			event.App = app
			return app.OnTerminate().Trigger(event, func(e *core.TerminateEvent) error {
				return e.App.ResetBootstrapState()
			})
		})

		return se.Next()
	})

	t.Cleanup(func() {
		_, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/stop", PORT))
		require.NoError(t, err)
	})

	server.RegisterApp(app)

	os.Args[1] = "serve"
	os.Args[2] = fmt.Sprintf("--http=127.0.0.1:%d", PORT)

	go app.Start()

	// Wait for the server to be ready
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		_, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/health", PORT))
		require.NoError(t, err)
	}, 5*time.Second, 20*time.Millisecond)

	return app
}

func assertEqualEvent(t *testing.T, expected application.Event, actual *core.Record) {
	require.Equal(t, expected.Name, actual.GetString("name"))
	require.Equal(t, expected.Begin.Unix(), actual.GetDateTime("begin").Time().Unix())
	require.Equal(t, expected.End.Unix(), actual.GetDateTime("end").Time().Unix())
	require.Equal(t, expected.Loc.Lat, actual.GetGeoPoint("loc").Lat)
	require.Equal(t, expected.Loc.Lon, actual.GetGeoPoint("loc").Lon)
	require.Equal(t, expected.Address, actual.GetString("address"))
	if expected.Price != nil {
		require.Equal(t, *expected.Price, actual.GetFloat("price"))
	} else {
		require.Zero(t, actual.GetFloat("price"))
	}
	if expected.PriceCurrency != nil {
		require.Equal(t, *expected.PriceCurrency, actual.GetString("price_currency"))
	} else {
		require.Zero(t, actual.GetString("price_currency"))
	}
	require.Equal(t, expected.Source, actual.GetString("source"))
	require.Equal(t, expected.Img, actual.GetString("img"))
	require.Equal(t, expected.Genres, actual.GetStringSlice("genres"))
	require.Equal(t, string(expected.Kind), actual.GetString("kind"))
	require.Equal(t, expected.Place, actual.GetString("place"))
}

func assertEqualEvents(t *testing.T, expected []application.Event, actual []*core.Record) {
	require.Equal(t, len(expected), len(actual))
	for i, event := range expected {
		assertEqualEvent(t, event, actual[i])
	}
}
