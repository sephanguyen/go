package controller

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/commands"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_repositories_report "github.com/manabie-com/backend/mock/lessonmgmt/lesson_report/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_LessonStudentSubscription_HandleDeleteLessonMemberAndReport(t *testing.T) {
	t.Parallel()
	lessonRepo := new(mock_repositories.MockLessonRepo)
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	lessonReportRepo := new(mock_repositories_report.MockLessonReportRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	masterRepo := new(mock_repositories.MockMasterDataRepo)
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	now := time.Now()
	var mockLessons []*domain.Lesson
	tcs := []struct {
		name     string
		event    interface{}
		setup    func(ctx context.Context)
		hasError bool
		ack      bool
	}{
		{
			name: "lesson_members has been removed successfully",
			event: &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: "student-1",
					IsActive:  true,
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:   "course-1",
						LocationId: "location-1",
						ClassId:    "class-1",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockLessons = []*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "name",
						LocationID:       "location-id",
						CourseID:         "course-id",
						CreatedAt:        now,
						UpdatedAt:        now,
						StartTime:        now,
						EndTime:          now,
						SchedulingStatus: "scheduling-status",
						TeachingMedium:   "teaching-medium",
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						ClassID:          "class-id-1",
						Learners:         domain.LessonLearners{},
						Teachers: domain.LessonTeachers{
							{
								TeacherID: "teacher-id",
							},
						},
						PreparationTime: 10,
						BreakTime:       15,
					},
					{
						LessonID:         "lesson-2",
						Name:             "name",
						LocationID:       "location-id",
						CourseID:         "course-id",
						CreatedAt:        now,
						UpdatedAt:        now,
						StartTime:        now,
						EndTime:          now,
						SchedulingStatus: "scheduling-status",
						TeachingMedium:   "teaching-medium",
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						ClassID:          "class-id-2",
						Learners:         domain.LessonLearners{},
						Teachers: domain.LessonTeachers{
							{
								TeacherID: "teacher-id",
							},
						},
						PreparationTime: 10,
						BreakTime:       15,
					},
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything, mock.Anything).
					Return(true, nil).Once()
				lessonRepo.On("GetLessonByIDs", mock.Anything, mock.Anything, []string{"lesson-1", "lesson-2"}).Return(mockLessons, nil).Once()
				lessonRepo.On("UpdateLessonsTeachingTime", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				lessonMemberRepo.
					On("GetLessonIDsByStudentCourseRemovedLocation", mock.Anything, mock.Anything, "course-1", "student-1", []string{"location-1"}).
					Once().Return([]string{"lesson-1", "lesson-2"}, nil)
				db.On("Begin", mock.Anything).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil).Once()
				lessonMemberRepo.
					On("SoftDelete", mock.Anything, mock.Anything, "student-1", []string{"lesson-1", "lesson-2"}).
					Once().Return(nil)
				lessonMemberRepo.
					On("GetLessonMembersInLessons", mock.Anything, mock.Anything, []string{"lesson-1", "lesson-2"}).
					Once().Return([]*domain.LessonMember{
					{
						StudentID: "student-2",
						LessonID:  "lesson-1",
					},
					{
						StudentID: "student-2",
						LessonID:  "lesson-2",
					},
				}, nil)
			},
			hasError: false,
			ack:      true,
		},
		{
			name: "lesson_members and lesson_report has been removed successfully",
			event: &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: "student-1",
					IsActive:  true,
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:   "course-1",
						LocationId: "location-1",
						ClassId:    "class-1",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockLessons = []*domain.Lesson{
					{
						LessonID:         "lesson-1",
						Name:             "name",
						LocationID:       "location-id",
						CourseID:         "course-id",
						CreatedAt:        now,
						UpdatedAt:        now,
						StartTime:        now,
						EndTime:          now,
						SchedulingStatus: "scheduling-status",
						TeachingMedium:   "teaching-medium",
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						ClassID:          "class-id-1",
						Learners:         domain.LessonLearners{},
						Teachers: domain.LessonTeachers{
							{
								TeacherID: "teacher-id",
							},
						},
						PreparationTime: 10,
						BreakTime:       15,
					},
					{
						LessonID:         "lesson-2",
						Name:             "name",
						LocationID:       "location-id",
						CourseID:         "course-id",
						CreatedAt:        now,
						UpdatedAt:        now,
						StartTime:        now,
						EndTime:          now,
						SchedulingStatus: "scheduling-status",
						TeachingMedium:   "teaching-medium",
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						ClassID:          "class-id-2",
						Learners:         domain.LessonLearners{},
						Teachers: domain.LessonTeachers{
							{
								TeacherID: "teacher-id",
							},
						},
						PreparationTime: 10,
						BreakTime:       15,
					},
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything, mock.Anything).
					Return(true, nil).Once()
				lessonRepo.On("GetLessonByIDs", mock.Anything, mock.Anything, []string{"lesson-1", "lesson-2"}).Return(mockLessons, nil).Once()
				lessonRepo.On("UpdateLessonsTeachingTime", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				lessonMemberRepo.
					On("GetLessonIDsByStudentCourseRemovedLocation", mock.Anything, mock.Anything, "course-1", "student-1", []string{"location-1"}).
					Once().Return([]string{"lesson-1", "lesson-2"}, nil)
				db.On("Begin", mock.Anything).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil).Once()
				lessonMemberRepo.
					On("SoftDelete", mock.Anything, mock.Anything, "student-1", []string{"lesson-1", "lesson-2"}).
					Once().Return(nil)
				lessonMemberRepo.
					On("GetLessonMembersInLessons", mock.Anything, mock.Anything, []string{"lesson-1", "lesson-2"}).
					Once().Return([]*domain.LessonMember{}, nil)
				lessonReportRepo.
					On("DeleteReportsBelongToLesson", mock.Anything, mock.Anything, []string{"lesson-1", "lesson-2"}).
					Once().Return(nil)
			},
			hasError: false,
			ack:      true,
		},
		{
			name: "lesson_members removed location of student course active nothing lessons removed location",
			event: &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: "student-1",
					IsActive:  true,
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:   "course-1",
						LocationId: "location-1",
						ClassId:    "class-1",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.
					On("GetLessonIDsByStudentCourseRemovedLocation", mock.Anything, mock.Anything, "course-1", "student-1", []string{"location-1"}).
					Once().Return([]string{}, nil)
			},
			hasError: false,
			ack:      true,
		},
		{
			name: "lesson_members removed location of student course active can't get lessons",
			event: &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: "student-1",
					IsActive:  true,
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:   "course-1",
						LocationId: "location-1",
						ClassId:    "class-1",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.
					On("GetLessonIDsByStudentCourseRemovedLocation", mock.Anything, mock.Anything, "course-1", "student-1", []string{"location-1"}).
					Once().Return([]string{}, errors.New("nothing"))
			},
			hasError: true,
			ack:      false,
		},
		{
			name: "lesson_members removed location of student course active can't soft delete",
			event: &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: "student-1",
					IsActive:  true,
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:   "course-1",
						LocationId: "location-1",
						ClassId:    "class-1",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.
					On("GetLessonIDsByStudentCourseRemovedLocation", mock.Anything, mock.Anything, "course-1", "student-1", []string{"location-1"}).
					Once().Return([]string{"lesson-1"}, nil)
				db.On("Begin", mock.Anything).Return(tx, nil).Once()
				tx.On("Rollback", mock.Anything).Return(nil).Once()
				lessonMemberRepo.
					On("SoftDelete", mock.Anything, mock.Anything, "student-1", []string{"lesson-1"}).
					Once().Return(errors.New("nothing"))
			},
			hasError: true,
			ack:      false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			tc.setup(ctx)

			s := &LessonStudentSubscription{
				Logger:            zap.NewNop(),
				wrapperConnection: wrapperConnection,
				LessonRepo:        lessonRepo,
				LessonMemberRepo:  lessonMemberRepo,
				LessonReportRepo:  lessonReportRepo,
				LessonCommandHandler: commands.LessonCommandHandler{
					MasterDataPort: masterRepo,
				},
				JSM:              jsm,
				UnleashClientIns: mockUnleashClient,
			}
			msg, err := proto.Marshal(tc.event.(*npb.EventStudentPackageV2))
			assert.NoError(t, err)

			ack, err := s.handleDeleteLessonMemberAndReport(context.Background(), msg)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, ack, tc.ack)
			mock.AssertExpectationsForObjects(t, s.LessonMemberRepo, s.LessonReportRepo, s.JSM, mockUnleashClient, lessonRepo, masterRepo)
		})
	}
}

