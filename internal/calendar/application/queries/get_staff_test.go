package queries

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/calendar/application/queries/payloads"
	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	mock_repositories "github.com/manabie-com/backend/mock/calendar/repositories"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetStaff_GetStaffsByLocation(t *testing.T) {
	t.Parallel()

	mockUserRepo := &mock_repositories.MockUserRepo{}
	mockDB := testutil.NewMockDB()
	status := []string{
		string(constants.Available),
		string(constants.OnLeave),
	}

	users := []*dto.User{
		{
			UserID: "user-id-1",
			Name:   "user1",
		},
		{
			UserID: "user-id-2",
			Name:   "user2",
		},
	}

	testCases := []struct {
		name     string
		req      *payloads.GetStaffRequest
		setup    func(context.Context)
		hasError bool
	}{
		{
			name: "success",
			req: &payloads.GetStaffRequest{
				LocationID:                "location-id",
				IsUsingUserBasicInfoTable: false,
			},
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetStaffsByLocationAndWorkingStatus", mock.Anything, mockDB.DB, "location-id", status, false).Once().Return(users, nil)
			},
		},
		{
			name: "success with using user basic info",
			req: &payloads.GetStaffRequest{
				LocationID:                "location-id",
				IsUsingUserBasicInfoTable: true,
			},
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetStaffsByLocationAndWorkingStatus", mock.Anything, mockDB.DB, "location-id", status, true).Once().Return(users, nil)
			},
		},
		{
			name: "failed",
			req: &payloads.GetStaffRequest{
				LocationID:                "location-id",
				IsUsingUserBasicInfoTable: false,
			},
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetStaffsByLocationAndWorkingStatus", mock.Anything, mockDB.DB, "location-id", status, false).Once().Return(nil, fmt.Errorf("error"))
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			getStaff := &GetStaff{
				UserRepo: mockUserRepo,
			}
			resp, err := getStaff.GetStaffsByLocation(ctx, mockDB.DB, tc.req)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NotEmpty(t, resp)
				assert.NoError(t, err)
				assert.Equal(t, users, resp.User)
				mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
			}
		})
	}
}
func TestGetStaff_GetStaffsByLocationIDsAndNameOrEmail(t *testing.T) {
	t.Parallel()

	mockUserRepo := &mock_repositories.MockUserRepo{}
	mockDB := testutil.NewMockDB()
	locationIds := []string{"location-1", "location-2"}
	teacherIDs := []string{"teacherID"}
	keyword := "name"
	users := []*dto.User{
		{
			UserID: "user-id-1",
			Name:   "user1",
			Email:  "email1",
		},
		{
			UserID: "user-id-2",
			Name:   "user2",
			Email:  "email2",
		},
	}

	testCases := []struct {
		name     string
		req      *payloads.GetStaffByLocationIDsAndNameOrEmailRequest
		setup    func(context.Context)
		hasError bool
	}{
		{
			name: "success",
			req: &payloads.GetStaffByLocationIDsAndNameOrEmailRequest{
				LocationIDs:        locationIds,
				Keyword:            keyword,
				FilteredTeacherIDs: teacherIDs,
				Limit:              10,
			},
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetStaffsByLocationIDsAndNameOrEmail", mock.Anything, mockDB.DB, locationIds, teacherIDs, keyword, 10).Once().Return(users, nil)
			},
		},
		{
			name: "failed",
			req: &payloads.GetStaffByLocationIDsAndNameOrEmailRequest{
				LocationIDs:        locationIds,
				Keyword:            keyword,
				FilteredTeacherIDs: teacherIDs,
				Limit:              10,
			},
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetStaffsByLocationIDsAndNameOrEmail", mock.Anything, mockDB.DB, locationIds, teacherIDs, keyword, 10).Once().Return(nil, fmt.Errorf("error"))
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			getStaff := &GetStaff{
				UserRepo: mockUserRepo,
			}
			resp, err := getStaff.GetStaffsByLocationIDsAndNameOrEmail(ctx, mockDB.DB, tc.req)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NotEmpty(t, resp)
				assert.NoError(t, err)
				assert.Equal(t, users, resp.User)
				mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
			}
		})
	}
}
