package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sprisa/west/util/ipconv"
	entmigrate "github.com/sprisa/west/westport/db/ent/migrate"
	"github.com/sprisa/west/westport/db/helpers"
)

func TestDatastoreType(t *testing.T) {
	tests := map[string]string{
		"sqlite":                           DatastoreSQLite,
		"sqlite:///tmp/west.db":            DatastoreSQLite,
		"postgres://west@localhost/west":   DatastorePostgres,
		"postgresql://west@localhost/west": DatastorePostgres,
		"mysql://west@localhost:3306/west": DatastoreMySQL,
	}
	for input, want := range tests {
		got, err := DatastoreType(input)
		if err != nil {
			t.Fatalf("DatastoreType(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("DatastoreType(%q) = %q, want %q", input, got, want)
		}
	}
	if _, err := DatastoreType("cockroach://localhost/west"); err == nil {
		t.Fatal("expected unsupported datastore to fail")
	}
}

func TestMySQLDSN(t *testing.T) {
	dsn, err := mysqlDSN("mysql://user:p%40ss@db.example:3306/west?tls=true")
	if err != nil {
		t.Fatal(err)
	}
	if dsn == "" {
		t.Fatal("mysql DSN is empty")
	}
	if _, err := mysqlDSN("mysql://db.example:3306"); err == nil {
		t.Fatal("expected missing database to fail")
	}
}

func TestDatastoreSchema(t *testing.T) {
	tests := map[string]string{
		"sqlite": "sqlite://" + filepath.Join(t.TempDir(), "west.db"),
	}
	if dsn := os.Getenv("WEST_TEST_POSTGRES_DSN"); dsn != "" {
		tests["postgres"] = dsn
	}
	if dsn := os.Getenv("WEST_TEST_MYSQL_DSN"); dsn != "" {
		tests["mysql"] = dsn
	}

	for name, dataSource := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			helpers.SetEncryptionPassword([]byte("test-password"))
			client, err := OpenDB(ctx, dataSource)
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			if err := client.Schema.Create(ctx, entmigrate.WithGlobalUniqueID(true)); err != nil {
				t.Fatal(err)
			}
			if _, err := client.Lighthouse.Delete().Exec(ctx); err != nil {
				t.Fatal(err)
			}
			if _, err := client.Device.Delete().Exec(ctx); err != nil {
				t.Fatal(err)
			}
			if _, err := client.Host.Delete().Exec(ctx); err != nil {
				t.Fatal(err)
			}
			if _, err := client.Settings.Delete().Exec(ctx); err != nil {
				t.Fatal(err)
			}
			cidr, err := helpers.NewIpCidr("10.10.10.1/24")
			if err != nil {
				t.Fatal(err)
			}
			if err := client.Settings.Create().
				SetCaCrt([]byte("ca")).
				SetCaKey([]byte("key")).
				SetCidr(cidr).
				Exec(ctx); err != nil {
				t.Fatal(err)
			}
			ip, err := ipconv.FromIPAddr(cidr.Addr())
			if err != nil {
				t.Fatal(err)
			}
			host, err := client.Host.Create().SetIP(ip).Save(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if err := client.Lighthouse.Create().
				SetIP(ip).
				SetEndpoint("127.0.0.1:4242").
				SetAPIEndpoint("127.0.0.1:80").
				SetCertificate([]byte("cert")).
				SetKey([]byte("key")).
				SetHostID(host.ID).
				Exec(ctx); err != nil {
				t.Fatal(err)
			}
			settings, err := client.Settings.Query().Only(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if settings.Cidr.String() != cidr.String() {
				t.Fatalf("cidr = %q, want %q", settings.Cidr, cidr)
			}
		})
	}
}
