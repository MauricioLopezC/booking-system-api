package slots

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// iv builds an interval on 2030-01-08 from "HH:MM" strings.
func iv(start, end string) Interval {
	return Interval{Start: hm(start), End: hm(end)}
}

func hm(s string) time.Time {
	t, err := time.Parse("15:04", s)
	if err != nil {
		panic(err)
	}
	return time.Date(2030, 1, 8, t.Hour(), t.Minute(), 0, 0, time.UTC)
}

func TestMergeIntervals(t *testing.T) {
	tests := []struct {
		name string
		in   []Interval
		want []Interval
	}{
		{"vacío", nil, nil},
		{"desordenados y disjuntos", []Interval{iv("14:00", "16:00"), iv("08:00", "10:00")}, []Interval{iv("08:00", "10:00"), iv("14:00", "16:00")}},
		{"solapados", []Interval{iv("08:00", "10:00"), iv("09:00", "11:00")}, []Interval{iv("08:00", "11:00")}},
		{"adyacentes se unen", []Interval{iv("08:00", "09:00"), iv("09:00", "10:00")}, []Interval{iv("08:00", "10:00")}},
		{"contenido", []Interval{iv("08:00", "12:00"), iv("09:00", "10:00")}, []Interval{iv("08:00", "12:00")}},
		{"vacíos e invertidos se descartan", []Interval{iv("08:00", "08:00"), iv("10:00", "09:00")}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, mergeIntervals(tc.in))
		})
	}
}

func TestSubtractIntervals(t *testing.T) {
	tests := []struct {
		name         string
		base, remove []Interval
		want         []Interval
	}{
		{"sin nada que restar", []Interval{iv("08:00", "10:00")}, nil, []Interval{iv("08:00", "10:00")}},
		{"hueco en el medio", []Interval{iv("08:00", "10:00")}, []Interval{iv("08:30", "09:00")}, []Interval{iv("08:00", "08:30"), iv("09:00", "10:00")}},
		{"recorta el inicio", []Interval{iv("08:00", "10:00")}, []Interval{iv("07:00", "08:30")}, []Interval{iv("08:30", "10:00")}},
		{"recorta el final", []Interval{iv("08:00", "10:00")}, []Interval{iv("09:30", "11:00")}, []Interval{iv("08:00", "09:30")}},
		{"cubre todo", []Interval{iv("08:00", "10:00")}, []Interval{iv("07:00", "11:00")}, nil},
		{"adyacente no recorta", []Interval{iv("08:00", "10:00")}, []Interval{iv("10:00", "11:00")}, []Interval{iv("08:00", "10:00")}},
		{
			"una resta abarca dos bases",
			[]Interval{iv("08:00", "10:00"), iv("14:00", "16:00")},
			[]Interval{iv("09:00", "15:00")},
			[]Interval{iv("08:00", "09:00"), iv("15:00", "16:00")},
		},
		{
			"varias restas en varias bases",
			[]Interval{iv("08:00", "12:00"), iv("14:00", "18:00")},
			[]Interval{iv("08:30", "09:00"), iv("10:00", "10:15"), iv("17:00", "19:00")},
			[]Interval{iv("08:00", "08:30"), iv("09:00", "10:00"), iv("10:15", "12:00"), iv("14:00", "17:00")},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, subtractIntervals(tc.base, tc.remove))
		})
	}
}

func TestSlotStarts(t *testing.T) {
	anchor := hm("00:00")
	past := time.Time{}

	tests := []struct {
		name     string
		free     []Interval
		duration time.Duration
		step     time.Duration
		now      time.Time
		want     []string
	}{
		{
			name:     "45 min con paso 15 aprovecha el final del bloque",
			free:     []Interval{iv("08:00", "10:00")},
			duration: 45 * time.Minute, step: 15 * time.Minute, now: past,
			want: []string{"08:00", "08:15", "08:30", "08:45", "09:00", "09:15"},
		},
		{
			name:     "hueco que empieza fuera de grilla: ofrece su inicio y vuelve a la grilla",
			free:     []Interval{iv("12:20", "13:30")},
			duration: 30 * time.Minute, step: 15 * time.Minute, now: past,
			want: []string{"12:20", "12:30", "12:45", "13:00"},
		},
		{
			name:     "hueco más corto que el servicio no ofrece nada",
			free:     []Interval{iv("08:00", "08:20")},
			duration: 30 * time.Minute, step: 15 * time.Minute, now: past,
			want: nil,
		},
		{
			name:     "descarta los inicios anteriores a ahora",
			free:     []Interval{iv("08:00", "10:00")},
			duration: 30 * time.Minute, step: 30 * time.Minute, now: hm("08:40"),
			want: []string{"09:00", "09:30"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got []string
			for _, s := range slotStarts(tc.free, tc.duration, tc.step, anchor, tc.now) {
				got = append(got, s.Format("15:04"))
			}
			assert.Equal(t, tc.want, got)
		})
	}
}
