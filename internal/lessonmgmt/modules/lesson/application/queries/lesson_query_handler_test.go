package queries

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_user_repo "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name     string
	ctx      context.Context
	payloads interface{}
	result   interface{}
	setup    func(ctx context.Context)
}

func TestLessonManagementService_RetrieveLessons(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	now := time.Now().UTC()
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := new(mock_repositories.MockLessonRepo)
	lessonTeacherRepo := new(mock_repositories.MockLessonTeacherRepo)

	s := &LessonQueryHandler{
		WrapperConnection: wrapperConnection,
		LessonRepo:        lessonRepo,
		LessonTeacherRepo: lessonTeacherRepo,
	}

	courses := []string{"course-1"}
	teachers := []string{"teacher-1"}
	students := []string{"student-1"}
	centers := []string{"center-1"}

	retrieveLessonError := fmt.Errorf("fail retrieve")
	testCases := []TestCase{
		{
			name: "School Admin get list lessons future successfully with filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetLessonListArg{
				Limit:       2,
				LessonID:    "",
				CurrentTime: now,
				Compare:     ">=",
				LessonTime:  "future",
				CourseIDs:   courses,
				TeacherIDs:  teachers,
				StudentIDs:  students,
				FromDate:    now.Add(-30 * time.Hour),
				ToDate:      now.Add(30 * time.Hour),
				FromTime:    "02:00:00",
				ToTime:      "10:00:00",
				KeyWord:     "Lesson Name",
				LocationIDs: centers,
				Dow:         []domain.DateOfWeek{0, 1, 2, 3, 4, 5, 6},
				Grades:      []int32{5, 6},
				TimeZone:    "UTC",
			},
			result: &RetrieveLessonsResponse{
				Lessons: []*domain.Lesson{
					{
						LessonID:       "lesson-1",
						Name:           "Lesson Name",
						LocationID:     "center-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodGroup,
						TeachingMedium: domain.LessonTeachingMediumOnline,
						Teachers: domain.LessonTeachers{
							{TeacherID: "teacher-1"},
							{TeacherID: "teacher-2"},
						},
					},
					{
						LessonID:       "lesson-2",
						Name:           "Lesson Name",
						LocationID:     "center-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodIndividual,
						TeachingMedium: domain.LessonTeachingMediumOffline,
						Teachers: domain.LessonTeachers{
							{TeacherID: "teacher-1"},
						},
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
				lessonRepo.On("Retrieve", mock.Anything, db, &payloads.GetLessonListArg{
					Limit:       2,
					LessonID:    "",
					CurrentTime: now,
					Compare:     ">=",
					LessonTime:  "future",
					CourseIDs:   courses,
					TeacherIDs:  teachers,
					StudentIDs:  students,
					FromDate:    now.Add(-30 * time.Hour),
					ToDate:      now.Add(30 * time.Hour),
					FromTime:    "02:00:00",
					ToTime:      "10:00:00",
					KeyWord:     "Lesson Name",
					LocationIDs: centers,
					Dow:         []domain.DateOfWeek{0, 1, 2, 3, 4, 5, 6},
					Grades:      []int32{5, 6},
					TimeZone:    "UTC",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:       "lesson-1",
						Name:           "Lesson Name",
						LocationID:     "center-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodGroup,
						TeachingMedium: domain.LessonTeachingMediumOnline,
					},
					{
						LessonID:       "lesson-2",
						Name:           "Lesson Name",
						LocationID:     "center-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodIndividual,
						TeachingMedium: domain.LessonTeachingMediumOffline,
					},
				}, uint32(99), "pre_id", uint32(2), nil)

				lessonTeacherRepo.On("GetTeachersByLessonIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1"},
					},
				}, nil)
			},
		},
		{
			name:     "Retrieve fail with getting lesson has error",
			ctx:      interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetLessonListArg{},
			result: &RetrieveLessonsResponse{
				Error: status.Error(codes.Internal, retrieveLessonError.Error()),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("Retrieve", mock.Anything, db, mock.Anything).Once().Return(nil, uint32(0), "", uint32(0), retrieveLessonError)
			},
		},
		{
			name: "Retrieve fail with get lesson teacher has error",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetLessonListArg{
				Limit:       2,
				LessonID:    "",
				CurrentTime: now,
				Compare:     ">=",
				LessonTime:  "future",
				CourseIDs:   courses,
				TeacherIDs:  teachers,
				StudentIDs:  students,
				FromDate:    now.Add(-30 * time.Hour),
				ToDate:      now.Add(30 * time.Hour),
				FromTime:    "02:00:00",
				ToTime:      "10:00:00",
				KeyWord:     "Lesson Name",
				LocationIDs: centers,
				Dow:         []domain.DateOfWeek{0, 1, 2, 3, 4, 5, 6},
				Grades:      []int32{5, 6},
				TimeZone:    "UTC",
			},
			result: &RetrieveLessonsResponse{
				Error: status.Error(codes.Internal, fmt.Errorf("LessonRepo.GetTeacherIDsByLessonIDs: %w", retrieveLessonError).Error()),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("Retrieve", mock.Anything, db, mock.Anything).Once().Return([]*domain.Lesson{
					{
						LessonID:       "lesson-1",
						Name:           "Lesson Name",
						LocationID:     "center-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodGroup,
						TeachingMedium: domain.LessonTeachingMediumOnline,
					},
					{
						LessonID:       "lesson-2",
						Name:           "Lesson Name",
						LocationID:     "center-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodIndividual,
						TeachingMedium: domain.LessonTeachingMediumOffline,
					},
				}, uint32(99), "pre_id", uint32(2), nil)

				lessonTeacherRepo.On("GetTeachersByLessonIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]domain.LessonTeachers{}, retrieveLessonError)
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp := s.RetrieveLesson(testCase.ctx, testCase.payloads.(*payloads.GetLessonListArg))
			expectedErr := testCase.result.(*RetrieveLessonsResponse).Error
			if expectedErr != nil {
				assert.Error(t, resp.Error)
				assert.Equal(t, expectedErr.Error(), resp.Error.Error())
			} else {
				assert.NoError(t, resp.Error)
				assert.Equal(t, testCase.result, resp)
			}

			mock.AssertExpectationsForObjects(t, lessonRepo, lessonTeacherRepo, mockUnleashClient)
		})
	}
}

