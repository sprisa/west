package main

import (
	"context"
	"os"

	l "github.com/sprisa/west/util/log"
	"github.com/sprisa/west/util/sig"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx := sig.ShutdownContext(context.Background())
	cmd := &cli.Command{
		Name:      "west",
		Usage:     "mesh networking",
		UsageText: "west start [jwt_token]",
		Commands: []*cli.Command{
			StartCommand,
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return cli.ShowAppHelp(cmd)
		},
	}

	err := cmd.Run(ctx, os.Args)
	if err != nil {
		l.Log.Error().Msg(err.Error())
		defer os.Exit(1)
	}
}
