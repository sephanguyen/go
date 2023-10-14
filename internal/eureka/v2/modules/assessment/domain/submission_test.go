package domain

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/stretchr/testify/assert"
)

func TestSubmissions_SortDescByAttemptedDate(t *testing.T) {
	t.Parallel()
	time1 := time.Date(2023, 01, 01, 1, 0, 0, 0, time.Local)
	time2 := time.Date(2023, 01, 01, 2, 0, 0, 0, time.Local)
	time3 := time.Date(2023, 01, 01, 3, 0, 0, 0, time.Local)

	t.Run("sort normally when there are different completed", func(t *testing.T) {
		// arrange
		unsorted := Submissions{
			{
				ID:          "ID1",
				CompletedAt: time1,
			},
			{
				ID:          "ID2",
				CompletedAt: time2,
			},
			{
				ID:          "ID3",
				CompletedAt: time3,
			},
		}
		sorted := Submissions{
			{
				ID:          "ID3",
				CompletedAt: time3,
			},
			{
				ID:          "ID2",
				CompletedAt: time2,
			},
			{
				ID:          "ID1",
				CompletedAt: time1,
			},
		}
		// act
		unsorted.SortDescByCompletedAt()

		// assert
		assert.Equal(t, sorted, unsorted)
	})

	t.Run("sort normally, keep order when date are the same", func(t *testing.T) {
		// arrange
		unsorted := Submissions{
			{
				ID:          "ID1",
				CompletedAt: time1,
			},
			{
				ID:          "ID2",
				CompletedAt: time2,
			},
			{
				ID:          "ID3",
				CompletedAt: time2,
			},
		}
		sorted := Submissions{
			{
				ID:          "ID2",
				CompletedAt: time2,
			},
			{
				ID:          "ID3",
				CompletedAt: time2,
			},
			{
				ID:          "ID1",
				CompletedAt: time1,
			},
		}
		// act
		unsorted.SortDescByCompletedAt()

		// assert
		assert.Equal(t, sorted, unsorted)
	})
}

func TestValidateSubmission(t *testing.T) {
	t.Parallel()

	t.Run("return ErrSubmissionIDRequired when id is missing", func(t *testing.T) {
		// arrange
		sut := Submission{
			ID:            "",
			AssessmentID:  idutil.ULIDNow(),
			StudentID:     idutil.ULIDNow(),
			GradingStatus: GradingStatusNotMarked,
			CompletedAt:   time.Now(),
		}
		// act
		err := sut.Validate()

		// assert
		assert.Equal(t, ErrSubmissionIDRequired, err)
	})

	t.Run("return ErrAssessmentIDRequired when id is missing", func(t *testing.T) {
		// arrange
		sut := Submission{
			ID:            idutil.ULIDNow(),
			AssessmentID:  "",
			StudentID:     idutil.ULIDNow(),
			GradingStatus: GradingStatusNotMarked,
			CompletedAt:   time.Now(),
		}
		// act
		err := sut.Validate()

		// assert
		assert.Equal(t, ErrAssessmentIDRequired, err)
	})

	t.Run("return ErrStudentIDRequired when id is missing", func(t *testing.T) {
		// arrange
		sut := Submission{
			ID:            idutil.ULIDNow(),
			AssessmentID:  idutil.ULIDNow(),
			StudentID:     "",
			GradingStatus: GradingStatusNotMarked,
			CompletedAt:   time.Now(),
		}
		// act
		err := sut.Validate()

		// assert
		assert.Equal(t, ErrStudentIDRequired, err)
	})

	t.Run("return ErrCompletedAtRequired when id is missing", func(t *testing.T) {
		// arrange
		var defTime time.Time
		sut := Submission{
			ID:            idutil.ULIDNow(),
			AssessmentID:  idutil.ULIDNow(),
			StudentID:     idutil.ULIDNow(),
			GradingStatus: GradingStatusNotMarked,
			CompletedAt:   defTime,
		}
		// act
		err := sut.Validate()

		// assert
		assert.Equal(t, ErrCompletedAtRequired, err)
	})

	t.Run("return ErrInvalidGradingStatus when id is missing", func(t *testing.T) {
		// arrange
		sut := Submission{
			ID:            idutil.ULIDNow(),
			AssessmentID:  idutil.ULIDNow(),
			StudentID:     idutil.ULIDNow(),
			GradingStatus: GradingStatusNone,
			CompletedAt:   time.Now(),
		}
		// act
		err := sut.Validate()

		// assert
		assert.Equal(t, ErrInvalidGradingStatus, err)
	})

	t.Run("return nil when no err", func(t *testing.T) {
		// arrange
		sut := Submission{
			ID:            idutil.ULIDNow(),
			AssessmentID:  idutil.ULIDNow(),
			StudentID:     idutil.ULIDNow(),
			GradingStatus: GradingStatusNotMarked,
			CompletedAt:   time.Now(),
		}
		// act
		err := sut.Validate()

		// assert
		assert.Nil(t, err)
	})
}
