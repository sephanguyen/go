package commands

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	calendar_constants "github.com/manabie-com/backend/internal/calendar/domain/constants"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	calendar_dto "github.com/manabie-com/backend/internal/calendar/domain/dto"
	calendar_entities "github.com/manabie-com/backend/internal/calendar/domain/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/producers"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	calendar_mock_repositories "github.com/manabie-com/backend/mock/calendar/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_media_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/media_module_adapter"
	mock_lesson_nats_repo "github.com/manabie-com/backend/mock/lessonmgmt/lesson/nats"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_user_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/usermodadapter"
	report_mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson_report/repositories"
	mock_user_repo "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"
	mock_clients "github.com/manabie-com/backend/mock/lessonmgmt/zoom/clients"
	mock_service "github.com/manabie-com/backend/mock/lessonmgmt/zoom/service"
	usermgmt_repo "github.com/manabie-com/backend/mock/usermgmt/repositories"
	mpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateLessonCommandHandler_UpdateLessonOneTime(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := new(mock_repositories.MockLessonRepo)
	lessonReportRepo := new(report_mock_repositories.MockLessonReportRepo)
	masterDataRepo := new(mock_repositories.MockMasterDataRepo)
	mediaModulePort := new(mock_media_module_adapter.MockMediaModuleAdapter)
	userModuleAdapter := new(mock_user_module_adapter.MockUserModuleAdapter)
	schedulerRepo := new(calendar_mock_repositories.MockSchedulerRepo)
	dateInfoRepo := new(calendar_mock_repositories.MockDateInfoRepo)
	studentSubscriptionRepo := new(mock_user_repo.MockStudentSubscriptionRepo)
	reallocationRepo := new(mock_repositories.MockReallocationRepo)
	classroomRepo := new(mock_repositories.MockClassroomRepo)
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	userAccessPathRepo := new(mock_user_repo.MockUserAccessPathRepo)
	enrollmentStatusHistoryRepo := &usermgmt_repo.MockDomainEnrollmentStatusHistoryRepo{}
	mockSchedulerClient := &mock_clients.MockSchedulerClient{}
	mockLessonPublisher := &mock_lesson_nats_repo.MockLessonPublisher{}

	testcases := []struct {
		name      string
		setup     func(ctx context.Context, ls *domain.Lesson)
		newLesson *domain.Lesson
		oldLesson *domain.Lesson
		hasError  bool
	}{
		{
			name: "happy case - normal update with course teaching time",
			newLesson: func() *domain.Lesson {
				builder := domain.NewLesson().
					WithMasterDataPort(masterDataRepo).
					WithUserModulePort(userModuleAdapter).
					WithMediaModulePort(mediaModulePort).
					WithDateInfoRepo(dateInfoRepo).
					WithClassroomRepo(classroomRepo).
					WithLessonRepo(lessonRepo).
					WithID("lesson-id-1").
					WithLocationID("center-id-1").
					WithTimeRange(time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC), time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC)).
					WithTeachingMedium(domain.LessonTeachingMediumOffline).
					WithTeachingMethod(domain.LessonTeachingMethodIndividual).
					WithTeacherIDs([]string{"teacher-id-1", "teacher-id-2"}).
					WithClassroomIDs([]string{"classroom-id-1", "classroom-id-2"}).
					WithSchedulingStatus(domain.LessonSchedulingStatusPublished).
					WithMaterials([]string{"media-id-1", "media-id-2"}).
					WithLearners(domain.LessonLearners{
						{
							LearnerID:        "user-id-1",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAttend,
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
					})
				return builder.BuildDraft()
			}(),
			oldLesson: &domain.Lesson{},
			setup: func(ctx context.Context, newLesson *domain.Lesson) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("GetLessonByID", ctx, tx, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(nil).
					Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-id-1", "media-id-2"}).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				dateInfoRepo.
					On("GetDateInfoByDateRangeAndLocationID",
						ctx, tx, mock.Anything, mock.Anything, mock.Anything,
					).Return(nil, nil).Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, tx, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				lessonRepo.
					On("UpdateLesson", ctx, tx, newLesson).Return(newLesson, nil).Once()
				reallocationRepo.
					On("GetByNewLessonID", ctx, tx, []string{}, mock.AnythingOfType("string")).
					Return([]*domain.Reallocation{}, nil).Once()
			},
		},
		{
			name: "happy case - normal update",
			newLesson: func() *domain.Lesson {
				builder := domain.NewLesson().
					WithMasterDataPort(masterDataRepo).
					WithUserModulePort(userModuleAdapter).
					WithMediaModulePort(mediaModulePort).
					WithDateInfoRepo(dateInfoRepo).
					WithClassroomRepo(classroomRepo).
					WithLessonRepo(lessonRepo).
					WithID("lesson-id-1").
					WithLocationID("center-id-1").
					WithTimeRange(time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC), time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC)).
					WithTeachingMedium(domain.LessonTeachingMediumOffline).
					WithTeachingMethod(domain.LessonTeachingMethodIndividual).
					WithTeacherIDs([]string{"teacher-id-1", "teacher-id-2"}).
					WithClassroomIDs([]string{"classroom-id-1", "classroom-id-2"}).
					WithSchedulingStatus(domain.LessonSchedulingStatusPublished).
					WithMaterials([]string{"media-id-1", "media-id-2"}).
					WithLearners(domain.LessonLearners{
						{
							LearnerID:        "user-id-1",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAttend,
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
					})
				return builder.BuildDraft()
			}(),
			oldLesson: &domain.Lesson{},
			setup: func(ctx context.Context, newLesson *domain.Lesson) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("GetLessonByID", ctx, tx, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(nil).
					Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-id-1", "media-id-2"}).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				dateInfoRepo.
					On("GetDateInfoByDateRangeAndLocationID",
						ctx, tx, mock.Anything, mock.Anything, mock.Anything,
					).Return(nil, nil).Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, tx, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				lessonRepo.
					On("UpdateLesson", ctx, tx, newLesson).Return(newLesson, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				reallocationRepo.On("GetByNewLessonID", ctx, tx, mock.Anything, mock.Anything).Return([]*domain.Reallocation{}, nil).Once()
			},
		},
		{
			name: "happy case - update lesson from weekly recurring to one time",
			newLesson: func() *domain.Lesson {
				builder := domain.NewLesson().
					WithMasterDataPort(masterDataRepo).
					WithUserModulePort(userModuleAdapter).
					WithMediaModulePort(mediaModulePort).
					WithDateInfoRepo(dateInfoRepo).
					WithClassroomRepo(classroomRepo).
					WithLessonRepo(lessonRepo).
					WithID("lesson-id-1").
					WithLocationID("center-id-1").
					WithTimeRange(time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC), time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC)).
					WithTeachingMedium(domain.LessonTeachingMediumOffline).
					WithTeachingMethod(domain.LessonTeachingMethodIndividual).
					WithTeacherIDs([]string{"teacher-id-1", "teacher-id-2"}).
					WithClassroomIDs([]string{"classroom-id-1", "classroom-id-2"}).
					WithSchedulingStatus(domain.LessonSchedulingStatusPublished).
					WithMaterials([]string{"media-id-1", "media-id-2"}).
					WithLearners(domain.LessonLearners{
						{
							LearnerID:        "user-id-1",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAttend,
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
					})
				return builder.BuildDraft()
			}(),
			oldLesson: &domain.Lesson{
				LocationID:  "center-id-2",
				SchedulerID: "scheduler-id",
			},
			setup: func(ctx context.Context, newLesson *domain.Lesson) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("GetLessonByID", ctx, tx, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(nil).
					Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-id-1", "media-id-2"}).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				dateInfoRepo.
					On("GetDateInfoByDateRangeAndLocationID", ctx, tx, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil).Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, tx, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				schedulerRepo.
					On("GetByID", ctx, tx, "scheduler-id").Return(&calendar_dto.Scheduler{
					SchedulerID: "scheduler-id",
					Frequency:   "weekly",
				}, nil).Once()

				mockSchedulerClient.On("CreateScheduler", ctx, mock.Anything).Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "new-scheduler",
				}, nil).Once()
				lessonRepo.
					On("UpdateLesson", ctx, tx, newLesson).Return(newLesson, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				reallocationRepo.On("GetByNewLessonID", ctx, tx, mock.Anything, mock.Anything).Return([]*domain.Reallocation{}, nil).Once()
			},
		},
		{
			name: "happy case - update lesson with reallocate student",
			newLesson: func() *domain.Lesson {
				builder := domain.NewLesson().
					WithMasterDataPort(masterDataRepo).
					WithUserModulePort(userModuleAdapter).
					WithMediaModulePort(mediaModulePort).
					WithDateInfoRepo(dateInfoRepo).
					WithClassroomRepo(classroomRepo).
					WithLessonRepo(lessonRepo).
					WithID("lesson-id-1").
					WithLocationID("center-id-1").
					WithTimeRange(time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC), time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC)).
					WithTeachingMedium(domain.LessonTeachingMediumOffline).
					WithTeachingMethod(domain.LessonTeachingMethodIndividual).
					WithTeacherIDs([]string{"teacher-id-1", "teacher-id-2"}).
					WithClassroomIDs([]string{"classroom-id-1", "classroom-id-2"}).
					WithSchedulingStatus(domain.LessonSchedulingStatusPublished).
					WithMaterials([]string{"media-id-1", "media-id-2"}).
					WithLearners(domain.LessonLearners{
						{
							LearnerID:        "user-id-1",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAbsent,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
					})
				return builder.BuildDraft()
			}(),
			oldLesson: &domain.Lesson{
				Learners: domain.LessonLearners{
					{
						LearnerID:        "user-id-1",
						CourseID:         "course-id-1",
						AttendStatus:     domain.StudentAttendStatusReallocate,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.NoticeEmpty,
						AttendanceReason: domain.ReasonEmpty,
					},
					{
						LearnerID:        "user-id-2",
						CourseID:         "course-id-2",
						AttendStatus:     domain.StudentAttendStatusReallocate,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.NoticeEmpty,
						AttendanceReason: domain.ReasonEmpty,
					},
				},
			},
			setup: func(ctx context.Context, newLesson *domain.Lesson) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("GetLessonByID", ctx, tx, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1"},
					).
					Return(nil).
					Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-id-1", "media-id-2"}).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				dateInfoRepo.
					On("GetDateInfoByDateRangeAndLocationID",
						ctx, tx, mock.Anything, mock.Anything, mock.Anything,
					).Return(nil, nil).Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, tx, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				lessonRepo.
					On("UpdateLesson", ctx, tx, newLesson).Return(newLesson, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				reallocationRepo.On("GetFollowingReallocation", ctx, tx, "lesson-id-1", []string{"user-id-1"}).
					Return([]*domain.Reallocation{
						{
							OriginalLessonID: "lesson-id-1",
							StudentID:        "user-id-1",
							NewLessonID:      "lesson-id-2",
						},
						{
							OriginalLessonID: "lesson-id-2",
							StudentID:        "user-id-1",
							NewLessonID:      "lesson-id-3",
						},
					}, nil).Once()
				reallocationRepo.On("SoftDelete", ctx, tx, []string{"user-id-1", "lesson-id-1", "user-id-1", "lesson-id-2"}, true).Return(nil).Once()
				lessonMemberRepo.On("DeleteLessonMembers", ctx, tx, []*domain.LessonMember{
					{
						LessonID:  "lesson-id-2",
						StudentID: "user-id-1",
					},
				}).Return(nil).Once()
				reallocationRepo.On("GetByNewLessonID", ctx, tx, mock.Anything, mock.Anything).Return([]*domain.Reallocation{}, nil).Once()
				reallocationRepo.On("SoftDelete", ctx, tx, []string{"user-id-2", "lesson-id-1"}, false).Return(nil).Once()
				reallocationRepo.On("CancelIfStudentReallocated", ctx, tx, []string{"user-id-2", "lesson-id-1"}).Return(nil).Once()
			},
		},
		{
			name: "happy case - create temporary location assignment",
			newLesson: func() *domain.Lesson {
				builder := domain.NewLesson().
					WithMasterDataPort(masterDataRepo).
					WithUserModulePort(userModuleAdapter).
					WithMediaModulePort(mediaModulePort).
					WithDateInfoRepo(dateInfoRepo).
					WithClassroomRepo(classroomRepo).
					WithLessonRepo(lessonRepo).
					WithID("lesson-id-1").
					WithLocationID("center-id-1").
					WithTimeRange(time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC), time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC)).
					WithTeachingMedium(domain.LessonTeachingMediumOffline).
					WithTeachingMethod(domain.LessonTeachingMethodIndividual).
					WithTeacherIDs([]string{"teacher-id-1", "teacher-id-2"}).
					WithClassroomIDs([]string{"classroom-id-1", "classroom-id-2"}).
					WithSchedulingStatus(domain.LessonSchedulingStatusPublished).
					WithMaterials([]string{"media-id-1", "media-id-2"}).
					WithLearners(domain.LessonLearners{
						{
							LearnerID:        "user-id-1",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAbsent,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-2",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAbsent,
							LocationID:       "center-id-2",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
							Reallocate: &domain.Reallocate{
								OriginalLessonID: "lesson-2",
							},
						},
					})
				return builder.BuildDraft()
			}(),
			oldLesson: &domain.Lesson{
				Learners: domain.LessonLearners{
					{
						LearnerID:        "user-id-1",
						CourseID:         "course-id-1",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.NoticeEmpty,
						AttendanceReason: domain.ReasonEmpty,
					},
				},
			},
			setup: func(ctx context.Context, newLesson *domain.Lesson) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("GetLessonByID", ctx, tx, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-1"},
					).
					Return(nil).
					Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-id-1", "media-id-2"}).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				dateInfoRepo.
					On("GetDateInfoByDateRangeAndLocationID",
						ctx, tx, mock.Anything, mock.Anything, mock.Anything,
					).Return(nil, nil).Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, tx, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				lessonRepo.
					On("UpdateLesson", ctx, tx, newLesson).Return(newLesson, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Twice()
				userAccessPathRepo.
					On("GetLocationAssignedByUserID", ctx, tx, []string{"user-id-2"}).Return(map[string][]string{
					"user-id-2": {"center-id-2"},
				}, nil).Once()
				enrollmentStatusHistoryRepo.
					On("Create", ctx, tx, mock.Anything).Return(nil).Once()
				userAccessPathRepo.
					On("Create", ctx, tx, mock.Anything).Return(nil).Once()
				reallocationRepo.On("GetByNewLessonID", ctx, tx, mock.Anything, mock.Anything).Return([]*domain.Reallocation{}, nil).Once()

			},
		},
		{
			name: "happy case - create temporary location assignment - publish event",
			newLesson: func() *domain.Lesson {
				builder := domain.NewLesson().
					WithMasterDataPort(masterDataRepo).
					WithUserModulePort(userModuleAdapter).
					WithMediaModulePort(mediaModulePort).
					WithDateInfoRepo(dateInfoRepo).
					WithClassroomRepo(classroomRepo).
					WithLessonRepo(lessonRepo).
					WithID("lesson-id-1").
					WithLocationID("center-id-1").
					WithTimeRange(time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC), time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC)).
					WithTeachingMedium(domain.LessonTeachingMediumOffline).
					WithTeachingMethod(domain.LessonTeachingMethodIndividual).
					WithTeacherIDs([]string{"teacher-id-1", "teacher-id-2"}).
					WithClassroomIDs([]string{"classroom-id-1", "classroom-id-2"}).
					WithSchedulingStatus(domain.LessonSchedulingStatusPublished).
					WithMaterials([]string{"media-id-1", "media-id-2"}).
					WithLearners(domain.LessonLearners{
						{
							LearnerID:        "user-id-1",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAbsent,
							LocationID:       "center-id-1",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
						},
						{
							LearnerID:        "user-id-2",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAbsent,
							LocationID:       "center-id-2",
							AttendanceNotice: domain.NoticeEmpty,
							AttendanceReason: domain.ReasonEmpty,
							Reallocate: &domain.Reallocate{
								OriginalLessonID: "lesson-2",
							},
						},
					})
				return builder.BuildDraft()
			}(),
			oldLesson: &domain.Lesson{
				Learners: domain.LessonLearners{
					{
						LearnerID:        "user-id-1",
						CourseID:         "course-id-1",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.NoticeEmpty,
						AttendanceReason: domain.ReasonEmpty,
					},
				},
			},
			setup: func(ctx context.Context, newLesson *domain.Lesson) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("GetLessonByID", ctx, tx, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-1"},
					).
					Return(nil).
					Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-id-1", "media-id-2"}).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				dateInfoRepo.
					On("GetDateInfoByDateRangeAndLocationID",
						ctx, tx, mock.Anything, mock.Anything, mock.Anything,
					).Return(nil, nil).Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, tx, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				lessonRepo.
					On("UpdateLesson", ctx, tx, newLesson).Return(newLesson, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Twice()
				userAccessPathRepo.
					On("GetLocationAssignedByUserID", ctx, tx, []string{"user-id-2"}).Return(map[string][]string{
					"user-id-2": {"center-id-2"},
				}, nil).Once()
				mockLessonPublisher.On("PublishTemporaryLocationAssignment", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				enrollmentStatusHistoryRepo.
					On("Create", ctx, tx, mock.Anything).Return(nil).Once()
				userAccessPathRepo.
					On("Create", ctx, tx, mock.Anything).Return(nil).Once()
				reallocationRepo.On("GetByNewLessonID", ctx, tx, mock.Anything, mock.Anything).Return([]*domain.Reallocation{}, nil).Once()

			},
		},
		{
			name: "happy case - normal update with course teaching time",
			newLesson: func() *domain.Lesson {
				builder := domain.NewLesson().
					WithMasterDataPort(masterDataRepo).
					WithUserModulePort(userModuleAdapter).
					WithMediaModulePort(mediaModulePort).
					WithDateInfoRepo(dateInfoRepo).
					WithClassroomRepo(classroomRepo).
					WithLessonRepo(lessonRepo).
					WithID("lesson-id-1").
					WithLocationID("center-id-1").
					WithTimeRange(time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC), time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC)).
					WithTeachingMedium(domain.LessonTeachingMediumOffline).
					WithTeachingMethod(domain.LessonTeachingMethodIndividual).
					WithTeacherIDs([]string{"teacher-id-1", "teacher-id-2"}).
					WithClassroomIDs([]string{"classroom-id-1", "classroom-id-2"}).
					WithSchedulingStatus(domain.LessonSchedulingStatusPublished).
					WithMaterials([]string{"media-id-1", "media-id-2"}).
					WithLearners(domain.LessonLearners{
						{
							LearnerID:        "user-id-1",
							CourseID:         "course-id-1",
							AttendStatus:     domain.StudentAttendStatusAttend,
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
					})
				return builder.BuildDraft()
			}(),
			oldLesson: &domain.Lesson{},
			setup: func(ctx context.Context, newLesson *domain.Lesson) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("GetLessonByID", ctx, tx, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, tx, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(nil).
					Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-id-1", "media-id-2"}).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				dateInfoRepo.
					On("GetDateInfoByDateRangeAndLocationID",
						ctx, tx, mock.Anything, mock.Anything, mock.Anything,
					).Return(nil, nil).Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, tx, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				lessonRepo.
					On("UpdateLesson", ctx, tx, newLesson).Return(newLesson, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				reallocationRepo.
					On("GetByNewLessonID", ctx, tx, []string{}, mock.AnythingOfType("string")).
					Return([]*domain.Reallocation{}, nil).Once()
			},
		},
	}
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.ManabieSchool),
		},
	}
	ctx := interceptors.ContextWithJWTClaims(context.Background(), claim)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx, tc.newLesson)
			handler := LessonCommandHandler{
				WrapperConnection:            wrapperConnection,
				LessonRepo:                   lessonRepo,
				SchedulerRepo:                schedulerRepo,
				LessonReportRepo:             lessonReportRepo,
				UnleashClientIns:             mockUnleashClient,
				Env:                          "local",
				StudentSubscriptionRepo:      studentSubscriptionRepo,
				ClassroomRepo:                classroomRepo,
				ReallocationRepo:             reallocationRepo,
				LessonMemberRepo:             lessonMemberRepo,
				UserAccessPathRepo:           userAccessPathRepo,
				StudentEnrollmentHistoryRepo: enrollmentStatusHistoryRepo,
				SchedulerClient:              mockSchedulerClient,
				LessonPublisher:              mockLessonPublisher,
				MasterDataPort:               masterDataRepo,
			}

			res, err := handler.UpdateLessonOneTime(ctx, UpdateLessonOneTimeCommandRequest{
				Lesson:        tc.newLesson,
				CurrentLesson: tc.oldLesson,
			})
			if tc.hasError {
				require.Error(t, err)
				require.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.Equal(t, res, tc.newLesson)
				mock.AssertExpectationsForObjects(t, db, tx, dateInfoRepo, lessonRepo, masterDataRepo, mediaModulePort, userModuleAdapter, mockUnleashClient)
			}
		})
	}
}

