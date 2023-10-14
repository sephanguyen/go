package domain_test

import (
	"context"
	"errors"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"testing"
	"time"

	calendar_dto "github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	calendar_mock_repositories "github.com/manabie-com/backend/mock/calendar/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_media_module "github.com/manabie-com/backend/mock/lessonmgmt/lesson/media_module_adapter"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_user_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/usermodadapter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLesson_IsValid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	masterDataRepo := new(mock_repositories.MockMasterDataRepo)
	mediaModulePort := new(mock_media_module.MockMediaModuleAdapter)
	userModuleAdapter := new(mock_user_module_adapter.MockUserModuleAdapter)
	repo := new(mock_repositories.MockLessonRepo)
	dateInfoRepo := new(calendar_mock_repositories.MockDateInfoRepo)
	dateInfos := []*calendar_dto.DateInfo{}
	classroomRepo := new(mock_repositories.MockClassroomRepo)

	tcs := []struct {
		name    string
		lesson  *domain.Lesson
		setup   func(ctx context.Context)
		isValid bool
	}{
		{
			name: "full fields",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				SchedulerID:      "scheduler-id",
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
				Persisted:        true,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, db, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
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
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						now,
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2", "user-id-3", "course-id-3"},
					).
					Return(nil).
					Once()
				classroomRepo.On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
			},
			isValid: true,
		},
		{
			name: "missing material",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
				Persisted:        true,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, db, "center-id-1").
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
						now,
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2", "user-id-3", "course-id-3"},
					).
					Return(nil).
					Once()
				classroomRepo.On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
			},
			isValid: true,
		},
		{
			name: "missing name",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusInformedAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, db, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
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
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						now,
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2", "user-id-3", "course-id-3"},
					).
					Return(nil).
					Once()
				classroomRepo.On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
			},
			isValid: true,
		},
		{
			name: "missing attendance note",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				SchedulerID:      "scheduler-id",
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
				Persisted:        true,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, db, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
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
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						now,
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2", "user-id-3", "course-id-3"},
					).
					Return(nil).
					Once()
				classroomRepo.On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
			},
			isValid: true,
		},
		{
			name: "lesson draft with student,teacher",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusDraft,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, db, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
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
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				userModuleAdapter.
					On(
						"CheckStudentCourseSubscriptions",
						ctx,
						now,
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2"},
					).
					Return(nil).
					Once()
				classroomRepo.On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
			},
			isValid: true,
		},
		{
			name: "lesson draft with no student,teacher",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusDraft,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
				Learners:         domain.LessonLearners{},
				Teachers:         domain.LessonTeachers{},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, db, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
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
			},
			isValid: true,
		},
		{
			name: "duplicated student and course id",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
						LearnerID:        "user-id-1",
						CourseID:         "course-id-2",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.OnTheDay,
						AttendanceReason: domain.FamilyReason,
					},
					{
						LearnerID:        "user-id-2",
						CourseID:         "course-id-2",
						AttendStatus:     domain.StudentAttendStatusEmpty,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.NoticeEmpty,
						AttendanceReason: domain.ReasonEmpty,
					},
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing center id",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing location id in learners",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "location id of learner not same center id",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1000000000",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing attendance notice in learners",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
						AttendanceReason: domain.ReasonEmpty,
					},
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing attendance reason in learners",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "end time before start time",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now.Add(-1 * time.Minute),
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing scheduling status",
			lesson: &domain.Lesson{
				LessonID:       "lesson-id-1",
				Name:           "lesson name 1",
				LocationID:     "center-id-1",
				CreatedAt:      now,
				UpdatedAt:      now,
				StartTime:      now,
				EndTime:        now,
				Persisted:      true,
				TeachingMedium: domain.LessonTeachingMediumOnline,
				TeachingMethod: domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing teaching medium",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing teaching method",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "there are no any teachers",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "there are no any learners",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, db, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
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
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(nil).
					Once()
				classroomRepo.On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
			},
			isValid: true,
		},
		{
			name: "missing course id in learners list",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
						AttendStatus:     domain.StudentAttendStatusEmpty,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.NoticeEmpty,
						AttendanceReason: domain.ReasonEmpty,
					},
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusInformedAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing user id in learners list",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
						CourseID:         "course-id-2",
						AttendStatus:     domain.StudentAttendStatusEmpty,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.NoticeEmpty,
						AttendanceReason: domain.ReasonEmpty,
					},
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusInformedAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing teacher id in teachers list",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusInformedAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
					},
				},
				Teachers: domain.LessonTeachers{
					{},
					{
						TeacherID: "teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
				DateInfos: dateInfos,
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "with not exist lesson id",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusInformedAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").
					Return(nil, errors.New("could not find center")).Once()
			},
			isValid: false,
		},
		{
			name: "with not exist center id",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusInformedAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, db, "center-id-1").
					Return(nil, errors.New("could not find center")).Once()
			},
			isValid: false,
		},
		{
			name: "with not exist media id",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusInformedAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, db, "center-id-1").
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
						now,
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2", "user-id-3", "course-id-3"},
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
					}, nil).Once()
			},
			isValid: false,
		},
		{
			name: "with not exist teacher id",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusInformedAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, db, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				userModuleAdapter.
					On("CheckTeacherIDs", ctx, []string{"teacher-id-1", "teacher-id-2"}).
					Return(errors.New("could not find teacher id")).
					Once()
			},
			isValid: false,
		},
		{
			name: "with not exist student course subscription",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				Persisted:        true,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:        "user-id-3",
						CourseID:         "course-id-3",
						AttendStatus:     domain.StudentAttendStatusInformedAbsent,
						LocationID:       "center-id-1",
						AttendanceNotice: domain.InAdvance,
						AttendanceReason: domain.PhysicalCondition,
						AttendanceNote:   "sample-attendance-note",
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
				DateInfos: dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, db, "center-id-1").
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
						now,
						[]string{"user-id-1", "course-id-1", "user-id-2", "course-id-2", "user-id-3", "course-id-3"},
					).
					Return(errors.New("could not find student course subscription")).
					Once()
			},
			isValid: false,
		},
		{
			name: "save draft lesson with group teaching method,missing learner,teacher,course_id,class_id fields",
			lesson: &domain.Lesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				LocationID:       "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusDraft,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodGroup,
				Learners:         domain.LessonLearners{},
				Teachers:         domain.LessonTeachers{},
				Persisted:        true,
				DateInfos:        dateInfos,
				Classrooms: domain.LessonClassrooms{
					{
						ClassroomID: "classroom-id-1",
					},
					{
						ClassroomID: "classroom-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					TeachingMethod: domain.LessonTeachingMethodGroup,
				}, nil).Once()
				dateInfoRepo.On("GetDateInfoByDateRangeAndLocationID",
					ctx, db, mock.Anything, mock.Anything, mock.Anything,
				).Return(dateInfos, nil).Once()
				masterDataRepo.
					On("GetLocationByID", ctx, db, "center-id-1").
					Return(&domain.Location{
						LocationID: "center-id-1",
						Name:       "center name 1",
						UpdatedAt:  now,
						CreatedAt:  now,
					}, nil).Once()
				classroomRepo.On("CheckClassroomIDs", ctx, mock.Anything, []string{"classroom-id-1", "classroom-id-2"}).
					Return(nil).Once()
			},
			isValid: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			builder := domain.NewLesson().
				WithID(tc.lesson.LessonID).
				WithName(tc.lesson.Name).
				WithLocationID(tc.lesson.LocationID).
				WithTimeRange(tc.lesson.StartTime, tc.lesson.EndTime).
				WithModificationTime(tc.lesson.CreatedAt, tc.lesson.UpdatedAt).
				WithTeachingMedium(tc.lesson.TeachingMedium).
				WithTeachingMethod(tc.lesson.TeachingMethod).
				WithSchedulingStatus(tc.lesson.SchedulingStatus).
				WithTeacherIDs(tc.lesson.Teachers.GetIDs()).
				WithLearners(tc.lesson.Learners).
				WithMasterDataPort(masterDataRepo).
				WithUserModulePort(userModuleAdapter).
				WithMediaModulePort(mediaModulePort).
				WithLessonRepo(repo).
				WithDateInfoRepo(dateInfoRepo).
				WithSchedulerID(tc.lesson.SchedulerID).
				WithClassroomIDs(tc.lesson.Classrooms.GetIDs()).
				WithClassroomRepo(classroomRepo)
			if tc.lesson.Material != nil {
				builder.WithMaterials(tc.lesson.Material.MediaIDs)
			}
			actual, err := builder.Build(ctx, db)
			if !tc.isValid {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				actual.MasterDataPort = nil
				actual.UserModulePort = nil
				actual.MediaModulePort = nil
				actual.DateInfoRepo = nil
				actual.Repo = nil
				actual.ClassroomRepo = nil
				assert.EqualValues(t, tc.lesson, actual)
			}
			mock.AssertExpectationsForObjects(
				t,
				db,
				masterDataRepo,
				userModuleAdapter,
				mediaModulePort,
				dateInfoRepo,
				repo,
			)
		})
	}
}

