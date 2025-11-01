package main

import (
	"log/slog"
	"os"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/leorolland/sortir.in/pkg/infrastructure/client/pb"
	"github.com/leorolland/sortir.in/pkg/infrastructure/collector"
)

func main() {
	if len(os.Args) < 2 {
		slog.Error("Missing city argument. Usage: allevents <city>")
		os.Exit(1)
	}
	city := os.Args[1]

	collector := collector.NewAllEventsCollector()
	eventSaver := pb.NewPBClient("http://localhost:8090")
	populator := application.NewPopulator(collector, eventSaver)

	slog.Info("Populate events for " + city)
	err := populator.Populate(application.CollectLocation{
		City:   city,
		Radius: 100000,
	})
	if err != nil {
		slog.Error("Failed to populate events for "+city, "error", err)
		os.Exit(1)
	}
	slog.Info("Events populated")
}
