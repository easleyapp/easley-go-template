package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsDotEnv(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("DATA_DIR=./from-file\nSOURCE_API_URL=https://api.example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	unsetEnv(t, "DATA_DIR")
	unsetEnv(t, "SOURCE_API_URL")
	unsetEnv(t, "PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DataDir != "./from-file" || cfg.SourceURL != "https://api.example.com" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestEnvWinsOverDotEnv(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("DATA_DIR=./from-file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	t.Setenv("DATA_DIR", "./from-env")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DataDir != "./from-env" {
		t.Fatalf("process environment should win: %+v", cfg)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	old, ok := os.LookupEnv(key)
	os.Unsetenv(key)
	t.Cleanup(func() {
		if ok {
			os.Setenv(key, old)
			return
		}
		os.Unsetenv(key)
	})
}