func TestUpdateLessonCommandHandler_UpdateRecurringLesson(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	masterDataRepo := new(mock_repositories.MockMasterDataRepo)
	mediaModulePort := new(mock_media_module_adapter.MockMediaModuleAdapter)
	userModuleAdapter := new(mock_user_module_adapter.MockUserModuleAdapter)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	courseRepo := new(mock_repositories.MockCourseRepo)
	schedulerRepo := new(calendar_mock_repositories.MockSchedulerRepo)
	studentSubscriptionRepo := new(mock_user_repo.MockStudentSubscriptionRepo)
	dateInfoRepo := new(calendar_mock_repositories.MockDateInfoRepo)
	classroomRepo := new(mock_repositories.MockClassroomRepo)
	jsm := new(mock_nats.JetStreamManagement)
	dateInfos := []*calendar_dto.DateInfo{}
	reallocationRepo := new(mock_repositories.MockReallocationRepo)
	mockSchedulerClient := &mock_clients.MockSchedulerClient{}

	selectedLesson := &domain.Lesson{
		LessonID:         "lesson-id-1",
		LocationID:       "center-id-1",
		CourseID:         "course-1",
		ClassID:          "class-1",
		StartTime:        time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
		EndTime:          time.Date(2022, 7, 2, 10, 0, 0, 0, time.UTC),
		CreatedAt:        time.Date(2022, 6, 2, 10, 0, 0, 0, time.UTC),
		UpdatedAt:        now,
		TeachingMedium:   domain.LessonTeachingMediumOffline,
		TeachingMethod:   domain.LessonTeachingMethodIndividual,
		SchedulingStatus: domain.LessonSchedulingStatusPublished,
		Learners: domain.LessonLearners{
			{
				LearnerID:        "user-id-1",
				CourseID:         "course-id-1",
				AttendStatus:     domain.StudentAttendStatusAttend,
				LocationID:       "center-id-1",
				AttendanceNotice: domain.NoticeEmpty,
				AttendanceReason: domain.ReasonEmpty,
			},
			{
				LearnerID:        "user-id-2",
				CourseID:         "course-id-2",
				AttendStatus:     domain.StudentAttendStatusAttend,
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
	tcs := []struct {
		name           string
		selectedLesson *domain.Lesson
		currentLesson  *domain.Lesson
		setup          func(ctx context.Context)
		untilDate      time.Time
		hasError       bool
	}{
		{
			name:           "learner info for following lesson must no change when saving this and following",
			selectedLesson: selectedLesson,
			currentLesson: &domain.Lesson{
				LessonID:    "lesson-id-1",
				LocationID:  "center-id-1",
				StartTime:   time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
				EndTime:     time.Date(2022, 7, 2, 10, 0, 0, 0, time.UTC),
				SchedulerID: "cur-scheduler-id",
			},
			untilDate: time.Date(2022, 7, 10, 9, 0, 0, 0, time.UTC),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.On("GetLessonBySchedulerID", ctx, tx, "cur-scheduler-id").Return([]*domain.Lesson{
					{
						LessonID:  "lesson-id-1",
						StartTime: time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
							{
								LearnerID:        "user-id-2",
								CourseID:         "course-id-2",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
						},
					},
					{
						LessonID:  "lesson-id-2",
						StartTime: time.Date(2022, 7, 9, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.FamilyReason,
								AttendanceNote:   "attendance note",
							},
							{
								LearnerID:        "user-id-2",
								CourseID:         "course-id-2",
								AttendStatus:     domain.StudentAttendStatusAbsent,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoContact,
								AttendanceReason: domain.SchoolEvent,
								AttendanceNote:   "attendance note",
							},
						},
					},
				}, nil).Once()
				dateInfoRepo.
					On("GetDateInfoByDateRangeAndLocationID",
						ctx, tx, mock.Anything, mock.Anything, mock.Anything,
					).Return(dateInfos, nil).Once()
				schedulerRepo.On("GetByID", mock.Anything, tx, mock.Anything).Once().Return(&dto.Scheduler{SchedulerID: "schedulerId", StartDate: time.Now(), EndDate: time.Now()}, nil)
				mockSchedulerClient.On("UpdateScheduler", ctx, mock.Anything).Once().Return(&mpb.UpdateSchedulerResponse{}, nil)
				studentSubscriptionRepo.
					On("GetStudentCourseSubscriptions", ctx, tx, []string{selectedLesson.LocationID}, []string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"}).
					Return(
						user_domain.StudentSubscriptions{
							{
								StudentID: "user-id-1",
								CourseID:  "course-id-1",
								StartAt:   time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 9, 1, 9, 0, 0, 0, time.UTC),
							},
							{
								StudentID: "user-id-2",
								CourseID:  "course-id-2",
								StartAt:   time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 9, 1, 9, 0, 0, 0, time.UTC),
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
						time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(nil).
					Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-id-1", "media-id-2"}).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				expectedLessons := []*domain.Lesson{
					{
						LessonID: "lesson-id-1",
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
							{
								LearnerID:        "user-id-2",
								CourseID:         "course-id-2",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
						},
					},
					{
						LessonID: "lesson-id-2",
						// learner info of this lesson don't change
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.FamilyReason,
								AttendanceNote:   "attendance note",
							},
							{
								LearnerID:        "user-id-2",
								CourseID:         "course-id-2",
								AttendStatus:     domain.StudentAttendStatusAbsent,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoContact,
								AttendanceReason: domain.SchoolEvent,
								AttendanceNote:   "attendance note",
							},
						},
					},
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				reallocationRepo.
					On("GetByNewLessonID", ctx, tx, []string{}, mock.AnythingOfType("string")).
					Return([]*domain.Reallocation{}, nil).Once()
				lessonRepo.
					On("UpsertLessons", ctx, tx, mock.MatchedBy(func(recurringLesson *domain.RecurringLesson) bool {
						lessons := recurringLesson.Lessons
						for idx, ls := range lessons {
							expLesson := expectedLessons[idx]
							if len(expLesson.Learners) != len(ls.Learners) {
								return false
							}
							for idx, l := range ls.Learners {
								expLearner := expLesson.Learners[idx]
								if l.AttendStatus != expLearner.AttendStatus {
									return false
								}
								if l.AttendanceNote != expLearner.AttendanceNote {
									return false
								}
								if l.AttendanceNotice != expLearner.AttendanceNotice {
									return false
								}
								if l.AttendanceReason != expLearner.AttendanceReason {
									return false
								}
							}
						}
						return true
					})).Return([]string{"lesson-id-1", "lesson-id-2"}, nil).Once()
			},
		},
		{
			name:           "change student info save by this and following lesson",
			selectedLesson: selectedLesson,
			currentLesson: &domain.Lesson{
				LessonID:    "lesson-id-1",
				LocationID:  "center-id-1",
				StartTime:   time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
				EndTime:     time.Date(2022, 7, 2, 10, 0, 0, 0, time.UTC),
				SchedulerID: "cur-scheduler-id",
				Learners: domain.LessonLearners{
					{
						LearnerID:        "user-id-1",
						CourseID:         "course-id-1",
						AttendStatus:     domain.StudentAttendStatusAttend,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.NoticeEmpty,
						AttendanceReason: domain.ReasonEmpty,
					},
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-2",
						AttendStatus:     domain.StudentAttendStatusAttend,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.NoticeEmpty,
						AttendanceReason: domain.ReasonEmpty,
					},
				},
			},
			untilDate: time.Date(2022, 7, 24, 9, 0, 0, 0, time.UTC),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				dateInfoRepo.
					On("GetDateInfoByDateRangeAndLocationID",
						ctx, tx, mock.Anything, mock.Anything, mock.Anything,
					).Return(dateInfos, nil).Once()
				lessonRepo.On("GetLessonBySchedulerID", ctx, tx, "cur-scheduler-id").Return([]*domain.Lesson{
					{
						LessonID:  "lesson-id-1",
						StartTime: time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
							{
								LearnerID:        "user-id-3",
								CourseID:         "course-id-2",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
						},
					},
					{
						LessonID:  "lesson-id-2",
						StartTime: time.Date(2022, 7, 9, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.FamilyReason,
								AttendanceNote:   "attendance note",
							},
							{
								LearnerID:        "user-id-3",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.FamilyReason,
								AttendanceNote:   "attendance note",
							},
							{
								LearnerID:        "user-id-4",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusInformedLate,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.OnTheDay,
								AttendanceReason: domain.PhysicalCondition,
								AttendanceNote:   "attendance note",
							},
						},
					},
					{
						LessonID:  "lesson-id-3",
						StartTime: time.Date(2022, 7, 9, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.FamilyReason,
								AttendanceNote:   "attendance note",
							},
							{
								LearnerID:        "user-id-2",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAbsent,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.SchoolEvent,
								AttendanceNote:   "attendance note",
								Reallocate: &domain.Reallocate{
									OriginalLessonID: "original-lesson-id",
								},
							},
						},
					},
				}, nil).Once()
				schedulerRepo.On("GetByID", mock.Anything, tx, mock.Anything).Once().Return(&dto.Scheduler{SchedulerID: "schedulerId", StartDate: time.Now(), EndDate: time.Now()}, nil)
				mockSchedulerClient.On("UpdateScheduler", ctx, mock.Anything).Once().Return(&mpb.UpdateSchedulerResponse{}, nil)
				studentSubscriptionRepo.
					On("GetStudentCourseSubscriptions", ctx, tx, []string{selectedLesson.LocationID}, []string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"}).
					Return(
						user_domain.StudentSubscriptions{
							{
								StudentID: "user-id-1",
								CourseID:  "course-id-1",
								StartAt:   time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 9, 1, 9, 0, 0, 0, time.UTC),
							},
							{
								StudentID: "user-id-2",
								CourseID:  "course-id-2",
								StartAt:   time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 9, 1, 9, 0, 0, 0, time.UTC),
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
						time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(nil).
					Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-id-1", "media-id-2"}).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				expectedLessons := []*domain.Lesson{
					{
						LessonID: "lesson-id-1",
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
							{
								LearnerID:        "user-id-2",
								CourseID:         "course-id-2",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
						},
					},
					{
						LessonID: "lesson-id-2",
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.FamilyReason,
								AttendanceNote:   "attendance note",
							},
							{
								LearnerID:        "user-id-4",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusInformedLate,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.OnTheDay,
								AttendanceReason: domain.PhysicalCondition,
								AttendanceNote:   "attendance note",
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
						LessonID: "lesson-id-3",
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.FamilyReason,
								AttendanceNote:   "attendance note",
							},
							{
								LearnerID:        "user-id-2",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAbsent,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.SchoolEvent,
								AttendanceNote:   "attendance note",
								Reallocate: &domain.Reallocate{
									OriginalLessonID: "original-lesson-id",
								},
							},
						},
					},
					{
						LessonID: "lesson-id-4",
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
				}
				reallocationRepo.
					On("GetByNewLessonID", ctx, tx, mock.Anything, mock.AnythingOfType("string")).
					Return([]*domain.Reallocation{}, nil).Once()
				reallocationRepo.
					On("SoftDelete", ctx, tx, mock.Anything, false).Return(nil).Once()
				reallocationRepo.
					On("CancelIfStudentReallocated", ctx, tx, mock.Anything).Return(nil).Once()
				lessonRepo.
					On("UpsertLessons", ctx, tx, mock.MatchedBy(func(recurringLesson *domain.RecurringLesson) bool {
						lessons := recurringLesson.Lessons
						if len(expectedLessons) != len(lessons) {
							return false
						}
						for idx, ls := range lessons {
							expLesson := expectedLessons[idx]
							compareLearnerFunc := func(expectedLearner, gotLearner domain.LessonLearners) bool {
								if len(expectedLearner) != len(gotLearner) {
									return false
								}
								gotLearnerMap := make(map[string]*domain.LessonLearner, len(gotLearner))
								for _, ln := range gotLearner {
									gotLearnerMap[ln.LearnerID] = ln
								}
								for _, ln := range expectedLearner {
									if learner, ok := gotLearnerMap[ln.LearnerID]; ok {
										if *learner != *ln {
											return false
										}
									} else {
										return false
									}
								}
								return true
							}
							return compareLearnerFunc(expLesson.Learners, ls.Learners)
						}
						return true
					})).Return([]string{"lesson-id-1", "lesson-id-2"}, nil).Once()
			},
		},
		{
			name: "updated lesson time successfully",
			selectedLesson: &domain.Lesson{
				LessonID:         "lesson-id-2",
				StartTime:        time.Date(2022, 7, 9, 9, 0, 0, 0, time.UTC),
				EndTime:          time.Date(2022, 7, 9, 10, 0, 0, 0, time.UTC),
				CreatedAt:        now,
				UpdatedAt:        now,
				LocationID:       "center-id-1",
				CourseID:         "course-1",
				ClassID:          "class-1",
				TeachingMedium:   domain.LessonTeachingMediumOffline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				Learners: domain.LessonLearners{
					{
						LearnerID:        "user-id-1",
						CourseID:         "course-id-1",
						AttendStatus:     domain.StudentAttendStatusAttend,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.NoticeEmpty,
						AttendanceReason: domain.ReasonEmpty,
					},
					{
						LearnerID:        "user-id-2",
						CourseID:         "course-id-2",
						AttendStatus:     domain.StudentAttendStatusAttend,
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
			},
			currentLesson: &domain.Lesson{
				LessonID:    "lesson-id-2",
				LocationID:  "center-1",
				CourseID:    "course-1",
				ClassID:     "class-1",
				StartTime:   time.Date(2022, 7, 8, 9, 0, 0, 0, time.UTC),
				EndTime:     time.Date(2022, 7, 8, 10, 0, 0, 0, time.UTC),
				SchedulerID: "cur-scheduler-id",
			},
			untilDate: time.Date(2022, 7, 16, 10, 0, 0, 0, time.UTC),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonDeleted, mock.Anything).Once().Return("", nil)
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				dateInfoRepo.
					On("GetDateInfoByDateRangeAndLocationID",
						ctx, tx, mock.Anything, mock.Anything, mock.Anything,
					).Return(dateInfos, nil).Once()
				lessonRepo.On("GetLessonBySchedulerID", ctx, tx, "cur-scheduler-id").Return([]*domain.Lesson{
					{
						LessonID: "lesson-id-1",
						EndTime:  time.Date(2022, 7, 1, 10, 0, 0, 0, time.UTC),
					},
					{
						LessonID:  "lesson-id-2",
						StartTime: time.Date(2022, 7, 8, 9, 0, 0, 0, time.UTC),
						EndTime:   time.Date(2022, 7, 8, 10, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
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
						SchedulingStatus: domain.LessonSchedulingStatusDraft,
					},
					{
						LessonID:  "lesson-id-3",
						StartTime: time.Date(2022, 7, 15, 9, 0, 0, 0, time.UTC),
						EndTime:   time.Date(2022, 7, 15, 10, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:    "user-id-1",
								CourseID:     "course-id-1",
								AttendStatus: domain.StudentAttendStatusAttend,
								LocationID:   "center-id-1",
							},
							{
								LearnerID:    "user-id-2",
								CourseID:     "course-id-2",
								AttendStatus: domain.StudentAttendStatusEmpty,
								LocationID:   "center-id-1",
							},
						},
						SchedulingStatus: domain.LessonSchedulingStatusDraft,
					},
				}, nil).Once()
				schedulerRepo.On("GetByID", mock.Anything, tx, mock.Anything).Once().Return(&dto.Scheduler{SchedulerID: "schedulerId", StartDate: time.Now(), EndDate: time.Now()}, nil)

				mockSchedulerClient.On("UpdateScheduler", ctx, &mpb.UpdateSchedulerRequest{
					SchedulerId: "cur-scheduler-id",
					EndDate:     timestamppb.New(time.Date(2022, 7, 1, 10, 0, 0, 0, time.UTC)),
				}).Return(&mpb.UpdateSchedulerResponse{}, nil).Once()

				newScheduler := calendar_entities.NewScheduler(
					time.Date(2022, 7, 9, 9, 0, 0, 0, time.UTC),
					time.Date(2022, 7, 16, 10, 0, 0, 0, time.UTC),
					"weekly",
					schedulerRepo,
				)

				mockSchedulerClient.On("CreateScheduler", ctx, mock.MatchedBy(func(sch *mpb.CreateSchedulerRequest) bool {
					if !sch.StartDate.AsTime().Equal(newScheduler.StartDate) {
						return false
					}
					if !sch.EndDate.AsTime().Equal(newScheduler.EndDate) {
						return false
					}
					return true
				})).Return(&mpb.CreateSchedulerResponse{
					SchedulerId: "new-scheduler-id",
				}, nil).Once()

				studentSubscriptionRepo.
					On("GetStudentCourseSubscriptions", ctx, tx, []string{"center-id-1"}, []string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"}).
					Return(
						user_domain.StudentSubscriptions{
							{
								StudentID: "user-id-1",
								CourseID:  "course-id-1",
								StartAt:   time.Date(2022, 6, 26, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 7, 30, 9, 0, 0, 0, time.UTC),
							},
							{
								StudentID: "user-id-2",
								CourseID:  "course-id-2",
								StartAt:   time.Date(2022, 6, 26, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 7, 30, 9, 0, 0, 0, time.UTC),
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
						time.Date(2022, 7, 9, 9, 0, 0, 0, time.UTC),
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
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				expectedLessons := []*domain.Lesson{
					{
						StartTime:        time.Date(2022, 7, 9, 9, 0, 0, 0, time.UTC),
						EndTime:          time.Date(2022, 7, 9, 10, 0, 0, 0, time.UTC),
						SchedulerID:      "new-scheduler-id",
						SchedulingStatus: domain.LessonSchedulingStatusPublished,
					},
					{
						StartTime:        time.Date(2022, 7, 16, 9, 0, 0, 0, time.UTC),
						EndTime:          time.Date(2022, 7, 16, 10, 0, 0, 0, time.UTC),
						SchedulerID:      "new-scheduler-id",
						SchedulingStatus: domain.LessonSchedulingStatusPublished,
					},
				}
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
							if expLesson.SchedulerID != ls.SchedulerID {
								return false
							}
							if expLesson.SchedulingStatus != ls.SchedulingStatus {
								return false
							}
						}
						return true
					})).Return([]string{"lesson-id-1", "lesson-id-2"}, nil).Once()
				reallocationRepo.On("GetByNewLessonID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]*domain.Reallocation{}, nil)
			},
		},
		{
			name:           "updated lesson weekly with the lesson is locked successfully",
			selectedLesson: selectedLesson,
			untilDate:      time.Date(2022, 7, 14, 9, 0, 0, 0, time.UTC),
			currentLesson: &domain.Lesson{
				LessonID:    "lesson-id-1",
				LocationID:  "center-1",
				CourseID:    "course-1",
				ClassID:     "class-1",
				StartTime:   time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
				EndTime:     time.Date(2022, 7, 1, 10, 0, 0, 0, time.UTC),
				SchedulerID: "cur-scheduler-id",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				dateInfoRepo.
					On("GetDateInfoByDateRangeAndLocationID",
						ctx, tx, mock.Anything, mock.Anything, mock.Anything,
					).Return(dateInfos, nil).Once()
				lessonRepo.On("GetLessonBySchedulerID", ctx, tx, "cur-scheduler-id").Return([]*domain.Lesson{
					{
						LessonID:  "lesson-id-1",
						StartTime: time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
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
						LessonID:  "lesson-id-2",
						StartTime: time.Date(2022, 7, 8, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
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
						IsLocked: true,
					}, // util lesson-id-3
					{
						LessonID:  "lesson-id-3",
						StartTime: time.Date(2022, 7, 15, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
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
						LessonID:  "lesson-id-4",
						StartTime: time.Date(2022, 7, 22, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
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
						IsLocked: true,
					},
				}, nil).Once()
				schedulerRepo.On("GetByID", mock.Anything, tx, mock.Anything).Once().Return(&dto.Scheduler{SchedulerID: "schedulerId", StartDate: time.Now(), EndDate: time.Now()}, nil)
				mockSchedulerClient.On("UpdateScheduler", ctx, mock.Anything).Once().Return(&mpb.UpdateSchedulerResponse{}, nil).Once()
				newScheduler := calendar_entities.NewScheduler(
					time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
					time.Date(2022, 7, 14, 9, 0, 0, 0, time.UTC),
					"weekly",
					schedulerRepo,
				)
				schedulerId := "new-scheduler-id"

				mockSchedulerClient.On("CreateScheduler", ctx, mock.MatchedBy(func(sch *mpb.CreateSchedulerRequest) bool {
					if !sch.StartDate.AsTime().Equal(newScheduler.StartDate) {
						return false
					}
					if !sch.EndDate.AsTime().Equal(newScheduler.EndDate) {
						return false
					}
					return true
				})).Once().Return(&mpb.CreateSchedulerResponse{
					SchedulerId: schedulerId,
				}, nil)

				lockedLessonIDs := []string{"lesson-id-2", "lesson-id-4"}
				lessonRepo.On("UpdateSchedulerID", mock.Anything, mock.Anything, lockedLessonIDs, schedulerId).Once().Return(nil)

				remainingLessonIDs := []string{"lesson-id-3"}
				lessonRepo.On("Delete", mock.Anything, mock.Anything, remainingLessonIDs).Once().Return(nil)

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
						time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
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
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				expectedLessons := []*domain.Lesson{
					selectedLesson,
					{
						StartTime:   time.Date(2022, 7, 9, 9, 0, 0, 0, time.UTC),
						EndTime:     time.Date(2022, 7, 9, 10, 0, 0, 0, time.UTC),
						SchedulerID: "new-scheduler-id",
					},
				}
				studentSubscriptionRepo.
					On("GetStudentCourseSubscriptions", ctx, tx, mock.Anything, []string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"}).
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
								StartAt:   time.Date(2022, 7, 9, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 9, 20, 9, 0, 0, 0, time.UTC),
							},
						}, nil).
					Once()
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
							if expLesson.SchedulerID != ls.SchedulerID {
								return false
							}
						}
						return true
					})).Return([]string{"lesson-id-1", "lesson-id-2"}, nil).Once()
			},
		},
		{
			name:           `no change lesson info,some student course not aligned with lesson date`,
			selectedLesson: selectedLesson,
			currentLesson: &domain.Lesson{
				LessonID:    "lesson-id-1",
				LocationID:  "center-id-1",
				StartTime:   time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
				EndTime:     time.Date(2022, 7, 2, 10, 0, 0, 0, time.UTC),
				SchedulerID: "cur-scheduler-id",
			},
			untilDate: time.Date(2022, 7, 10, 9, 0, 0, 0, time.UTC),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonBySchedulerID", ctx, tx, "cur-scheduler-id").Return([]*domain.Lesson{
					{
						LessonID:  "lesson-id-1",
						StartTime: time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
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
						LessonID:  "lesson-id-2",
						StartTime: time.Date(2022, 7, 8, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
						},
					},
				}, nil).Once()
				dateInfoRepo.
					On("GetDateInfoByDateRangeAndLocationID",
						ctx, tx, mock.Anything, mock.Anything, mock.Anything,
					).Return(dateInfos, nil).Once()
				schedulerRepo.On("GetByID", mock.Anything, tx, mock.Anything).Once().Return(&dto.Scheduler{SchedulerID: "schedulerId", StartDate: time.Now(), EndDate: time.Now()}, nil)
				mockSchedulerClient.On("UpdateScheduler", ctx, mock.Anything).Once().Return(&mpb.UpdateSchedulerResponse{}, nil)

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
						time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(nil).
					Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-id-1", "media-id-2"}).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()

				expectedLessons := []*domain.Lesson{
					{
						Learners:    selectedLesson.Learners,
						SchedulerID: "cur-scheduler-id",
					},
					//only 1 student add to 2nd lesson
					{
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
						SchedulerID: "cur-scheduler-id",
					},
				}
				studentSubscriptionRepo.
					On("GetStudentCourseSubscriptions", ctx, tx, []string{selectedLesson.LocationID}, []string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"}).
					Return(
						user_domain.StudentSubscriptions{
							{
								StudentID: "user-id-1",
								CourseID:  "course-id-1",
								StartAt:   time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 7, 10, 9, 0, 0, 0, time.UTC),
							},
							{
								StudentID: "user-id-2",
								CourseID:  "course-id-2",
								StartAt:   time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 7, 8, 9, 0, 0, 0, time.UTC),
							},
						}, nil).
					Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Once()
				lessonRepo.
					On("UpsertLessons", ctx, tx, mock.MatchedBy(func(recurringLesson *domain.RecurringLesson) bool {
						lessons := recurringLesson.Lessons
						for idx, ls := range lessons {
							expLesson := expectedLessons[idx]
							if len(expLesson.Learners) != len(ls.Learners) {
								return false
							}
							if expLesson.SchedulerID != ls.SchedulerID {
								return false
							}
						}
						return true
					})).Return([]string{"lesson-id-1", "lesson-id-2"}, nil).Once()
			},
		},
		{
			name:           "no change status for following status if selected lesson have completed or cancelled status",
			selectedLesson: selectedLesson,
			currentLesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				LocationID:       "center-id-1",
				StartTime:        time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
				EndTime:          time.Date(2022, 7, 2, 10, 0, 0, 0, time.UTC),
				SchedulerID:      "cur-scheduler-id",
				SchedulingStatus: domain.LessonSchedulingStatusCompleted,
			},
			untilDate: time.Date(2022, 7, 10, 9, 0, 0, 0, time.UTC),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				dateInfoRepo.
					On("GetDateInfoByDateRangeAndLocationID",
						ctx, tx, mock.Anything, mock.Anything, mock.Anything,
					).Return(dateInfos, nil).Once()
				lessonRepo.On("GetLessonBySchedulerID", ctx, tx, "cur-scheduler-id").Return([]*domain.Lesson{
					{
						LessonID:  "lesson-id-1",
						StartTime: time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
							{
								LearnerID:        "user-id-2",
								CourseID:         "course-id-2",
								AttendStatus:     domain.StudentAttendStatusAttend,
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
						},
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
					},
					{
						LessonID:  "lesson-id-2",
						StartTime: time.Date(2022, 7, 9, 9, 0, 0, 0, time.UTC),
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
						SchedulingStatus: domain.LessonSchedulingStatusPublished,
					},
				}, nil).Once()
				schedulerRepo.On("GetByID", mock.Anything, tx, mock.Anything).Once().Return(&dto.Scheduler{SchedulerID: "schedulerId", StartDate: time.Now(), EndDate: time.Now()}, nil)
				mockSchedulerClient.On("UpdateScheduler", ctx, mock.Anything).Once().Return(&mpb.UpdateSchedulerResponse{}, nil)

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
						time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(nil).
					Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-id-1", "media-id-2"}).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()

				expectedLessons := []*domain.Lesson{
					{
						SchedulerID:      "cur-scheduler-id",
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
					},
					{
						SchedulerID:      "cur-scheduler-id",
						SchedulingStatus: domain.LessonSchedulingStatusPublished,
					},
				}
				studentSubscriptionRepo.
					On("GetStudentCourseSubscriptions", ctx, tx, []string{selectedLesson.LocationID}, []string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"}).
					Return(
						user_domain.StudentSubscriptions{
							{
								StudentID: "user-id-1",
								CourseID:  "course-id-1",
								StartAt:   time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 7, 26, 9, 0, 0, 0, time.UTC),
							},
							{
								StudentID: "user-id-2",
								CourseID:  "course-id-2",
								StartAt:   time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 7, 26, 9, 0, 0, 0, time.UTC),
							},
						}, nil).
					Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				lessonRepo.
					On("UpsertLessons", ctx, tx, mock.MatchedBy(func(recurringLesson *domain.RecurringLesson) bool {
						lessons := recurringLesson.Lessons
						for idx, ls := range lessons {
							expLesson := expectedLessons[idx]
							if expLesson.SchedulingStatus != ls.SchedulingStatus {
								return false
							}
							if expLesson.SchedulerID != ls.SchedulerID {
								return false
							}
						}
						return true
					})).Return([]string{"lesson-id-1", "lesson-id-2"}, nil).Once()
				reallocationRepo.On("GetByNewLessonID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]*domain.Reallocation{}, nil)
			},
		},
		{
			name:           `update lesson date with first day is closed date`,
			selectedLesson: selectedLesson,
			currentLesson: &domain.Lesson{
				LessonID:    "lesson-id-1",
				LocationID:  "center-id-1",
				StartTime:   time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
				EndTime:     time.Date(2022, 7, 2, 10, 0, 0, 0, time.UTC),
				SchedulerID: "cur-scheduler-id",
			},
			untilDate: time.Date(2022, 7, 10, 9, 0, 0, 0, time.UTC),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonBySchedulerID", ctx, tx, "cur-scheduler-id").Return([]*domain.Lesson{
					{
						LessonID:  "lesson-id-1",
						StartTime: time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:    "user-id-1",
								CourseID:     "course-id-1",
								AttendStatus: domain.StudentAttendStatusAttend,
								LocationID:   "center-id-1",
							},
							{
								LearnerID:    "user-id-2",
								CourseID:     "course-id-2",
								AttendStatus: domain.StudentAttendStatusEmpty,
								LocationID:   "center-id-1",
							},
						},
					},
					{
						LessonID:  "lesson-id-2",
						StartTime: time.Date(2022, 7, 8, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:    "user-id-1",
								CourseID:     "course-id-1",
								AttendStatus: domain.StudentAttendStatusAttend,
								LocationID:   "center-id-1",
							},
						},
					},
				}, nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, tx, mock.Anything, mock.Anything, mock.Anything,
				).Return([]*calendar_dto.DateInfo{
					{
						Date:        time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
						LocationID:  "center-id-1",
						DateTypeID:  string(calendar_constants.ClosedDay),
						OpeningTime: "",
						Status:      string(calendar_constants.Draft),
					},
				}, nil).Once()
				schedulerRepo.On("GetByID", mock.Anything, tx, mock.Anything).Once().Return(&dto.Scheduler{SchedulerID: "schedulerId", StartDate: time.Now(), EndDate: time.Now()}, nil)
				mockSchedulerClient.On("UpdateScheduler", ctx, mock.Anything).Once().Return(&mpb.UpdateSchedulerResponse{}, nil)

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
						time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(nil).
					Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-id-1", "media-id-2"}).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()

				studentSubscriptionRepo.
					On("GetStudentCourseSubscriptions", ctx, tx, []string{selectedLesson.LocationID}, []string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"}).
					Return(
						user_domain.StudentSubscriptions{
							{
								StudentID: "user-id-1",
								CourseID:  "course-id-1",
								StartAt:   time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 7, 10, 9, 0, 0, 0, time.UTC),
							},
							{
								StudentID: "user-id-2",
								CourseID:  "course-id-2",
								StartAt:   time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 7, 8, 9, 0, 0, 0, time.UTC),
							},
						}, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				expectedLessons := []*domain.Lesson{
					{
						Learners:    selectedLesson.Learners,
						SchedulerID: "cur-scheduler-id",
					},
					{
						Learners: domain.LessonLearners{
							{
								LearnerID:    "user-id-1",
								CourseID:     "course-id-1",
								AttendStatus: domain.StudentAttendStatusEmpty,
								LocationID:   "center-id-1",
							},
						},
						SchedulerID: "cur-scheduler-id"},
				}
				lessonRepo.
					On("UpsertLessons", ctx, tx, mock.MatchedBy(func(recurringLesson *domain.RecurringLesson) bool {
						lessons := recurringLesson.Lessons
						for idx, ls := range lessons {
							expLesson := expectedLessons[idx]
							if len(expLesson.Learners) != len(ls.Learners) {
								return false
							}
							if expLesson.SchedulerID != ls.SchedulerID {
								return false
							}
						}
						return true
					})).Return(nil, errors.New("could not create lesson in closed date")).Once()
			},
			hasError: true,
		},
		{
			name:           "change student info and course teaching time save by this and following lesson",
			selectedLesson: selectedLesson,
			currentLesson: &domain.Lesson{
				LessonID:    "lesson-id-1",
				LocationID:  "center-id-1",
				StartTime:   time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
				EndTime:     time.Date(2022, 7, 2, 10, 0, 0, 0, time.UTC),
				SchedulerID: "cur-scheduler-id",
				Learners: domain.LessonLearners{
					{
						LearnerID:        "user-id-1",
						CourseID:         "course-id-1",
						AttendStatus:     domain.StudentAttendStatusAttend,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.NoticeEmpty,
						AttendanceReason: domain.ReasonEmpty,
					},
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-2",
						AttendStatus:     domain.StudentAttendStatusAttend,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.NoticeEmpty,
						AttendanceReason: domain.ReasonEmpty,
					},
				},
			},
			untilDate: time.Date(2022, 7, 24, 9, 0, 0, 0, time.UTC),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				dateInfoRepo.
					On("GetDateInfoByDateRangeAndLocationID",
						ctx, tx, mock.Anything, mock.Anything, mock.Anything,
					).Return(dateInfos, nil).Once()
				lessonRepo.On("GetLessonBySchedulerID", ctx, tx, "cur-scheduler-id").Return([]*domain.Lesson{
					{
						LessonID:  "lesson-id-1",
						StartTime: time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
							{
								LearnerID:        "user-id-3",
								CourseID:         "course-id-2",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
						},
					},
					{
						LessonID:  "lesson-id-2",
						StartTime: time.Date(2022, 7, 9, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.FamilyReason,
								AttendanceNote:   "attendance note",
							},
							{
								LearnerID:        "user-id-3",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.FamilyReason,
								AttendanceNote:   "attendance note",
							},
							{
								LearnerID:        "user-id-4",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusInformedLate,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.OnTheDay,
								AttendanceReason: domain.PhysicalCondition,
								AttendanceNote:   "attendance note",
							},
						},
					},
					{
						LessonID:  "lesson-id-3",
						StartTime: time.Date(2022, 7, 9, 9, 0, 0, 0, time.UTC),
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.FamilyReason,
								AttendanceNote:   "attendance note",
							},
							{
								LearnerID:        "user-id-2",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAbsent,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.SchoolEvent,
								AttendanceNote:   "attendance note",
								Reallocate: &domain.Reallocate{
									OriginalLessonID: "original-lesson-id",
								},
							},
						},
					},
				}, nil).Once()
				schedulerRepo.On("GetByID", mock.Anything, tx, mock.Anything).Once().Return(&dto.Scheduler{SchedulerID: "schedulerId", StartDate: time.Now(), EndDate: time.Now()}, nil)
				mockSchedulerClient.On("UpdateScheduler", ctx, mock.Anything).Once().Return(&mpb.UpdateSchedulerResponse{}, nil)
				studentSubscriptionRepo.
					On("GetStudentCourseSubscriptions", ctx, tx, []string{selectedLesson.LocationID}, []string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"}).
					Return(
						user_domain.StudentSubscriptions{
							{
								StudentID: "user-id-1",
								CourseID:  "course-id-1",
								StartAt:   time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 9, 1, 9, 0, 0, 0, time.UTC),
							},
							{
								StudentID: "user-id-2",
								CourseID:  "course-id-2",
								StartAt:   time.Date(2022, 7, 1, 9, 0, 0, 0, time.UTC),
								EndAt:     time.Date(2022, 9, 1, 9, 0, 0, 0, time.UTC),
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
						time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(nil).
					Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-id-1", "media-id-2"}).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
				classroomRepo.
					On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				expectedLessons := []*domain.Lesson{
					{
						LessonID: "lesson-id-1",
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
							{
								LearnerID:        "user-id-2",
								CourseID:         "course-id-2",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.NoticeEmpty,
								AttendanceReason: domain.ReasonEmpty,
							},
						},
					},
					{
						LessonID: "lesson-id-2",
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.FamilyReason,
								AttendanceNote:   "attendance note",
							},
							{
								LearnerID:        "user-id-4",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusInformedLate,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.OnTheDay,
								AttendanceReason: domain.PhysicalCondition,
								AttendanceNote:   "attendance note",
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
						LessonID: "lesson-id-3",
						Learners: domain.LessonLearners{
							{
								LearnerID:        "user-id-1",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAttend,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.FamilyReason,
								AttendanceNote:   "attendance note",
							},
							{
								LearnerID:        "user-id-2",
								CourseID:         "course-id-1",
								AttendStatus:     domain.StudentAttendStatusAbsent,
								LocationID:       "center-id-1",
								AttendanceNotice: domain.InAdvance,
								AttendanceReason: domain.SchoolEvent,
								AttendanceNote:   "attendance note",
								Reallocate: &domain.Reallocate{
									OriginalLessonID: "original-lesson-id",
								},
							},
						},
					},
					{
						LessonID: "lesson-id-4",
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
				}
				reallocationRepo.
					On("SoftDelete", ctx, tx, mock.Anything, false).Return(nil).Once()
				reallocationRepo.
					On("CancelIfStudentReallocated", ctx, tx, mock.Anything).Return(nil).Once()
				lessonRepo.
					On("UpsertLessons", ctx, tx, mock.MatchedBy(func(recurringLesson *domain.RecurringLesson) bool {
						lessons := recurringLesson.Lessons
						if len(expectedLessons) != len(lessons) {
							return false
						}
						for idx, ls := range lessons {
							expLesson := expectedLessons[idx]
							compareLearnerFunc := func(expectedLearner, gotLearner domain.LessonLearners) bool {
								if len(expectedLearner) != len(gotLearner) {
									return false
								}
								gotLearnerMap := make(map[string]*domain.LessonLearner, len(gotLearner))
								for _, ln := range gotLearner {
									gotLearnerMap[ln.LearnerID] = ln
								}
								for _, ln := range expectedLearner {
									if learner, ok := gotLearnerMap[ln.LearnerID]; ok {
										if *learner != *ln {
											return false
										}
									} else {
										return false
									}
								}
								return true
							}
							return compareLearnerFunc(expLesson.Learners, ls.Learners)
						}
						return true
					})).Return([]string{"lesson-id-1", "lesson-id-2"}, nil).Once()
			},
		},
	}

	handler := LessonCommandHandler{
		WrapperConnection:       wrapperConnection,
		LessonRepo:              lessonRepo,
		SchedulerRepo:           schedulerRepo,
		StudentSubscriptionRepo: studentSubscriptionRepo,
		UnleashClientIns:        mockUnleashClient,
		LessonProducer: producers.LessonProducer{
			JSM: jsm,
		},
		ReallocationRepo: reallocationRepo,
		SchedulerClient:  mockSchedulerClient,
		MasterDataPort:   masterDataRepo,
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			upsertedLesson, _, err := handler.UpdateRecurringLesson(ctx, UpdateRecurringLessonCommandRequest{
				SelectedLesson: tc.selectedLesson,
				CurrentLesson:  tc.currentLesson,
				RRuleCmd: RecurrenceRuleCommand{
					StartTime: tc.selectedLesson.StartTime,
					EndTime:   tc.selectedLesson.EndTime,
					UntilDate: tc.untilDate,
				},
			})
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, upsertedLesson)
			}
			mock.AssertExpectationsForObjects(
				t,
				db,
				tx,
				masterDataRepo,
				userModuleAdapter,
				mediaModulePort,
				lessonRepo,
				courseRepo,
				dateInfoRepo,
				mockUnleashClient,
			)
		})
	}
}

