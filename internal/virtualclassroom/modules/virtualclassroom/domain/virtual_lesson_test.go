package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLesson_IsValid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	repo := new(mock_repositories.MockVirtualLessonRepo)

	tcs := []struct {
		name    string
		lesson  *domain.VirtualLesson
		setup   func(ctx context.Context)
		isValid bool
	}{
		{
			name: "full fields",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				SchedulerID:      "scheduler-id",
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.VirtualLesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
			},
			isValid: true,
		},
		{
			name: "missing material",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.VirtualLesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
			},
			isValid: true,
		},
		{
			name: "missing name",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.VirtualLesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
			},
			isValid: true,
		},
		{
			name: "lesson draft with student,teacher",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusDraft,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
				Teachers: domain.LessonTeachers{
					{
						TeacherID: "teacher-id-1",
					},
					{
						TeacherID: "teacher-id-2",
					},
				},
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.VirtualLesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
			},
			isValid: true,
		},
		{
			name: "lesson draft with no student,teacher",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusDraft,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
				Learners:         domain.LessonLearners{},
				Teachers:         domain.LessonTeachers{},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.VirtualLesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
			},
			isValid: true,
		},

		// failed cases
		{
			name: "duplicated student and course id",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
				Learners: domain.LessonLearners{
					{
						LearnerID:    "user-id-1",
						CourseID:     "course-id-1",
						AttendStatus: domain.StudentAttendStatusAttend,
						LocationID:   "center-id-1",
					},
					{
						LearnerID:    "user-id-1",
						CourseID:     "course-id-2",
						AttendStatus: domain.StudentAttendStatusAbsent,
						LocationID:   "center-id-1",
					},
					{
						LearnerID:    "user-id-2",
						CourseID:     "course-id-2",
						AttendStatus: domain.StudentAttendStatusEmpty,
						LocationID:   "center-id-1",
					},
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing center id",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing location id in learners",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "location id of learner not same center id",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1000000000",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "end time before start time",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now.Add(-1 * time.Minute),
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing scheduling status",
			lesson: &domain.VirtualLesson{
				LessonID:       "lesson-id-1",
				Name:           "lesson name 1",
				CenterID:       "center-id-1",
				CreatedAt:      now,
				UpdatedAt:      now,
				StartTime:      now,
				EndTime:        now,
				TeachingMedium: domain.LessonTeachingMediumOnline,
				TeachingMethod: domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing teaching medium",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing teaching method",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "there are no any learners",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing course id in learners list",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
				Learners: domain.LessonLearners{
					{
						LearnerID:    "user-id-1",
						CourseID:     "course-id-1",
						AttendStatus: domain.StudentAttendStatusAttend,
						LocationID:   "center-id-1",
					},
					{
						LearnerID:    "user-id-2",
						AttendStatus: domain.StudentAttendStatusEmpty,
						LocationID:   "center-id-1",
					},
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing user id in learners list",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
				Learners: domain.LessonLearners{
					{
						LearnerID:    "user-id-1",
						CourseID:     "course-id-1",
						AttendStatus: domain.StudentAttendStatusAttend,
						LocationID:   "center-id-1",
					},
					{
						CourseID:     "course-id-2",
						AttendStatus: domain.StudentAttendStatusEmpty,
						LocationID:   "center-id-1",
					},
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing teacher id in teachers list",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
					},
				},
				Teachers: domain.LessonTeachers{
					{},
					{
						TeacherID: "teacher-id-2",
					},
				},
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "with not exist lesson id",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
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
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.VirtualLesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
			},
			isValid: false,
		},
		{
			name: "with not exist media id",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.VirtualLesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
			},
			isValid: false,
		},
		{
			name: "with not exist teacher id",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.VirtualLesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
			},
			isValid: false,
		},
		{
			name: "with not exist student course subscription",
			lesson: &domain.VirtualLesson{
				LessonID:         "lesson-id-1",
				Name:             "lesson name 1",
				CenterID:         "center-id-1",
				CreatedAt:        now,
				UpdatedAt:        now,
				StartTime:        now,
				EndTime:          now,
				SchedulingStatus: domain.LessonSchedulingStatusPublished,
				TeachingMedium:   domain.LessonTeachingMediumOnline,
				TeachingMethod:   domain.LessonTeachingMethodIndividual,
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
					{
						LearnerID:    "user-id-3",
						CourseID:     "course-id-3",
						AttendStatus: domain.StudentAttendStatusInformedAbsent,
						LocationID:   "center-id-1",
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
				TeacherIDs: domain.TeacherIDs{
					[]string{
						"teacher-id-1",
						"teacher-id-2",
					},
				},
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			setup: func(ctx context.Context) {
				repo.On("GetLessonByID", ctx, db, "lesson-id-1").Return(&domain.VirtualLesson{
					TeachingMethod: domain.LessonTeachingMethodIndividual,
				}, nil).Once()
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			builder := domain.NewVirtualLesson().
				WithLessonID(tc.lesson.LessonID).
				WithName(tc.lesson.Name).
				WithCenterID(tc.lesson.CenterID).
				WithTimeRange(tc.lesson.StartTime, tc.lesson.EndTime).
				WithModificationTime(tc.lesson.CreatedAt, tc.lesson.UpdatedAt).
				WithTeachingMedium(tc.lesson.TeachingMedium).
				WithTeachingMethod(tc.lesson.TeachingMethod).
				WithSchedulingStatus(tc.lesson.SchedulingStatus).
				WithTeacherIDs(tc.lesson.Teachers.GetIDs()).
				WithLearners(tc.lesson.Learners).
				WithSchedulerID(tc.lesson.SchedulerID)
			if tc.lesson.Material != nil {
				builder.WithMaterials(tc.lesson.Material.MediaIDs)
			}
			actual := builder.BuildDraft()
			assert.EqualValues(t, tc.lesson, actual)

			mock.AssertExpectationsForObjects(
				t,
				db,
			)
		})
	}
}
