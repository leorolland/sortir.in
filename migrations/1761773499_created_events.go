package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `{
			"createRule": null,
			"deleteRule": null,
			"fields": [
				{
					"autogeneratePattern": "[a-z0-9]{15}",
					"hidden": false,
					"id": "text3208210256",
					"max": 15,
					"min": 15,
					"name": "id",
					"pattern": "^[a-z0-9]+$",
					"presentable": false,
					"primaryKey": true,
					"required": true,
					"system": true,
					"type": "text"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text1579384326",
					"max": 0,
					"min": 0,
					"name": "name",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": false,
					"system": false,
					"type": "text"
				},
				{
					"hidden": false,
					"id": "date2055574805",
					"max": "",
					"min": "",
					"name": "begin",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "date"
				},
				{
					"hidden": false,
					"id": "date16528305",
					"max": "",
					"min": "",
					"name": "end",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "date"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text2287119580",
					"max": 0,
					"min": 0,
					"name": "loc",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": false,
					"system": false,
					"type": "text"
				}
			],
			"id": "pbc_1687431684",
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_Fzakmt1s3P` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `name` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_rQgmK6d8MG` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `begin` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_CgRFiDxNlc` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `end` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_eMYMur2IyZ` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `loc` + "`" + `)"
			],
			"listRule": null,
			"name": "events",
			"system": false,
			"type": "base",
			"updateRule": null,
			"viewRule": null
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1687431684")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
