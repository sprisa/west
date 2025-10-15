package westport

import (
	"context"

	"github.com/urfave/cli/v3"
)

var WestPortCommand = &cli.Command{
	Name:      "port",
	Usage:     "West port coordination server",
	UsageText: "west port install",
	Commands: []*cli.Command{
		InstallCommand,
		StartCommand,
		AddCommand,
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		return cli.ShowSubcommandHelp(cmd)
	},
}
