package queries

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/domain"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	masterdata_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/assigned_student/repositories"
	lesson_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	academic_year_repositories "github.com/manabie-com/backend/mock/lessonmgmt/master_data/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name     string
	ctx      context.Context
	payloads *payloads.GetAssignedStudentListArg
	result   *GetAssignedStudentListResponse
	setup    func(ctx context.Context)
}

func TestAssignedStudentGRPCService_GetAssignedStudentList(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	assignedStudentRepo := new(mock_repositories.MockAssignedStudentRepo)

	s := &AssignedStudentQueryHandler{
		WrapperConnection:   wrapperConnection,
		AssignedStudentRepo: assignedStudentRepo,
	}
	courseId := "course-1"
	courses := []string{courseId}
	students := []string{"student-1"}
	locations := []string{"center-1"}
	getAssignedStudentError := fmt.Errorf("fail retrieve")
	timezone := "timezone-1"

	testCases := []TestCase{
		{
			name: "School Admin get list assigned student purchase slot successfully with filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetAssignedStudentListArg{
				Limit:                     2,
				CourseIDs:                 courses,
				StudentIDs:                students,
				KeyWord:                   "student name",
				LocationIDs:               locations,
				PurchaseMethod:            string(domain.PurchaseMethodSlot),
				Timezone:                  timezone,
			},
			result: &GetAssignedStudentListResponse{
				AsgStudents: []*domain.AssignedStudent{
					{
						StudentID:      "student-1",
						CourseID:       courseId,
						LocationID:     "center-1",
						Duration:       mock.Anything,
						PurchasedSlot:  int32(2),
						AssignedSlot:   int32(2),
						SlotGap:        int32(0),
						AssignedStatus: domain.AssignedStudentStatusJustAssigned,
					},
					{
						StudentID:      "student-2",
						CourseID:       courseId,
						LocationID:     "center-2",
						Duration:       mock.Anything,
						PurchasedSlot:  int32(2),
						AssignedSlot:   int32(2),
						SlotGap:        int32(0),
						AssignedStatus: domain.AssignedStudentStatusJustAssigned,
					},
				},
				Total:    uint32(99),
				OffsetID: "",
				Error:    nil,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				assignedStudentRepo.On("GetAssignedStudentList", mock.Anything, db, &payloads.GetAssignedStudentListArg{
					Limit:                     2,
					CourseIDs:                 courses,
					StudentIDs:                students,
					KeyWord:                   "student name",
					LocationIDs:               locations,
					PurchaseMethod:            string(domain.PurchaseMethodSlot),
					Timezone:                  timezone,
				}).Once().Return([]*domain.AssignedStudent{
					{
						StudentID:      "student-1",
						CourseID:       courseId,
						LocationID:     "center-1",
						Duration:       mock.Anything,
						PurchasedSlot:  int32(2),
						AssignedSlot:   int32(2),
						SlotGap:        int32(0),
						AssignedStatus: domain.AssignedStudentStatusJustAssigned,
					},
					{
						StudentID:      "student-2",
						CourseID:       courseId,
						LocationID:     "center-2",
						Duration:       mock.Anything,
						PurchasedSlot:  int32(2),
						AssignedSlot:   int32(2),
						SlotGap:        int32(0),
						AssignedStatus: domain.AssignedStudentStatusJustAssigned,
					},
				}, uint32(99), "pre_id", uint32(2), nil)
			},
		},
		{
			name: "School Admin get list assigned student recurring slot successfully with filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetAssignedStudentListArg{
				Limit:                     2,
				CourseIDs:                 courses,
				StudentIDs:                students,
				KeyWord:                   "student name",
				LocationIDs:               locations,
				PurchaseMethod:            string(domain.PurchaseMethodRecurring),
				Timezone:                  timezone,
			},
			result: &GetAssignedStudentListResponse{
				AsgStudents: []*domain.AssignedStudent{
					{
						StudentID:      "student-1",
						CourseID:       courseId,
						LocationID:     "center-1",
						Duration:       mock.Anything,
						PurchasedSlot:  int32(2),
						AssignedSlot:   int32(2),
						SlotGap:        int32(0),
						AssignedStatus: domain.AssignedStudentStatusJustAssigned,
					},
					{
						StudentID:      "student-2",
						CourseID:       courseId,
						LocationID:     "center-2",
						Duration:       mock.Anything,
						PurchasedSlot:  int32(2),
						AssignedSlot:   int32(2),
						SlotGap:        int32(0),
						AssignedStatus: domain.AssignedStudentStatusJustAssigned,
					},
				},
				Total:    uint32(99),
				OffsetID: "",
				Error:    nil,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				assignedStudentRepo.On("GetAssignedStudentList", mock.Anything, db, &payloads.GetAssignedStudentListArg{
					Limit:                     2,
					CourseIDs:                 courses,
					StudentIDs:                students,
					KeyWord:                   "student name",
					LocationIDs:               locations,
					PurchaseMethod:            string(domain.PurchaseMethodRecurring),
					Timezone:                  timezone,
				}).Once().Return([]*domain.AssignedStudent{
					{
						StudentID:      "student-1",
						CourseID:       courseId,
						LocationID:     "center-1",
						Duration:       mock.Anything,
						PurchasedSlot:  int32(2),
						AssignedSlot:   int32(2),
						SlotGap:        int32(0),
						AssignedStatus: domain.AssignedStudentStatusJustAssigned,
					},
					{
						StudentID:      "student-2",
						CourseID:       courseId,
						LocationID:     "center-2",
						Duration:       mock.Anything,
						PurchasedSlot:  int32(2),
						AssignedSlot:   int32(2),
						SlotGap:        int32(0),
						AssignedStatus: domain.AssignedStudentStatusJustAssigned,
					},
				}, uint32(99), "pre_id", uint32(2), nil)
			},
		},
		{
			name:     "Retrieve fail with getting assigned students has error",
			ctx:      interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetAssignedStudentListArg{},
			result: &GetAssignedStudentListResponse{
				Error: status.Error(codes.Internal, getAssignedStudentError.Error()),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				assignedStudentRepo.On("GetAssignedStudentList", mock.Anything, db, mock.Anything).Once().Return(nil, uint32(0), "", uint32(0), getAssignedStudentError)
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp := s.GetAssignedStudentList(testCase.ctx, testCase.payloads)
			expectedErr := testCase.result.Error
			if expectedErr != nil {
				assert.Error(t, resp.Error)
				assert.Equal(t, expectedErr.Error(), resp.Error.Error())
			} else {
				assert.NoError(t, resp.Error)
				assert.Equal(t, testCase.result, resp)
			}

			mock.AssertExpectationsForObjects(t, assignedStudentRepo, mockUnleashClient)
		})
	}
}

