// Package config config for env variables
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"
	_ "time/tzdata" // BUSINESS_TIMEZONE works even on images without system tzdata

	"github.com/joho/godotenv"
)

var requiredVars = []string{
	"DB_HOST",
	"DB_PORT",
	"DB_USER",
	"DB_PASSWORD",
	"DB_NAME",
	"SECRET",
}

// Load reads the .env file (optional: variables may come from the real
// environment, e.g. in a container) and validates all required environment variables.
// Must be called once at startup before anything else.
func Load() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatalf("Error loading .env file: %v", err)
	}

	var missing []string
	for _, key := range requiredVars {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		log.Fatalf("Missing required environment variables: %v", missing)
	}
}

// Config returns the value of an environment variable.
func Config(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("config: %q not set (was config.Load() called?)", key))
	}
	return v
}

// Get returns the value of an optional environment variable, or fallback when it's not set.
func Get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// BusinessLocation returns the business timezone from BUSINESS_TIMEZONE (IANA name, default UTC).
// Schedules and exceptions are wall-clock times in this zone.
func BusinessLocation() (*time.Location, error) {
	return time.LoadLocation(Get("BUSINESS_TIMEZONE", "UTC"))
}
