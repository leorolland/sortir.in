package migrations

import (
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// normalizeEventDate parses a stored event date (RFC3339 with any offset, or
// any layout accepted by types.ParseDateTime) and returns it in the canonical
// PocketBase UTC layout, so that string comparisons in SQL queries are valid.
func normalizeEventDate(value string) (string, bool) {
	if value == "" {
		return "", false
	}

	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		dateTime, err := types.ParseDateTime(value)
		if err != nil {
			return "", false
		}
		t = dateTime.Time()
	}

	return t.UTC().Format(types.DefaultDateLayout), true
}

func init() {
	m.Register(func(app core.App) error {
		var rows []struct {
			Id    string `db:"id"`
			Name  string `db:"name"`
			Begin string `db:"begin"`
			End   string `db:"end"`
		}

		err := app.DB().Select("id", "name", "begin", "end").From("events").All(&rows)
		if err != nil {
			return err
		}

		seen := make(map[string]bool, len(rows))

		for _, row := range rows {
			begin, beginOk := normalizeEventDate(row.Begin)
			end, endOk := normalizeEventDate(row.End)
			if !beginOk && !endOk {
				continue
			}

			// The (name, begin, end) unique index can be violated once two rows
			// written with different time zones normalize to the same instant:
			// keep the first one and drop the duplicates.
			key := row.Name + "\x00" + begin + "\x00" + end
			if seen[key] {
				_, err := app.DB().Delete("events", dbx.HashExp{"id": row.Id}).Execute()
				if err != nil {
					return err
				}
				continue
			}
			seen[key] = true

			normalized := make(map[string]any, 2)
			if beginOk {
				normalized["begin"] = begin
			}
			if endOk {
				normalized["end"] = end
			}

			_, err := app.DB().Update("events", normalized, dbx.HashExp{"id": row.Id}).Execute()
			if err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}
