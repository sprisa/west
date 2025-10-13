package db

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/sprisa/west/westport/db/ent"
	_ "github.com/sprisa/west/westport/db/ent/runtime"
)

func OpenDB() (*ent.Client, error) {
	return ent.Open("sqlite3", "file:westdb?mode=rwc&cache=shared&_fk=1")
}
