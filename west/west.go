package west

import (
	"context"
	"os"

	l "github.com/sprisa/west/util/log"
	"github.com/sprisa/west/util/sig"
	"github.com/sprisa/west/westport"
	"github.com/urfave/cli/v3"
)

func Main() {
	ctx := sig.ShutdownContext(context.Background())
	err := WestCommand.Run(ctx, os.Args)
	if err != nil {
		l.Log.Error().Msg(err.Error())
		defer os.Exit(1)
	}
}

var WestCommand = &cli.Command{
	Name:      "west",
	Usage:     "mesh networking",
	UsageText: "west start [jwt_token]",
	Commands: []*cli.Command{
		StartCommand,
		westport.WestPortCommand,
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		return cli.ShowAppHelp(cmd)
	},
}
