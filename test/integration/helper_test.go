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
