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
	// Label is the short place title displayed on the pin ("Saint André").
	// Empty when the events have no known place: the pin renders without label.
	Label string `json:"label,omitempty"`
	// Place is the raw venue name the label is derived from. It is only
	// meaningful on the per-event rows returned by the repository; it never
	// reaches the API payload.
	Place string `json:"-"`
}

type PinsService interface {
	GetPins(bounds Bounds, minDate, maxDate time.Time) ([]Pin, error)
}

type pins struct {
	eventRepository EventRepository
}

func NewPins(eventRepository EventRepository) PinsService {
	return &pins{
		eventRepository: eventRepository,
	}
}

func (p *pins) GetPins(bounds Bounds, minDate, maxDate time.Time) ([]Pin, error) {
	events, err := p.eventRepository.ByBoundsAndDateRange(bounds, minDate, maxDate)
	if err != nil {
		return nil, err
	}

	type pinGroup struct {
		pin    Pin
		places map[string]int
	}

	groups := make(map[string]*pinGroup)

	// group pins at the same location and kind together, remembering how
	// often each place name shows up so the pin can carry a label
	for _, event := range events {
		key := getLocationKindKey(event.Loc, event.Kind)

		group, exists := groups[key]
		if !exists {
			group = &pinGroup{
				pin: Pin{
					Loc:  event.Loc,
					Kind: event.Kind,
				},
				places: make(map[string]int),
			}
			groups[key] = group
		}

		group.pin.Amount++
		// Address-like places ("13 Rue Henri") make poor labels: skip them,
		// so a real venue name at the same location wins the count.
		if event.Place != "" && !isAddressLikePlace(event.Place) {
			group.places[event.Place]++
		}
	}

	pins := make([]Pin, 0, len(groups))
	for _, group := range groups {
		pin := group.pin
		pin.Label = ShortPlaceTitle(bestPlace(group.places))
		pins = append(pins, pin)
	}

	slices.SortFunc(pins, func(a, b Pin) int {
		return cmp.Compare(getLocationKindKey(a.Loc, a.Kind), getLocationKindKey(b.Loc, b.Kind))
	})

	return pins, nil
}

// bestPlace returns the most frequent place among the counted ones, breaking
// ties lexicographically so the label is stable across runs.
func bestPlace(places map[string]int) string {
	best, bestCount := "", 0
	for place, count := range places {
		if count > bestCount || (count == bestCount && place < best) {
			best, bestCount = place, count
		}
	}
	return best
}

// getLocationKindKey creates a unique key for a location and kind combination
func getLocationKindKey(loc EventLocation, kind Kind) string {
	return fmt.Sprintf("%f:%f:%s", loc.Lat, loc.Lon, kind)
}