func TestUpdateLessonCommandHandler_UpdateSchedulingStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := new(mock_repositories.MockLessonRepo)
	mockZoomService := &mock_service.MockZoomService{}
	lesson := &domain.Lesson{
		LessonID:  "lesson-id-1",
		StartTime: time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
	}
	testCases := []struct {
		name       string
		newStatus  string
		setup      func(ctx context.Context)
		lesson     *domain.Lesson
		savingType lpb.SavingType
		hasError   bool
	}{
		{
			name:      "change from published to canceled",
			newStatus: string(domain.LessonSchedulingStatusPublished),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("GetLessonWithSchedulerInfoByLessonID", mock.Anything, db, mock.AnythingOfType("string")).Return(lesson, nil).Once()
				lesson.SchedulingStatus = domain.LessonSchedulingStatusCanceled
				lessonRepo.
					On("UpdateLessonSchedulingStatus", mock.Anything, db, lesson).Return(lesson, nil).Once()
			},
		},
		{
			name:      "should call api remove zoom when cancel a lesson once time have zoom ID",
			newStatus: string(domain.LessonSchedulingStatusCanceled),
			setup: func(ctx context.Context) {
				zoomID := "zoom-id"
				lesson_zoom := &domain.Lesson{
					LessonID:         "lesson-id-1",
					StartTime:        time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
					EndTime:          time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
					SchedulingStatus: domain.LessonSchedulingStatusPublished,
					SchedulerInfo: &domain.SchedulerInfo{
						SchedulerID: "schedulerID",
						Freq:        "once",
					},
					ZoomID: zoomID,
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("GetLessonWithSchedulerInfoByLessonID", mock.Anything, db, mock.AnythingOfType("string")).Return(lesson_zoom, nil).Once()
				mockZoomService.
					On("RetryDeleteZoomLink", mock.Anything, zoomID).Return(true, nil).Once()
				lessonRepo.
					On("RemoveZoomLinkByLessonID", mock.Anything, db, "lesson-id-1").Return(nil).Once()

				lessonRepo.
					On("UpdateLessonSchedulingStatus", mock.Anything, db, lesson_zoom).Return(lesson_zoom, nil).Once()
			},
		},
		{
			name:      "change from canceled to published",
			newStatus: string(domain.LessonSchedulingStatusCanceled),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("GetLessonWithSchedulerInfoByLessonID", mock.Anything, db, mock.AnythingOfType("string")).Return(lesson, nil).Once()
				lesson.SchedulingStatus = domain.LessonSchedulingStatusPublished
				lessonRepo.
					On("UpdateLessonSchedulingStatus", mock.Anything, db, lesson).Return(lesson, nil).Once()
			},
		},
		{
			name:      "cannot change from canceled to completed",
			newStatus: string(domain.LessonSchedulingStatusCompleted),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("GetLessonWithSchedulerInfoByLessonID", mock.Anything, db, mock.AnythingOfType("string")).Return(lesson, nil).Once()
				lesson.SchedulingStatus = domain.LessonSchedulingStatusCanceled
			},
			hasError: true,
		},
		{
			name:      "cannot change from canceled to canceled",
			newStatus: string(domain.LessonSchedulingStatusCanceled),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("GetLessonWithSchedulerInfoByLessonID", mock.Anything, db, mock.AnythingOfType("string")).Return(lesson, nil).Once()
				lesson.SchedulingStatus = domain.LessonSchedulingStatusCanceled
			},
			hasError: true,
		},
		{
			name:       "change status from draft to published by saving this and following",
			newStatus:  string(domain.LessonSchedulingStatusPublished),
			savingType: lpb.SavingType_THIS_AND_FOLLOWING,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("GetLessonWithSchedulerInfoByLessonID", mock.Anything, db, mock.AnythingOfType("string")).Return(&domain.Lesson{
					LessonID:         "lesson-id-1",
					StartTime:        time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
					EndTime:          time.Date(2022, 7, 2, 10, 0, 0, 0, time.UTC),
					SchedulingStatus: domain.LessonSchedulingStatusDraft,
					SchedulerID:      "cur-scheduler-id",
				}, nil).Once()
				lessonRepo.On("GetLessonBySchedulerID", ctx, db, "cur-scheduler-id").Return([]*domain.Lesson{
					{
						LessonID:         "lesson-id-1",
						StartTime:        time.Date(2022, 7, 2, 9, 0, 0, 0, time.UTC),
						EndTime:          time.Date(2022, 7, 2, 10, 0, 0, 0, time.UTC),
						SchedulingStatus: domain.LessonSchedulingStatusDraft,
					},
					{
						LessonID:         "lesson-id-2",
						StartTime:        time.Date(2022, 7, 9, 9, 0, 0, 0, time.UTC),
						EndTime:          time.Date(2022, 7, 9, 10, 0, 0, 0, time.UTC),
						SchedulingStatus: domain.LessonSchedulingStatusCanceled,
					},
					{
						LessonID:         "lesson-id-3",
						StartTime:        time.Date(2022, 7, 16, 9, 0, 0, 0, time.UTC),
						EndTime:          time.Date(2022, 7, 16, 10, 0, 0, 0, time.UTC),
						SchedulingStatus: domain.LessonSchedulingStatusCompleted,
					},
					{
						LessonID:         "lesson-id-4",
						StartTime:        time.Date(2022, 7, 23, 9, 0, 0, 0, time.UTC),
						EndTime:          time.Date(2022, 7, 23, 10, 0, 0, 0, time.UTC),
						SchedulingStatus: domain.LessonSchedulingStatusPublished,
					},
				}, nil).Once()
				lessonStatus := map[string]domain.LessonSchedulingStatus{
					"lesson-id-1": domain.LessonSchedulingStatusPublished,
				}
				lessonRepo.On("UpdateSchedulingStatus", ctx, db, lessonStatus).Return(nil).Once()
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := LessonCommandHandler{
				WrapperConnection: wrapperConnection,
				LessonRepo:        lessonRepo,
				ZoomService:       mockZoomService,
			}
			res, err := handler.UpdateLessonSchedulingStatus(ctx, &UpdateLessonStatusCommandRequest{
				LessonID:         "lesson-id",
				SchedulingStatus: tc.newStatus,
				SavingType:       tc.savingType,
			})
			if tc.hasError {
				require.Error(t, err)
				require.Nil(t, res)
			} else {
				require.NoError(t, err)
				mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, mockUnleashClient)
			}
		})
	}
}

