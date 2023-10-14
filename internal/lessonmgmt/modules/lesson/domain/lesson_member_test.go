package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLessonMembers_NewLessonLearner(t *testing.T) {
	t.Parallel()
	t.Run("have at least one student", func(t *testing.T) {
		lessonMembers := LessonMembers{
			{
				LessonID:  "ls1",
				StudentID: "st1",
				CourseID:  "c1",
			},
			{
				LessonID:  "ls2",
				StudentID: "st2",
				CourseID:  "c2",
			},
		}
		studentID := lessonMembers.GetStudentIDs()
		require.Equal(t, []string{"st1", "st2"}, studentID)
	})
	t.Run("empty student", func(t *testing.T) {
		lessonMembers := LessonMembers{}
		studentID := lessonMembers.GetStudentIDs()
		require.Equal(t, []string{}, studentID)
	})
}

func TestLessonMembers_GetMapFieldValuesOfStudent(t *testing.T) {
	t.Parallel()
	t.Run("have at least one student", func(t *testing.T) {
		lessonMembers := LessonMembers{
			{
				LessonID:  "ls1",
				StudentID: "st1",
				CourseID:  "c1",
			},
			{
				LessonID:  "ls2",
				StudentID: "st2",
				CourseID:  "c2",
			},
		}
		expected := map[string]*LessonMember{
			"st1": {
				LessonID:  "ls1",
				StudentID: "st1",
				CourseID:  "c1",
			},
			"st2": {
				LessonID:  "ls2",
				StudentID: "st2",
				CourseID:  "c2",
			},
		}
		lessonMemberMap := lessonMembers.GetMapFieldValuesOfStudent()
		require.Equal(t, expected, lessonMemberMap)
	})
}
