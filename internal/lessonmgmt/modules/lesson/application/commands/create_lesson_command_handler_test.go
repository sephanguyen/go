package commands

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	calendar_constants "github.com/manabie-com/backend/internal/calendar/domain/constants"
	calendar_dto "github.com/manabie-com/backend/internal/calendar/domain/dto"
	calendar_entities "github.com/manabie-com/backend/internal/calendar/domain/entities"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	calendar_mock_repositories "github.com/manabie-com/backend/mock/calendar/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_media_module "github.com/manabie-com/backend/mock/lessonmgmt/lesson/media_module_adapter"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_user_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/usermodadapter"
	mock_user_repo "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"
	mock_clients "github.com/manabie-com/backend/mock/lessonmgmt/zoom/clients"
	mpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateLessonCommandHandler_CreateRecurringLesson(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	schedulerRepo := new(calendar_mock_repositories.MockSchedulerRepo)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	studentSubscriptionRepo := new(mock_user_repo.MockStudentSubscriptionRepo)
	masterDataRepo := new(mock_repositories.MockMasterDataRepo)
	mediaModulePort := new(mock_media_module.MockMediaModuleAdapter)
	userModuleAdapter := new(mock_user_module_adapter.MockUserModuleAdapter)
	dateInfoRepo := new(calendar_mock_repositories.MockDateInfoRepo)
	dateInfos := []*calendar_dto.DateInfo{}
	classroomRepo := new(mock_repositories.MockClassroomRepo)
	userRepo := new(mock_user_repo.MockUserRepo)
	mockSchedulerClient := &mock_clients.MockSchedulerClient{}

	baseLesson := &domain.Lesson{
		LessonID:         "test-id-1",
		Name:             "lesson name",
		LocationID:       "center-id-1",
		CreatedAt:        now,
		UpdatedAt:        now,
		StartTime:        time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
		EndTime:          time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
		SchedulingStatus: domain.LessonSchedulingStatusPublished,
		TeachingMedium:   domain.LessonTeachingMediumOffline,
		TeachingMethod:   domain.LessonTeachingMethodIndividual,
		Learners: domain.LessonLearners{
			{
				LearnerID:        "user-id-1",
				CourseID:         "course-id-1",
				AttendStatus:     domain.StudentAttendStatusAttend,
				LocationID:       "center-id-1",
				AttendanceNotice: domain.NoticeEmpty,
				AttendanceReason: domain.ReasonEmpty,
				AttendanceNote:   "sample-attendance-note",
			},
			{
				LearnerID:        "user-id-2",
				CourseID:         "course-id-2",
				AttendStatus:     domain.StudentAttendStatusEmpty,
				LocationID:       "center-id-1",
				AttendanceNotice: domain.NoticeEmpty,
				AttendanceReason: domain.ReasonEmpty,
			},
		},
		Teachers: domain.LessonTeachers{
			{
				TeacherID: "teacher-id-1",
			},
			{
				TeacherID: "teacher-id-2",
			},
		},
		Material: &domain.LessonMaterial{
			MediaIDs: []string{"media-id-1", "media-id-2"},
		},
		Classrooms: domain.LessonClassrooms{
			{
				ClassroomID: "classroom-id-1",
			},
			{
				ClassroomID: "classroom-id-2",
			},
		},
		MasterDataPort:  masterDataRepo,
		UserModulePort:  userModuleAdapter,
		MediaModulePort: mediaModulePort,
		DateInfoRepo:    dateInfoRepo,
		ClassroomRepo:   classroomRepo,
	}
	expectedLessons := []*domain.Lesson{
		baseLesson,
		{
			LessonID:  "test-id-2",
			StartTime: time.Date(2022, 7, 4, 9, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2022, 7, 4, 10, 0, 0, 0, time.UTC),
			Learners: domain.LessonLearners{
				{
					LearnerID:        "user-id-1",
					CourseID:         "course-id-1",
					AttendStatus:     domain.StudentAttendStatusEmpty,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
				{
					LearnerID:        "user-id-2",
					CourseID:         "course-id-2",
					AttendStatus:     domain.StudentAttendStatusEmpty,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
			},
		},
		{
			LessonID:  "test-id-3",
			StartTime: time.Date(2022, 7, 11, 9, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2022, 7, 11, 10, 0, 0, 0, time.UTC),
			Learners: domain.LessonLearners{
				{
					LearnerID:        "user-id-1",
					CourseID:         "course-id-1",
					AttendStatus:     domain.StudentAttendStatusEmpty,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
				{
					LearnerID:        "user-id-2",
					CourseID:         "course-id-2",
					AttendStatus:     domain.StudentAttendStatusEmpty,
					AttendanceNotice: domain.NoticeEmpty,
					AttendanceReason: domain.ReasonEmpty,
				},
			},
		},
	}
	var getLessonIDs = func(lessons []*domain.Lesson) []string {
		lessonIDs := []string{}
		for i := 0; i < len(expectedLessons); i++ {
			lessonIDs = append(lessonIDs, expectedLessons[i].LessonID)
		}
		return lessonIDs
	}
	testCases := []struct {
		name     string
		setup    func(ctx context.Context)
		lesson   *domain.Lesson
		rule     RecurrenceRuleCommand
		hasError bool
	}{
		{
			name: "success",
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				scheduler := calendar_entities.NewScheduler(
					time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
					time.Date(2022, 7, 12, 10, 0, 0, 0, time.UTC),
					calendar_constants.FrequencyWeekly,
					schedulerRepo,
				)

				mockSchedulerClient.On("CreateScheduler", ctx, mock.MatchedBy(func(sch *mpb.CreateSchedulerRequest) bool {
					if !sch.StartDate.AsTime().Equal(scheduler.StartDate) {
						return false
					}
					if !sch.EndDate.AsTime().Equal(scheduler.EndDate) {
						return false
					}
					return true
				})).Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "scheduler-id",
				}, nil).Once()
				studentSubscriptionRepo.
					On("GetStudentCourseSubscriptions", ctx, db, mock.Anything, []string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"}).
					Return(
						user_domain.StudentSubscriptions{
							{
								StudentID: "user-id-1",
								CourseID:  "course-id-1",
								StartAt:   time.Date(2022, 6, 26, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 9, 12, 9, 0, 0, 0, time.UTC),
							},
							{
								StudentID: "user-id-2",
								CourseID:  "course-id-2",
								StartAt:   time.Date(2022, 6, 26, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 7, 10, 9, 0, 0, 0, time.UTC),
							},
						}, nil).
					Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				classroomRepo.On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(nil).
					Once()
				// student
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-1",
					).
					Return(&user_domain.User{ID: "user-id-1", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-2",
					).
					Return(&user_domain.User{ID: "user-id-2", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-3",
					).
					Return(&user_domain.User{ID: "user-id-3", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				// teacher
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-1",
					).
					Return(&user_domain.User{ID: "teacher-id-1", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-2",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				mediaModulePort.
					On(
						"RetrieveMediasByIDs",
						ctx,
						[]string{"media-id-1", "media-id-2"},
					).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.
					On("UpsertLessons", ctx, tx, mock.MatchedBy(func(recurringLesson *domain.RecurringLesson) bool {
						lessons := recurringLesson.Lessons
						for idx, ls := range lessons {
							expLesson := expectedLessons[idx]
							expLesson.LessonID = "test-id-" + fmt.Sprint(idx+1)
							ls.LessonID = "test-id-" + fmt.Sprint(idx+1)
							if !expLesson.StartTime.Equal(ls.StartTime) {
								return false
							}
							if !expLesson.EndTime.Equal(ls.EndTime) {
								return false
							}
						}
						return true
					})).Return(getLessonIDs(expectedLessons), nil).Once()
			},
			lesson: baseLesson,
			rule: RecurrenceRuleCommand{
				StartTime: time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
				UntilDate: time.Date(2022, 7, 12, 10, 0, 0, 0, time.UTC),
			},
			hasError: false,
		},
		{
			name: "success with course teaching time",
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				scheduler := calendar_entities.NewScheduler(
					time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
					time.Date(2022, 7, 12, 10, 0, 0, 0, time.UTC),
					calendar_constants.FrequencyWeekly,
					schedulerRepo,
				)

				mockSchedulerClient.On("CreateScheduler", ctx, mock.MatchedBy(func(sch *mpb.CreateSchedulerRequest) bool {
					if !sch.StartDate.AsTime().Equal(scheduler.StartDate) {
						return false
					}
					if !sch.EndDate.AsTime().Equal(scheduler.EndDate) {
						return false
					}
					return true
				})).Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "scheduler-id",
				}, nil).Once()
				studentSubscriptionRepo.
					On("GetStudentCourseSubscriptions", ctx, db, mock.Anything, []string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"}).
					Return(
						user_domain.StudentSubscriptions{
							{
								StudentID: "user-id-1",
								CourseID:  "course-id-1",
								StartAt:   time.Date(2022, 6, 26, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 9, 12, 9, 0, 0, 0, time.UTC),
							},
							{
								StudentID: "user-id-2",
								CourseID:  "course-id-2",
								StartAt:   time.Date(2022, 6, 26, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 7, 10, 9, 0, 0, 0, time.UTC),
							},
						}, nil).
					Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				classroomRepo.On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(nil).
					Once()
				// student
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-1",
					).
					Return(&user_domain.User{ID: "user-id-1", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-2",
					).
					Return(&user_domain.User{ID: "user-id-2", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"user-id-3",
					).
					Return(&user_domain.User{ID: "user-id-3", Group: "USER_GROUP_STUDENT"}, nil).
					Once()
				// teacher
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-1",
					).
					Return(&user_domain.User{ID: "teacher-id-1", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				userRepo.
					On(
						"GetUserByUserID",
						ctx,
						db,
						"teacher-id-2",
					).
					Return(&user_domain.User{ID: "teacher-id-2", Group: "USER_GROUP_TEACHER"}, nil).
					Once()
				mediaModulePort.
					On(
						"RetrieveMediasByIDs",
						ctx,
						[]string{"media-id-1", "media-id-2"},
					).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				masterDataRepo.
					On("GetCourseTeachingTimeByIDs", ctx, tx, []string{"course-id-1", "course-id-2"}).
					Return(map[string]*domain.Course{
						"course-id-1": {
							CourseID:        "course-id-1",
							PreparationTime: 120,
							BreakTime:       10,
						},
						"course-id-2": {
							CourseID:        "course-id-2",
							PreparationTime: 100,
							BreakTime:       20,
						},
					}, nil).Once()
				lessonRepo.
					On("UpsertLessons", ctx, tx, mock.MatchedBy(func(recurringLesson *domain.RecurringLesson) bool {
						lessons := recurringLesson.Lessons
						for idx, ls := range lessons {
							expLesson := expectedLessons[idx]
							expLesson.LessonID = "test-id-" + fmt.Sprint(idx+1)
							ls.LessonID = "test-id-" + fmt.Sprint(idx+1)
							if !expLesson.StartTime.Equal(ls.StartTime) {
								return false
							}
							if !expLesson.EndTime.Equal(ls.EndTime) {
								return false
							}
						}
						return true
					})).Return(getLessonIDs(expectedLessons), nil).Once()
			},
			lesson: baseLesson,
			rule: RecurrenceRuleCommand{
				StartTime: time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
				UntilDate: time.Date(2022, 7, 12, 10, 0, 0, 0, time.UTC),
			},
			hasError: false,
		},
		{
			name:   "some student course duration not aligned with lesson date",
			lesson: baseLesson,
			rule: RecurrenceRuleCommand{
				StartTime: time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
				UntilDate: time.Date(2022, 7, 12, 10, 0, 0, 0, time.UTC),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				scheduler := calendar_entities.NewScheduler(
					time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
					time.Date(2022, 7, 12, 10, 0, 0, 0, time.UTC),
					calendar_constants.FrequencyWeekly,
					schedulerRepo,
				)

				mockSchedulerClient.On("CreateScheduler", ctx, mock.MatchedBy(func(sch *mpb.CreateSchedulerRequest) bool {
					if !sch.StartDate.AsTime().Equal(scheduler.StartDate) {
						return false
					}
					if !sch.EndDate.AsTime().Equal(scheduler.EndDate) {
						return false
					}
					return true
				})).Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "scheduler-id",
				}, nil).Once()
				studentSubscriptionRepo.
					On("GetStudentCourseSubscriptions", ctx, db, mock.Anything, []string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"}).
					Return(
						user_domain.StudentSubscriptions{
							{
								StudentID: "user-id-1",
								CourseID:  "course-id-1",
								StartAt:   time.Date(2022, 6, 26, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 9, 12, 9, 0, 0, 0, time.UTC),
							},
							{
								StudentID: "user-id-2",
								CourseID:  "course-id-2",
								StartAt:   time.Date(2022, 6, 26, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 7, 10, 9, 0, 0, 0, time.UTC),
							},
						}, nil).
					Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(nil).
					Once()
				mediaModulePort.
					On(
						"RetrieveMediasByIDs",
						ctx,
						[]string{"media-id-1", "media-id-2"},
					).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				classroomRepo.On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				expectedLessons := []*domain.Lesson{
					{
						StartTime: time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
						EndTime:   time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
						Learners:  baseLesson.Learners,
					},
					{
						StartTime: time.Date(2022, 7, 4, 9, 0, 0, 0, time.UTC),
						EndTime:   time.Date(2022, 7, 4, 10, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusEmpty,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
							{
								LearnerID:        "user-id-2",
								CourseID:         "course-id-2",
								AttendStatus:     domain.StudentAttendStatusEmpty,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
						},
					},
					{
						StartTime: time.Date(2022, 7, 11, 9, 0, 0, 0, time.UTC),
						EndTime:   time.Date(2022, 7, 11, 10, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusEmpty,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
						},
					},
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.
					On("UpsertLessons", ctx, tx, mock.MatchedBy(func(recurringLesson *domain.RecurringLesson) bool {
						lessons := recurringLesson.Lessons
						for idx, ls := range lessons {
							expLesson := expectedLessons[idx]
							if !expLesson.StartTime.Equal(ls.StartTime) {
								return false
							}
							if !expLesson.EndTime.Equal(ls.EndTime) {
								return false
							}
							require.EqualValues(t, expLesson.Learners, ls.Learners)
						}
						return true
					})).Return(getLessonIDs(expectedLessons), nil).Once()
			},
		},
		{
			name:   "any student course duration not aligned with selected lesson date",
			lesson: baseLesson,
			rule: RecurrenceRuleCommand{
				StartTime: time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
				UntilDate: time.Date(2022, 7, 12, 10, 0, 0, 0, time.UTC),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				scheduler := calendar_entities.NewScheduler(
					time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
					time.Date(2022, 7, 12, 10, 0, 0, 0, time.UTC),
					calendar_constants.FrequencyWeekly,
					schedulerRepo,
				)

				mockSchedulerClient.On("CreateScheduler", ctx, mock.MatchedBy(func(sch *mpb.CreateSchedulerRequest) bool {
					if !sch.StartDate.AsTime().Equal(scheduler.StartDate) {
						return false
					}
					if !sch.EndDate.AsTime().Equal(scheduler.EndDate) {
						return false
					}
					return true
				})).Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "scheduler-id",
				}, nil).Once()

				studentSubscriptionRepo.
					On("GetStudentCourseSubscriptions", ctx, db, mock.Anything, []string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"}).
					Return(
						user_domain.StudentSubscriptions{
							{
								StudentID: "user-id-1",
								CourseID:  "course-id-1",
								StartAt:   time.Date(2022, 6, 10, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 9, 26, 9, 0, 0, 0, time.UTC),
							},
							{
								StudentID: "user-id-2",
								CourseID:  "course-id-2",
								StartAt:   time.Date(2022, 6, 28, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 7, 10, 9, 0, 0, 0, time.UTC),
							},
						}, nil).
					Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(errors.New("student course invalid")).
					Once()

			},
			hasError: true,
		},
		{
			name:   "create scheduler fail",
			lesson: baseLesson,
			rule: RecurrenceRuleCommand{
				StartTime: time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
				UntilDate: time.Date(2022, 7, 12, 10, 0, 0, 0, time.UTC),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				scheduler := calendar_entities.NewScheduler(
					time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
					time.Date(2022, 7, 12, 10, 0, 0, 0, time.UTC),
					calendar_constants.FrequencyWeekly,
					schedulerRepo,
				)

				mockSchedulerClient.On("CreateScheduler", ctx, mock.MatchedBy(func(sch *mpb.CreateSchedulerRequest) bool {
					if !sch.StartDate.AsTime().Equal(scheduler.StartDate) {
						return false
					}
					if !sch.EndDate.AsTime().Equal(scheduler.EndDate) {
						return false
					}
					return true
				})).Return(nil, fmt.Errorf("insert scheduler fail")).Once()

				studentSubscriptionRepo.
					On("GetStudentCourseSubscriptions", ctx, db, mock.Anything, []string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"}).
					Return(
						user_domain.StudentSubscriptions{
							{
								StudentID: "user-id-1",
								CourseID:  "course-id-1",
								StartAt:   time.Date(2022, 6, 10, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 9, 26, 9, 0, 0, 0, time.UTC),
							},
							{
								StudentID: "user-id-2",
								CourseID:  "course-id-2",
								StartAt:   time.Date(2022, 6, 28, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 7, 10, 9, 0, 0, 0, time.UTC),
							},
						}, nil).
					Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(errors.New("student course invalid")).
					Once()

			},
			hasError: true,
		},
		{
			name:   "create lesson date with first day is closed date",
			lesson: baseLesson,
			rule: RecurrenceRuleCommand{
				StartTime: time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
				UntilDate: time.Date(2022, 7, 12, 10, 0, 0, 0, time.UTC),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return([]*calendar_dto.DateInfo{
					{
						Date:        time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
						LocationID:  "center-id-1",
						DateTypeID:  string(calendar_constants.ClosedDay),
						OpeningTime: "",
						Status:      string(calendar_constants.Draft),
					},
				}, nil).Once()
				lessonRepo.
					On("UpsertLessons", ctx, tx, mock.MatchedBy(func(recurringLesson *domain.RecurringLesson) bool {
						lessons := recurringLesson.Lessons
						for idx, ls := range lessons {
							expLesson := expectedLessons[idx]
							if !expLesson.StartTime.Equal(ls.StartTime) {
								return false
							}
							if !expLesson.EndTime.Equal(ls.EndTime) {
								return false
							}
							require.EqualValues(t, expLesson.Learners, ls.Learners)
						}
						return true
					})).Return(nil, errors.New("could not create lesson in closed date")).Once()
			},
			hasError: true,
		},
	}
	handler := LessonCommandHandler{
		WrapperConnection:       wrapperConnection,
		LessonRepo:              lessonRepo,
		SchedulerRepo:           schedulerRepo,
		StudentSubscriptionRepo: studentSubscriptionRepo,
		UnleashClientIns:        mockUnleashClient,
		SchedulerClient:         mockSchedulerClient,
		MasterDataPort:          masterDataRepo,
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			res, err := handler.CreateRecurringLesson(ctx, CreateRecurringLesson{
				Lesson:   tc.lesson,
				RRuleCmd: tc.rule,
			})
			if err != nil {
				require.True(t, tc.hasError)
			} else {
				require.False(t, tc.hasError)
				require.NoError(t, err)
				require.Equal(t, "scheduler-id", res.ID)
				require.Len(t, res.Lessons, len(expectedLessons))
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}

func TestLessonExecutorService_ImportLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	now := time.Now()
	db := new(mock_database.Ext)
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	masterDataRepo := new(mock_repositories.MockMasterDataRepo)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	dateInfoRepo := new(calendar_mock_repositories.MockDateInfoRepo)
	schedulerRepo := new(calendar_mock_repositories.MockSchedulerRepo)
	dateInfos := []*calendar_dto.DateInfo{}
	studentSubscriptionRepo := new(mock_user_repo.MockStudentSubscriptionRepo)
	expectedlocationVN, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	mockSchedulerClient := &mock_clients.MockSchedulerClient{}

	testCases := []struct {
		name     string
		setup    func(ctx context.Context)
		input    interface{}
		lessons  []*domain.Lesson
		hasError bool
	}{
		{
			name: "import lesson ver1 - successfully",
			input: &lpb.ImportLessonRequest{
				Payload: []byte(fmt.Sprintf(`partner_internal_id,start_date_time,end_date_time,teaching_method
					pid_1,2023-01-02 05:40:00,2023-01-02 06:45:00,1
					pid_2,2023-01-05 08:30:00,2023-01-05 14:45:00,2`)),
				TimeZone: "Asia/Ho_Chi_Minh",
			},

			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Twice()
				tx.On("Commit", ctx).Return(nil).Twice()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(3)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				masterDataRepo.On("GetLowestLocationsByPartnerInternalIDs", ctx, mock.Anything, []string{"pid_1", "pid_2"}).Once().Return(map[string]*domain.Location{
					"pid_1": {LocationID: "center-1", Name: "Center 1"},
					"pid_2": {LocationID: "center-2", Name: "Center 2"},
				}, nil)

				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return([]*calendar_dto.DateInfo{}, nil).Twice()

				mockSchedulerClient.On("CreateScheduler", ctx, mock.Anything).Twice().Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "SchedulerID",
				}, nil)

				masterDataRepo.On("GetLocationByID", ctx, tx, "center-1").
					Return(&domain.Location{
						LocationID: "center-1",
						Name:       "Center 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()

				masterDataRepo.On("GetLocationByID", ctx, tx, "center-2").
					Return(&domain.Location{
						LocationID: "center-2",
						Name:       "Center 2",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()

				expectedLessons := []*domain.Lesson{
					{
						LessonID:         "test-id-1",
						LocationID:       "center-1",
						StartTime:        time.Date(2023, 01, 02, 05, 40, 0, 0, expectedlocationVN),
						EndTime:          time.Date(2023, 01, 02, 06, 45, 0, 0, expectedlocationVN),
						SchedulingStatus: domain.LessonSchedulingStatusDraft,
						TeachingMedium:   domain.LessonTeachingMediumOffline,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						DateInfos:        dateInfos,
						MasterDataPort:   masterDataRepo,
						DateInfoRepo:     dateInfoRepo,
					},
					{
						LessonID:         "test-id-2",
						LocationID:       "center-2",
						StartTime:        time.Date(2023, 01, 05, 8, 30, 0, 0, expectedlocationVN),
						EndTime:          time.Date(2023, 01, 05, 14, 45, 0, 0, expectedlocationVN),
						SchedulingStatus: domain.LessonSchedulingStatusDraft,
						TeachingMedium:   domain.LessonTeachingMediumOffline,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						DateInfos:        dateInfos,
						MasterDataPort:   masterDataRepo,
						DateInfoRepo:     dateInfoRepo,
					},
				}
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Times(4)
				lessonRepo.On("InsertLesson", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actualLesson := args[2].(*domain.Lesson)
						actualLesson.LessonID = expectedLessons[0].LessonID
						actualLesson.MasterDataPort = masterDataRepo
						actualLesson.UserModulePort = nil
						actualLesson.MediaModulePort = nil
						actualLesson.DateInfoRepo = dateInfoRepo
						actualLesson.ClassroomRepo = nil

						assert.NotEmpty(t, actualLesson.LessonID)
						expectedLessons[0].SchedulerID = actualLesson.SchedulerID
						expectedLessons[0].CreatedAt = actualLesson.CreatedAt
						expectedLessons[0].UpdatedAt = actualLesson.UpdatedAt
						assert.EqualValues(t, expectedLessons[0], actualLesson)
					}).Return(expectedLessons[0], nil).Once()

				lessonRepo.On("InsertLesson", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actualLesson := args[2].(*domain.Lesson)
						actualLesson.LessonID = expectedLessons[1].LessonID
						actualLesson.MasterDataPort = masterDataRepo
						actualLesson.UserModulePort = nil
						actualLesson.MediaModulePort = nil
						actualLesson.DateInfoRepo = dateInfoRepo
						actualLesson.ClassroomRepo = nil

						assert.NotEmpty(t, actualLesson.LessonID)
						expectedLessons[1].SchedulerID = actualLesson.SchedulerID
						expectedLessons[1].CreatedAt = actualLesson.CreatedAt
						expectedLessons[1].UpdatedAt = actualLesson.UpdatedAt
						assert.EqualValues(t, expectedLessons[1], actualLesson)
					}).Return(expectedLessons[1], nil).Once()
			},
		},
		{
			name: "import lesson ver2 - successfully",
			input: &lpb.ImportLessonRequest{
				Payload: []byte(fmt.Sprintf(`partner_internal_id,start_date_time,end_date_time,teaching_method,teaching_medium,teacher_ids,student_course_ids
					pid_1,2023-01-02 05:40:00,2023-01-02 06:45:00,1,1,,
					pid_2,2023-01-05 08:30:00,2023-01-05 14:45:00,2,2,,`)),
				TimeZone: "Asia/Ho_Chi_Minh",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Twice()
				tx.On("Commit", ctx).Return(nil).Twice()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(3)
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				masterDataRepo.On("GetLowestLocationsByPartnerInternalIDs", ctx, mock.Anything, []string{"pid_1", "pid_2"}).Once().Return(map[string]*domain.Location{
					"pid_1": {LocationID: "center-1", Name: "Center 1"},
					"pid_2": {LocationID: "center-2", Name: "Center 2"},
				}, nil)

				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return([]*calendar_dto.DateInfo{}, nil).Once()

				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return([]*calendar_dto.DateInfo{}, nil).Twice()

				mockSchedulerClient.On("CreateScheduler", ctx, mock.Anything).Twice().Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "SchedulerID",
				}, nil)

				masterDataRepo.On("GetLocationByID", ctx, tx, "center-1").
					Return(&domain.Location{
						LocationID: "center-1",
						Name:       "Center 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()

				masterDataRepo.On("GetLocationByID", ctx, tx, "center-2").
					Return(&domain.Location{
						LocationID: "center-2",
						Name:       "Center 2",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()

				expectedLessons := []*domain.Lesson{
					{
						LessonID:         "test-id-1",
						LocationID:       "center-1",
						StartTime:        time.Date(2023, 01, 02, 05, 40, 0, 0, expectedlocationVN),
						EndTime:          time.Date(2023, 01, 02, 06, 45, 0, 0, expectedlocationVN),
						SchedulingStatus: domain.LessonSchedulingStatusDraft,
						TeachingMedium:   domain.LessonTeachingMediumOffline,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						DateInfos:        dateInfos,
						MasterDataPort:   masterDataRepo,
						DateInfoRepo:     dateInfoRepo,
					},
					{
						LessonID:         "test-id-2",
						LocationID:       "center-2",
						StartTime:        time.Date(2023, 01, 05, 8, 30, 0, 0, expectedlocationVN),
						EndTime:          time.Date(2023, 01, 05, 14, 45, 0, 0, expectedlocationVN),
						SchedulingStatus: domain.LessonSchedulingStatusDraft,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						DateInfos:        dateInfos,
						MasterDataPort:   masterDataRepo,
						DateInfoRepo:     dateInfoRepo,
					},
				}
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Times(4)
				lessonRepo.On("InsertLesson", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actualLesson := args[2].(*domain.Lesson)
						actualLesson.LessonID = expectedLessons[0].LessonID
						actualLesson.MasterDataPort = masterDataRepo
						actualLesson.UserModulePort = nil
						actualLesson.MediaModulePort = nil
						actualLesson.DateInfoRepo = dateInfoRepo
						actualLesson.ClassroomRepo = nil

						assert.NotEmpty(t, actualLesson.LessonID)
						expectedLessons[0].SchedulerID = actualLesson.SchedulerID
						expectedLessons[0].CreatedAt = actualLesson.CreatedAt
						expectedLessons[0].UpdatedAt = actualLesson.UpdatedAt
						assert.EqualValues(t, expectedLessons[0], actualLesson)
					}).Return(expectedLessons[0], nil).Once()

				lessonRepo.On("InsertLesson", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actualLesson := args[2].(*domain.Lesson)
						actualLesson.LessonID = expectedLessons[1].LessonID
						actualLesson.MasterDataPort = masterDataRepo
						actualLesson.UserModulePort = nil
						actualLesson.MediaModulePort = nil
						actualLesson.DateInfoRepo = dateInfoRepo
						actualLesson.ClassroomRepo = nil

						assert.NotEmpty(t, actualLesson.LessonID)
						expectedLessons[1].SchedulerID = actualLesson.SchedulerID
						expectedLessons[1].CreatedAt = actualLesson.CreatedAt
						expectedLessons[1].UpdatedAt = actualLesson.UpdatedAt
						assert.EqualValues(t, expectedLessons[1], actualLesson)
					}).Return(expectedLessons[1], nil).Once()
			},
		},
	}

	handler := LessonCommandHandler{
		WrapperConnection:       wrapperConnection,
		LessonRepo:              lessonRepo,
		SchedulerRepo:           schedulerRepo,
		StudentSubscriptionRepo: studentSubscriptionRepo,
		UnleashClientIns:        mockUnleashClient,
		MasterDataPort:          masterDataRepo,
		DateInfoRepo:            dateInfoRepo,
		SchedulerClient:         mockSchedulerClient,
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			res, csvErr, err := handler.ImportLesson(ctx, tc.input.(*lpb.ImportLessonRequest))
			if err != nil {
				require.True(t, tc.hasError)
			} else {
				require.False(t, tc.hasError)
				require.NoError(t, err)
				require.Nil(t, csvErr)
				require.NotNil(t, res)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}

func TestLessonExecutorService_ImportLessonV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	now := time.Now()
	db := new(mock_database.Ext)
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	masterDataRepo := new(mock_repositories.MockMasterDataRepo)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	dateInfoRepo := new(calendar_mock_repositories.MockDateInfoRepo)
	schedulerRepo := new(calendar_mock_repositories.MockSchedulerRepo)
	dateInfos := []*calendar_dto.DateInfo{}
	studentSubscriptionRepo := new(mock_user_repo.MockStudentSubscriptionRepo)
	expectedlocationVN, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	mockSchedulerClient := &mock_clients.MockSchedulerClient{}

	mapSchedulers := map[string]string{
		"test-id-1": "scheduler_id_01",
		"test-id-2": "scheduler_id_02",
	}

	testCases := []struct {
		name     string
		setup    func(ctx context.Context)
		input    interface{}
		lessons  []*domain.Lesson
		hasError bool
	}{
		{
			name: "import lesson ver2 - successfully",
			input: &lpb.ImportLessonRequest{
				Payload: []byte(fmt.Sprintf(`partner_internal_id,start_date_time,end_date_time,teaching_method,teaching_medium,teacher_ids,student_course_ids
					pid_1,2023-01-02 05:40:00,2023-01-02 06:45:00,1,1,,
					pid_2,2023-01-05 08:30:00,2023-01-05 14:45:00,2,2,,`)),
				TimeZone: "Asia/Ho_Chi_Minh",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Twice()
				tx.On("Commit", ctx).Return(nil).Twice()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
					mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Twice()
				masterDataRepo.On("GetLowestLocationsByPartnerInternalIDs", ctx, mock.Anything, []string{"pid_1", "pid_2"}).Once().Return(map[string]*domain.Location{
					"pid_1": {LocationID: "center-1", Name: "Center 1"},
					"pid_2": {LocationID: "center-2", Name: "Center 2"},
				}, nil)

				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return([]*calendar_dto.DateInfo{}, nil).Twice()

				mockSchedulerClient.On("CreateManySchedulers", ctx, mock.Anything).Once().Return(&mpb.CreateManySchedulersResponse{
					MapSchedulers: mapSchedulers,
				}, nil)

				masterDataRepo.On("GetLocationByID", ctx, tx, "center-1").
					Return(&domain.Location{
						LocationID: "center-1",
						Name:       "Center 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()

				masterDataRepo.On("GetLocationByID", ctx, tx, "center-2").
					Return(&domain.Location{
						LocationID: "center-2",
						Name:       "Center 2",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()

				expectedLessons := []*domain.Lesson{
					{
						LessonID:         "test-id-1",
						LocationID:       "center-1",
						StartTime:        time.Date(2023, 01, 02, 05, 40, 0, 0, expectedlocationVN),
						EndTime:          time.Date(2023, 01, 02, 06, 45, 0, 0, expectedlocationVN),
						SchedulingStatus: domain.LessonSchedulingStatusDraft,
						TeachingMedium:   domain.LessonTeachingMediumOffline,
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						DateInfos:        dateInfos,
						MasterDataPort:   masterDataRepo,
						DateInfoRepo:     dateInfoRepo,
					},
					{
						LessonID:         "test-id-2",
						LocationID:       "center-2",
						StartTime:        time.Date(2023, 01, 05, 8, 30, 0, 0, expectedlocationVN),
						EndTime:          time.Date(2023, 01, 05, 14, 45, 0, 0, expectedlocationVN),
						SchedulingStatus: domain.LessonSchedulingStatusDraft,
						TeachingMedium:   domain.LessonTeachingMediumOnline,
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						DateInfos:        dateInfos,
						MasterDataPort:   masterDataRepo,
						DateInfoRepo:     dateInfoRepo,
					},
				}
				lessonRepo.On("InsertLesson", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actualLesson := args[2].(*domain.Lesson)
						actualLesson.LessonID = expectedLessons[0].LessonID
						actualLesson.MasterDataPort = masterDataRepo
						actualLesson.UserModulePort = nil
						actualLesson.MediaModulePort = nil
						actualLesson.DateInfoRepo = dateInfoRepo
						actualLesson.ClassroomRepo = nil

						assert.NotEmpty(t, actualLesson.LessonID)
						expectedLessons[0].SchedulerID = actualLesson.SchedulerID
						expectedLessons[0].CreatedAt = actualLesson.CreatedAt
						expectedLessons[0].UpdatedAt = actualLesson.UpdatedAt
						assert.EqualValues(t, expectedLessons[0], actualLesson)
					}).Return(expectedLessons[0], nil).Once()

				lessonRepo.On("InsertLesson", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actualLesson := args[2].(*domain.Lesson)
						actualLesson.LessonID = expectedLessons[1].LessonID
						actualLesson.MasterDataPort = masterDataRepo
						actualLesson.UserModulePort = nil
						actualLesson.MediaModulePort = nil
						actualLesson.DateInfoRepo = dateInfoRepo
						actualLesson.ClassroomRepo = nil

						assert.NotEmpty(t, actualLesson.LessonID)
						expectedLessons[1].SchedulerID = actualLesson.SchedulerID
						expectedLessons[1].CreatedAt = actualLesson.CreatedAt
						expectedLessons[1].UpdatedAt = actualLesson.UpdatedAt
						assert.EqualValues(t, expectedLessons[1], actualLesson)
					}).Return(expectedLessons[1], nil).Once()
			},
		},
	}

	handler := LessonCommandHandler{
		WrapperConnection:       wrapperConnection,
		LessonRepo:              lessonRepo,
		SchedulerRepo:           schedulerRepo,
		StudentSubscriptionRepo: studentSubscriptionRepo,
		UnleashClientIns:        mockUnleashClient,
		MasterDataPort:          masterDataRepo,
		DateInfoRepo:            dateInfoRepo,
		SchedulerClient:         mockSchedulerClient,
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			res, csvErr, err := handler.ImportLessonV2(ctx, tc.input.(*lpb.ImportLessonRequest))
			if err != nil {
				require.True(t, tc.hasError)
			} else {
				require.False(t, tc.hasError)
				require.NoError(t, err)
				require.Nil(t, csvErr)
				require.NotNil(t, res)
			}
			mock.AssertExpectationsForObjects(t, masterDataRepo, dateInfoRepo, mockSchedulerClient, lessonRepo, mockUnleashClient)
		})
	}
}
