package scheduleexceptions

import (
	"testing"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/stretchr/testify/require"
)

func TestHasScheduleExceptionOverlap(t *testing.T) {
	// Helper to parse time strings into time.Time objects with zero date
	mustParseTimeOnly := func(value string) time.Time {
		// Use "15:04" layout for HH:MM format
		parsedTime, err := time.ParseInLocation("15:04", value, time.Local)
		if err != nil {
			panic(err)
		}
		return parsedTime
	}

	// Helper to create a ScheduleOverride with time pointers
	createScheduleOverride := func(startTime, endTime string, allDay bool) model.ScheduleOverride {
		so := model.ScheduleOverride{AllDay: allDay}
		if !allDay {
			st := mustParseTimeOnly(startTime)
			et := mustParseTimeOnly(endTime)
			so.StartTime = &st
			so.EndTime = &et
		}
		return so
	}

	testCases := []struct {
		name                 string
		newStart             string
		newEnd               string
		existing             []model.ScheduleOverride
		shouldOverlap        bool
		expectedOverlapIndex int // Index of the expected overlapping schedule in 'existing', -1 if no overlap
	}{
		{
			name:                 "When new schedule overlaps with an existing AllDay exception",
			newStart:             "09:00",
			newEnd:               "10:00",
			existing:             []model.ScheduleOverride{createScheduleOverride("", "", true)},
			shouldOverlap:        true,
			expectedOverlapIndex: 0,
		},
		{
			name:                 "When no existing exceptions",
			newStart:             "09:00",
			newEnd:               "10:00",
			existing:             []model.ScheduleOverride{},
			shouldOverlap:        false,
			expectedOverlapIndex: -1,
		},
		{
			name:                 "When new schedule is contained within an existing partial exception",
			newStart:             "10:00",
			newEnd:               "11:00",
			existing:             []model.ScheduleOverride{createScheduleOverride("09:00", "12:00", false)},
			shouldOverlap:        true,
			expectedOverlapIndex: 0,
		},
		{
			name:                 "When new schedule contains an existing partial exception",
			newStart:             "08:00",
			newEnd:               "13:00",
			existing:             []model.ScheduleOverride{createScheduleOverride("10:00", "11:00", false)},
			shouldOverlap:        true,
			expectedOverlapIndex: 0,
		},
		{
			name:                 "When new schedule partially overlaps at the beginning of an existing partial exception",
			newStart:             "08:00",
			newEnd:               "10:00",
			existing:             []model.ScheduleOverride{createScheduleOverride("09:00", "12:00", false)},
			shouldOverlap:        true,
			expectedOverlapIndex: 0,
		},
		{
			name:                 "When new schedule partially overlaps at the end of an existing partial exception",
			newStart:             "11:00",
			newEnd:               "13:00",
			existing:             []model.ScheduleOverride{createScheduleOverride("09:00", "12:00", false)},
			shouldOverlap:        true,
			expectedOverlapIndex: 0,
		},
		{
			name:                 "When new schedule is identical to an existing partial exception",
			newStart:             "09:00",
			newEnd:               "12:00",
			existing:             []model.ScheduleOverride{createScheduleOverride("09:00", "12:00", false)},
			shouldOverlap:        true,
			expectedOverlapIndex: 0,
		},
		{
			name:                 "When new schedule starts at the same time as an existing one but ends earlier",
			newStart:             "09:00",
			newEnd:               "11:00",
			existing:             []model.ScheduleOverride{createScheduleOverride("09:00", "12:00", false)},
			shouldOverlap:        true,
			expectedOverlapIndex: 0,
		},
		{
			name:                 "When new schedule ends at the same time as an existing one but starts later",
			newStart:             "10:00",
			newEnd:               "12:00",
			existing:             []model.ScheduleOverride{createScheduleOverride("09:00", "12:00", false)},
			shouldOverlap:        true,
			expectedOverlapIndex: 0,
		},
		{
			name:                 "When new schedule is completely before all existing partial exceptions",
			newStart:             "07:00",
			newEnd:               "08:00",
			existing:             []model.ScheduleOverride{createScheduleOverride("09:00", "12:00", false)},
			shouldOverlap:        false,
			expectedOverlapIndex: -1,
		},
		{
			name:                 "When new schedule is completely after all existing partial exceptions",
			newStart:             "13:00",
			newEnd:               "14:00",
			existing:             []model.ScheduleOverride{createScheduleOverride("09:00", "12:00", false)},
			shouldOverlap:        false,
			expectedOverlapIndex: -1,
		},
		{
			name:                 "When new schedule ends exactly when an existing one begins",
			newStart:             "08:00",
			newEnd:               "09:00",
			existing:             []model.ScheduleOverride{createScheduleOverride("09:00", "12:00", false)},
			shouldOverlap:        false,
			expectedOverlapIndex: -1,
		},
		{
			name:                 "When new schedule starts exactly when an existing one ends",
			newStart:             "12:00",
			newEnd:               "13:00",
			existing:             []model.ScheduleOverride{createScheduleOverride("09:00", "12:00", false)},
			shouldOverlap:        false,
			expectedOverlapIndex: -1,
		},
		{
			name:                 "When multiple existing partial exceptions, new overlaps with the first one",
			newStart:             "09:30",
			newEnd:               "10:30",
			existing:             []model.ScheduleOverride{createScheduleOverride("09:00", "10:00", false), createScheduleOverride("11:00", "12:00", false)},
			shouldOverlap:        true,
			expectedOverlapIndex: 0,
		},
		{
			name:                 "When multiple existing partial exceptions, new overlaps with the second one",
			newStart:             "11:30",
			newEnd:               "12:30",
			existing:             []model.ScheduleOverride{createScheduleOverride("09:00", "10:00", false), createScheduleOverride("11:00", "12:00", false)},
			shouldOverlap:        true,
			expectedOverlapIndex: 1,
		},
		{
			name:                 "When multiple existing partial exceptions, new does not overlap with any",
			newStart:             "10:00",
			newEnd:               "11:00",
			existing:             []model.ScheduleOverride{createScheduleOverride("09:00", "10:00", false), createScheduleOverride("11:00", "12:00", false)},
			shouldOverlap:        false,
			expectedOverlapIndex: -1,
		},
		{
			name:                 "When existing contains a non-AllDay with nil StartTime/EndTime (should be skipped)",
			newStart:             "09:00",
			newEnd:               "10:00",
			existing:             []model.ScheduleOverride{{AllDay: false, StartTime: nil, EndTime: nil}},
			shouldOverlap:        false,
			expectedOverlapIndex: -1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dtoStart := mustParseTimeOnly(tc.newStart)
			dtoEnd := mustParseTimeOnly(tc.newEnd)

			hasOverlap, match := hasScheduleExceptionOverlap(dtoStart, dtoEnd, tc.existing)

			require.Equal(t, tc.shouldOverlap, hasOverlap)
			if tc.expectedOverlapIndex == -1 {
				require.Nil(t, match)
			} else {
				require.NotNil(t, match)
				require.Equal(t, &tc.existing[tc.expectedOverlapIndex], match)
			}
		})
	}
}
