package domain

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/stretchr/testify/assert"
)

func TestSessions_SortDescByCompletedAt(t *testing.T) {
	t.Parallel()

	completedAt1 := time.Date(2023, 01, 01, 1, 0, 0, 0, time.Local)
	completedAt2 := time.Date(2023, 01, 01, 2, 0, 0, 0, time.Local)
	completedAt3 := time.Date(2023, 01, 01, 3, 0, 0, 0, time.Local)

	testCases := []struct {
		Name             string
		Request          any
		ExpectedResponse any
	}{
		{
			Name: "happy case",
			Request: Sessions{
				{
					ID:          "session_id_1",
					CreatedAt:   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
					CompletedAt: &completedAt1,
				},
				{
					ID:          "session_id_2",
					CreatedAt:   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
					CompletedAt: &completedAt2,
				},
				{
					ID:          "session_id_3",
					CreatedAt:   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
					CompletedAt: &completedAt3,
				},
			},
			ExpectedResponse: Sessions{
				{
					ID:          "session_id_3",
					CreatedAt:   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
					CompletedAt: &completedAt3,
				},
				{
					ID:          "session_id_2",
					CreatedAt:   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
					CompletedAt: &completedAt2,
				},
				{
					ID:          "session_id_1",
					CreatedAt:   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
					CompletedAt: &completedAt1,
				},
			},
		},
		{
			Name: "happy case: one session with completedAt is nil",
			Request: Sessions{
				{
					ID:          "session_id_1",
					CreatedAt:   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
					CompletedAt: &completedAt1,
				},
				{
					ID:          "session_id_2",
					CreatedAt:   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
					CompletedAt: &completedAt2,
				},
				{
					ID:        "session_id_3",
					CreatedAt: time.Date(2023, 01, 01, 1, 0, 0, 0, time.Local),
				},
				{
					ID:          "session_id_4",
					CreatedAt:   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
					CompletedAt: &completedAt3,
				},
			},
			ExpectedResponse: Sessions{
				{
					ID:        "session_id_3",
					CreatedAt: time.Date(2023, 01, 01, 1, 0, 0, 0, time.Local),
				},
				{
					ID:          "session_id_4",
					CreatedAt:   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
					CompletedAt: &completedAt3,
				},
				{
					ID:          "session_id_2",
					CreatedAt:   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
					CompletedAt: &completedAt2,
				},
				{
					ID:          "session_id_1",
					CreatedAt:   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
					CompletedAt: &completedAt1,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Request.(Sessions).SortDescByCompletedAt()
			assert.Equal(t, tc.ExpectedResponse, tc.Request.(Sessions))
		})
	}
}

func TestSession_Validate(t *testing.T) {
	t.Parallel()

	t.Run("return err ErrIDRequired when no ID", func(t *testing.T) {
		// arrange
		sut := Session{
			ID:                 "",
			AssessmentID:       idutil.ULIDNow(),
			UserID:             idutil.ULIDNow(),
			CourseID:           idutil.ULIDNow(),
			LearningMaterialID: idutil.ULIDNow(),
			Status:             SessionStatusNone,
		}

		// act
		err := sut.Validate()

		// assert
		assert.Equal(t, ErrIDRequired, err)
	})

	t.Run("return err ErrAssessmentIDRequired when no assessment ID", func(t *testing.T) {
		// arrange
		sut := Session{
			ID:                 idutil.ULIDNow(),
			AssessmentID:       "",
			UserID:             idutil.ULIDNow(),
			CourseID:           idutil.ULIDNow(),
			LearningMaterialID: idutil.ULIDNow(),
			Status:             SessionStatusNone,
		}

		// act
		err := sut.Validate()

		// assert
		assert.Equal(t, ErrAssessmentIDRequired, err)
	})

	t.Run("return err ErrUserIDRequired when no user ID", func(t *testing.T) {
		// arrange
		sut := Session{
			ID:                 idutil.ULIDNow(),
			AssessmentID:       idutil.ULIDNow(),
			UserID:             "",
			CourseID:           idutil.ULIDNow(),
			LearningMaterialID: idutil.ULIDNow(),
			Status:             SessionStatusCompleted,
		}

		// act
		err := sut.Validate()

		// assert
		assert.Equal(t, ErrUserIDRequired, err)
	})

	t.Run("return err ErrInvalidSessionStatus when wrong status", func(t *testing.T) {
		// arrange
		sut := Session{
			ID:                 idutil.ULIDNow(),
			AssessmentID:       idutil.ULIDNow(),
			UserID:             idutil.ULIDNow(),
			CourseID:           idutil.ULIDNow(),
			LearningMaterialID: idutil.ULIDNow(),
			Status:             "NONE A",
		}

		// act
		err := sut.Validate()

		// assert
		assert.Equal(t, ErrInvalidSessionStatus, err)
	})

	t.Run("return nil when all data fit", func(t *testing.T) {
		// arrange
		sut := Session{
			ID:                 idutil.ULIDNow(),
			AssessmentID:       idutil.ULIDNow(),
			UserID:             idutil.ULIDNow(),
			CourseID:           idutil.ULIDNow(),
			LearningMaterialID: idutil.ULIDNow(),
			Status:             SessionStatusCompleted,
		}

		// act
		err := sut.Validate()

		// assert
		assert.Nil(t, err)
	})
}
