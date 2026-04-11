package schedule

import (
	"testing"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/stretchr/testify/require"
)

func TestHasScheduleOverlap(t *testing.T) {
	layout := "15:04"

	existingSchedules := []model.Schedule{
		{
			StartTime: mustParse(layout, "09:00"),
			EndTime:   mustParse(layout, "12:00"),
		},
		{
			StartTime: mustParse(layout, "14:00"),
			EndTime:   mustParse(layout, "16:00"),
		},
	}

	testCases := []struct {
		name                string
		newStart            string
		newEnd              string
		shouldOverlap       bool
		expectedOverlapWith *model.Schedule
	}{
		{
			name:                "When new schedule is contained within an existing schedule",
			newStart:            "10:00",
			newEnd:              "11:00",
			shouldOverlap:       true,
			expectedOverlapWith: &existingSchedules[0],
		},
		{
			name:                "When new schedule contains an existing schedule",
			newStart:            "08:00",
			newEnd:              "13:00",
			shouldOverlap:       true,
			expectedOverlapWith: &existingSchedules[0],
		},
		{
			name:                "When new schedule partially overlaps at the beginning of an existing schedule",
			newStart:            "08:00",
			newEnd:              "10:00",
			shouldOverlap:       true,
			expectedOverlapWith: &existingSchedules[0],
		},
		{
			name:                "When new schedule partially overlaps at the end of an existing schedule",
			newStart:            "11:00",
			newEnd:              "13:00",
			shouldOverlap:       true,
			expectedOverlapWith: &existingSchedules[0],
		},
		{
			name:                "When new schedule is identical to an existing schedule",
			newStart:            "09:00",
			newEnd:              "12:00",
			shouldOverlap:       true,
			expectedOverlapWith: &existingSchedules[0],
		},
		{
			name:                "When new schedule starts at the same time as an existing one but ends earlier",
			newStart:            "09:00",
			newEnd:              "11:00",
			shouldOverlap:       true,
			expectedOverlapWith: &existingSchedules[0],
		},
		{
			name:                "When new schedule ends at the same time as an existing one but starts later",
			newStart:            "10:00",
			newEnd:              "12:00",
			shouldOverlap:       true,
			expectedOverlapWith: &existingSchedules[0],
		},
		{
			name:                "When new schedule is completely before all existing schedules",
			newStart:            "07:00",
			newEnd:              "08:00",
			shouldOverlap:       false,
			expectedOverlapWith: nil,
		},
		{
			name:                "When new schedule is completely after all existing schedules",
			newStart:            "17:00",
			newEnd:              "18:00",
			shouldOverlap:       false,
			expectedOverlapWith: nil,
		},
		{
			name:                "When new schedule is between two existing schedules",
			newStart:            "12:30",
			newEnd:              "13:30",
			shouldOverlap:       false,
			expectedOverlapWith: nil,
		},
		{
			name:                "When new schedule ends exactly when an existing one begins",
			newStart:            "08:00",
			newEnd:              "09:00",
			shouldOverlap:       false,
			expectedOverlapWith: nil,
		},
		{
			name:                "When new schedule starts exactly when an existing one ends",
			newStart:            "12:00",
			newEnd:              "13:00",
			shouldOverlap:       false,
			expectedOverlapWith: nil,
		},
		{
			name:                "When new schedule overlaps with the second existing schedule",
			newStart:            "15:00",
			newEnd:              "17:00",
			shouldOverlap:       true,
			expectedOverlapWith: &existingSchedules[1],
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dtoStart := mustParse(layout, tc.newStart)
			dtoEnd := mustParse(layout, tc.newEnd)

			hasOverlap, match := hasScheduleOverlap(dtoStart, dtoEnd, existingSchedules)

			require.Equal(t, tc.shouldOverlap, hasOverlap)
			require.Equal(t, tc.expectedOverlapWith, match)
		})
	}
}

func mustParse(layout, value string) time.Time {
	t, err := time.ParseInLocation(layout, value, time.UTC)
	if err != nil {
		panic(err)
	}
	return t
}