func TestUpdateLessonCommandHandler_BulkUpdateSchedulingStatus_Cancel(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := new(mock_repositories.MockLessonRepo)
	lessonListCancel := []*domain.Lesson{
		{
			LessonID:         "lesson-id-1",
			StartTime:        time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
			EndTime:          time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
			SchedulingStatus: domain.LessonSchedulingStatusPublished,
		},
		{
			LessonID:         "lesson-id-2",
			StartTime:        time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
			EndTime:          time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
			SchedulingStatus: domain.LessonSchedulingStatusCompleted,
		},
	}
	lessonListPublish := []*domain.Lesson{
		{
			LessonID:         "lesson-id-3",
			StartTime:        time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
			EndTime:          time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
			SchedulingStatus: domain.LessonSchedulingStatusDraft,
			TeachingMethod:   domain.LessonTeachingMethodGroup,
			Teachers:         domain.LessonTeachers{&domain.LessonTeacher{TeacherID: "test-teacher-id-1"}},
			CourseID:         "test-course-id-1",
		},
		{
			LessonID:         "lesson-id-4",
			StartTime:        time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC),
			EndTime:          time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC),
			SchedulingStatus: domain.LessonSchedulingStatusDraft,
			TeachingMethod:   domain.LessonTeachingMethodIndividual,
			Teachers:         domain.LessonTeachers{&domain.LessonTeacher{TeacherID: "test-teacher-id-1"}},
		},
	}
	testCases := []struct {
		name     string
		req      BulkUpdateLessonSchedulingStatusCommandRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "update successfully with action cancel",
			req: BulkUpdateLessonSchedulingStatusCommandRequest{
				Action:  lpb.LessonBulkAction_LESSON_BULK_ACTION_CANCEL,
				Lessons: lessonListCancel,
			},
			setup: func(ctx context.Context) {
				lessonStatus := map[string]domain.LessonSchedulingStatus{
					"lesson-id-1": domain.LessonSchedulingStatusCanceled,
					"lesson-id-2": domain.LessonSchedulingStatusCanceled,
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("UpdateSchedulingStatus", mock.Anything, db, lessonStatus).Return(nil).Once()
			},
			hasError: false,
		},
		{
			name: "update successfully with action publish",
			req: BulkUpdateLessonSchedulingStatusCommandRequest{
				Action:  lpb.LessonBulkAction_LESSON_BULK_ACTION_PUBLISH,
				Lessons: lessonListPublish,
			},
			setup: func(ctx context.Context) {
				lessonStatus := map[string]domain.LessonSchedulingStatus{
					"lesson-id-3": domain.LessonSchedulingStatusPublished,
					"lesson-id-4": domain.LessonSchedulingStatusPublished,
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("UpdateSchedulingStatus", mock.Anything, db, lessonStatus).Return(nil).Once()
			},
			hasError: false,
		},
		{
			name: "update failed with invalid action",
			req: BulkUpdateLessonSchedulingStatusCommandRequest{
				Action:  10000,
				Lessons: lessonListCancel,
			},
			setup: func(ctx context.Context) {
				lessonStatus := map[string]domain.LessonSchedulingStatus{
					"lesson-id-1": domain.LessonSchedulingStatusCanceled,
					"lesson-id-2": domain.LessonSchedulingStatusCanceled,
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.
					On("UpdateSchedulingStatus", mock.Anything, db, lessonStatus).Return().Once()
			},
			hasError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := LessonCommandHandler{
				WrapperConnection: wrapperConnection,
				LessonRepo:        lessonRepo,
			}
			res, err := handler.BulkUpdateLessonSchedulingStatus(ctx, &tc.req)
			if tc.hasError {
				require.Error(t, err)
				require.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, mockUnleashClient)
			}
		})
	}
}

