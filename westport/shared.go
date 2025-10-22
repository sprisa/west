package westport

import (
	"bytes"
	"io"
	"os"

	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/input"
	"github.com/sprisa/west/util/ioutil"
	"github.com/sprisa/west/westport/db/helpers"
)

func readEncryptionPassword() (err error) {
	var pswd string
	// Read from stdin if available
	if ioutil.StdinAvailable() {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		pswd = string(bytes.TrimSpace(b))
	} else {
		pswd, err = prompt.New().Ask("password:").
			Input("", input.WithEchoMode(input.EchoPassword), input.WithHelp(true))
		if err != nil {
			return err
		}
	}

	copy(helpers.EncryptionKey[:], pswd)
	// l.Log.Info().Msg(pswd)
	// l.Log.Info().Msgf("key: %s", string(helpers.EncryptionKey[:]))
	return nil
}
