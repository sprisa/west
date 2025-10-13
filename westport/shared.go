package main

import (
	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/input"
	"github.com/sprisa/west/westport/db/helpers"
)

func promptEncryptionPassword() error {
	pswd, err := prompt.New().Ask("password:").
		Input("", input.WithEchoMode(input.EchoPassword), input.WithHelp(true))
	if err != nil {
		return err
	}
	copy(helpers.EncryptionKey[:], pswd)
	// l.Log.Info().Msg(pswd)
	// l.Log.Info().Msgf("key: %s", string(helpers.EncryptionKey[:]))
	return nil
}
