package application

import (
	"fmt"
	"log/slog"
)

type EventSaver interface {
	SaveEvents(events []Event) error
}

// globalSaveBatchSize bounds each SaveEvents call of a nationwide collection:
// the write path upserts one row per event, so tens of thousands of events
// must not arrive in a single request.
const globalSaveBatchSize = 1000

type populator struct {
	collector  Collector
	eventSaver EventSaver
}

func NewPopulator(collector Collector, eventSaver EventSaver) populator {
	return populator{
		collector:  collector,
		eventSaver: eventSaver,
	}
}
func (c *populator) Populate(location CollectLocation) error {
	slog.Info("Collecting events", "city", location.City)
	events, err := c.collector.Collect(location)
	if err != nil {
		return err
	}

	slog.Info("Saving events", "count", len(events))
	return c.eventSaver.SaveEvents(events)
}

type globalPopulator struct {
	collector  GlobalCollector
	eventSaver EventSaver
}

func NewGlobalPopulator(collector GlobalCollector, eventSaver EventSaver) globalPopulator {
	return globalPopulator{
		collector:  collector,
		eventSaver: eventSaver,
	}
}

func (p *globalPopulator) PopulateAll() error {
	events, err := p.collector.CollectAll()
	if err != nil {
		return err
	}

	slog.Info("Saving nationwide events", "count", len(events))
	for start := 0; start < len(events); start += globalSaveBatchSize {
		end := min(start+globalSaveBatchSize, len(events))
		if err := p.eventSaver.SaveEvents(events[start:end]); err != nil {
			return fmt.Errorf("failed to save nationwide events batch %d-%d: %w", start, end, err)
		}
	}
	return nil
}
