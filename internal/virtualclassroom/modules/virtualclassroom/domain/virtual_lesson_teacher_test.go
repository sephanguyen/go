package domain_test

import (
	"testing"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLessonTeacher_IsValid(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name          string
		lessonTeacher *domain.LessonTeacher
		hasError      bool
	}{
		{
			name: "full fields",
			lessonTeacher: &domain.LessonTeacher{
				TeacherID: "teacher-1",
			},
			hasError: false,
		},
		{
			name: "missing teacherID",
			lessonTeacher: &domain.LessonTeacher{
				TeacherID: "",
			},
			hasError: true,
		},
	}

	for i, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.lessonTeacher.IsValid()
			if tcs[i].hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(
				t,
			)
		})
	}
}
