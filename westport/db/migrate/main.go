//go:build ignore
// +build ignore

package main

import (
	"github.com/sprisa/west/westport/db/migrate"
	l "github.com/sprisa/x/log"
)

func main() {
	err := migrate.Migrate()
	if err != nil {
		l.Log.Err(err).Msg("error migrating")
		return
	}
	l.Log.Print("✨ Successfully ran migration on db.")
}
