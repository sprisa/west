package main

import (
	"context"
	"fmt"

	"reship/util/print"

	"github.com/golang-jwt/jwt/v5"
	"github.com/urfave/cli/v3"
)

type TokenClaims struct {
	Endpoint string `json:"endpoint"`
	jwt.RegisteredClaims
}

var StartCommand = &cli.Command{
	Name:      "start",
	Usage:     "Start west device",
	UsageText: "west start [jwt_token]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			// TODO: Would be better if this was set in a different command
			// then stored in a secure enclave for start.
			Name:     "token",
			Aliases:  []string{"t"},
			Required: true,
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		token := c.String("token")
		parser := jwt.NewParser()
		claims := &TokenClaims{}
		info, _, err := parser.ParseUnverified(token, claims)
		if err != nil {
			return fmt.Errorf("error parsing token: %w", err)
		}
		print.PrettyPrint(info)

		print.PrettyPrint(claims)

		return nil
	},
}
