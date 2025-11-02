package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1687431684")
		if err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(6, []byte(`{
			"hidden": false,
			"id": "geoPoint2287119580",
			"name": "loc",
			"presentable": true,
			"required": false,
			"system": false,
			"type": "geoPoint"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1687431684")
		if err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(6, []byte(`{
			"hidden": false,
			"id": "geoPoint2287119580",
			"name": "loc",
			"presentable": true,
			"required": true,
			"system": false,
			"type": "geoPoint"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
