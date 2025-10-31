package application

import "log/slog"

type EventSaver interface {
	SaveEvents(events []Event) error
}

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
