package application

//go:generate go run github.com/golang/mock/mockgen -destination=mocks/mock_event_repository.go -package=applicationmocks github.com/leorolland/sortir.in/pkg/application EventRepository

import "time"

type Bounds struct {
	North float64
	South float64
	East  float64
	West  float64
}

type EventRepository interface {
	ByBoundsAndMaxDate(bounds Bounds, maxDate time.Time) ([]Event, error)
}
