package slots

import (
	"slices"
	"time"
)

// Interval is a half-open time range [Start, End).
type Interval struct {
	Start, End time.Time
}

// mergeIntervals sorts the intervals and joins the ones that overlap or touch.
// Empty intervals (End <= Start) are dropped. The input is not modified.
func mergeIntervals(in []Interval) []Interval {
	sorted := make([]Interval, 0, len(in))
	for _, iv := range in {
		if iv.End.After(iv.Start) {
			sorted = append(sorted, iv)
		}
	}
	slices.SortFunc(sorted, func(a, b Interval) int { return a.Start.Compare(b.Start) })

	var out []Interval
	for _, iv := range sorted {
		last := len(out) - 1
		if last >= 0 && !iv.Start.After(out[last].End) {
			if iv.End.After(out[last].End) {
				out[last].End = iv.End
			}
			continue
		}
		out = append(out, iv)
	}
	return out
}

// subtractIntervals returns the parts of base not covered by remove.
// Both inputs must be merged (sorted, non-overlapping), as returned by mergeIntervals.
func subtractIntervals(base, remove []Interval) []Interval {
	var out []Interval
	j := 0
	for _, b := range base {
		cur := b.Start
		// skip removals that end before this base interval starts
		for j < len(remove) && !remove[j].End.After(b.Start) {
			j++
		}
		for k := j; k < len(remove) && remove[k].Start.Before(b.End); k++ {
			if remove[k].Start.After(cur) {
				out = append(out, Interval{Start: cur, End: remove[k].Start})
			}
			if remove[k].End.After(cur) {
				cur = remove[k].End
			}
		}
		if b.End.After(cur) {
			out = append(out, Interval{Start: cur, End: b.End})
		}
	}
	return out
}

// slotStarts lists every start time t inside the free intervals such that
// [t, t+duration) fits and t is not before now. Each free interval offers its own
// start (so gaps left by bookings are usable) and then the points of a grid of
// `step` anchored at `anchor` (so customers see tidy times like 10:00, 10:15...).
func slotStarts(free []Interval, duration, step time.Duration, anchor, now time.Time) []time.Time {
	var out []time.Time
	for _, f := range free {
		for t := f.Start; !t.Add(duration).After(f.End); t = nextGridPoint(t, anchor, step) {
			if !t.Before(now) {
				out = append(out, t)
			}
		}
	}
	return out
}

// nextGridPoint returns the first anchor + k*step strictly after t (t >= anchor).
func nextGridPoint(t, anchor time.Time, step time.Duration) time.Time {
	k := t.Sub(anchor)/step + 1
	return anchor.Add(k * step)
}
