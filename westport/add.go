package main

import (
	"context"
	"errors"

	"github.com/sprisa/west/util/errutil"
	"github.com/sprisa/west/util/ipconv"
	"github.com/sprisa/west/westport/db"
	"github.com/sprisa/west/westport/db/ent"
	"github.com/sprisa/west/westport/db/migrate"
	"github.com/urfave/cli/v3"
)

var AddCommand = &cli.Command{
	Name:      "add",
	Usage:     "Register a new west device",
	UsageText: "west-port add",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Required: true,
			Usage:    "Device name. Must be unique.",
		},
		&cli.StringFlag{
			Name:     "ip",
			Usage:    "IP for device. Must be unique within existing cidr.",
			Required: true,
			Validator: func(s string) error {
				_, err := ipconv.ParseToIP(s)
				return err
			},
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		client, err := db.OpenDB()
		if err != nil {
			return errutil.WrapError(err, "error opening db")
		}
		defer client.Close()
		err = migrate.MigrateClient(ctx, client)
		if err != nil {
			return errutil.WrapError(err, "error migrating db")
		}

		err = promptEncryptionPassword()
		if err != nil {
			return err
		}

		settings, err := client.Settings.Query().Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return errors.New("error finding settings. Trying installing first.")
			}
			return errutil.WrapError(err, "error initializing settings")
		}


		print(settings.Cidr.String())

		return nil
	},
}
