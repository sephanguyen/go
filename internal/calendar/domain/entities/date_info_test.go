package entities

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_repositories "github.com/manabie-com/backend/mock/calendar/repositories"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDateInfo_GetDateInfoStatus(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		status         string
		expectedStatus constants.DateInfoStatus
		hasError       bool
	}{
		{
			name:           "valid date info status - none",
			status:         "none",
			expectedStatus: constants.None,
			hasError:       false,
		},
		{
			name:           "valid date info status - draft",
			status:         "draft",
			expectedStatus: constants.Draft,
			hasError:       false,
		},
		{
			name:           "valid date info status - published",
			status:         "published",
			expectedStatus: constants.Published,
			hasError:       false,
		},
		{
			name:           "valid date info status but with caps",
			status:         "PubliSHed",
			expectedStatus: constants.Published,
			hasError:       false,
		},
		{
			name:           "valid date info status but with space",
			status:         "published ",
			expectedStatus: "",
			hasError:       true,
		},
		{
			name:           "invalid date info status",
			status:         "hello",
			expectedStatus: "",
			hasError:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dateInfoStatus, err := GetDateInfoStatus(tc.status)

			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expectedStatus, dateInfoStatus)
		})
	}
}

func TestDateInfo_Validate(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	locationRepo := &mock_repositories.MockLocationRepo{}
	locationID := idutil.ULIDNow()

	testCases := []struct {
		name     string
		dateInfo *DateInfo
		setup    func(context.Context)
		hasError bool
	}{
		{
			name: "date info with empty date",
			dateInfo: &DateInfo{
				LocationID: locationID,
				Status:     constants.Draft,
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name: "date info with empty location id",
			dateInfo: &DateInfo{
				Date:   now,
				Status: constants.Draft,
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name: "date info closed date type and with opening time",
			dateInfo: &DateInfo{
				Date:        now,
				LocationID:  locationID,
				DateTypeID:  constants.ClosedDay,
				OpeningTime: "09:00",
				Status:      constants.Draft,
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name: "failed to fetch location ID in DB",
			dateInfo: &DateInfo{
				Date:       now,
				LocationID: locationID,
				Status:     constants.Draft,
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetLocationByID", mock.Anything, mock.Anything, locationID).Once().Return(nil, errors.New("error"))
			},
			hasError: true,
		},
		{
			name: "date info with complete required details",
			dateInfo: &DateInfo{
				Date:       now,
				LocationID: locationID,
				Status:     constants.Draft,
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetLocationByID", mock.Anything, mock.Anything, locationID).Once().Return(&dto.Location{}, nil)
			},
			hasError: false,
		},
		{
			name: "date info with all info",
			dateInfo: &DateInfo{
				Date:        now,
				LocationID:  locationID,
				DateTypeID:  constants.RegularDay,
				OpeningTime: "09:00",
				Status:      constants.Draft,
				TimeZone:    "sample-timezone",
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetLocationByID", mock.Anything, mock.Anything, locationID).Once().Return(&dto.Location{}, nil)
			},
			hasError: false,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()

		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			tc.dateInfo.LocationRepo = locationRepo

			err := tc.dateInfo.Validate(ctx)

			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDateInfo_Upsert(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	dateInfoRepo := &mock_repositories.MockDateInfoRepo{}
	locationRepo := &mock_repositories.MockLocationRepo{}
	locationID := idutil.ULIDNow()

	testCases := []struct {
		name     string
		dateInfo *DateInfo
		setup    func(context.Context)
		hasError bool
	}{
		{
			name: "upsert date info successfully",
			dateInfo: &DateInfo{
				Date:        now,
				LocationID:  locationID,
				DateTypeID:  constants.RegularDay,
				OpeningTime: "09:00",
				Status:      constants.Draft,
				TimeZone:    "sample-timezone",
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetLocationByID", mock.Anything, mock.Anything, locationID).Once().Return(&dto.Location{}, nil)

				dateInfoRepo.On("UpsertDateInfo", mock.Anything, mock.Anything, &dto.UpsertDateInfoParams{
					DateInfo: &dto.DateInfo{
						Date:        now,
						LocationID:  locationID,
						DateTypeID:  string(constants.RegularDay),
						OpeningTime: "09:00",
						Status:      string(constants.Draft),
						TimeZone:    "sample-timezone",
					},
				}).Once().Return(nil)
			},
			hasError: false,
		},
		{
			name: "upsert date info failed",
			dateInfo: &DateInfo{
				Date:        now,
				LocationID:  locationID,
				DateTypeID:  constants.RegularDay,
				OpeningTime: "09:00",
				Status:      constants.Draft,
				TimeZone:    "sample-timezone",
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetLocationByID", mock.Anything, mock.Anything, locationID).Once().Return(&dto.Location{}, nil)

				dateInfoRepo.On("UpsertDateInfo", mock.Anything, mock.Anything, &dto.UpsertDateInfoParams{
					DateInfo: &dto.DateInfo{
						Date:        now,
						LocationID:  locationID,
						DateTypeID:  string(constants.RegularDay),
						OpeningTime: "09:00",
						Status:      string(constants.Draft),
						TimeZone:    "sample-timezone",
					},
				}).Once().Return(errors.New("error"))
			},
			hasError: true,
		},
		{
			name: "fetch location id failed",
			dateInfo: &DateInfo{
				Date:        now,
				LocationID:  locationID,
				DateTypeID:  constants.RegularDay,
				OpeningTime: "09:00",
				Status:      constants.Draft,
				TimeZone:    "sample-timezone",
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetLocationByID", mock.Anything, mock.Anything, locationID).Once().Return(nil, errors.New("error"))
			},
			hasError: true,
		},
		{
			name: "invalid date info",
			dateInfo: &DateInfo{
				Date:        now,
				LocationID:  locationID,
				DateTypeID:  constants.ClosedDay,
				OpeningTime: "09:00",
				Status:      constants.Draft,
				TimeZone:    "sample-timezone",
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			tc.dateInfo.DateInfoRepo = dateInfoRepo
			tc.dateInfo.LocationRepo = locationRepo
			err := tc.dateInfo.Upsert(ctx)
			if err != nil {
				require.True(t, tc.hasError)
			} else {
				require.False(t, tc.hasError)
			}
		})
	}
}

func TestDateInfo_Duplicate(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	dateInfoRepo := &mock_repositories.MockDateInfoRepo{}
	locationRepo := &mock_repositories.MockLocationRepo{}
	locationID := idutil.ULIDNow()
	dates := []time.Time{
		now.Add(1 * 24 * time.Hour),
		now.Add(2 * 24 * time.Hour),
		now.Add(3 * 24 * time.Hour),
	}

	testCases := []struct {
		name     string
		dateInfo *DateInfo
		dates    []time.Time
		setup    func(context.Context)
		hasError bool
	}{
		{
			name: "duplicate date info successfully",
			dateInfo: &DateInfo{
				Date:        now,
				LocationID:  locationID,
				DateTypeID:  constants.RegularDay,
				OpeningTime: "09:00",
				Status:      constants.Draft,
				TimeZone:    "sample-timezone",
			},
			dates: dates,
			setup: func(ctx context.Context) {
				locationRepo.On("GetLocationByID", mock.Anything, mock.Anything, locationID).Once().Return(&dto.Location{}, nil)

				dateInfoRepo.On("GetDateInfoByDateAndLocationID", mock.Anything, mock.Anything, now, locationID).Once().
					Return(&dto.DateInfo{
						Date:        now,
						LocationID:  locationID,
						DateTypeID:  string(constants.RegularDay),
						OpeningTime: "09:00",
						Status:      string(constants.Draft),
						TimeZone:    "sample-timezone",
					}, nil)

				dateInfoRepo.On("DuplicateDateInfo", mock.Anything, mock.Anything, &dto.DuplicateDateInfoParams{
					DateInfo: &dto.DateInfo{
						Date:        now,
						LocationID:  locationID,
						DateTypeID:  string(constants.RegularDay),
						OpeningTime: "09:00",
						Status:      string(constants.Draft),
						TimeZone:    "sample-timezone",
					},
					Dates: dates,
				}).Once().Return(nil)
			},
			hasError: false,
		},
		{
			name: "duplicate date info failed",
			dateInfo: &DateInfo{
				Date:        now,
				LocationID:  locationID,
				DateTypeID:  constants.RegularDay,
				OpeningTime: "09:00",
				Status:      constants.Draft,
				TimeZone:    "sample-timezone",
			},
			dates: dates,
			setup: func(ctx context.Context) {
				locationRepo.On("GetLocationByID", mock.Anything, mock.Anything, locationID).Once().Return(&dto.Location{}, nil)

				dateInfoRepo.On("GetDateInfoByDateAndLocationID", mock.Anything, mock.Anything, now, locationID).Once().
					Return(&dto.DateInfo{
						Date:        now,
						LocationID:  locationID,
						DateTypeID:  string(constants.RegularDay),
						OpeningTime: "09:00",
						Status:      string(constants.Draft),
						TimeZone:    "sample-timezone",
					}, nil)

				dateInfoRepo.On("DuplicateDateInfo", mock.Anything, mock.Anything, &dto.DuplicateDateInfoParams{
					DateInfo: &dto.DateInfo{
						Date:        now,
						LocationID:  locationID,
						DateTypeID:  string(constants.RegularDay),
						OpeningTime: "09:00",
						Status:      string(constants.Draft),
						TimeZone:    "sample-timezone",
					},
					Dates: dates,
				}).Once().Return(errors.New("error"))
			},
			hasError: true,
		},
		{
			name: "failed to fetch date info",
			dateInfo: &DateInfo{
				Date:        now,
				LocationID:  locationID,
				DateTypeID:  constants.RegularDay,
				OpeningTime: "09:00",
				Status:      constants.Draft,
				TimeZone:    "sample-timezone",
			},
			dates: dates,
			setup: func(ctx context.Context) {
				locationRepo.On("GetLocationByID", mock.Anything, mock.Anything, locationID).Once().Return(&dto.Location{}, nil)

				dateInfoRepo.On("GetDateInfoByDateAndLocationID", mock.Anything, mock.Anything, now, locationID).Once().
					Return(nil, errors.New("error"))
			},
			hasError: true,
		},
		{
			name: "invalid date info",
			dateInfo: &DateInfo{
				Date:        now,
				LocationID:  locationID,
				DateTypeID:  constants.ClosedDay,
				OpeningTime: "09:00",
				Status:      constants.Draft,
				TimeZone:    "sample-timezone",
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			tc.dateInfo.DateInfoRepo = dateInfoRepo
			tc.dateInfo.LocationRepo = locationRepo
			err := tc.dateInfo.Duplicate(ctx, tc.dates)
			if err != nil {
				require.True(t, tc.hasError)
			} else {
				require.False(t, tc.hasError)
			}
		})
	}
}
