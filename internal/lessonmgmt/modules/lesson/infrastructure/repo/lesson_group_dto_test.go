package repo

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestNewLessonGroupFromLessonEntity(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name        string
		CourseID    pgtype.Text
		Lesson      *domain.Lesson
		LessonGroup *LessonGroup
	}{
		{
			name:     "full fields",
			CourseID: database.Text("course-id-1"),
			Lesson: &domain.Lesson{
				Material: &domain.LessonMaterial{
					MediaIDs: []string{"media-id-1", "media-id-2"},
				},
			},
			LessonGroup: &LessonGroup{
				LessonGroupID: pgtype.Text{Status: pgtype.Null},
				CourseID:      database.Text("course-id-1"),
				MediaIDs:      database.TextArray([]string{"media-id-1", "media-id-2"}),
				CreatedAt:     pgtype.Timestamptz{Status: pgtype.Null},
				UpdatedAt:     pgtype.Timestamptz{Status: pgtype.Null},
			},
		},
		{
			name:     "missing material field",
			CourseID: database.Text("course-id-1"),
			Lesson:   &domain.Lesson{},
			LessonGroup: &LessonGroup{
				LessonGroupID: pgtype.Text{Status: pgtype.Null},
				CourseID:      database.Text("course-id-1"),
				MediaIDs:      pgtype.TextArray{Status: pgtype.Null},
				CreatedAt:     pgtype.Timestamptz{Status: pgtype.Null},
				UpdatedAt:     pgtype.Timestamptz{Status: pgtype.Null},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewLessonGroupFromLessonEntity(tc.Lesson, tc.CourseID)
			assert.EqualValues(t, tc.LessonGroup, actual)
		})
	}
}
