package repository

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

type DBGetter interface {
	Get() dbx.Builder
}

type dbGetter struct {
	app *pocketbase.PocketBase
}

func NewDBGetter(app *pocketbase.PocketBase) dbGetter {
	return dbGetter{app}
}

func (g dbGetter) Get() dbx.Builder {
	return g.app.DB()
}
