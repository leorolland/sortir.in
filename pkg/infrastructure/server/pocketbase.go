package server

import (
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/leorolland/sortir.in/pkg/infrastructure/repository"
	"github.com/leorolland/sortir.in/pkg/infrastructure/server/requests"
	"github.com/leorolland/sortir.in/ui"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

func RegisterApp(app *pocketbase.PocketBase) {
	initServices(app)
	bindRoutes(app)
	bindCrons(app)
}

func initServices(app *pocketbase.PocketBase) {
	dbGetter := repository.NewDBGetter(app)
	eventRepository := repository.NewEventRepository(dbGetter)
	app.Store().Set("pinsService", application.NewPins(eventRepository))
}

func bindRoutes(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/{path...}", apis.Static(ui.BuildDirFS, true)).Bind(apis.Gzip())
		se.Router.PUT("/api/events", requests.PutEvents)
		se.Router.GET("/api/pins", requests.GetPins)
		se.Router.GET("/api/version", requests.GetVersion)
		return se.Next()
	})
}

func bindCrons(app *pocketbase.PocketBase) {
	app.Cron().MustAdd("delete_expired_events_cron", "* * * * *", func() {
		if err := DeleteExpiredEvents(app); err != nil {
			app.Logger().Error("failed to delete expired events", "error", err)
		}
	})
}

// DeleteExpiredEvents removes events that have already ended.
func DeleteExpiredEvents(app core.App) error {
	_, err := app.DB().Delete("events",
		dbx.NewExp("end < {:now}", dbx.Params{"now": time.Now().UTC().Format(types.DefaultDateLayout)}),
	).Execute()
	return err
}
