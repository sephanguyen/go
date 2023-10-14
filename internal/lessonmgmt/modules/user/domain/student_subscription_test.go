package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStudentSubscriptions_IsValid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	tcs := []struct {
		name    string
		input   StudentSubscriptions
		isValid bool
	}{
		{
			name: "full fields",
			input: StudentSubscriptions{
				{
					SubscriptionID: "id-1",
					StudentID:      "student-id-1",
					CourseID:       "course-id-1",
					StartAt:        now,
					EndAt:          now,
					CreatedAt:      now,
					UpdatedAt:      now,
					LocationIDs:    []string{"location-id-1", "location-id-3", "location-id-5"},
				},
				{
					SubscriptionID: "id-2",
					StudentID:      "student-id-2",
					CourseID:       "course-id-3",
					StartAt:        now,
					EndAt:          now,
					CreatedAt:      now,
					UpdatedAt:      now.Add(2 * time.Minute),
					LocationIDs:    []string{"location-id-1", "location-id-3"},
				},
			},
			isValid: true,
		},
		{
			name: "missing location ids",
			input: StudentSubscriptions{
				{
					SubscriptionID: "id-1",
					StudentID:      "student-id-1",
					CourseID:       "course-id-1",
					StartAt:        now,
					EndAt:          now,
					CreatedAt:      now,
					UpdatedAt:      now,
				},
				{
					SubscriptionID: "id-2",
					StudentID:      "student-id-2",
					CourseID:       "course-id-3",
					StartAt:        now,
					EndAt:          now,
					CreatedAt:      now,
					UpdatedAt:      now.Add(2 * time.Minute),
				},
			},
			isValid: true,
		},
		{
			name: "missing studentID",
			input: StudentSubscriptions{
				{
					SubscriptionID: "id-1",
					CourseID:       "course-id-1",
					StartAt:        now,
					EndAt:          now,
					CreatedAt:      now,
					UpdatedAt:      now,
					LocationIDs:    []string{"location-id-1", "location-id-3", "location-id-5"},
				},
				{
					SubscriptionID: "id-2",
					StudentID:      "student-id-2",
					CourseID:       "course-id-3",
					StartAt:        now,
					EndAt:          now,
					CreatedAt:      now,
					UpdatedAt:      now.Add(2 * time.Minute),
					LocationIDs:    []string{"location-id-1", "location-id-3"},
				},
			},
			isValid: false,
		},
		{
			name: "missing courseID",
			input: StudentSubscriptions{
				{
					SubscriptionID: "id-1",
					StudentID:      "student-id-1",
					CourseID:       "course-id-1",
					StartAt:        now,
					EndAt:          now,
					CreatedAt:      now,
					UpdatedAt:      now,
					LocationIDs:    []string{"location-id-1", "location-id-3", "location-id-5"},
				},
				{
					SubscriptionID: "id-2",
					StudentID:      "student-id-2",
					StartAt:        now,
					EndAt:          now,
					CreatedAt:      now,
					UpdatedAt:      now.Add(2 * time.Minute),
					LocationIDs:    []string{"location-id-1", "location-id-3"},
				},
			},
			isValid: false,
		},
		{
			name: "end time before start time",
			input: StudentSubscriptions{
				{
					SubscriptionID: "id-1",
					StudentID:      "student-id-1",
					CourseID:       "course-id-1",
					StartAt:        now,
					EndAt:          now.Add(-1 * time.Minute),
					CreatedAt:      now,
					UpdatedAt:      now,
					LocationIDs:    []string{"location-id-1", "location-id-3", "location-id-5"},
				},
				{
					SubscriptionID: "id-2",
					StudentID:      "student-id-2",
					CourseID:       "course-id-3",
					StartAt:        now,
					EndAt:          now,
					CreatedAt:      now,
					UpdatedAt:      now.Add(2 * time.Minute),
					LocationIDs:    []string{"location-id-1", "location-id-3"},
				},
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.IsValid()
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
