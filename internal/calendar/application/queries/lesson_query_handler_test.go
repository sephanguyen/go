package queries

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/calendar/application/queries/payloads"
	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	lesson_payloads "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries/payloads"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	mock_repositories "github.com/manabie-com/backend/mock/calendar/repositories"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLessonQueryHandler_GetLessonDetail(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	mockLessonRepo := &mock_repositories.MockLessonRepo{}
	mockLessonTeacherRepo := &mock_repositories.MockLessonTeacherRepo{}
	mockLessonMemeberRepo := &mock_repositories.MockLessonMemberRepo{}
	mockLessonClassroomRepo := &mock_repositories.MockLessonClassroomRepo{}
	mockLessonGroupRepo := &mock_repositories.MockLessonGroupRepo{}
	mockSchedulerRepo := &mock_repositories.MockSchedulerRepo{}
	mockUserRepo := &mock_repositories.MockUserRepo{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	request := &payloads.GetLessonDetailRequest{
		LessonID: "lesson-id-1",
	}
	lessonIDs := []string{request.LessonID}
	now := time.Now()
	mediaArgs := &lesson_domain.ListMediaByLessonArgs{
		LessonID: request.LessonID,
		Limit:    50,
	}

	testCases := []struct {
		name         string
		req          *payloads.GetLessonDetailRequest
		expectedResp *payloads.GetLessonDetailResponse
		setup        func(context.Context)
		hasError     bool
	}{
		{
			name: "success",
			req:  request,
			expectedResp: &payloads.GetLessonDetailResponse{
				Lesson: &lesson_domain.Lesson{
					LessonID:         "lesson-id-1",
					Name:             "Lesson Name",
					LocationID:       "location-id-1",
					LocationName:     "Location Name",
					StartTime:        now,
					EndTime:          now,
					TeachingMethod:   lesson_domain.LessonTeachingMethodIndividual,
					TeachingMedium:   lesson_domain.LessonTeachingMediumOnline,
					SchedulingStatus: lesson_domain.LessonSchedulingStatusCompleted,
					ClassID:          "class-id-1",
					ClassName:        "Class Name",
					CourseID:         "course-id-1",
					CourseName:       "Course Name",
					IsLocked:         false,
					SchedulerID:      "scheduler-id-1",
					Teachers: lesson_domain.LessonTeachers{
						{TeacherID: "teacher-1", Name: "teacher-name-1"},
						{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
					Learners: lesson_domain.LessonLearners{
						{
							LearnerID:        "student-id-1",
							LearnerName:      "student-name 1",
							CourseID:         "course-1",
							CourseName:       "course-name-1",
							Grade:            "Grade 5",
							AttendStatus:     lesson_domain.StudentAttendStatusAbsent,
							AttendanceNotice: lesson_domain.OnTheDay,
							AttendanceReason: lesson_domain.FamilyReason,
							AttendanceNote:   "sample-note",
						},
						{
							LearnerID:        "student-id-2",
							LearnerName:      "student-name 2",
							CourseID:         "course-2",
							CourseName:       "course-name-2",
							Grade:            "Grade 5",
							AttendStatus:     lesson_domain.StudentAttendStatusAbsent,
							AttendanceNotice: lesson_domain.OnTheDay,
							AttendanceReason: lesson_domain.FamilyReason,
							AttendanceNote:   "sample-note",
						},
						{
							LearnerID:        "student-id-3",
							LearnerName:      "student-name 3",
							CourseID:         "course-1",
							CourseName:       "course-name-1",
							Grade:            "Grade 6",
							AttendStatus:     lesson_domain.StudentAttendStatusAbsent,
							AttendanceNotice: lesson_domain.OnTheDay,
							AttendanceReason: lesson_domain.FamilyReason,
							AttendanceNote:   "sample-note",
						},
					},
					Classrooms: lesson_domain.LessonClassrooms{
						{
							ClassroomID:   "classroom-id-1",
							ClassroomName: "classroom-name-1",
						},
						{
							ClassroomID:   "classroom-id-2",
							ClassroomName: "classroom-name-2",
						},
					},
					Material: &lesson_domain.LessonMaterial{
						MediaIDs: []string{"media-id-1", "media-id-2", "media-id-3"},
					},
					ZoomLink:    "zoom-link",
					ZoomID:      "zoom-id1",
					ZoomOwnerID: "zoom-owner-id1",
				},
				Scheduler: &dto.Scheduler{
					SchedulerID: "scheduler-id-1",
					StartDate:   now,
					EndDate:     now,
					Frequency:   string(constants.FrequencyOnce),
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).Return(false, nil).Once()

				mockLessonRepo.On("GetLessonWithNamesByID", mock.Anything, mockDB.DB, request.LessonID).Once().Return(&lesson_domain.Lesson{
					LessonID:         "lesson-id-1",
					Name:             "Lesson Name",
					LocationID:       "location-id-1",
					LocationName:     "Location Name",
					StartTime:        now,
					EndTime:          now,
					TeachingMethod:   lesson_domain.LessonTeachingMethodIndividual,
					TeachingMedium:   lesson_domain.LessonTeachingMediumOnline,
					SchedulingStatus: lesson_domain.LessonSchedulingStatusCompleted,
					ClassID:          "class-id-1",
					ClassName:        "Class Name",
					CourseID:         "course-id-1",
					CourseName:       "Course Name",
					IsLocked:         false,
					SchedulerID:      "scheduler-id-1",
					ZoomLink:         "zoom-link",
					ZoomID:           "zoom-id1",
					ZoomOwnerID:      "zoom-owner-id1",
				}, nil)

				mockLessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mockDB.DB, lessonIDs, false).Once().Return(map[string]lesson_domain.LessonTeachers{
					"lesson-id-1": {
						&lesson_domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&lesson_domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
				}, nil)

				mockLessonMemeberRepo.On("GetLessonLearnersWithCourseAndNamesByLessonIDs", mock.Anything, mockDB.DB, lessonIDs, false).Once().Return(map[string]lesson_domain.LessonLearners{
					"lesson-id-1": {
						&lesson_domain.LessonLearner{
							LearnerID:        "student-id-1",
							LearnerName:      "student-name 1",
							CourseID:         "course-1",
							CourseName:       "course-name-1",
							AttendStatus:     lesson_domain.StudentAttendStatusAbsent,
							AttendanceNotice: lesson_domain.OnTheDay,
							AttendanceReason: lesson_domain.FamilyReason,
							AttendanceNote:   "sample-note",
						},
						&lesson_domain.LessonLearner{
							LearnerID:        "student-id-2",
							LearnerName:      "student-name 2",
							CourseID:         "course-2",
							CourseName:       "course-name-2",
							AttendStatus:     lesson_domain.StudentAttendStatusAbsent,
							AttendanceNotice: lesson_domain.OnTheDay,
							AttendanceReason: lesson_domain.FamilyReason,
							AttendanceNote:   "sample-note",
						},
						&lesson_domain.LessonLearner{
							LearnerID:        "student-id-3",
							LearnerName:      "student-name 3",
							CourseID:         "course-1",
							CourseName:       "course-name-1",
							AttendStatus:     lesson_domain.StudentAttendStatusAbsent,
							AttendanceNotice: lesson_domain.OnTheDay,
							AttendanceReason: lesson_domain.FamilyReason,
							AttendanceNote:   "sample-note",
						},
					},
				}, nil)

				mockUserRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mockDB.DB, []string{"student-id-1", "student-id-2", "student-id-3"}, false).Once().Return(map[string]string{
					"student-id-1": "Grade 5",
					"student-id-2": "Grade 5",
					"student-id-3": "Grade 6",
				}, nil)

				mockLessonClassroomRepo.On("GetLessonClassroomsWithNamesByLessonIDs", mock.Anything, mockDB.DB, lessonIDs).Once().Return(map[string]lesson_domain.LessonClassrooms{
					"lesson-id-1": {
						&lesson_domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&lesson_domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
				}, nil)

				mockLessonGroupRepo.On("ListMediaByLessonArgs", mock.Anything, mockDB.DB, mediaArgs).Once().Return(media_domain.Medias{
					&media_domain.Media{ID: "media-id-1"},
					&media_domain.Media{ID: "media-id-2"},
					&media_domain.Media{ID: "media-id-3"},
				}, nil)

				mockSchedulerRepo.On("GetByID", mock.Anything, mockDB.DB, "scheduler-id-1").Once().Return(&dto.Scheduler{
					SchedulerID: "scheduler-id-1",
					StartDate:   now,
					EndDate:     now,
					Frequency:   string(constants.FrequencyOnce),
				}, nil)
			},
			hasError: false,
		},
		{
			name: "failed to get lesson",
			req: &payloads.GetLessonDetailRequest{
				LessonID: "lesson-id-1",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).Return(false, nil).Once()

				mockLessonRepo.On("GetLessonWithNamesByID", mock.Anything, mockDB.DB, request.LessonID).Once().Return(nil, fmt.Errorf("error"))
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			lessonQueryHandler := &LessonQueryHandler{
				LessonRepo:          mockLessonRepo,
				LessonTeacherRepo:   mockLessonTeacherRepo,
				LessonMemberRepo:    mockLessonMemeberRepo,
				LessonClassroomRepo: mockLessonClassroomRepo,
				LessonGroupRepo:     mockLessonGroupRepo,
				SchedulerRepo:       mockSchedulerRepo,
				UserRepo:            mockUserRepo,
				Env:                 "local",
				UnleashClient:       mockUnleashClient,
			}

			resp, err := lessonQueryHandler.GetLessonDetail(ctx, mockDB.DB, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, resp)
				require.Equal(t, tc.expectedResp, resp)
			}
		})
	}
}

func TestLessonQueryHandler_GetLessonIDsForBulkStatusUpdate(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	mockLessonRepo := &mock_repositories.MockLessonRepo{}
	mockLessonTeacherRepo := &mock_repositories.MockLessonTeacherRepo{}

	now := time.Now()
	requestPublish := &payloads.GetLessonIDsForBulkStatusUpdateRequest{
		LocationID: "location-id1",
		Action:     lesson_domain.LessonBulkActionPublish,
		StartDate:  now,
		EndDate:    now.Add(12 * 24 * time.Hour),
		StartTime:  now,
		EndTime:    now.Add(4 * time.Hour),
		Timezone:   "timezone",
	}

	requestCancel := &payloads.GetLessonIDsForBulkStatusUpdateRequest{
		LocationID: "location-id1",
		Action:     lesson_domain.LessonBulkActionCancel,
		StartDate:  now,
		EndDate:    now.Add(12 * 24 * time.Hour),
		StartTime:  now,
		EndTime:    now.Add(4 * time.Hour),
		Timezone:   "timezone",
	}

	lessonIDs := []string{
		"lesson-id1", "lesson-id2",
	}

	testCases := []struct {
		name         string
		req          *payloads.GetLessonIDsForBulkStatusUpdateRequest
		expectedResp []*payloads.GetLessonIDsForBulkStatusUpdateResponse
		setup        func(context.Context)
		hasError     bool
	}{
		{
			name: "success with bulk cancel",
			req:  requestCancel,
			expectedResp: []*payloads.GetLessonIDsForBulkStatusUpdateResponse{
				{
					LessonStatus:           lesson_domain.LessonSchedulingStatusCompleted,
					ModifiableLessonsCount: uint32(2),
					LessonsCount:           uint32(2),
					LessonIDs:              lessonIDs,
				},
				{
					LessonStatus:           lesson_domain.LessonSchedulingStatusPublished,
					ModifiableLessonsCount: uint32(2),
					LessonsCount:           uint32(2),
					LessonIDs:              lessonIDs,
				},
			},
			setup: func(ctx context.Context) {
				mockLessonRepo.On("GetLessonsByLocationStatusAndDateTimeRange", mock.Anything, mockDB.DB, &lesson_payloads.GetLessonsByLocationStatusAndDateTimeRangeArgs{
					LocationID:   requestPublish.LocationID,
					LessonStatus: lesson_domain.LessonSchedulingStatusCompleted,
					StartDate:    requestPublish.StartDate,
					EndDate:      requestPublish.EndDate,
					StartTime:    requestPublish.StartTime,
					EndTime:      requestPublish.EndTime,
					Timezone:     "timezone",
				}).Once().Return([]*lesson_domain.Lesson{
					{
						LessonID:         "lesson-id1",
						LocationID:       "location-id1",
						SchedulingStatus: lesson_domain.LessonSchedulingStatusCompleted,
					},
					{
						LessonID:         "lesson-id2",
						LocationID:       "location-id1",
						SchedulingStatus: lesson_domain.LessonSchedulingStatusCompleted,
					},
				}, nil)

				mockLessonRepo.On("GetLessonsByLocationStatusAndDateTimeRange", mock.Anything, mockDB.DB, &lesson_payloads.GetLessonsByLocationStatusAndDateTimeRangeArgs{
					LocationID:   requestPublish.LocationID,
					LessonStatus: lesson_domain.LessonSchedulingStatusPublished,
					StartDate:    requestPublish.StartDate,
					EndDate:      requestPublish.EndDate,
					StartTime:    requestPublish.StartTime,
					EndTime:      requestPublish.EndTime,
					Timezone:     "timezone",
				}).Once().Return([]*lesson_domain.Lesson{
					{
						LessonID:         "lesson-id1",
						LocationID:       "location-id1",
						SchedulingStatus: lesson_domain.LessonSchedulingStatusPublished,
					},
					{
						LessonID:         "lesson-id2",
						LocationID:       "location-id1",
						SchedulingStatus: lesson_domain.LessonSchedulingStatusPublished,
					},
				}, nil)
			},
			hasError: false,
		},
		{
			name: "success with bulk publish",
			req:  requestPublish,
			expectedResp: []*payloads.GetLessonIDsForBulkStatusUpdateResponse{
				{
					LessonStatus:           lesson_domain.LessonSchedulingStatusDraft,
					ModifiableLessonsCount: uint32(2),
					LessonsCount:           uint32(3),
					LessonIDs:              lessonIDs,
				},
			},
			setup: func(ctx context.Context) {
				mockLessonRepo.On("GetLessonsByLocationStatusAndDateTimeRange", mock.Anything, mockDB.DB, &lesson_payloads.GetLessonsByLocationStatusAndDateTimeRangeArgs{
					LocationID:   requestPublish.LocationID,
					LessonStatus: lesson_domain.LessonSchedulingStatusDraft,
					StartDate:    requestPublish.StartDate,
					EndDate:      requestPublish.EndDate,
					StartTime:    requestPublish.StartTime,
					EndTime:      requestPublish.EndTime,
					Timezone:     "timezone",
				}).Once().Return([]*lesson_domain.Lesson{
					{
						LessonID:         "lesson-id1",
						LocationID:       "location-id1",
						TeachingMethod:   lesson_domain.LessonTeachingMethodIndividual,
						SchedulingStatus: lesson_domain.LessonSchedulingStatusDraft,
						CourseID:         "course-id1",
					},
					{
						LessonID:         "lesson-id2",
						LocationID:       "location-id1",
						TeachingMethod:   lesson_domain.LessonTeachingMethodGroup,
						SchedulingStatus: lesson_domain.LessonSchedulingStatusDraft,
						CourseID:         "course-id2",
					},
					{
						LessonID:         "lesson-id3",
						LocationID:       "location-id1",
						TeachingMethod:   lesson_domain.LessonTeachingMethodIndividual,
						SchedulingStatus: lesson_domain.LessonSchedulingStatusDraft,
						CourseID:         "course-id3",
					},
				}, nil)

				mockLessonTeacherRepo.On("GetTeachersByLessonIDs", mock.Anything, mockDB.DB, append(lessonIDs, "lesson-id3")).Once().Return(map[string]lesson_domain.LessonTeachers{
					"lesson-id1": {
						&lesson_domain.LessonTeacher{TeacherID: "teacher-1"},
					},
					"lesson-id2": {
						&lesson_domain.LessonTeacher{TeacherID: "teacher-1"},
					},
				}, nil)
			},
			hasError: false,
		},
		{
			name: "failed with bulk cancel",
			req:  requestCancel,
			setup: func(ctx context.Context) {
				mockLessonRepo.On("GetLessonsByLocationStatusAndDateTimeRange", mock.Anything, mockDB.DB, &lesson_payloads.GetLessonsByLocationStatusAndDateTimeRangeArgs{
					LocationID:   requestPublish.LocationID,
					LessonStatus: lesson_domain.LessonSchedulingStatusCompleted,
					StartDate:    requestPublish.StartDate,
					EndDate:      requestPublish.EndDate,
					StartTime:    requestPublish.StartTime,
					EndTime:      requestPublish.EndTime,
					Timezone:     "timezone",
				}).Once().Return(nil, fmt.Errorf("error"))
			},
			hasError: true,
		},
		{
			name: "failed with bulk publish",
			req:  requestPublish,
			setup: func(ctx context.Context) {
				mockLessonRepo.On("GetLessonsByLocationStatusAndDateTimeRange", mock.Anything, mockDB.DB, &lesson_payloads.GetLessonsByLocationStatusAndDateTimeRangeArgs{
					LocationID:   requestPublish.LocationID,
					LessonStatus: lesson_domain.LessonSchedulingStatusDraft,
					StartDate:    requestPublish.StartDate,
					EndDate:      requestPublish.EndDate,
					StartTime:    requestPublish.StartTime,
					EndTime:      requestPublish.EndTime,
					Timezone:     "timezone",
				}).Once().Return(nil, fmt.Errorf("error"))
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			lessonQueryHandler := &LessonQueryHandler{
				LessonRepo:        mockLessonRepo,
				LessonTeacherRepo: mockLessonTeacherRepo,
			}

			resp, err := lessonQueryHandler.GetLessonIDsForBulkStatusUpdate(ctx, mockDB.DB, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, resp)
				require.Equal(t, tc.expectedResp, resp)
			}
		})
	}
}
