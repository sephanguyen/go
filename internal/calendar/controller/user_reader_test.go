package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/manabie-com/backend/internal/calendar/application/queries/payloads"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_queries "github.com/manabie-com/backend/mock/calendar/application/queries"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	"github.com/manabie-com/backend/mock/testutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserReaderService_GetStaffsByLocation(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	staffQueryHandler := &mock_queries.MockGetStaff{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	service := &UserReaderService{
		UserPort:         staffQueryHandler,
		db:               mockDB.DB,
		unleashClientIns: mockUnleashClient,
		env:              "local",
	}

	locationID := idutil.ULIDNow()

	t.Run("success", func(t *testing.T) {
		req := &cpb.GetStaffsByLocationRequest{
			LocationId: locationID,
		}
		mockUnleashClient.On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).Return(false, nil).Once()
		staffQueryHandler.On("GetStaffsByLocation", mock.Anything, mockDB.DB, &payloads.GetStaffRequest{
			LocationID:                req.LocationId,
			IsUsingUserBasicInfoTable: false,
		}).Once().Return(&payloads.GetStaffResponse{
			User: []*dto.User{
				{
					UserID: "user-id1",
					Name:   "user-name1",
				},
				{
					UserID: "user-id2",
					Name:   "user-name2",
				},
			},
		}, nil)

		res, err := service.GetStaffsByLocation(context.Background(), req)
		require.Nil(t, err)
		require.NotNil(t, res)
	})

	t.Run("success with using basic info", func(t *testing.T) {
		req := &cpb.GetStaffsByLocationRequest{
			LocationId: locationID,
		}
		mockUnleashClient.On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Once()
		staffQueryHandler.On("GetStaffsByLocation", mock.Anything, mockDB.DB, &payloads.GetStaffRequest{
			LocationID:                req.LocationId,
			IsUsingUserBasicInfoTable: true,
		}).Once().Return(&payloads.GetStaffResponse{
			User: []*dto.User{
				{
					UserID: "user-id1",
					Name:   "user-name1",
				},
				{
					UserID: "user-id2",
					Name:   "user-name2",
				},
			},
		}, nil)

		res, err := service.GetStaffsByLocation(context.Background(), req)
		require.Nil(t, err)
		require.NotNil(t, res)
	})

	t.Run("failed", func(t *testing.T) {
		req := &cpb.GetStaffsByLocationRequest{
			LocationId: locationID,
		}
		mockUnleashClient.On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).Return(false, nil).Once()
		staffQueryHandler.On("GetStaffsByLocation", mock.Anything, mockDB.DB, &payloads.GetStaffRequest{
			LocationID:                req.LocationId,
			IsUsingUserBasicInfoTable: false,
		}).Once().Return(nil, errors.New("error"))

		res, err := service.GetStaffsByLocation(context.Background(), req)
		require.NotNil(t, err)
		require.Nil(t, res)
	})
}
func TestUserReaderService_GetStaffsByLocationIDsAndNameOrEmail(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	staffQueryHandler := &mock_queries.MockGetStaff{}
	service := &UserReaderService{
		UserPort: staffQueryHandler,
		db:       mockDB.DB,
		env:      "local",
	}

	locationIDs := []string{idutil.ULIDNow()}
	teacherIDs := []string{"teacherID"}
	keyword := "name"

	t.Run("success", func(t *testing.T) {
		req := &cpb.GetStaffsByLocationIDsAndNameOrEmailRequest{
			LocationIds:        locationIDs,
			Keyword:            keyword,
			FilteredTeacherIds: teacherIDs,
		}
		staffQueryHandler.On("GetStaffsByLocationIDsAndNameOrEmail", mock.Anything, mockDB.DB, &payloads.GetStaffByLocationIDsAndNameOrEmailRequest{
			LocationIDs:        req.LocationIds,
			Keyword:            req.Keyword,
			FilteredTeacherIDs: req.FilteredTeacherIds,
		}).Once().Return(&payloads.GetStaffByLocationIDsAndNameOrEmailResponse{
			User: []*dto.User{
				{
					UserID: "user-id1",
					Name:   "user-name1",
					Email:  "user-email1",
				},
				{
					UserID: "user-id2",
					Name:   "user-name2",
					Email:  "user-email2",
				},
			},
		}, nil)

		res, err := service.GetStaffsByLocationIDsAndNameOrEmail(context.Background(), req)
		require.Nil(t, err)
		require.NotNil(t, res)
	})

	t.Run("failed", func(t *testing.T) {
		req := &cpb.GetStaffsByLocationIDsAndNameOrEmailRequest{
			LocationIds:        locationIDs,
			Keyword:            keyword,
			FilteredTeacherIds: teacherIDs,
		}
		staffQueryHandler.On("GetStaffsByLocationIDsAndNameOrEmail", mock.Anything, mockDB.DB, &payloads.GetStaffByLocationIDsAndNameOrEmailRequest{
			LocationIDs:        req.LocationIds,
			Keyword:            req.Keyword,
			FilteredTeacherIDs: req.FilteredTeacherIds,
		}).Once().Return(nil, errors.New("error"))

		res, err := service.GetStaffsByLocationIDsAndNameOrEmail(context.Background(), req)
		require.NotNil(t, err)
		require.Nil(t, res)
	})
}
