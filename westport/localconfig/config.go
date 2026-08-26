package localconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sprisa/west/westport/db/helpers"
)

var FilePath = "west-port.config.json"

type Config struct {
	Datastore    string `json:"datastore"`
	LighthouseIP string `json:"lighthouse_ip"`
}

func Exists() (bool, error) {
	_, err := os.Stat(FilePath)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func Load() (*Config, error) {
	ciphertext, err := os.ReadFile(FilePath)
	if err != nil {
		return nil, err
	}
	plaintext, err := helpers.Decrypt(strings.TrimSpace(string(ciphertext)))
	if err != nil {
		return nil, fmt.Errorf("decrypt west-port config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(plaintext, &cfg); err != nil {
		return nil, fmt.Errorf("decode west-port config: %w", err)
	}
	if cfg.Datastore == "" || cfg.LighthouseIP == "" {
		return nil, errors.New("west-port config is missing datastore or lighthouse_ip")
	}
	return &cfg, nil
}

func Save(cfg Config) error {
	plaintext, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	ciphertext, err := helpers.Encrypt(plaintext)
	if err != nil {
		return err
	}
	dir := filepath.Dir(FilePath)
	tmp, err := os.CreateTemp(dir, ".west-port.config.*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.WriteString(ciphertext); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, 0600); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, FilePath)
}
