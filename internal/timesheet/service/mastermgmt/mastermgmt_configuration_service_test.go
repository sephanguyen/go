package mastermgmt

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mock_mastermgmt "github.com/manabie-com/backend/mock/timesheet/service/mastermgmt"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func TestMasterConfigurationServiceImpl_CheckPartnerTimesheetServiceStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		masterMgmtConfigurationServiceClient = new(mock_mastermgmt.MasterMgmtConfigurationServiceClient)
	)

	s := MasterConfigurationServiceImpl{
		MasterMgmtConfigurationServiceClient: masterMgmtConfigurationServiceClient,
	}

	getConfigurationByKeyResponseWithOnStatus := &mpb.GetConfigurationByKeyResponse{
		Configuration: &mpb.Configuration{
			Id:          "test_id",
			ConfigKey:   "hcm.timesheet_management",
			ConfigValue: "on",
		},
	}

	getConfigurationByKeyResponseWithOffStatus := &mpb.GetConfigurationByKeyResponse{
		Configuration: &mpb.Configuration{
			Id:          "test_id",
			ConfigKey:   "hcm.timesheet_management",
			ConfigValue: "off",
		},
	}

	testCases := []struct {
		name         string
		ctx          context.Context
		expectedResp bool
		expectedErr  error
		setup        func(ctx context.Context)
	}{
		{
			name:         "check partner timesheet service status success with on status",
			ctx:          ctx,
			expectedResp: true,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				masterMgmtConfigurationServiceClient.On("GetConfigurationByKey", mock.Anything, mock.Anything).
					Return(getConfigurationByKeyResponseWithOnStatus, nil).Once()

			},
		},
		{
			name:         "happy case when timesheet service status off",
			ctx:          ctx,
			expectedResp: false,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				masterMgmtConfigurationServiceClient.On("GetConfigurationByKey", mock.Anything, mock.Anything).
					Return(getConfigurationByKeyResponseWithOffStatus, nil).Once()

			},
		},

		{
			name:         "error when check partner timesheet service status",
			ctx:          ctx,
			expectedResp: false,
			expectedErr:  fmt.Errorf("s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey: %s", status.Error(codes.FailedPrecondition, "configuration key cannot be empty")),
			setup: func(ctx context.Context) {
				masterMgmtConfigurationServiceClient.On("GetConfigurationByKey", mock.Anything, mock.Anything).
					Return(&mpb.GetConfigurationByKeyResponse{}, status.Error(codes.FailedPrecondition, "configuration key cannot be empty")).Once()

			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.CheckPartnerTimesheetServiceIsOn(testCase.ctx)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}
