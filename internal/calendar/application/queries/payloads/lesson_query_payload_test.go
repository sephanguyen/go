package payloads

import (
	"testing"
	"time"

	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/stretchr/testify/require"
)

func TestGetLessonDetailRequest_Validate(t *testing.T) {
	t.Parallel()
	lessonID := "lesson-id1"

	testCases := []struct {
		name     string
		req      *GetLessonDetailRequest
		hasError bool
	}{
		{
			name: "request full fields",
			req: &GetLessonDetailRequest{
				LessonID: lessonID,
			},
			hasError: false,
		},
		{
			name:     "request lesson id empty",
			req:      &GetLessonDetailRequest{},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()

			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetLessonIDsForBulkStatusUpdateRequest_Validate(t *testing.T) {
	t.Parallel()
	locationID := "location-id1"
	action := lesson_domain.LessonBulkActionPublish
	now := time.Now()

	testCases := []struct {
		name     string
		req      *GetLessonIDsForBulkStatusUpdateRequest
		hasError bool
	}{
		{
			name: "request full fields",
			req: &GetLessonIDsForBulkStatusUpdateRequest{
				LocationID: locationID,
				Action:     action,
				StartDate:  now,
				EndDate:    now.Add(2 * 24 * time.Hour),
				StartTime:  now,
				EndTime:    now.Add(2 * time.Hour),
			},
			hasError: false,
		},
		{
			name: "request start and end times empty",
			req: &GetLessonIDsForBulkStatusUpdateRequest{
				LocationID: locationID,
				Action:     action,
				StartDate:  now,
				EndDate:    now.Add(2 * 24 * time.Hour),
			},
			hasError: false,
		},
		{
			name: "request end date before start date",
			req: &GetLessonIDsForBulkStatusUpdateRequest{
				Action:    action,
				StartDate: now.Add(2 * 24 * time.Hour),
				EndDate:   now,
				StartTime: now,
				EndTime:   now.Add(2 * time.Hour),
			},
			hasError: true,
		},
		{
			name: "request location id empty",
			req: &GetLessonIDsForBulkStatusUpdateRequest{
				Action:    action,
				StartDate: now,
				EndDate:   now.Add(2 * 24 * time.Hour),
				StartTime: now,
				EndTime:   now.Add(2 * time.Hour),
			},
			hasError: true,
		},
		{
			name: "request start date empty",
			req: &GetLessonIDsForBulkStatusUpdateRequest{
				LocationID: locationID,
				Action:     action,
				EndDate:    now.Add(2 * 24 * time.Hour),
				StartTime:  now,
				EndTime:    now.Add(2 * time.Hour),
			},
			hasError: true,
		},
		{
			name: "request end date empty",
			req: &GetLessonIDsForBulkStatusUpdateRequest{
				LocationID: locationID,
				Action:     action,
				StartDate:  now,
				StartTime:  now,
				EndTime:    now.Add(2 * time.Hour),
			},
			hasError: true,
		},
		{
			name: "request start time empty",
			req: &GetLessonIDsForBulkStatusUpdateRequest{
				LocationID: locationID,
				Action:     action,
				StartDate:  now,
				EndDate:    now.Add(2 * 24 * time.Hour),
				EndTime:    now.Add(2 * time.Hour),
			},
			hasError: true,
		},
		{
			name: "request end time empty",
			req: &GetLessonIDsForBulkStatusUpdateRequest{
				LocationID: locationID,
				Action:     action,
				StartDate:  now,
				EndDate:    now.Add(2 * 24 * time.Hour),
				StartTime:  now,
			},
			hasError: true,
		},
		{
			name: "request end time before start time",
			req: &GetLessonIDsForBulkStatusUpdateRequest{
				Action:    action,
				StartDate: now,
				EndDate:   now.Add(2 * 24 * time.Hour),
				StartTime: now.Add(2 * time.Hour),
				EndTime:   now,
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()

			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
