package bookings

import (
	"os"
	"testing"
	"time"
)

// Same as cmd/api/main.go: postgres time columns are read in time.Local.
func TestMain(m *testing.M) {
	time.Local = time.UTC
	os.Exit(m.Run())
}
