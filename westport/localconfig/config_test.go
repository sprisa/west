package localconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sprisa/west/westport/db/helpers"
)

func TestSaveLoadAndRejectDuplicate(t *testing.T) {
	originalPath := FilePath
	FilePath = filepath.Join(t.TempDir(), "west-port.config.json")
	t.Cleanup(func() { FilePath = originalPath })
	helpers.SetEncryptionPassword([]byte("correct horse battery staple"))

	want := Config{Datastore: "postgres://west@example/west", LighthouseIP: "10.10.10.1"}
	if err := Save(want); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(FilePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(contents), want.Datastore) || strings.HasPrefix(string(contents), "{") {
		t.Fatal("config file contains plaintext JSON")
	}
	info, err := os.Stat(FilePath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("config permissions = %o", info.Mode().Perm())
	}

	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if *got != want {
		t.Fatalf("Load() = %#v, want %#v", *got, want)
	}
	if err := Save(want); err != nil {
		t.Fatalf("atomic overwrite should succeed: %v", err)
	}
}

func TestLoadRejectsWrongPassword(t *testing.T) {
	originalPath := FilePath
	FilePath = filepath.Join(t.TempDir(), "west-port.config.json")
	t.Cleanup(func() { FilePath = originalPath })
	helpers.SetEncryptionPassword([]byte("correct"))
	if err := Save(Config{Datastore: "sqlite", LighthouseIP: "10.10.10.1"}); err != nil {
		t.Fatal(err)
	}
	helpers.SetEncryptionPassword([]byte("wrong"))
	if _, err := Load(); err == nil {
		t.Fatal("expected wrong password to fail")
	}
}
