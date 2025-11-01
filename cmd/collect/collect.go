package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/leorolland/sortir.in/pkg/infrastructure/collector"
)

func main() {
	if len(os.Args) < 2 {
		slog.Error("Missing city argument. Usage: allevents <city>")
		os.Exit(1)
	}
	city := os.Args[1]

	collector := getCollectorByName(os.Args[2])
	if collector == nil {
		slog.Error("Invalid collector name. Usage: collect <city> <collector>")
		os.Exit(1)
	}

	events, err := collector.Collect(application.CollectLocation{
		City:   city,
		Radius: 100000,
	})
	if err != nil {
		slog.Error("Failed to collect events for "+city, "error", err)
		os.Exit(1)
	}

	// show as json
	jsonData, err := json.Marshal(events)
	if err != nil {
		slog.Error("Failed to marshal events for "+city, "error", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
}

func getCollectorByName(name string) application.Collector {
	switch name {
	case "allevents":
		return collector.NewAllEventsCollector()
	default:
		return nil
	}
}
