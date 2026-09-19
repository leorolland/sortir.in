package application

import "time"

type Bounds struct {
	North float64
	South float64
	East  float64
	West  float64
}

//go:generate go run github.com/golang/mock/mockgen -destination=mocks/mock_event_repository.go -package=applicationmocks github.com/leorolland/sortir.in/pkg/application EventRepository
type EventRepository interface {
	ByBoundsAndDateRange(bounds Bounds, minDate, maxDate time.Time) ([]Pin, error)
}
