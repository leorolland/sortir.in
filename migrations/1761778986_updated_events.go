package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1687431684")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_tuX5bqYCzb` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `name` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_OD9hQVHwr1` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `begin` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_WAV5wspmIS` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `end` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_h2O7hIvG2R` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `loc` + "`" + `)"
			]
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1687431684")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_tuX5bqYCzb` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `name` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_OD9hQVHwr1` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `begin` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_WAV5wspmIS` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `end` + "`" + `)"
			]
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
