package db

import (
	"github.com/sprisa/west/westport/db/ent"
	_ "github.com/sprisa/west/westport/db/ent/runtime"
)

func OpenDB() (*ent.Client, error) {
	return ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
}
