package domain_test

import (
	"testing"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLessonLearner_IsValid(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name          string
		lessonLearner *domain.LessonLearner
		hasError      bool
	}{
		{
			name: "full fields",
			lessonLearner: &domain.LessonLearner{
				LearnerID:    "learner-1",
				CourseID:     "course-1",
				AttendStatus: domain.StudentAttendStatusAbsent,
				LocationID:   "location-1",
			},
			hasError: false,
		},
		{
			name: "missing learnerID",
			lessonLearner: &domain.LessonLearner{
				LearnerID:    "",
				CourseID:     "course-1",
				AttendStatus: domain.StudentAttendStatusAbsent,
				LocationID:   "location-1",
			},
			hasError: true,
		},
		{
			name: "missing courseID",
			lessonLearner: &domain.LessonLearner{
				LearnerID:    "learner-1",
				CourseID:     "",
				AttendStatus: domain.StudentAttendStatusAbsent,
				LocationID:   "location-1",
			},
			hasError: true,
		},
		{
			name: "missing attendStatus",
			lessonLearner: &domain.LessonLearner{
				LearnerID:    "learner-1",
				CourseID:     "course-1",
				AttendStatus: "",
				LocationID:   "location-1",
			},
			hasError: true,
		},
		{
			name: "missing locationID",
			lessonLearner: &domain.LessonLearner{
				LearnerID:    "learner-1",
				CourseID:     "course-1",
				AttendStatus: domain.StudentAttendStatusAbsent,
				LocationID:   "",
			},
			hasError: true,
		},
	}

	for i, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.lessonLearner.IsValid()
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
