package application

import (
	"cmp"
	"fmt"
	"slices"
	"time"
)

// Pin represents a pin on the map
// It contains the location, kind and amount of events at that location
// It has two goals:
//  1. Reduce the size of payloads sent to the client (compared to sending all events)
//  2. Display events which are in the same place in one pin instead of a stack of pins
type Pin struct {
	Loc    EventLocation `json:"loc"`
	Kind   Kind          `json:"kind"`
	Amount int           `json:"amount"`
}

// PinsService defines the interface for the pins service
type PinsService interface {
	GetPins(bounds Bounds, maxDate time.Time) ([]Pin, error)
}

type pins struct {
	eventRepository EventRepository
}

func NewPins(eventRepository EventRepository) PinsService {
	return &pins{
		eventRepository: eventRepository,
	}
}

func (p *pins) GetPins(bounds Bounds, maxDate time.Time) ([]Pin, error) {
	events, err := p.eventRepository.ByBoundsAndMaxDate(bounds, maxDate)
	if err != nil {
		return nil, err
	}

	pinsMap := make(map[string]Pin)

	for _, event := range events {
		key := getLocationKindKey(event.Loc, event.Kind)

		pin, exists := pinsMap[key]
		if exists {
			pin.Amount++
		} else {
			pin = Pin{
				Loc:    event.Loc,
				Kind:   event.Kind,
				Amount: 1,
			}
		}

		pinsMap[key] = pin
	}

	pins := make([]Pin, 0, len(pinsMap))
	for _, pin := range pinsMap {
		pins = append(pins, pin)
	}

	slices.SortFunc(pins, func(a, b Pin) int {
		return cmp.Compare(getLocationKindKey(a.Loc, a.Kind), getLocationKindKey(b.Loc, b.Kind))
	})

	return pins, nil
}

// getLocationKindKey creates a unique key for a location and kind combination
func getLocationKindKey(loc EventLocation, kind Kind) string {
	return fmt.Sprintf("%f:%f:%s", loc.Lat, loc.Lon, kind)
}
