//go:build ignore
// +build ignore

package main

import (
	"os"

	"github.com/sprisa/west/westport/db/migrate"
	l "github.com/sprisa/x/log"
)

func main() {
	dataSource := "sqlite"
	if len(os.Args) > 1 {
		dataSource = os.Args[1]
	}
	err := migrate.Migrate(dataSource)
	if err != nil {
		l.Log.Err(err).Msg("error migrating")
		return
	}
	l.Log.Print("✨ Successfully ran migration on db.")
}
