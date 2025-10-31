package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/leorolland/sortir.in/cmd/webauthn"
	"github.com/leorolland/sortir.in/ui"

	_ "github.com/leorolland/sortir.in/migrations"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/types"
)

func main() {
	app := pocketbase.New()

	var hooksDir string
	app.RootCmd.PersistentFlags().StringVar(
		&hooksDir,
		"hooksDir",
		"",
		"the directory with the JS app hooks",
	)

	var hooksWatch bool
	app.RootCmd.PersistentFlags().BoolVar(
		&hooksWatch,
		"hooksWatch",
		true,
		"auto restart the app on pb_hooks file change",
	)

	var hooksPool int
	app.RootCmd.PersistentFlags().IntVar(
		&hooksPool,
		"hooksPool",
		25,
		"the total prewarm goja.Runtime instances for the JS app hooks execution",
	)

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

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/{path...}", apis.Static(ui.BuildDirFS, true)).Bind(apis.Gzip())

		se.Router.GET("/api/go-hello", func(e *core.RequestEvent) error {
			return e.JSON(http.StatusOK, map[string]string{"message": "Hello world from Go!"})
		})

		se.Router.GET("/api/add-events", func(e *core.RequestEvent) error {
			// Insert in smaller batches to avoid too large statements
			batchSize := 100
			totalCount := 100_000

			for i := 0; i < totalCount; i += batchSize {
				// Calculate actual batch size (last batch might be smaller)
				currentBatchSize := batchSize
				if i+currentBatchSize > totalCount {
					currentBatchSize = totalCount - i
				}

				// Build a multi-row insert for this batch
				sql := "INSERT INTO events (name, loc) VALUES "
				params := dbx.Params{}

				for j := 0; j < currentBatchSize; j++ {
					if j > 0 {
						sql += ", "
					}
					paramName1 := fmt.Sprintf("name%d", j)
					paramName2 := fmt.Sprintf("loc%d", j)
					sql += fmt.Sprintf("({:%s}, {:%s})", paramName1, paramName2)

					params[paramName1] = "Event 1"
					params[paramName2] = types.GeoPoint{
						Lat: rand.Float64()*180 - 90,
						Lon: rand.Float64()*360 - 180,
					}
				}

				// Execute this batch
				_, err := app.DB().NewQuery(sql).Bind(params).Execute()
				if err != nil {
					return apis.NewBadRequestError("Failed to insert events: "+err.Error(), nil)
				}
			}

			return e.JSON(http.StatusOK, map[string]string{"message": "ok"})
		})

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
