package l

import (
	"os"

	"github.com/rs/zerolog"
)

var Log = zerolog.New(zerolog.NewConsoleWriter())

func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" {
		level, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			Log.Err(err).Msg("invalid LOG_LEVEL")
		} else {
			Log = Log.Level(level)
		}
	}
}
