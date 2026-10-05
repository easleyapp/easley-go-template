package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const defaultEnvFile = ".env"

type Config struct {
	Port         int
	DataDir      string
	SourceURL    string
	SourceAPIKey string
}

func Load() (Config, error) {
	loadDotEnv(defaultEnvFile)

	port := 8080
	if v := os.Getenv("PORT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 65535 {
			return Config{}, fmt.Errorf("invalid PORT %q", v)
		}
		port = n
	}
	return Config{
		Port:         port,
		DataDir:      getenv("DATA_DIR", "./data"),
		SourceURL:    os.Getenv("SOURCE_API_URL"),
		SourceAPIKey: os.Getenv("SOURCE_API_KEY"),
	}, nil
}

// loadDotEnv fills variables from a local .env for development. A variable that
// is already present in the process environment always wins, so production
// configuration is never overridden by a stray file.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, strings.Trim(strings.TrimSpace(value), `"'`))
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
