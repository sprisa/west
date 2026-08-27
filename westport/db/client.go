package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sprisa/west/westport/db/ent"
	_ "github.com/sprisa/west/westport/db/ent/runtime"
	"github.com/sprisa/x/errutil"
	l "github.com/sprisa/x/log"
	"modernc.org/sqlite"
)

const (
	DatastoreSQLite   = "sqlite"
	DatastorePostgres = "postgres"
	DatastoreMySQL    = "mysql"
)

var DBFilePath = "westdb"

var registerSQLite sync.Once

func DatastoreType(dataSource string) (string, error) {
	switch {
	case dataSource == DatastoreSQLite, strings.HasPrefix(dataSource, "sqlite://"):
		return DatastoreSQLite, nil
	case strings.HasPrefix(dataSource, "postgres://"), strings.HasPrefix(dataSource, "postgresql://"):
		return DatastorePostgres, nil
	case strings.HasPrefix(dataSource, "mysql://"):
		return DatastoreMySQL, nil
	default:
		return "", fmt.Errorf("unsupported datastore %q; use sqlite, sqlite://, postgres://, postgresql://, or mysql://", dataSource)
	}
}

func OpenDB(ctx context.Context, dataSource string) (*ent.Client, error) {
	typeName, err := DatastoreType(dataSource)
	if err != nil {
		return nil, err
	}

	driverName, entDialect, dsn := "", "", ""
	switch typeName {
	case DatastoreSQLite:
		registerSQLite.Do(func() {
			sql.Register("west-sqlite", &sqliteDriver{})
		})
		driverName, entDialect = "west-sqlite", dialect.SQLite
		path := DBFilePath
		if dataSource != DatastoreSQLite {
			path = strings.TrimPrefix(dataSource, "sqlite://")
			if path == "" {
				return nil, errors.New("sqlite datastore path cannot be empty")
			}
		}
		dsn = fmt.Sprintf("file:%s?mode=rwc&cache=shared&_fk=1", path)
	case DatastorePostgres:
		u, err := url.Parse(dataSource)
		if err != nil || strings.TrimPrefix(u.Path, "/") == "" {
			return nil, errors.New("postgres datastore must include a database name")
		}
		driverName, entDialect, dsn = "pgx", dialect.Postgres, dataSource
	case DatastoreMySQL:
		driverName, entDialect = "mysql", dialect.MySQL
		dsn, err = mysqlDSN(dataSource)
		if err != nil {
			return nil, err
		}
	}

	l.Log.Debug().Str("datastore", typeName).Msg("Opening database")
	sqlDB, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("connect to %s datastore: %w", typeName, err)
	}

	drv := entsql.OpenDB(entDialect, sqlDB)
	return ent.NewClient(ent.Driver(drv)), nil
}

func mysqlDSN(dataSource string) (string, error) {
	u, err := url.Parse(dataSource)
	if err != nil {
		return "", fmt.Errorf("parse mysql datastore: %w", err)
	}
	database := strings.TrimPrefix(u.Path, "/")
	if database == "" {
		return "", errors.New("mysql datastore must include a database name")
	}
	username, password := "", ""
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}
	config := &mysql.Config{
		User:      username,
		Passwd:    password,
		Net:       "tcp",
		Addr:      u.Host,
		DBName:    database,
		ParseTime: true,
		Loc:       time.UTC,
	}
	dsn := config.FormatDSN()
	if u.RawQuery != "" {
		separator := "?"
		if strings.Contains(dsn, "?") {
			separator = "&"
		}
		dsn += separator + u.RawQuery
	}
	config, err = mysql.ParseDSN(dsn)
	if err != nil {
		return "", err
	}
	config.ParseTime = true
	return config.FormatDSN(), nil
}

type sqliteDriver struct {
	sqlite.Driver
}

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
		return nil, errutil.WrapErr(err, "error enabling foreign_keys")
	}
	return conn, nil
}
