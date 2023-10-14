package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	usermgmt_entities "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	bob_repo "github.com/manabie-com/backend/mock/bob/repositories"
	search_repo "github.com/manabie-com/backend/mock/lessonmgmt/lesson/elasticsearch"
	lesson_repo "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	usermgmt_repo "github.com/manabie-com/backend/mock/usermgmt/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLessonSearchIndexer_indexLessonDocument(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	mockLessonRepo := &lesson_repo.MockLessonRepo{}
	mockSearchRepo := &search_repo.MockSearchRepo{}
	mockStudentRepo := &usermgmt_repo.MockStudentRepo{}
	mockUserRepo := &bob_repo.MockUserRepo{}

	s := &LessonSearchIndexer{
		LessonRepo:  mockLessonRepo,
		StudentRepo: mockStudentRepo,
		UserRepo:    mockUserRepo,
		SearchRepo:  mockSearchRepo,
	}
	type TestCase struct {
		name        string
		input       []string
		expectedErr error
		setup       func(ctx context.Context)
	}
	now := time.Now()

	student := []*usermgmt_entities.LegacyStudent{
		{
			ID:           database.Text("user-1"),
			CurrentGrade: database.Int2(12),
		},
	}
	user := []*bob_entities.User{
		{
			ID:       database.Text("user-1"),
			LastName: database.Text("Name"),
		},
	}
	lessons := []*domain.Lesson{
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
			TeachingMethod:   domain.LessonTeachingMethodGroup,
			ClassID:          "class-id-1",
			Learners: domain.LessonLearners{
				{
					LearnerID: "user-1",
					CourseID:  "course-id",
				},
			},
			Teachers: domain.LessonTeachers{
				{
					TeacherID: "teacher-id",
				},
			},
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
			Learners: domain.LessonLearners{
				{
					LearnerID:    "user-1",
					CourseID:     "course-id",
					AttendStatus: domain.StudentAttendStatusAbsent,
				},
			},
			Teachers: domain.LessonTeachers{
				{
					TeacherID: "teacher-id",
				},
			},
		},
	}
	userIds := []string{}
	for _, ls := range lessons {
		userIds = ls.GetLearnersIDs()
	}

	lessonSearchs, lessonIDs := newLessonSearch(lessons)

	testCases := []TestCase{
		{
			name:  "success",
			input: lessonIDs,
			setup: func(ctx context.Context) {
				mockLessonRepo.On("GetLessonByIDs", mock.Anything, mock.Anything, lessonIDs).
					Once().
					Return(lessons, nil)
				mockStudentRepo.On("FindStudentProfilesByIDs", mock.Anything, mock.Anything, database.TextArray(userIds)).
					Times(len(lessons)).
					Return(student, nil)
				mockUserRepo.On("Retrieve", mock.Anything, mock.Anything, database.TextArray(userIds), mock.Anything).
					Times(len(lessons)).
					Return(user, nil)
				mockSearchRepo.On("BulkUpsert", mock.Anything, mock.MatchedBy(func(docs []*domain.LessonSearch) bool {
					for _, doc := range docs {
						for _, ls := range lessonSearchs {
							if ls.LessonID == doc.LessonID {
								return true
							}
						}
					}
					return false
				})).
					Once().Return(2, nil)
			},
		},
		{
			name:        "failed",
			input:       lessonIDs,
			expectedErr: fmt.Errorf("cannot get lessons: %w", errors.New("Internal Error")),
			setup: func(ctx context.Context) {
				mockLessonRepo.On("GetLessonByIDs", mock.Anything, mock.Anything, lessonIDs).
					Return(nil, errors.New("Internal Error"))
			},
		},
		{
			name:        "failed to bulkupsert",
			input:       lessonIDs,
			expectedErr: fmt.Errorf("cannot upsert lesson document: %w", errors.New("Internal Error")),
			setup: func(ctx context.Context) {
				mockLessonRepo.On("GetLessonByIDs", mock.Anything, mock.Anything, lessonIDs).
					Once().
					Return(lessons, nil)
				mockStudentRepo.On("FindStudentProfilesByIDs", mock.Anything, mock.Anything, database.TextArray(userIds)).
					Times(len(lessons)).
					Return(student, nil)
				mockUserRepo.On("Retrieve", mock.Anything, mock.Anything, database.TextArray(userIds), mock.Anything).
					Times(len(lessons)).
					Return(user, nil)
				mockSearchRepo.On("BulkUpsert", mock.Anything, mock.MatchedBy(func(docs []*domain.LessonSearch) bool {
					for _, doc := range docs {
						for _, ls := range lessonSearchs {
							if ls.LessonID == doc.LessonID {
								return true
							}
						}
					}
					return false
				})).Once().Return(0, errors.New("Internal Error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctxRP := golibs.ResourcePathToCtx(context.Background(), "school-id")
			tc.setup(ctxRP)
			err := s.indexLessonDocument(ctx, tc.input)
			if tc.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func newLessonSearch(lessons []*domain.Lesson) (domain.LessonSearchs, []string) {
	lessonIDs := []string{}
	lessonSearchs := domain.LessonSearchs{}

	for _, ls := range lessons {
		lessonIDs = append(lessonIDs, ls.LessonID)
		lessonSearch := &domain.LessonSearch{
			LessonID:       ls.LessonID,
			LocationID:     ls.LocationID,
			TeachingMedium: string(ls.TeachingMedium),
			TeachingMethod: string(ls.TeachingMethod),
			CreatedAt:      ls.CreatedAt,
			UpdatedAt:      ls.UpdatedAt,
			StartTime:      ls.StartTime,
			EndTime:        ls.EndTime,
			DeletedAt:      ls.DeletedAt,
			LessonTeacher:  ls.Teachers.GetIDs(),
		}
		if ls.TeachingMethod == domain.LessonTeachingMethodGroup {
			lessonSearch.ClassID = ls.ClassID
			lessonSearch.CourseID = ls.CourseID
		}
		lm := []*domain.LessonMemberEs{}
		for _, ls := range ls.Learners {
			lm = append(lm, &domain.LessonMemberEs{
				ID:           ls.LearnerID,
				CourseID:     ls.CourseID,
				Name:         "Name",
				CurrentGrade: 12,
			})
		}
		lessonSearch.AddLessonMembers(lm)
		lessonSearchs = append(lessonSearchs, lessonSearch)
	}
	return lessonSearchs, lessonIDs
}
