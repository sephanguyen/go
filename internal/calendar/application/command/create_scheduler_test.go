package command

import (
	"context"
	"fmt"
	"testing"
	"time"

	mock_repositories "github.com/manabie-com/backend/mock/calendar/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreateSchedulerCommand_CreateManySchedulers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mockDB := &mock_database.Ext{}
	mockSchedulerRepo := new(mock_repositories.MockSchedulerRepo)

	createSchedulerCmd := CreateSchedulerCommand{
		SchedulerRepo: mockSchedulerRepo,
	}

	testCases := []struct {
		name        string
		req         *cpb.CreateManySchedulersRequest
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name: "invalid params",
			req: &cpb.CreateManySchedulersRequest{
				Schedulers: []*cpb.CreateSchedulerWithIdentityRequest{
					{
						Identity: "lesson_id_01",
						Request:  nil,
					},
					{
						Identity: "lesson_id_02",
						Request: &cpb.CreateSchedulerRequest{
							StartDate: timestamppb.New(time.Date(2023, 4, 10, 0, 0, 0, 0, time.UTC)),
							EndDate:   timestamppb.New(time.Date(2023, 4, 8, 1, 0, 0, 0, time.UTC)),
							Frequency: cpb.Frequency_ONCE,
						},
					},
					{
						Identity: "",
						Request:  nil,
					},
				},
			},
			expectedErr: fmt.Errorf("build params to create many schedulers fail: lesson_id_01 missing param to create scheduler; end date of lesson_id_02 is earlier than start date; missing Identity"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "CreateMany error",
			req: &cpb.CreateManySchedulersRequest{
				Schedulers: []*cpb.CreateSchedulerWithIdentityRequest{
					{
						Identity: "lesson_id_01",
						Request: &cpb.CreateSchedulerRequest{
							StartDate: timestamppb.New(time.Date(2023, 4, 8, 0, 0, 0, 0, time.UTC)),
							EndDate:   timestamppb.New(time.Date(2023, 4, 8, 1, 0, 0, 0, time.UTC)),
							Frequency: cpb.Frequency_ONCE,
						},
					},
				},
			},
			expectedErr: fmt.Errorf("create many schedulers fail: error"),
			setup: func(ctx context.Context) {
				mockSchedulerRepo.On("CreateMany", ctx, mockDB, mock.Anything).Once().Return(map[string]string{}, fmt.Errorf("error"))
			},
		},
		{
			name: "success",
			req: &cpb.CreateManySchedulersRequest{
				Schedulers: []*cpb.CreateSchedulerWithIdentityRequest{
					{
						Identity: "lesson_id_01",
						Request: &cpb.CreateSchedulerRequest{
							StartDate: timestamppb.New(time.Date(2023, 4, 8, 0, 0, 0, 0, time.UTC)),
							EndDate:   timestamppb.New(time.Date(2023, 4, 8, 1, 0, 0, 0, time.UTC)),
							Frequency: cpb.Frequency_ONCE,
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockSchedulerRepo.On("CreateMany", ctx, mockDB, mock.Anything).Once().Return(map[string]string{
					"lesson_id_01": "scheduler_id_01",
				}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			resp, err := createSchedulerCmd.CreateManySchedulers(ctx, mockDB, tc.req)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
				assert.Empty(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, resp)
			}
		})
	}
}