func TestLessonManagementService_RetrieveLessonsOnCalendar(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	now := time.Now().UTC()
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := new(mock_repositories.MockLessonRepo)
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	lessonTeacherRepo := new(mock_repositories.MockLessonTeacherRepo)
	lessonClassroomRepo := new(mock_repositories.MockLessonClassroomRepo)
	userRepo := new(mock_user_repo.MockUserRepo)
	lessonIDs := []string{"lesson-1", "lesson-2", "lesson-3"}

	lessonQueryHandler := &LessonQueryHandler{
		WrapperConnection:   wrapperConnection,
		LessonRepo:          lessonRepo,
		LessonMemberRepo:    lessonMemberRepo,
		LessonTeacherRepo:   lessonTeacherRepo,
		LessonClassroomRepo: lessonClassroomRepo,
		UserRepo:            userRepo,
	}

	testCases := []TestCase{
		{
			name: "user successfully retrieves lessons on calendar",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetLessonListOnCalendarArgs{
				View:       payloads.Weekly,
				FromDate:   now,
				ToDate:     now.Add(7 * 24 * time.Hour),
				LocationID: "location-id-1",
				Timezone:   "sample-timezone",
			},
			result: &RetrieveLessonsResponse{
				Lessons: []*domain.Lesson{
					{
						LessonID:       "lesson-1",
						Name:           "Lesson Name 1",
						LocationID:     "location-id-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodIndividual,
						TeachingMedium: domain.LessonTeachingMediumOnline,
						Teachers: domain.LessonTeachers{
							{TeacherID: "teacher-1", Name: "teacher-name-1"},
							{TeacherID: "teacher-2", Name: "teacher-name-2"},
						},
						Learners: []*domain.LessonLearner{
							{
								LearnerID:   "student-id-1",
								CourseID:    "course-1",
								Grade:       "Grade 5",
								LearnerName: "student-name 1",
								CourseName:  "course-name-1",
							},
							{
								LearnerID:   "student-id-2",
								CourseID:    "course-2",
								Grade:       "Grade 5",
								LearnerName: "student-name 2",
								CourseName:  "course-name-2",
							},
							{
								LearnerID:   "student-id-3",
								CourseID:    "course-1",
								Grade:       "Grade 6",
								LearnerName: "student-name 3",
								CourseName:  "course-name-1",
							},
						},
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						Classrooms: domain.LessonClassrooms{
							{
								ClassroomID:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomID:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
						SchedulerID: "scheduler-id-1",
					},
					{
						LessonID:       "lesson-2",
						Name:           "Lesson Name 2",
						LocationID:     "location-id-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodIndividual,
						TeachingMedium: domain.LessonTeachingMediumOnline,
						Teachers: domain.LessonTeachers{
							{TeacherID: "teacher-1", Name: "teacher-name-1"},
						},
						Learners: []*domain.LessonLearner{
							{
								LearnerID:   "student-id-4",
								CourseID:    "course-2",
								Grade:       "Grade 4",
								LearnerName: "student-name 4",
								CourseName:  "course-name-2",
							},
							{
								LearnerID:   "student-id-5",
								CourseID:    "course-2",
								Grade:       "Grade 5",
								LearnerName: "student-name 5",
								CourseName:  "course-name-2",
							},
						},
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						Classrooms: domain.LessonClassrooms{
							{
								ClassroomID:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomID:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
						SchedulerID: "scheduler-id-2",
					},
					{
						LessonID:       "lesson-3",
						Name:           "Lesson Name 3",
						LocationID:     "location-id-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodGroup,
						TeachingMedium: domain.LessonTeachingMediumOnline,
						Teachers: domain.LessonTeachers{
							{TeacherID: "teacher-1", Name: "teacher-name-1"},
							{TeacherID: "teacher-2", Name: "teacher-name-2"},
						},
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "class-name-1",
						CourseName:       "course-name-2",
						Classrooms: domain.LessonClassrooms{
							{
								ClassroomID:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomID:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
					},
				},
				Error: nil,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "class-name-1",
						CourseName:       "course-name-2",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
					},
					"lesson-3": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
				}, nil)

				lessonMemberRepo.On("GetLessonLearnersWithCourseAndNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonLearners{
					"lesson-1": {
						&domain.LessonLearner{
							LearnerID:   "student-id-1",
							CourseID:    "course-1",
							LearnerName: "student-name 1",
							CourseName:  "course-name-1",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-2",
							CourseID:    "course-2",
							LearnerName: "student-name 2",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-3",
							CourseID:    "course-1",
							LearnerName: "student-name 3",
							CourseName:  "course-name-1",
						},
					},
					"lesson-2": {
						&domain.LessonLearner{
							LearnerID:   "student-id-4",
							CourseID:    "course-2",
							LearnerName: "student-name 4",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-5",
							CourseID:    "course-2",
							LearnerName: "student-name 5",
							CourseName:  "course-name-2",
						},
					},
				}, nil)

				lessonClassroomRepo.On("GetLessonClassroomsWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs).Once().Return(map[string]domain.LessonClassrooms{
					"lesson-1": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-2": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-3": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-1", "student-id-2", "student-id-3"}).Once().Return(map[string]string{
					"student-id-1": "Grade 5",
					"student-id-2": "Grade 5",
					"student-id-3": "Grade 6",
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-4", "student-id-5"}).Once().Return(map[string]string{
					"student-id-4": "Grade 4",
					"student-id-5": "Grade 5",
				}, nil)
			},
		},
		{
			name: "user successfully retrieves lessons on calendar with filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetLessonListOnCalendarArgs{
				View:       payloads.Weekly,
				FromDate:   now,
				ToDate:     now.Add(7 * 24 * time.Hour),
				LocationID: "location-id-1",
				Timezone:   "sample-timezone",
				StudentIDs: []string{"student-id-1", "student-id-2", "student-id-3", "student-id-4", "student-id-5"},
				CourseIDs:  []string{"course-1", "course-2"},
				TeacherIDs: []string{"teacher-1", "teacher-2"},
				ClassIDs:   []string{"class-1"},
			},
			result: &RetrieveLessonsResponse{
				Lessons: []*domain.Lesson{
					{
						LessonID:       "lesson-1",
						Name:           "Lesson Name 1",
						LocationID:     "location-id-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodIndividual,
						TeachingMedium: domain.LessonTeachingMediumOnline,
						Teachers: domain.LessonTeachers{
							{TeacherID: "teacher-1", Name: "teacher-name-1"},
							{TeacherID: "teacher-2", Name: "teacher-name-2"},
						},
						Learners: []*domain.LessonLearner{
							{
								LearnerID:   "student-id-1",
								CourseID:    "course-1",
								Grade:       "Grade 5",
								LearnerName: "student-name 1",
								CourseName:  "course-name-1",
							},
							{
								LearnerID:   "student-id-2",
								CourseID:    "course-2",
								Grade:       "Grade 5",
								LearnerName: "student-name 2",
								CourseName:  "course-name-2",
							},
							{
								LearnerID:   "student-id-3",
								CourseID:    "course-1",
								Grade:       "Grade 6",
								LearnerName: "student-name 3",
								CourseName:  "course-name-1",
							},
						},
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						Classrooms: domain.LessonClassrooms{
							{
								ClassroomID:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomID:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
						SchedulerID: "scheduler-id-1",
					},
					{
						LessonID:       "lesson-2",
						Name:           "Lesson Name 2",
						LocationID:     "location-id-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodIndividual,
						TeachingMedium: domain.LessonTeachingMediumOnline,
						Teachers: domain.LessonTeachers{
							{TeacherID: "teacher-1", Name: "teacher-name-1"},
						},
						Learners: []*domain.LessonLearner{
							{
								LearnerID:   "student-id-4",
								CourseID:    "course-2",
								Grade:       "Grade 4",
								LearnerName: "student-name 4",
								CourseName:  "course-name-2",
							},
							{
								LearnerID:   "student-id-5",
								CourseID:    "course-2",
								Grade:       "Grade 5",
								LearnerName: "student-name 5",
								CourseName:  "course-name-2",
							},
						},
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						Classrooms: domain.LessonClassrooms{
							{
								ClassroomID:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomID:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
						SchedulerID: "scheduler-id-2",
					},
					{
						LessonID:       "lesson-3",
						Name:           "Lesson Name 3",
						LocationID:     "location-id-1",
						StartTime:      now,
						EndTime:        now,
						TeachingMethod: domain.LessonTeachingMethodGroup,
						TeachingMedium: domain.LessonTeachingMediumOnline,
						Teachers: domain.LessonTeachers{
							{TeacherID: "teacher-1", Name: "teacher-name-1"},
							{TeacherID: "teacher-2", Name: "teacher-name-2"},
						},
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "class-name-1",
						CourseName:       "course-name-2",
						Classrooms: domain.LessonClassrooms{
							{
								ClassroomID:   "classroom-id-1",
								ClassroomName: "classroom-name-1",
							},
							{
								ClassroomID:   "classroom-id-2",
								ClassroomName: "classroom-name-2",
							},
						},
					},
				},
				Error: nil,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
					StudentIDs: []string{"student-id-1", "student-id-2", "student-id-3", "student-id-4", "student-id-5"},
					CourseIDs:  []string{"course-1", "course-2"},
					TeacherIDs: []string{"teacher-1", "teacher-2"},
					ClassIDs:   []string{"class-1"},
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "class-name-1",
						CourseName:       "course-name-2",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
					},
					"lesson-3": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
				}, nil)

				lessonMemberRepo.On("GetLessonLearnersWithCourseAndNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonLearners{
					"lesson-1": {
						&domain.LessonLearner{
							LearnerID:   "student-id-1",
							CourseID:    "course-1",
							LearnerName: "student-name 1",
							CourseName:  "course-name-1",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-2",
							CourseID:    "course-2",
							LearnerName: "student-name 2",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-3",
							CourseID:    "course-1",
							LearnerName: "student-name 3",
							CourseName:  "course-name-1",
						},
					},
					"lesson-2": {
						&domain.LessonLearner{
							LearnerID:   "student-id-4",
							CourseID:    "course-2",
							LearnerName: "student-name 4",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-5",
							CourseID:    "course-2",
							LearnerName: "student-name 5",
							CourseName:  "course-name-2",
						},
					},
				}, nil)

				lessonClassroomRepo.On("GetLessonClassroomsWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs).Once().Return(map[string]domain.LessonClassrooms{
					"lesson-1": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-2": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-3": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-1", "student-id-2", "student-id-3"}).Once().Return(map[string]string{
					"student-id-1": "Grade 5",
					"student-id-2": "Grade 5",
					"student-id-3": "Grade 6",
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-4", "student-id-5"}).Once().Return(map[string]string{
					"student-id-4": "Grade 4",
					"student-id-5": "Grade 5",
				}, nil)
			},
		},
		{
			name: "returns empty list",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetLessonListOnCalendarArgs{
				View:       payloads.Weekly,
				FromDate:   now,
				ToDate:     now.Add(7 * 24 * time.Hour),
				LocationID: "location-id-1",
				Timezone:   "sample-timezone",
			},
			result: &RetrieveLessonsResponse{
				Lessons: []*domain.Lesson{},
				Error:   nil,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, mock.Anything).Once().Return([]*domain.Lesson{}, nil)
			},
		},
		{
			name: "failed to fetch lessons on calendar from lesson repo",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetLessonListOnCalendarArgs{
				View:       payloads.Weekly,
				FromDate:   now,
				ToDate:     now.Add(7 * 24 * time.Hour),
				LocationID: "location-id-1",
				Timezone:   "sample-timezone",
			},
			result: &RetrieveLessonsResponse{
				Error: fmt.Errorf("rpc error: code = Internal desc = LessonRepo.GetLessonsOnCalendar: some-lesson-repo-error"),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return(nil, fmt.Errorf("some-lesson-repo-error"))
			},
		},
		{
			name: "failed to fetch teachers by lesson IDs from lesson teacher repo",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetLessonListOnCalendarArgs{
				View:       payloads.Weekly,
				FromDate:   now,
				ToDate:     now.Add(7 * 24 * time.Hour),
				LocationID: "location-id-1",
				Timezone:   "sample-timezone",
			},
			result: &RetrieveLessonsResponse{
				Error: fmt.Errorf("rpc error: code = Internal desc = LessonTeacherRepo.GetTeachersWithNamesByLessonIDs: some-lesson-teacher-repo-error"),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "class-name-1",
						CourseName:       "course-name-2",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().
					Return(map[string]domain.LessonTeachers{}, fmt.Errorf("some-lesson-teacher-repo-error"))
			},
		},
		{
			name: "failed to fetch lesson learners by lesson IDs from lesson member repo",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetLessonListOnCalendarArgs{
				View:       payloads.Weekly,
				FromDate:   now,
				ToDate:     now.Add(7 * 24 * time.Hour),
				LocationID: "location-id-1",
				Timezone:   "sample-timezone",
			},
			result: &RetrieveLessonsResponse{
				Error: fmt.Errorf("rpc error: code = Internal desc = LessonMemberRepo.GetLessonLearnersWithCourseAndNamesByLessonIDs: some-lesson-member-repo-error"),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "class-name-1",
						CourseName:       "course-name-2",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
					},
					"lesson-3": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
				}, nil)

				lessonMemberRepo.On("GetLessonLearnersWithCourseAndNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().
					Return(map[string]domain.LessonLearners{}, fmt.Errorf("some-lesson-member-repo-error"))
			},
		},
		{
			name: "failed to fetch lesson classrooms by lesson IDs from lesson classroom repo",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetLessonListOnCalendarArgs{
				View:       payloads.Weekly,
				FromDate:   now,
				ToDate:     now.Add(7 * 24 * time.Hour),
				LocationID: "location-id-1",
				Timezone:   "sample-timezone",
			},
			result: &RetrieveLessonsResponse{
				Error: fmt.Errorf("rpc error: code = Internal desc = LessonClassroomRepo.GetLessonClassroomsWithNamesByLessonIDs: some-lesson-classroom-repo-error"),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "class-name-1",
						CourseName:       "course-name-2",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
					},
					"lesson-3": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
				}, nil)

				lessonMemberRepo.On("GetLessonLearnersWithCourseAndNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonLearners{
					"lesson-1": {
						&domain.LessonLearner{
							LearnerID:   "student-id-1",
							CourseID:    "course-1",
							LearnerName: "student-name 1",
							CourseName:  "course-name-1",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-2",
							CourseID:    "course-2",
							LearnerName: "student-name 2",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-3",
							CourseID:    "course-1",
							LearnerName: "student-name 3",
							CourseName:  "course-name-1",
						},
					},
					"lesson-2": {
						&domain.LessonLearner{
							LearnerID:   "student-id-4",
							CourseID:    "course-2",
							LearnerName: "student-name 4",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-5",
							CourseID:    "course-2",
							LearnerName: "student-name 5",
							CourseName:  "course-name-2",
						},
					},
				}, nil)

				lessonClassroomRepo.On("GetLessonClassroomsWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs).Once().
					Return(map[string]domain.LessonClassrooms{}, fmt.Errorf("some-lesson-classroom-repo-error"))
			},
		},
		{
			name: "failed to fetch student grades of individual lessons from user repo",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetLessonListOnCalendarArgs{
				View:       payloads.Weekly,
				FromDate:   now,
				ToDate:     now.Add(7 * 24 * time.Hour),
				LocationID: "location-id-1",
				Timezone:   "sample-timezone",
			},
			result: &RetrieveLessonsResponse{
				Error: fmt.Errorf("rpc error: code = Internal desc = UserRepo.GetStudentCurrentGradeByUserIDs: some-user-repo-error"),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "class-name-1",
						CourseName:       "course-name-2",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
					},
					"lesson-3": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
				}, nil)

				lessonMemberRepo.On("GetLessonLearnersWithCourseAndNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonLearners{
					"lesson-1": {
						&domain.LessonLearner{
							LearnerID:   "student-id-1",
							CourseID:    "course-1",
							LearnerName: "student-name 1",
							CourseName:  "course-name-1",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-2",
							CourseID:    "course-2",
							LearnerName: "student-name 2",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-3",
							CourseID:    "course-1",
							LearnerName: "student-name 3",
							CourseName:  "course-name-1",
						},
					},
					"lesson-2": {
						&domain.LessonLearner{
							LearnerID:   "student-id-4",
							CourseID:    "course-2",
							LearnerName: "student-name 4",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-5",
							CourseID:    "course-2",
							LearnerName: "student-name 5",
							CourseName:  "course-name-2",
						},
					},
				}, nil)

				lessonClassroomRepo.On("GetLessonClassroomsWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs).Once().Return(map[string]domain.LessonClassrooms{
					"lesson-1": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-2": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-3": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-1", "student-id-2", "student-id-3"}).Once().
					Return(map[string]string{}, fmt.Errorf("some-user-repo-error"))
			},
		},
		{
			name: "failed to fetch student grades of individual lessons from user repo in second loop iteration",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			payloads: &payloads.GetLessonListOnCalendarArgs{
				View:       payloads.Weekly,
				FromDate:   now,
				ToDate:     now.Add(7 * 24 * time.Hour),
				LocationID: "location-id-1",
				Timezone:   "sample-timezone",
			},
			result: &RetrieveLessonsResponse{
				Error: fmt.Errorf("rpc error: code = Internal desc = UserRepo.GetStudentCurrentGradeByUserIDs: some-user-repo-error"),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsOnCalendar", mock.Anything, db, &payloads.GetLessonListOnCalendarArgs{
					View:       payloads.Weekly,
					FromDate:   now,
					ToDate:     now.Add(7 * 24 * time.Hour),
					LocationID: "location-id-1",
					Timezone:   "sample-timezone",
				}).Once().Return([]*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "Lesson Name 1",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-1",
					},
					{
						LessonID:         "lesson-2",
						Name:             "Lesson Name 2",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "",
						CourseID:         "",
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
						ClassName:        "",
						CourseName:       "",
						SchedulerID:      "scheduler-id-2",
					},
					{
						LessonID:         "lesson-3",
						Name:             "Lesson Name 3",
						LocationID:       "location-id-1",
						StartTime:        now,
						EndTime:          now,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						ClassID:          "class-1",
						CourseID:         "course-2",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
						ClassName:        "class-name-1",
						CourseName:       "course-name-2",
					},
				}, nil)

				lessonTeacherRepo.On("GetTeachersWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonTeachers{
					"lesson-1": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
					"lesson-2": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
					},
					"lesson-3": {
						&domain.LessonTeacher{TeacherID: "teacher-1", Name: "teacher-name-1"},
						&domain.LessonTeacher{TeacherID: "teacher-2", Name: "teacher-name-2"},
					},
				}, nil)

				lessonMemberRepo.On("GetLessonLearnersWithCourseAndNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs, true).Once().Return(map[string]domain.LessonLearners{
					"lesson-1": {
						&domain.LessonLearner{
							LearnerID:   "student-id-1",
							CourseID:    "course-1",
							LearnerName: "student-name 1",
							CourseName:  "course-name-1",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-2",
							CourseID:    "course-2",
							LearnerName: "student-name 2",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-3",
							CourseID:    "course-1",
							LearnerName: "student-name 3",
							CourseName:  "course-name-1",
						},
					},
					"lesson-2": {
						&domain.LessonLearner{
							LearnerID:   "student-id-4",
							CourseID:    "course-2",
							LearnerName: "student-name 4",
							CourseName:  "course-name-2",
						},
						&domain.LessonLearner{
							LearnerID:   "student-id-5",
							CourseID:    "course-2",
							LearnerName: "student-name 5",
							CourseName:  "course-name-2",
						},
					},
				}, nil)

				lessonClassroomRepo.On("GetLessonClassroomsWithNamesByLessonIDs", mock.Anything, mock.Anything, lessonIDs).Once().Return(map[string]domain.LessonClassrooms{
					"lesson-1": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-2": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
					"lesson-3": {
						&domain.LessonClassroom{ClassroomID: "classroom-id-1", ClassroomName: "classroom-name-1"},
						&domain.LessonClassroom{ClassroomID: "classroom-id-2", ClassroomName: "classroom-name-2"},
					},
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-1", "student-id-2", "student-id-3"}).Once().Return(map[string]string{
					"student-id-1": "Grade 5",
					"student-id-2": "Grade 5",
					"student-id-3": "Grade 6",
				}, nil)

				userRepo.On("GetStudentCurrentGradeByUserIDs", mock.Anything, mock.Anything, []string{"student-id-4", "student-id-5"}).Once().
					Return(map[string]string{}, fmt.Errorf("some-user-repo-error"))
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp := lessonQueryHandler.RetrieveLessonsOnCalendar(testCase.ctx, testCase.payloads.(*payloads.GetLessonListOnCalendarArgs))
			expectedErr := testCase.result.(*RetrieveLessonsResponse).Error
			if expectedErr != nil {
				assert.Error(t, resp.Error)
				assert.Equal(t, expectedErr.Error(), resp.Error.Error())
			} else {
				assert.NoError(t, resp.Error)
				assert.Equal(t, testCase.result, resp)
			}

			mock.AssertExpectationsForObjects(t, lessonRepo, lessonTeacherRepo, mockUnleashClient)
		})
	}
}
