package collector_test

import (
	"testing"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/leorolland/sortir.in/pkg/infrastructure/collector"
)

func TestRandomCollectorSuccess(t *testing.T) {
	collector := collector.NewRandomCollector()
	events, err := collector.Collect(application.CollectLocation{
		City: "Paris",
		Lon: 2.3522,
		Lat:  48.8566,
		Radius: 10000,
	})
	if err != nil {
		t.Fatalf("failed to collect events: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
}
