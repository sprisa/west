package main

import (
	"os"
	"path/filepath"

	"github.com/sprisa/west/west"
	"github.com/sprisa/west/westport/db"
	"github.com/sprisa/x/env"
	"github.com/sprisa/x/errutil"
	l "github.com/sprisa/x/log"
)

var isDaemon = env.Parse("WEST_SNAP_MODE", func(val string) bool {
	return val == "daemon"
})
var snapCommonDir = env.Assert("SNAP_COMMON")

var configDirPath = filepath.Join(snapCommonDir, "config")

func init() {
	// The db should be located under the user common dir, not root
	// Use snap common dir for configs
	db.DBFilePath = filepath.Join(configDirPath, db.DBFilePath)
}

func main() {
	if isDaemon {
		l.Log.Info().Msg("Setting up config directory")
		// Setup config dir
		_, err := os.Stat(configDirPath)
		if os.IsNotExist(err) {
			// Give permissions to all since `west port install` is ran with non root
			err = os.Mkdir(configDirPath, 0777)
			errutil.InvariantErr(err, "error creating config dir")
			// Need to set up again with chmod because Go using umask on Linux.
			// This caused the permission in mkdir to be set incorrectly.
			// https://stackoverflow.com/a/61645606/6635914
			err = os.Chmod(configDirPath, 0777)
			errutil.InvariantErr(err, "error chmod config dir")
			l.Log.Info().Str("dir", configDirPath).Msg("Created common config dir")
		}

		// Create db file
		_, err = os.Stat(db.DBFilePath)
		if os.IsNotExist(err) {
			_, err = os.Create(db.DBFilePath)
			errutil.InvariantErr(err, "error creating db file")
			err = os.Chmod(db.DBFilePath, 0777)
			errutil.InvariantErr(err, "error chmod db file")
			l.Log.Info().Str("path", db.DBFilePath).Msg("Created common db file")
		}

		return
	}

	west.Main()
}
