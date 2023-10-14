package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLessonTeacher_Validate(t *testing.T) {
	t.Run("lesson teacher is valid", func(t *testing.T) {
		lt := &LessonTeacher{
			TeacherID: "teacher-1",
			Name:      "teacherA",
		}
		err := lt.Validate()
		require.NoError(t, err)
	})
}
