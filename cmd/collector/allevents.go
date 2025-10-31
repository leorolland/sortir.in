package main

import (
	"log/slog"
	"os"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/leorolland/sortir.in/pkg/infrastructure/client/pb"
	"github.com/leorolland/sortir.in/pkg/infrastructure/collector"
)

func main() {
	collector := collector.NewAllEventsCollector()
	eventSaver := pb.NewPBClient("http://localhost:8090")
	populator := application.NewPopulator(collector, eventSaver)

	slog.Info("Populate events for Rennes")
	err := populator.Populate(application.CollectLocation{
		City:   "rennes",
		Radius: 100000,
	})
	if err != nil {
		slog.Error("Failed to populate events for Rennes", "error", err)
		os.Exit(1)
	}
	slog.Info("Events populated")
}
