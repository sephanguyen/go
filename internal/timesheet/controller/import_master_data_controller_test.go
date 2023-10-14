package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_services "github.com/manabie-com/backend/mock/timesheet/services/import_master_data"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestImportMasterDataController_ImportTimesheetConfig(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	timesheetSV := new(mock_services.MockImportTimesheetConfigService)

	ctl := &ImportMasterDataController{
		ImportTimesheetConfigService: timesheetSV,
	}

	testCases := []TestCase{
		{
			name:         "happy case",
			ctx:          interceptors.ContextWithUserID(ctx, "user_id"),
			req:          &pb.ImportTimesheetConfigRequest{Payload: []byte("not empty")},
			expectedErr:  nil,
			expectedResp: &pb.ImportTimesheetConfigResponse{Errors: []*pb.ImportTimesheetConfigError{}},
			setup: func(ctx context.Context) {
				timesheetSV.
					On("ImportTimesheetConfig", ctx, mock.Anything).
					Return([]*pb.ImportTimesheetConfigError{}, nil).
					Once()
			},
		},
		{
			name:         "error case Import timesheet config failed",
			ctx:          interceptors.ContextWithUserID(ctx, "user_id"),
			req:          &pb.ImportTimesheetConfigRequest{Payload: []byte("not empty")},
			expectedErr:  errors.New("import failed"),
			expectedResp: (*pb.ImportTimesheetConfigResponse)(nil),
			setup: func(ctx context.Context) {
				timesheetSV.
					On("ImportTimesheetConfig", ctx, mock.Anything).
					Return(nil, errors.New("import failed")).
					Once()
			},
		},
		{
			name:         "error case empty payload",
			ctx:          interceptors.ContextWithUserID(ctx, "user_id"),
			req:          &pb.ImportTimesheetConfigRequest{},
			expectedErr:  status.Error(codes.InvalidArgument, "missing payload for import timesheet config"),
			expectedResp: (*pb.ImportTimesheetConfigResponse)(nil),
			setup: func(ctx context.Context) {
			},
		},
	}

	// Do Test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.ImportTimesheetConfigRequest)
			resp, err := ctl.ImportTimesheetConfig(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}
