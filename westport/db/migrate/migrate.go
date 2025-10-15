package migrate

import (
	"context"
	"fmt"

	"github.com/sprisa/west/util/sig"
	"github.com/sprisa/west/westport/db"
	"github.com/sprisa/west/westport/db/ent"
	"github.com/sprisa/west/westport/db/ent/migrate"
)

func Migrate() error {
	ctx := sig.ShutdownContext(context.Background())
	client, err := db.OpenDB()
	if err != nil {
		return err
	}

	// Run the auto migration tool.
	err = MigrateClient(ctx, client)
	if err != nil {
		return fmt.Errorf("failed migrating db: %v", err)
	}

	return nil
}

func MigrateClient(ctx context.Context, client *ent.Client) error {
	return client.Schema.Create(
		ctx,
		migrate.WithGlobalUniqueID(true),
		migrate.WithForeignKeys(false),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	)
}
