package main

import (
	"context"

	"github.com/sprisa/west/util/errutil"
	l "github.com/sprisa/west/util/log"
	"github.com/sprisa/west/util/sig"
	"github.com/sprisa/west/westport/db"
	"github.com/sprisa/west/westport/db/ent"
	"github.com/sprisa/west/westport/db/migrate"
)

func main() {
	ctx := sig.ShutdownContext(context.Background())
	client, err := db.OpenDB()
	errutil.InvariantError(err, "error opening db")
	defer client.Close()
	err = migrate.MigrateClient(ctx, client)
	errutil.InvariantError(err, "error migrating db")

	settings, err := client.Settings.Query().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			settings, err = client.Settings.Create().Save(ctx)
			errutil.InvariantError(err, "error initializing settings")
		} else {
			errutil.InvariantError(err, "error fetching settings")
		}
	}

	l.Log.Info().Msgf("settings: %+v", settings)
}
