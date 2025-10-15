package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"

	"entgo.io/ent/dialect"
	"github.com/sprisa/west/util/errutil"
	l "github.com/sprisa/west/util/log"
	"github.com/sprisa/west/westport/db/ent"
	_ "github.com/sprisa/west/westport/db/ent/runtime"
	"modernc.org/sqlite"
)

var DBFilePath string = "westdb"

func OpenDB() (*ent.Client, error) {
	sql.Register("sqlite3", &sqliteDriver{})
	l.Log.Debug().Msgf("DB Open: %s", DBFilePath)
	_, err := os.Stat(DBFilePath)
	if err != nil && os.IsNotExist(err) == false {
		return nil, err
	}

	return ent.Open(
		dialect.SQLite,
		fmt.Sprintf("file:%s?mode=rwc&cache=shared&_fk=1", DBFilePath),
	)
}

type sqliteDriver struct {
	sqlite.Driver
}

// https://github.com/ent/ent/discussions/1667#discussioncomment-1132296
func (d sqliteDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.Driver.Open(name)
	if err != nil {
		return conn, err
	}
	c := conn.(interface {
		Exec(stmt string, args []driver.Value) (driver.Result, error)
	})
	if _, err := c.Exec("PRAGMA foreign_keys = on;", nil); err != nil {
		conn.Close()
		return nil, errutil.WrapError(err, "error enabled foreign_keys")
	}
	return conn, nil
}