func TestUpdateLessonCommandHandler_MarkStudentAsReallocate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonId, studentId := "lessonId", "studentId"
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	reallocationRepo := new(mock_repositories.MockReallocationRepo)
	req := &MarkStudentAsReallocateRequest{
		Member: &domain.LessonMember{
			LessonID:         lessonId,
			StudentID:        studentId,
			AttendanceStatus: string(lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_REALLOCATE),
		},
		ReAllocations: &domain.Reallocation{
			OriginalLessonID: lessonId,
			StudentID:        studentId,
		},
	}
	testCases := []struct {
		name       string
		newStatus  string
		setup      func(ctx context.Context)
		lesson     *domain.Lesson
		savingType lpb.SavingType
		hasError   bool
	}{
		{
			name: "Query lesson member failed",
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.
					On("FindByID", ctx, tx, lessonId, studentId).Return(nil, errors.New("failed")).Once()
			},
			hasError: true,
		},
		{
			name: "Student's attendance status is not absent",
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.
					On("FindByID", ctx, tx, lessonId, studentId).Return(&domain.LessonMember{
					AttendanceStatus: string(lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND),
				}, nil).Once()
			},
			hasError: true,
		},
		{
			name: "Update student attendance status failed",
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.
					On("FindByID", ctx, tx, lessonId, studentId).Return(&domain.LessonMember{
					AttendanceStatus: string(domain.StudentAttendStatusAbsent),
				}, nil).Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, tx, mock.MatchedBy(func(members []*domain.LessonMember) bool {
						if len(members) != 1 {
							return false
						}
						if members[0].StudentID != studentId || members[0].LessonID != lessonId || members[0].AttendanceStatus != string(lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_REALLOCATE) {
							return false
						}
						return true
					}), repo.UpdateLessonMemberFields{"attendance_status"}).Return(errors.New("UpdateLessonMembersFields failed")).Once()
			},
			hasError: true,
		},
		{
			name: "Upsert new record to reallocation table failed",
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.
					On("FindByID", ctx, tx, lessonId, studentId).Return(&domain.LessonMember{
					AttendanceStatus: string(domain.StudentAttendStatusAbsent),
				}, nil).Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, tx, mock.MatchedBy(func(members []*domain.LessonMember) bool {
						if len(members) != 1 {
							return false
						}
						if members[0].StudentID != studentId || members[0].LessonID != lessonId || members[0].AttendanceStatus != string(lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_REALLOCATE) {
							return false
						}
						return true
					}), repo.UpdateLessonMemberFields{"attendance_status"}).Return(nil).Once()
				reallocationRepo.
					On("UpsertReallocation", ctx, tx, lessonId, mock.MatchedBy(func(reallocations []*domain.Reallocation) bool {
						if len(reallocations) != 1 {
							return false
						}
						if reallocations[0].StudentID != studentId || reallocations[0].OriginalLessonID != lessonId {
							return false
						}
						return true
					})).Return(errors.New("UpsertReallocation failed")).Once()
			},
			hasError: true,
		},
		{
			name: "Mark student as reallocate successfully",
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.
					On("FindByID", ctx, tx, lessonId, studentId).Return(&domain.LessonMember{
					AttendanceStatus: string(domain.StudentAttendStatusAbsent),
				}, nil).Once()
				lessonMemberRepo.
					On("UpdateLessonMembersFields", ctx, tx, mock.MatchedBy(func(members []*domain.LessonMember) bool {
						if len(members) != 1 {
							return false
						}
						if members[0].StudentID != studentId || members[0].LessonID != lessonId || members[0].AttendanceStatus != string(lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_REALLOCATE) {
							return false
						}
						return true
					}), repo.UpdateLessonMemberFields{"attendance_status"}).Return(nil).Once()
				reallocationRepo.
					On("UpsertReallocation", ctx, tx, lessonId, mock.MatchedBy(func(reallocations []*domain.Reallocation) bool {
						if len(reallocations) != 1 {
							return false
						}
						if reallocations[0].StudentID != studentId || reallocations[0].OriginalLessonID != lessonId {
							return false
						}
						return true
					})).Return(nil).Once()
			},
			hasError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := LessonCommandHandler{
				WrapperConnection: wrapperConnection,
				LessonMemberRepo:  lessonMemberRepo,
				ReallocationRepo:  reallocationRepo,
			}
			err := handler.MarkStudentAsReallocate(ctx, req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				mock.AssertExpectationsForObjects(t, db, tx, lessonMemberRepo, reallocationRepo, mockUnleashClient)
			}
		})
	}
}
