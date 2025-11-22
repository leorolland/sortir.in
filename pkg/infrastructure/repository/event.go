package repository

import (
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/pocketbase/dbx"
)

type EventRepository interface {
	ByBoundsAndMaxDate(bounds application.Bounds, maxDate time.Time) ([]application.Event, error)
}

type eventRepository struct {
	db dbx.Builder
}

func NewEventRepository(db dbx.Builder) eventRepository {
	return eventRepository{db: db}
}

func (r eventRepository) ByBoundsAndMaxDate(bounds application.Bounds, maxDate time.Time) ([]application.Event, error) {
	query := r.db.Select("kind", "loc").From("events").Where(dbx.And(
		dbx.NewExp("loc.lat >= :south", dbx.Params{"south": bounds.South}),
		dbx.NewExp("loc.lat <= :north", dbx.Params{"north": bounds.North}),
		dbx.NewExp("loc.lon >= :west", dbx.Params{"west": bounds.West}),
		dbx.NewExp("loc.lon <= :east", dbx.Params{"east": bounds.East}),
		dbx.NewExp("end <= :maxDate", dbx.Params{"maxDate": maxDate}),
	)).Limit(1000)

	var events []application.Event

	err := query.All(&events)
	if err != nil {
		return nil, err
	}

	return events, nil
}
