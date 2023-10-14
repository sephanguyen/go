package queries

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/calendar/application/queries/payloads"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	mock_repositories "github.com/manabie-com/backend/mock/calendar/repositories"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDateInfoQueryHandler_FetchDateInfoByDateRangeAndLocationID(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	mockDateInfoRepo := &mock_repositories.MockDateInfoRepo{}

	now := time.Now()
	request := &payloads.FetchDateInfoByDateRangeRequest{
		StartDate:  now.Add(-48 * time.Hour),
		EndDate:    now.Add(48 * time.Hour),
		LocationID: "location-id1",
		Timezone:   "sample-timezone",
	}

	testCases := []struct {
		name         string
		req          *payloads.FetchDateInfoByDateRangeRequest
		expectedResp *payloads.FetchDateInfoByDateRangeResponse
		setup        func(context.Context)
		hasError     bool
	}{
		{
			name: "success",
			req:  request,
			expectedResp: &payloads.FetchDateInfoByDateRangeResponse{
				DateInfos: []*dto.DateInfo{
					{
						Date:                now,
						LocationID:          "location-id1",
						DateTypeID:          "regular",
						DateTypeDisplayName: "REGULAR",
						OpeningTime:         "09:00",
						Status:              "draft",
						TimeZone:            "sample-timezone",
					},
					{
						Date:                now.Add(24 * time.Hour),
						LocationID:          "location-id1",
						DateTypeID:          "regular",
						DateTypeDisplayName: "REGULAR",
						OpeningTime:         "09:00",
						Status:              "draft",
						TimeZone:            "sample-timezone",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDateInfoRepo.On("GetDateInfoDetailedByDateRangeAndLocationID", mock.Anything, mock.Anything, request.StartDate, request.EndDate, request.LocationID, request.Timezone).Once().
					Return([]*dto.DateInfo{
						{
							Date:                now,
							LocationID:          "location-id1",
							DateTypeID:          "regular",
							DateTypeDisplayName: "REGULAR",
							OpeningTime:         "09:00",
							Status:              "draft",
							TimeZone:            "sample-timezone",
						},
						{
							Date:                now.Add(24 * time.Hour),
							LocationID:          "location-id1",
							DateTypeID:          "regular",
							DateTypeDisplayName: "REGULAR",
							OpeningTime:         "09:00",
							Status:              "draft",
							TimeZone:            "sample-timezone",
						},
					}, nil)
			},
			hasError: false,
		},
		{
			name: "failed to get date info detailed",
			req:  request,
			setup: func(ctx context.Context) {
				mockDateInfoRepo.On("GetDateInfoDetailedByDateRangeAndLocationID", mock.Anything, mockDB.DB, request.StartDate, request.EndDate, request.LocationID, request.Timezone).Once().
					Return(nil, fmt.Errorf("error"))
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			dateInfoQueryHandler := &DateInfoQueryHandler{
				DB:           mockDB.DB,
				DateInfoRepo: mockDateInfoRepo,
			}

			resp, err := dateInfoQueryHandler.FetchDateInfoByDateRangeAndLocationID(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, resp)
				require.Equal(t, tc.expectedResp, resp)
			}
		})
	}
}