func TestAssignedStudentGRPCService_GetStudentAttendance(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	assignedStudentRepo := new(mock_repositories.MockAssignedStudentRepo)
	reallocationRepo := &lesson_repositories.MockReallocationRepo{}
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	academicYeerRepo := new(academic_year_repositories.MockAcademicYearRepository)

	testCases := []struct {
		name   string
		req    *GetStudentAttendanceRequest
		resp   *GetStudentAttendanceResponse
		setup  func(ctx context.Context)
		hasErr bool
	}{
		{
			name: "list student attendance that no filter",
			req: &GetStudentAttendanceRequest{
				SearchKey: "",
				Timezone:  "UTC",
				Filter:    domain.Filter{},
				Paging: support.Paging[int]{
					Limit:  5,
					Offset: 0,
				},
			},
			resp: &GetStudentAttendanceResponse{
				Total: 3,
				StudentAttendance: []*domain.StudentAttendance{
					{
						LessonID:     "lesson-1",
						StudentID:    "student-1",
						AttendStatus: string(lesson_domain.StudentAttendStatusAttend),
						CourseID:     "course-1",
						LocationID:   "location-1",
					},
					{
						LessonID:     "lesson-1",
						StudentID:    "student-2",
						AttendStatus: string(lesson_domain.StudentAttendStatusAttend),
						CourseID:     "course-1",
						LocationID:   "location-1",
					},
					{
						LessonID:            "lesson-1",
						StudentID:           "student-3",
						AttendStatus:        string(lesson_domain.StudentAttendStatusReallocate),
						CourseID:            "course-1",
						LocationID:          "location-1",
						ReallocatedLessonID: "lesson-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				academicYeerRepo.On("GetCurrentAcademicYear", mock.Anything, mock.Anything).Once().Return(&masterdata_domain.AcademicYear{
					AcademicYearID: database.Text("y1"),
				}, nil)
				assignedStudentRepo.On("GetStudentAttendance", ctx, mock.Anything, mock.Anything).Return(
					[]*domain.StudentAttendance{
						{
							LessonID:     "lesson-1",
							StudentID:    "student-1",
							AttendStatus: string(lesson_domain.StudentAttendStatusAttend),
							CourseID:     "course-1",
							LocationID:   "location-1",
						},
						{
							LessonID:     "lesson-1",
							StudentID:    "student-2",
							AttendStatus: string(lesson_domain.StudentAttendStatusAttend),
							CourseID:     "course-1",
							LocationID:   "location-1",
						},
						{
							LessonID:     "lesson-1",
							StudentID:    "student-3",
							AttendStatus: string(lesson_domain.StudentAttendStatusReallocate),
							CourseID:     "course-1",
							LocationID:   "location-1",
						},
					}, uint32(3), nil).Once()
				reallocationRepo.On("GetReallocatedLesson", ctx, mock.Anything, []string{"lesson-1", "student-3"}).Return(
					[]*lesson_domain.Reallocation{
						{
							OriginalLessonID: "lesson-1",
							StudentID:        "student-3",
							NewLessonID:      "lesson-2",
							CourseID:         "course-1",
						},
					}, nil).Once()
			},
		},
	}

	handler := &AssignedStudentQueryHandler{
		WrapperConnection:   wrapperConnection,
		AssignedStudentRepo: assignedStudentRepo,
		ReallocationRepo:    reallocationRepo,
		AcademicYearRepo:    academicYeerRepo,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			res, err := handler.GetStudentAttendance(ctx, tc.req)
			if err != nil {
				require.True(t, tc.hasErr)
			} else {
				require.Equal(t, tc.resp.Total, res.Total)
				require.Equal(t, tc.resp.StudentAttendance, res.StudentAttendance)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}
