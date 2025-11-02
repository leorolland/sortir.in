package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/leorolland/sortir.in/pkg/infrastructure/collector"
)

func main() {
	if len(os.Args) < 2 {
		slog.Error("Missing number of locations argument. Usage: collect <num_locations> <collector>")
		os.Exit(1)
	}

	numLocations, err := strconv.Atoi(os.Args[1])
	if err != nil {
		slog.Error("Invalid number of locations. Must be an integer.", "error", err)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		slog.Error("Missing collector name. Usage: collect <num_locations> <collector>")
		os.Exit(1)
	}

	collector := getCollectorByName(os.Args[2])
	if collector == nil {
		slog.Error("Invalid collector name. Usage: collect <num_locations> <collector>")
		os.Exit(1)
	}

	locationsIterator := application.NewFrenchCitiesIterator()

	var allEvents []application.Event
	for i := 0; i < numLocations; i++ {
		location := locationsIterator.Next()
		if location == nil {
			// No more locations available
			break
		}

		events, err := collector.Collect(*location)
		if err != nil {
			slog.Error("Failed to collect events", "city", location.City, "error", err)
			continue
		}

		allEvents = append(allEvents, events...)
	}

	jsonData, err := json.Marshal(allEvents)
	if err != nil {
		slog.Error("Failed to marshal events", "error", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
}

func getCollectorByName(name string) application.Collector {
	switch name {
	case "allevents":
		return collector.NewAllEventsCollector()
	case "bobine":
		return collector.NewBobineCollector()
	default:
		return nil
	}
}
