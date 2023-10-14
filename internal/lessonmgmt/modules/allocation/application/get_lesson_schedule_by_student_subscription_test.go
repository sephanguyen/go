package application

import (
	"context"
	"testing"
	"time"

	golibs_db "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/domain"
	course_location_schedule_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	master_data_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	course_location_schedule_repositories "github.com/manabie-com/backend/mock/lessonmgmt/course_location_schedule/repositories"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson_allocation/repositories"
	academic_week_repositories "github.com/manabie-com/backend/mock/lessonmgmt/master_data/repositories"
	user_mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetLessonScheduleByStudentSubscriptionHandler_GetLessonScheduleByStudentSubscription(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonAllocationRepo := new(mock_repositories.MockLessonAllocationRepo)
	studentSubscriptionRepo := new(user_mock_repositories.MockStudentSubscriptionRepo)
	studentSubscriptionAccessPathRepo := new(user_mock_repositories.MockStudentSubscriptionAccessPathRepo)
	academicWeekRepo := new(academic_week_repositories.MockAcademicWeekRepository)
	courseLocationScheduleRepo := new(course_location_schedule_repositories.MockCourseLocationScheduleRepo)

	testCase := []struct {
		name     string
		setup    func(context.Context)
		req      *lpb.GetLessonScheduleByStudentSubscriptionRequest
		resp     *lpb.GetLessonScheduleByStudentSubscriptionResponse
		hasError bool
	}{
		{
			name: "happy case",
			req: &lpb.GetLessonScheduleByStudentSubscriptionRequest{
				StudentSubscriptionId: "student-subscription-1",
				Paging: &cpb.Paging{
					Limit: 5,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			resp: &lpb.GetLessonScheduleByStudentSubscriptionResponse{
				TotalItems:            2,
				AllocatedLessonsCount: 3,
				TotalLesson:           10,
				Items: []*lpb.GetLessonScheduleByStudentSubscriptionResponse_WeeklyLessonList{
					{
						AcademicWeekId: "week_1",
						WeekOrder:      1,
						WeekName:       "Week 1",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentSubscriptionRepo.On("GetByStudentSubscriptionID", mock.Anything, mock.Anything, "student-subscription-1").Once().Return(&user_domain.StudentSubscription{
					StudentSubscriptionID: "student-subscription-1",
					CourseID:              "course-1",
					StudentID:             "student-1",
					LocationIDs:           []string{"location-1"},
					StartAt:               time.Date(2023, 6, 5, 0, 0, 0, 0, time.UTC),
					EndAt:                 time.Date(2023, 7, 5, 0, 0, 0, 0, time.UTC),
				}, nil)
				studentSubscriptionAccessPathRepo.On("FindLocationsByStudentSubscriptionIDs", mock.Anything, mock.Anything, []string{"student-subscription-1"}).
					Once().Return(map[string][]string{
					"student-subscription-1": {"location-1"},
				}, nil)

				totalLesson := 10

				courseLocationScheduleRepo.On("GetByCourseIDAndLocationID", mock.Anything, mock.Anything, "course-1", "location-1").
					Once().Return(&course_location_schedule_domain.CourseLocationSchedule{
					ProductTypeSchedule: course_location_schedule_domain.OneTime,
					TotalNoLesson:       &totalLesson,
					AcademicWeeks:       []string{"1,2,3,4,5,6"},
				}, nil)
				academicWeekRepo.On("GetByDateRange", mock.Anything, mock.Anything, "location-1", []string{"1,2,3,4,5,6"}, mock.Anything, mock.Anything).
					Once().Return([]*master_data_domain.AcademicWeek{
					{
						AcademicWeekID: golibs_db.Text("week_1"),
						WeekOrder:      golibs_db.Int2(1),
						Name:           golibs_db.Text("Week 1"),
						StartDate:      pgtype.Date(golibs_db.Timestamptz(time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC))),
						EndDate:        pgtype.Date(golibs_db.Timestamptz(time.Date(2023, 6, 8, 0, 0, 0, 0, time.UTC))),
					},
					{
						AcademicWeekID: golibs_db.Text("week_2"),
						WeekOrder:      golibs_db.Int2(2),
						Name:           golibs_db.Text("Week 2"),
						StartDate:      pgtype.Date(golibs_db.Timestamptz(time.Date(2023, 6, 9, 0, 0, 0, 0, time.UTC))),
						EndDate:        pgtype.Date(golibs_db.Timestamptz(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC))),
					},
				}, nil)
				lessonAllocationRepo.On("GetByStudentSubscriptionAndWeek", mock.Anything, mock.Anything, "student-1", "course-1", []string{"week_1", "week_2"}).
					Once().Return(map[string][]*domain.LessonAllocationInfo{
					"week_1": {
						{
							LessonID:         "lesson-1",
							StartTime:        time.Date(2023, 6, 5, 0, 0, 0, 0, time.UTC),
							EndTime:          time.Date(2023, 6, 5, 1, 0, 0, 0, time.UTC),
							LocationID:       "location-1",
							AttendanceStatus: lesson_domain.StudentAttendStatusAttend,
							Status:           lesson_domain.LessonSchedulingStatusCompleted,
							TeachingMethod:   lesson_domain.LessonTeachingMethodIndividual,
							LessonReportID:   "",
						},
					},
					"week_2": {},
				}, nil)
				lessonAllocationRepo.On("CountAssignedSlotPerStudentCourse", mock.Anything, mock.Anything, "student-1", "course-1").
					Once().Return(uint32(3), nil)
				lessonAllocationRepo.On("CountPurchasedSlotPerStudentSubscription", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(uint32(10), nil)
			},
		},
	}
	handler := GetLessonScheduleByStudentSubscriptionHandler{
		LessonAllocationRepo:              lessonAllocationRepo,
		WrapperConnection:                 wrapperConnection,
		StudentSubscriptionRepo:           studentSubscriptionRepo,
		StudentSubscriptionAccessPathRepo: studentSubscriptionAccessPathRepo,
		AcademicWeekRepo:                  academicWeekRepo,
		CourseLocationScheduleRepo:        courseLocationScheduleRepo,
	}
	for _, tc := range testCase {
		tc.setup(ctx)
		t.Run(tc.name, func(t *testing.T) {
			res, err := handler.GetLessonScheduleByStudentSubscription(ctx, tc.req)
			if err != nil {
				require.True(t, tc.hasError)
			} else {
				require.Equal(t, tc.resp.TotalItems, res.TotalItems)
				require.Equal(t, tc.resp.TotalLesson, res.TotalLesson)
				require.Equal(t, tc.resp.AllocatedLessonsCount, res.AllocatedLessonsCount)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}
