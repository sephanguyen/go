package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_lesson_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
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

func TestCourseLocationScheduleService_ImportCourseLocationSchedule(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	courseLocationScheduleRepo := new(mock_lesson_repositories.MockCourseLocationScheduleRepo)

	s := NewCourseLocationScheduleService(wrapperConnection, courseLocationScheduleRepo)

	testCases := []TestCase{
		{
			name: "parsing valid file (without validation error lines in response)",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &lpb.ImportCourseLocationScheduleRequest{
				Payload: []byte(fmt.Sprintf(`course_location_schedule_id,course_id,location_id,academic_week,product_type_schedule,frequency,total_no_lessons
				1,01GWK3QSBD4D0PKBXENQN1CPY6,01GWK39VWTEFF3DMH8MJCPARE0,1_2,1,,3
				2,01GWK3QSBD4D0PKBXENQN1CPY6,01GWK39VWX66PNQGQQMEXT8MZ0,1_2,2,3,
				3,01GWK5J5V1S3PDX2Z459X77YXR,01GWK5HJ1YAZEK3D9SPN08CT81,1_2,3,,
				,01GWK3NF2ZF2TKNGEN26YJQMPZ,01GWK39VVZ1FBCK4740RJQA0N8,1_2,4,,`)),
			},
			expectedResp: &lpb.ImportCourseLocationScheduleResponse{
				Errors: []*lpb.ImportCourseLocationScheduleResponse_ImportError{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				courseLocationScheduleRepo.On("UpsertMultiCourseLocationSchedule", ctx, db, mock.Anything).Once().Return(nil)
				courseLocationScheduleRepo.On("GetAcademicWeekValid", ctx, db, mock.Anything, mock.Anything).Once().Return(map[string]bool{
					"01GWK39VWTEFF3DMH8MJCPARE0-1": true, "01GWK39VWTEFF3DMH8MJCPARE0-2": true,
					"01GWK39VWX66PNQGQQMEXT8MZ0-1": true, "01GWK39VWX66PNQGQQMEXT8MZ0-2": true,
					"01GWK5HJ1YAZEK3D9SPN08CT81-1": true, "01GWK5HJ1YAZEK3D9SPN08CT81-2": true,
				}, nil)

			},
			expectedErr: nil,
		},
		{
			name: "parsing invalid when academic_week empty",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &lpb.ImportCourseLocationScheduleRequest{
				Payload: []byte(fmt.Sprintf(`course_location_schedule_id,course_id,location_id,academic_week,product_type_schedule,frequency,total_no_lessons
				,01GWK3QSBD4D0PKBXENQN1CPY6,01GWK39VWX66PNQGQQMEXT8MZ0,,2,3,
				`)),
			},
			expectedResp: &lpb.ImportCourseLocationScheduleResponse{
				Errors: []*lpb.ImportCourseLocationScheduleResponse_ImportError{
					{
						RowNumber: 1,
						Error:     "academic_week is required",
					},
				},
			},
			expectedErr: nil,
			setup:       func(ctx context.Context) {},
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
			mock.AssertExpectationsForObjects(t, db, courseLocationScheduleRepo, mockUnleashClient)
		})
	}
}

func TestCourseLocationScheduleService_ExportCourseLocationSchedule(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	courseLocationScheduleRepo := new(mock_lesson_repositories.MockCourseLocationScheduleRepo)

	s := NewCourseLocationScheduleService(wrapperConnection, courseLocationScheduleRepo)

	testCases := []TestCase{
		{
			name: "export data success",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			expectedResp: &lpb.ExportCourseLocationScheduleResponse{
				Data: []byte(fmt.Sprintf(`course_location_schedule_id,course_id,location_id,academic_week,product_type_schedule,frequency,total_no_lessons
				,01GWK3QSBD4D0PKBXENQN1CPY6,01GWK39VWTEFF3DMH8MJCPARE0,1_2_3_4_5,1,,3
				`)),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				totalNoLesson := 3
				courseLocationScheduleRepo.On("ExportCourseLocationSchedule", ctx, db, mock.Anything).Once().Return([]*domain.CourseLocationSchedule{
					{CourseID: "01GWK3QSBD4D0PKBXENQN1CPY6", LocationID: "01GWK39VWTEFF3DMH8MJCPARE0",
						AcademicWeeks: []string{"1", "2", "3", "4", "5"}, ProductTypeSchedule: domain.OneTime, Frequency: nil, TotalNoLesson: &totalNoLesson},
				}, nil)

			},
			expectedErr: nil,
		},

		{
			name:        "error",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: fmt.Errorf("no rows in result set"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				courseLocationScheduleRepo.On("ExportCourseLocationSchedule", ctx, db, mock.Anything).Once().Return(nil, fmt.Errorf("no rows in result set"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.ExportCourseLocationSchedule(testCase.ctx)
			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
			}
			mock.AssertExpectationsForObjects(t, db, courseLocationScheduleRepo)
		})
	}
}
