package watcher

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
)

func TestConfigReloadDoesNotMarkNewerDiskWriteAsApplied(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("port: 8317\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var ports []int
	w, err := NewWatcher(path, dir, func(cfg *config.Config) {
		ports = append(ports, cfg.Port)
		if len(ports) == 1 {
			if errWrite := os.WriteFile(path, []byte("port: 8318\n"), 0o600); errWrite != nil {
				t.Fatal(errWrite)
			}
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Stop() }()
	w.reloadConfigIfChanged()
	w.reloadConfigIfChanged()
	if !reflect.DeepEqual(ports, []int{8317, 8318}) {
		t.Fatalf("applied ports = %v, want both snapshots", ports)
	}
}
