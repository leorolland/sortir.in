package main

import (
	"log"

	"github.com/leorolland/sortir.in/pkg/infrastructure/server"

	_ "github.com/leorolland/sortir.in/migrations"

	"github.com/pocketbase/pocketbase"
)

func main() {
	app := pocketbase.New()

	server.RegisterApp(app)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
