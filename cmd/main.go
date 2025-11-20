package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/leorolland/sortir.in/cmd/requests"
	"github.com/leorolland/sortir.in/cmd/webauthn"
	"github.com/leorolland/sortir.in/ui"

	_ "github.com/leorolland/sortir.in/migrations"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	app := pocketbase.New()

	var migrationsDir string
	app.RootCmd.PersistentFlags().StringVar(
		&migrationsDir,
		"migrationsDir",
		"pb_migrations",
		"the directory with the user defined migrations",
	)

	var automigrate bool
	app.RootCmd.PersistentFlags().BoolVar(
		&automigrate,
		"automigrate",
		true,
		"enable/disable auto migrations",
	)

	var indexFallback bool
	app.RootCmd.PersistentFlags().BoolVar(
		&indexFallback,
		"indexFallback",
		true,
		"fallback the request to index.html on missing static path (eg. when pretty urls are used with SPA)",
	)

	app.RootCmd.ParseFlags(os.Args[1:])

	// register the `migrate` command
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
	})

	webauthn.Register(app)

	app.Cron().MustAdd("hourly_delete_expired_events", "* * * * *", func() {
		_, err := app.DB().Delete("events", dbx.NewExp("end < {:now}", dbx.Params{"now": time.Now().Format("2006-01-02 15:04:05")})).Execute()
		if err != nil {
			app.Logger().Error("failed to delete expired events", "error", err)
		}
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/{path...}", apis.Static(ui.BuildDirFS, true)).Bind(apis.Gzip())
		se.Router.PUT("/api/events", requests.PutEvents)

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
