package main

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/leorolland/sortir.in/pkg/infrastructure/client/pb"
	"github.com/leorolland/sortir.in/pkg/infrastructure/collector"
)

func main() {
	if len(os.Args) < 2 {
		slog.Error("Missing limit argument. Usage: populate <limit>")
		os.Exit(1)
	}

	limit, err := strconv.Atoi(os.Args[1])
	if err != nil {
		slog.Error("Invalid limit argument. Must be an integer", "error", err)
		os.Exit(1)
	}

	compositeCollector := collector.NewCompositeCollector(
		collector.NewAllEventsCollector(),
		collector.NewBobineCollector(),
		collector.NewParisEventsCollector(),
	)

	eventSaver := pb.NewPBClient("http://localhost:8090")
	populator := application.NewPopulator(compositeCollector, eventSaver)

	slog.Info("Populating events", "location_limit", limit)

	iterator := application.NewFrenchCitiesIterator()
	locationsProcessed := 0

	for locationsProcessed < limit {
		location := iterator.Next()
		if location == nil {
			slog.Info("No more locations available")
			break
		}

		slog.Info("Processing location", "city", location.City)

		err := populator.Populate(*location)
		if err != nil {
			slog.Error("Failed to populate events", "city", location.City, "error", err)
			continue
		}

		locationsProcessed++
		slog.Info("Events populated successfully", "city", location.City)
	}

	slog.Info("All locations processed", "count", locationsProcessed)
}
