//go:build ignore
// +build ignore

package main

import (
	l "github.com/sprisa/west/util/log"
	"github.com/sprisa/west/westport/db/migrate"
)

func main() {
	err := migrate.Migrate()
	if err != nil {
		l.Log.Err(err).Msg("error migrating")
		return
	}
	l.Log.Print("✨ Successfully ran migration on db.")
}
