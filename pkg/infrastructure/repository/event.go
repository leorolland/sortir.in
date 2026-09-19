package repository

import (
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/tools/types"
)

type eventRepository struct {
	db DBGetter
}

func NewEventRepository(db DBGetter) eventRepository {
	return eventRepository{db: db}
}

func (r eventRepository) ByBoundsAndDateRange(bounds application.Bounds, minDate, maxDate time.Time) ([]application.Pin, error) {
	// Dates are stored in the canonical PocketBase layout (UTC), so binding the
	// bounds as strings in the same layout makes the comparisons exact.
	maxDateParam := maxDate.UTC().Format(types.DefaultDateLayout)
	conditions := []dbx.Expression{
		dbx.NewExp("json_extract(loc, '$.lat') >= {:south}", dbx.Params{"south": bounds.South}),
		dbx.NewExp("json_extract(loc, '$.lat') <= {:north}", dbx.Params{"north": bounds.North}),
		dbx.NewExp("json_extract(loc, '$.lon') >= {:west}", dbx.Params{"west": bounds.West}),
		dbx.NewExp("json_extract(loc, '$.lon') <= {:east}", dbx.Params{"east": bounds.East}),
		dbx.NewExp("`begin` <= {:maxDate}", dbx.Params{"maxDate": maxDateParam}),
	}

	if !minDate.IsZero() {
		minDateParam := minDate.UTC().Format(types.DefaultDateLayout)
		conditions = append(conditions, dbx.NewExp("`end` >= {:minDate}", dbx.Params{"minDate": minDateParam}))
	}

	query := r.db.Get().Select("kind", "loc").From("events").Where(dbx.And(conditions...)).Limit(5000)

	var rows []struct {
		Kind string                 `db:"kind"`
		Loc  types.JSONMap[float64] `db:"loc"`
	}

	err := query.All(&rows)
	if err != nil {
		return nil, err
	}

	pins := make([]application.Pin, len(rows))
	for i, row := range rows {
		pins[i] = application.Pin{
			Kind:   application.Kind(row.Kind),
			Loc:    application.EventLocation{Lat: row.Loc.Get("lat"), Lon: row.Loc.Get("lon")},
			Amount: 1,
		}
	}

	return pins, nil
}
