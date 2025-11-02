package collector

import (
	"log/slog"
	"reflect"

	"github.com/leorolland/sortir.in/pkg/application"
)

type compositeCollector struct {
	collectors []application.Collector
}

func NewCompositeCollector(collectors ...application.Collector) application.Collector {
	return &compositeCollector{collectors: collectors}
}

func (c *compositeCollector) Collect(location application.CollectLocation) ([]application.Event, error) {
	allEvents := []application.Event{}
	for _, collector := range c.collectors {
		collectorEvents, err := collector.Collect(location)
		if err != nil {
			return nil, err
		}
		slog.Info("Collected events", "count", len(collectorEvents), "location", location.City, "collector", reflect.TypeOf(collector))
		allEvents = append(allEvents, collectorEvents...)
	}
	return allEvents, nil
}
