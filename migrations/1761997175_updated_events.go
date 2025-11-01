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

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(2, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text1002749145",
			"max": 0,
			"min": 0,
			"name": "kind",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": true,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(6, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text1948079053",
			"max": 255,
			"min": 0,
			"name": "place",
			"pattern": "",
			"presentable": true,
			"primaryKey": false,
			"required": false,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(7, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text223244161",
			"max": 255,
			"min": 0,
			"name": "address",
			"pattern": "",
			"presentable": true,
			"primaryKey": false,
			"required": false,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(8, []byte(`{
			"hidden": false,
			"id": "number3402113753",
			"max": null,
			"min": 0,
			"name": "price",
			"onlyInt": false,
			"presentable": true,
			"required": false,
			"system": false,
			"type": "number"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(9, []byte(`{
			"exceptDomains": null,
			"hidden": false,
			"id": "url1602912115",
			"name": "source",
			"onlyDomains": null,
			"presentable": false,
			"required": false,
			"system": false,
			"type": "url"
		}`)); err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(4, []byte(`{
			"hidden": false,
			"id": "date16528305",
			"max": "",
			"min": "",
			"name": "end",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "date"
		}`)); err != nil {
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
				"-- Create the composite functional index\nCREATE INDEX idx_events_lat_lon_bbox ON events (\n    (CASE WHEN json_valid(loc) THEN JSON_EXTRACT(loc, '$.lat') ELSE JSON_EXTRACT(json_object('pb', loc), '$.pb.lat') END),\n    (CASE WHEN json_valid(loc) THEN JSON_EXTRACT(loc, '$.lon') ELSE JSON_EXTRACT(json_object('pb', loc), '$.pb.lon') END)\n);"
			]
		}`), &collection); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("text1002749145")

		// remove field
		collection.Fields.RemoveById("text1948079053")

		// remove field
		collection.Fields.RemoveById("text223244161")

		// remove field
		collection.Fields.RemoveById("number3402113753")

		// remove field
		collection.Fields.RemoveById("url1602912115")

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(3, []byte(`{
			"hidden": false,
			"id": "date16528305",
			"max": "",
			"min": "",
			"name": "end",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "date"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
