package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_services "github.com/manabie-com/backend/mock/lessonmgmt/lesson/course_location_schedule"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestCourseLocationScheduleController_ImportCourseLocationSchedule(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	courseLocationScheduleService := new(mock_services.MockCourseLocationScheduleService)

	s := NewCourseLocationControllerController(wrapperConnection, courseLocationScheduleService)

	testCases := []TestCase{
		{
			name: "parsing valid file (without validation error lines in response)",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &lpb.ImportCourseLocationScheduleRequest{
				Payload: []byte(fmt.Sprintf(`course_location_schedule_id,course_id,location_id,academic_week,product_type_schedule,frequency,
				,01GWK3QSBD4D0PKBXENQN1CPY6,01GWK39VWTEFF3DMH8MJCPARE0,1_2_3_4_5,1,,3
				,01GWK3QSBD4D0PKBXENQN1CPY6,01GWK39VWX66PNQGQQMEXT8MZ0,1_2_3_4_5,2,3,`)),
			},
			expectedResp: &lpb.ImportCourseLocationScheduleResponse{
				Errors: []*lpb.ImportCourseLocationScheduleResponse_ImportError{},
			},
			setup: func(ctx context.Context) {
				courseLocationScheduleService.On("ImportCourseLocationSchedule", ctx, mock.Anything).Once().Return(&lpb.ImportCourseLocationScheduleResponse{
					Errors: []*lpb.ImportCourseLocationScheduleResponse_ImportError{},
				}, nil)

			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.ImportCourseLocationSchedule(testCase.ctx, testCase.req.(*lpb.ImportCourseLocationScheduleRequest))
			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.NotNil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.expectedResp.(*lpb.ImportCourseLocationScheduleResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, err.RowNumber, expectedResp.Errors[i].RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}
			mock.AssertExpectationsForObjects(t, db, courseLocationScheduleService, mockUnleashClient)
		})
	}
}
