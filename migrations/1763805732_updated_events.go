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
				"CREATE UNIQUE INDEX ` + "`" + `idx_tuX5bqYCzb` + "`" + ` ON ` + "`" + `events` + "`" + ` (\n  ` + "`" + `name` + "`" + `,\n  ` + "`" + `begin` + "`" + `,\n  ` + "`" + `end` + "`" + `\n)",
				"-- Create the composite functional index\nCREATE INDEX idx_events_lat_lon_bbox ON events (\n    (CASE WHEN json_valid(loc) THEN JSON_EXTRACT(loc, '$.lat') ELSE JSON_EXTRACT(json_object('pb', loc), '$.pb.lat') END),\n    (CASE WHEN json_valid(loc) THEN JSON_EXTRACT(loc, '$.lon') ELSE JSON_EXTRACT(json_object('pb', loc), '$.pb.lon') END)\n);"
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
				"CREATE UNIQUE INDEX ` + "`" + `idx_tuX5bqYCzb` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `name` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_OD9hQVHwr1` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `begin` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_WAV5wspmIS` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `end` + "`" + `)",
				"-- Create the composite functional index\nCREATE INDEX idx_events_lat_lon_bbox ON events (\n    (CASE WHEN json_valid(loc) THEN JSON_EXTRACT(loc, '$.lat') ELSE JSON_EXTRACT(json_object('pb', loc), '$.pb.lat') END),\n    (CASE WHEN json_valid(loc) THEN JSON_EXTRACT(loc, '$.lon') ELSE JSON_EXTRACT(json_object('pb', loc), '$.pb.lon') END)\n);",
				"CREATE INDEX ` + "`" + `idx_zhZ64dCrB8` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `address` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_kPSe8tdJ5U` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `price` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_VXjaZv1E3x` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `kind` + "`" + `)"
			]
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
