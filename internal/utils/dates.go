package utils

import "time"

// Comprare two times ignoring the dates, compare only hour and minutes
// Return 1 if t1 > t2, -1 if t2 > t1, 0 if t1 = t2
func CompareTimeOnly(t1, t2 time.Time) int {
	//compare hour
	if h1, h2 := t1.Hour(), t2.Hour(); h1 != h2 {
		if h1 > h2 {
			return 1
		}
		return -1
	}
	// if hours are equal, compare minutes
	if m1, m2 := t1.Minute(), t2.Minute(); m1 != m2 {
		if m1 > m2 {
			return 1
		}
		return -1
	}
	//if all are equal
	return 0
}

// Return a time.Time type with a fixed date 0000/01/01 in order to compare times with
// other date fixed times using available methods
func TimeOfDate(t time.Time) time.Time {
	t = t.In(time.UTC)
	return time.Date(0, 1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
}
