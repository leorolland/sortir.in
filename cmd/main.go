package main

import (
	"log"

	"github.com/leorolland/sortir.in/pkg/infrastructure/server"

	_ "github.com/leorolland/sortir.in/migrations"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	server.RegisterApp(app)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
