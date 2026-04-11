package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTimeOfDay(t *testing.T) {
	ts := time.Date(2026, 4, 11, 14, 30, 45, 0, time.UTC)
	tod := NewTimeOfDay(ts)
	assert.Equal(t, 14, tod.Hour)
	assert.Equal(t, 30, tod.Minute)
	assert.Equal(t, 45, tod.Second)
}

func TestParseTimeOfDay(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		want    TimeOfDay
	}{
		{"09:00:00", false, TimeOfDay{9, 0, 0}},
		{"23:59:59", false, TimeOfDay{23, 59, 59}},
		{"00:00:00", false, TimeOfDay{0, 0, 0}},
		{"24:00:00", true, TimeOfDay{}},
		{"09:60:00", true, TimeOfDay{}},
		{"abc", true, TimeOfDay{}},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseTimeOfDay(tc.input)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestTimeOfDay_String(t *testing.T) {
	assert.Equal(t, "09:05:00", TimeOfDay{9, 5, 0}.String())
	assert.Equal(t, "23:59:59", TimeOfDay{23, 59, 59}.String())
}

func TestTimeOfDay_ToTimeOnDate(t *testing.T) {
	day := time.Date(2030, 1, 8, 0, 0, 0, 0, time.UTC)
	tod := TimeOfDay{Hour: 14, Minute: 30, Second: 0}
	got := tod.ToTimeOnDate(day)

	assert.Equal(t, 2030, got.Year())
	assert.Equal(t, time.January, got.Month())
	assert.Equal(t, 8, got.Day())
	assert.Equal(t, 14, got.Hour())
	assert.Equal(t, 30, got.Minute())
	assert.Equal(t, time.UTC, got.Location())
}

func TestTimeOfDay_Before(t *testing.T) {
	tests := []struct {
		a, b TimeOfDay
		want bool
	}{
		{TimeOfDay{8, 0, 0}, TimeOfDay{9, 0, 0}, true},
		{TimeOfDay{9, 0, 0}, TimeOfDay{8, 0, 0}, false},
		{TimeOfDay{8, 0, 0}, TimeOfDay{8, 0, 0}, false},
		{TimeOfDay{8, 0, 0}, TimeOfDay{8, 30, 0}, true},
		{TimeOfDay{8, 30, 0}, TimeOfDay{8, 0, 0}, false},
		{TimeOfDay{8, 0, 0}, TimeOfDay{8, 0, 1}, true},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, tc.a.Before(tc.b), "%v.Before(%v)", tc.a, tc.b)
	}
}

func TestTimeOfDay_After(t *testing.T) {
	assert.True(t, TimeOfDay{9, 0, 0}.After(TimeOfDay{8, 0, 0}))
	assert.False(t, TimeOfDay{8, 0, 0}.After(TimeOfDay{9, 0, 0}))
	assert.False(t, TimeOfDay{8, 0, 0}.After(TimeOfDay{8, 0, 0}))
}

func TestTimeOfDay_Equal(t *testing.T) {
	assert.True(t, TimeOfDay{9, 30, 0}.Equal(TimeOfDay{9, 30, 0}))
	assert.False(t, TimeOfDay{9, 30, 0}.Equal(TimeOfDay{9, 31, 0}))
}

func TestTimeOfDay_Between(t *testing.T) {
	start := TimeOfDay{8, 0, 0}
	end := TimeOfDay{10, 0, 0}

	tests := []struct {
		t    TimeOfDay
		want bool
		name string
	}{
		{TimeOfDay{8, 0, 0}, true, "equal to start (inclusive)"},
		{TimeOfDay{10, 0, 0}, true, "equal to end (inclusive)"},
		{TimeOfDay{9, 0, 0}, true, "in the middle"},
		{TimeOfDay{7, 59, 59}, false, "before start"},
		{TimeOfDay{10, 0, 1}, false, "after end"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.t.Between(start, end))
		})
	}
}
