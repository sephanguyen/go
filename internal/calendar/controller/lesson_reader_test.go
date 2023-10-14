package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/calendar/application/queries/payloads"
	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/calendar/support"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	mock_queries "github.com/manabie-com/backend/mock/calendar/application/queries"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	"github.com/manabie-com/backend/mock/testutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLessonReaderService_GetLessonDetailOnCalendar(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(mockDB.DB, mockDB.DB, mockUnleashClient, "local")
	lessonQueryHandler := &mock_queries.MockLessonQueryHandler{}
	service := &LessonReaderService{
		wrapperConnection:  wrapperConnection,
		lessonQueryHandler: lessonQueryHandler,
	}

	req := &cpb.GetLessonDetailOnCalendarRequest{
		LessonId: "lesson-id1",
	}
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		lessonQueryHandler.On("GetLessonDetail", mock.Anything, mockDB.DB, &payloads.GetLessonDetailRequest{
			LessonID: "lesson-id1",
		}).Once().Return(&payloads.GetLessonDetailResponse{
			Lesson: &lesson_domain.Lesson{
				LessonID:         "lesson-id1",
				Name:             "lesson-name",
				StartTime:        now,
				EndTime:          now.Add(1 * 24 * time.Hour),
				LocationID:       "location-id1",
				LocationName:     "location-name",
				CourseID:         "course-id1",
				CourseName:       "course-name",
				ClassID:          "class-id1",
				ClassName:        "class-name",
				SchedulingStatus: lesson_domain.LessonSchedulingStatusDraft,
				TeachingMedium:   lesson_domain.LessonTeachingMediumOnline,
				TeachingMethod:   lesson_domain.LessonTeachingMethodGroup,
				IsLocked:         false,
				Learners: []*lesson_domain.LessonLearner{
					{
						LearnerID:        "student-id1",
						LearnerName:      "student-name",
						CourseID:         "course-id1",
						CourseName:       "course-name",
						Grade:            "grade",
						AttendStatus:     lesson_domain.StudentAttendStatusAbsent,
						AttendanceNotice: lesson_domain.OnTheDay,
						AttendanceReason: lesson_domain.FamilyReason,
						AttendanceNote:   "sample-note",
					},
					{
						LearnerID:        "student-id2",
						LearnerName:      "student-name",
						CourseID:         "course-id1",
						CourseName:       "course-name",
						Grade:            "grade",
						AttendStatus:     lesson_domain.StudentAttendStatusAbsent,
						AttendanceNotice: lesson_domain.OnTheDay,
						AttendanceReason: lesson_domain.FamilyReason,
						AttendanceNote:   "sample-note",
					},
				},
				Teachers: []*lesson_domain.LessonTeacher{
					{
						TeacherID: "teacher-id1",
						Name:      "teacher-name",
					},
					{
						TeacherID: "teacher-id2",
						Name:      "teacher-name",
					},
				},
				Classrooms: []*lesson_domain.LessonClassroom{
					{
						ClassroomID:   "classroom-id1",
						ClassroomName: "classroom-name",
					},
					{
						ClassroomID:   "classroom-id2",
						ClassroomName: "classroom-name",
					},
				},
				Material: &lesson_domain.LessonMaterial{
					MediaIDs: []string{"media-id1, media-id2, media-id3"},
				},
				ZoomLink:    "zoom-link",
				ZoomID:      "zoom-id1",
				ZoomOwnerID: "zoom-owner-id1",
			},
			Scheduler: &dto.Scheduler{
				SchedulerID: "scheduler-id1",
				StartDate:   now,
				EndDate:     now.Add(1 * 24 * time.Hour),
				Frequency:   string(constants.FrequencyOnce),
			},
		}, nil)

		res, err := service.GetLessonDetailOnCalendar(context.Background(), req)
		require.NotNil(t, res)
		require.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})
	t.Run("failed", func(t *testing.T) {
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		lessonQueryHandler.On("GetLessonDetail", mock.Anything, mockDB.DB, &payloads.GetLessonDetailRequest{
			LessonID: "lesson-id1",
		}).Once().Return(nil, errors.New("error"))

		res, err := service.GetLessonDetailOnCalendar(context.Background(), req)
		require.Nil(t, res)
		require.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})
}

func TestLessonReaderService_GetLessonIDsForBulkStatusUpdate(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(mockDB.DB, mockDB.DB, mockUnleashClient, "local")
	lessonQueryHandler := &mock_queries.MockLessonQueryHandler{}
	service := &LessonReaderService{
		wrapperConnection:  wrapperConnection,
		lessonQueryHandler: lessonQueryHandler,
	}

	now := time.Now()
	reqPb := &cpb.GetLessonIDsForBulkStatusUpdateRequest{
		LocationId: "location-id1",
		Action:     lpb.LessonBulkAction_LESSON_BULK_ACTION_PUBLISH,
		StartDate:  timestamppb.New(now),
		EndDate:    timestamppb.New(now.Add(12 * 24 * time.Hour)),
		StartTime:  timestamppb.New(now),
		EndTime:    timestamppb.New(now.Add(4 * time.Hour)),
		Timezone:   "timezone",
	}

	req := &payloads.GetLessonIDsForBulkStatusUpdateRequest{
		LocationID: reqPb.LocationId,
		Action:     lesson_domain.LessonBulkActionPublish,
		StartDate:  reqPb.StartDate.AsTime(),
		EndDate:    reqPb.EndDate.AsTime(),
		StartTime:  reqPb.StartDate.AsTime(),
		EndTime:    reqPb.EndTime.AsTime(),
		Timezone:   "timezone",
	}

	t.Run("success", func(t *testing.T) {
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		lessonQueryHandler.On("GetLessonIDsForBulkStatusUpdate", mock.Anything, mockDB.DB, req).Once().
			Return([]*payloads.GetLessonIDsForBulkStatusUpdateResponse{
				{
					LessonStatus:           lesson_domain.LessonSchedulingStatusDraft,
					ModifiableLessonsCount: uint32(2),
					LessonsCount:           uint32(10),
					LessonIDs:              []string{"test1", "test2"},
				},
			}, nil)

		res, err := service.GetLessonIDsForBulkStatusUpdate(context.Background(), reqPb)
		require.NotNil(t, res)
		require.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})
	t.Run("failed", func(t *testing.T) {
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		lessonQueryHandler.On("GetLessonIDsForBulkStatusUpdate", mock.Anything, mockDB.DB, req).Once().Return(nil, errors.New("error"))

		res, err := service.GetLessonIDsForBulkStatusUpdate(context.Background(), reqPb)
		require.Nil(t, res)
		require.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})
}
