package migrate

import (
	"context"
	"fmt"

	"github.com/sprisa/west/westport/db"
	"github.com/sprisa/west/westport/db/ent/migrate"
)

func Migrate() error {
	client, err := db.OpenDB()
	if err != nil {
		return err
	}

	// Run the auto migration tool.
	err = client.Schema.Create(
		context.Background(),
		migrate.WithGlobalUniqueID(true),
		migrate.WithForeignKeys(false),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	)
	if err != nil {
		return fmt.Errorf("failed migrating db: %v", err)
	}

	return nil
}
