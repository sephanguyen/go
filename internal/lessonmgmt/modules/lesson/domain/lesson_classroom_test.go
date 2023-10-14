package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLessonClassroom_IsValid(t *testing.T) {
	t.Parallel()
	t.Run("lesson classroom is valid", func(t *testing.T) {
		lc := &LessonClassroom{
			ClassroomID: "cr1",
		}
		err := lc.IsValid()
		require.Nil(t, err)
	})
	t.Run("lesson classroom is invalid", func(t *testing.T) {
		lc := &LessonClassroom{}
		err := lc.IsValid()
		require.NotNil(t, err)
	})
}

func TestLessonClassroom_WithClassName(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		lc := &LessonClassroom{
			ClassroomID: "cr1",
		}
		classroomName := "className"
		lc.WithClassroomName(classroomName)
		require.Equal(t, classroomName, lc.ClassroomName)
	})
}

func TestLessonClassroom_WithClassArea(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		lc := &LessonClassroom{
			ClassroomID: "cr1",
		}
		classroomArea := "classArea"
		lc.WithClassroomArea(classroomArea)
		require.Equal(t, classroomArea, lc.ClassroomArea)
	})
}