func Test_LessonStudentSubscription_HandleRemovingInactiveStudentFromLesson(t *testing.T) {
	t.Parallel()
	lessonRepo := new(mock_repositories.MockLessonRepo)
	masterRepo := new(mock_repositories.MockMasterDataRepo)
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	lessonReportRepo := new(mock_repositories_report.MockLessonReportRepo)
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	jsm := &mock_nats.JetStreamManagement{}
	reallocationRepo := &mock_repositories.MockReallocationRepo{}
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	now := time.Now()
	var mockLessons []*domain.Lesson

	testCase := []struct {
		name     string
		setup    func(ctx context.Context)
		data     *npb.EventStudentPackageV2
		hasError bool
		ack      bool
	}{
		{
			name: "remove inactive student successfully",
			data: &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: "student-id",
					IsActive:  true,
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:  "course-id",
						StartDate: timestamppb.New(time.Date(2022, 8, 1, 9, 0, 0, 0, time.UTC)),
						EndDate:   timestamppb.New(time.Date(2022, 8, 30, 9, 0, 0, 0, time.UTC)),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockLessons = []*domain.Lesson{
					{
						LessonID:         "lesson-id-1",
						Name:             "name",
						LocationID:       "location-id",
						CourseID:         "course-id",
						CreatedAt:        now,
						UpdatedAt:        now,
						StartTime:        now,
						EndTime:          now,
						SchedulingStatus: "scheduling-status",
						TeachingMedium:   "teaching-medium",
						TeachingMethod:   domain.LessonTeachingMethodIndividual,
						ClassID:          "class-id-1",
						Learners:         domain.LessonLearners{},
						Teachers: domain.LessonTeachers{
							{
								TeacherID: "teacher-id",
							},
						},
						PreparationTime: 10,
						BreakTime:       15,
					},
					{
						LessonID:         "lesson-id-2",
						Name:             "name",
						LocationID:       "location-id",
						CourseID:         "course-id",
						CreatedAt:        now,
						UpdatedAt:        now,
						StartTime:        now,
						EndTime:          now,
						SchedulingStatus: "scheduling-status",
						TeachingMedium:   "teaching-medium",
						TeachingMethod:   domain.LessonTeachingMethodGroup,
						ClassID:          "class-id-2",
						Learners:         domain.LessonLearners{},
						Teachers: domain.LessonTeachers{
							{
								TeacherID: "teacher-id",
							},
						},
						PreparationTime: 10,
						BreakTime:       15,
					},
				}
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything, mock.Anything).
					Return(true, nil).Once()
				lessonRepo.On("GetLessonByIDs", mock.Anything, mock.Anything, []string{"lesson-id-1", "lesson-id-2"}).Return(mockLessons, nil).Once()
				lessonRepo.On("UpdateLessonsTeachingTime", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.On("GetLessonsOutOfStudentCourse", mock.Anything, db, &user_domain.StudentSubscription{
					StudentID: "student-id",
					CourseID:  "course-id",
					StartAt:   time.Date(2022, 8, 1, 9, 0, 0, 0, time.UTC),
					EndAt:     time.Date(2022, 8, 30, 9, 0, 0, 0, time.UTC),
				}).Return([]string{"lesson-id-1", "lesson-id-2"}, nil)
				db.On("Begin", mock.Anything).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				lessonMemberRepo.On("SoftDelete", mock.Anything, tx, "student-id", []string{"lesson-id-1", "lesson-id-2"}).
					Return(nil).Once()
				reallocationRepo.
					On("SoftDelete", mock.Anything, mock.Anything, []string{"student-id", "lesson-id-1", "student-id", "lesson-id-2"}, false).
					Return(nil).Once()
				reallocationRepo.
					On("CancelIfStudentReallocated", mock.Anything, mock.Anything, []string{"student-id", "lesson-id-1", "student-id", "lesson-id-2"}).
					Return(nil).Once()
				lessonReportRepo.
					On("DeleteReportsBelongToLesson", mock.Anything, mock.Anything, []string{"lesson-id-1", "lesson-id-2"}).
					Return(nil).Once()
				lessonMemberRepo.
					On("GetLessonMembersInLessons", mock.Anything, mock.Anything, []string{"lesson-id-1", "lesson-id-2"}).
					Return([]*domain.LessonMember{}, nil).Once()
			},
			ack: true,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			tc.setup(ctx)
			s := &LessonStudentSubscription{
				Logger:            zap.NewNop(),
				wrapperConnection: wrapperConnection,
				LessonRepo:        lessonRepo,
				LessonMemberRepo:  lessonMemberRepo,
				LessonReportRepo:  lessonReportRepo,
				Env:               "local",
				UnleashClientIns:  mockUnleashClient,
				JSM:               jsm,
				ReallocationRepo:  reallocationRepo,
				LessonCommandHandler: commands.LessonCommandHandler{
					MasterDataPort: masterRepo,
				},
			}
			msg, err := proto.Marshal(tc.data)
			assert.NoError(t, err)
			ack, err := s.handleRemovingStudentFromLesson(ctx, msg)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.ack, ack)
			mock.AssertExpectationsForObjects(t, mockUnleashClient, lessonRepo, masterRepo)
		})
	}
}
