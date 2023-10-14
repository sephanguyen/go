package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestNewClassFromEntity(t *testing.T) {
	now := time.Now()
	t.Run("success", func(t *testing.T) {
		classEntity := &domain.Class{
			ClassID:    "class-id",
			Name:       "class-name",
			CourseID:   "course_id",
			LocationID: "location-id",
			SchoolID:   "1",
			UpdatedAt:  now,
			CreatedAt:  now,
		}
		expectedClass := &Class{
			ClassID:    database.Text("class-id"),
			Name:       database.Text("class-name"),
			CourseID:   database.Text("course_id"),
			LocationID: database.Text("location-id"),
			SchoolID:   database.Text("1"),
			CreatedAt:  database.Timestamptz(now),
			UpdatedAt:  database.Timestamptz(now),
			DeletedAt:  pgtype.Timestamptz{Status: pgtype.Null},
		}
		gotClass, err := NewClassFromEntity(classEntity)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedClass, gotClass)
	})

}

func TestNewCourseStudentFromEntity(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		courseStudentEntity := &domain.ClassWithCourseStudent{
			ClassID:   "class-id",
			CourseID:  "course-id",
			StudentID: "student-id",
		}
		expectedCourseStudent := &ClassWithCourseStudent{
			ClassID:   database.Text("class-id"),
			CourseID:  database.Text("course-id"),
			StudentID: database.Text("student-id"),
		}
		gotCourseStudent, err := NewCourseStudentFromEntity(courseStudentEntity)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedCourseStudent, gotCourseStudent)
	})

}
