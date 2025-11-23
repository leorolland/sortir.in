package server

import (
	"time"

	"github.com/leorolland/sortir.in/pkg/infrastructure/server/requests"
	"github.com/leorolland/sortir.in/ui"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func RegisterApp(app *pocketbase.PocketBase) {
	bindRoutes(app)
	bindCrons(app)
}

func bindRoutes(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/{path...}", apis.Static(ui.BuildDirFS, true)).Bind(apis.Gzip())
		se.Router.PUT("/api/events", requests.PutEvents)
		return se.Next()
	})
}

func bindCrons(app *pocketbase.PocketBase) {
	app.Cron().MustAdd("delete_expired_events_cron", "* * * * *", func() {
		_, err := app.DB().Delete("events", dbx.NewExp("end < {:now}", dbx.Params{"now": time.Now().Format("2006-01-02 15:04:05")})).Execute()
		if err != nil {
			app.Logger().Error("failed to delete expired events", "error", err)
		}
	})
}
