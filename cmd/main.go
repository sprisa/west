package main

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
)

var log = zerolog.New(zerolog.NewConsoleWriter())

func main() {
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

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Error().Msg(err.Error())
		defer os.Exit(1)
	}
}