func TestLesson_AddTeachers(t *testing.T) {
	lesson := &domain.Lesson{}
	lesson.AddTeachers(domain.LessonTeachers{
		{
			TeacherID: "teacher-id-1",
		},
		{
			TeacherID: "teacher-id-2",
		},
		{
			TeacherID: "teacher-id-3",
		},
	})

	teacherIDs := lesson.GetTeacherIDs()
	assert.ElementsMatch(t, []string{"teacher-id-1", "teacher-id-2", "teacher-id-3"}, teacherIDs)
}

func TestLesson_AddLearners(t *testing.T) {
	lesson := &domain.Lesson{}
	lesson.AddLearners(domain.LessonLearners{
		{
			LearnerID:        "learner-id-1",
			CourseID:         "course-id-1",
			AttendStatus:     domain.StudentAttendStatusAttend,
			LocationID:       "location-id-1",
			AttendanceNotice: domain.NoticeEmpty,
			AttendanceReason: domain.ReasonEmpty,
		},
		{
			LearnerID:        "learner-id-2",
			CourseID:         "course-id-1",
			AttendStatus:     domain.StudentAttendStatusLate,
			LocationID:       "location-id-1",
			AttendanceNotice: domain.NoContact,
			AttendanceReason: domain.ReasonOther,
			AttendanceNote:   "sample-attendance-note",
		},
		{
			LearnerID:        "learner-id-3",
			CourseID:         "course-id-2",
			AttendStatus:     domain.StudentAttendStatusEmpty,
			LocationID:       "location-id-1",
			AttendanceNotice: domain.NoticeEmpty,
			AttendanceReason: domain.ReasonEmpty,
		},
	})

	learnerIDs := lesson.GetLearnersIDs()
	assert.ElementsMatch(t, []string{"learner-id-1", "learner-id-2", "learner-id-3"}, learnerIDs)
}
