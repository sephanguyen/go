package dto

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestSubmission_ToEntity(t *testing.T) {
	t.Parallel()

	t.Run("happy case", func(t *testing.T) {
		// arrange
		now := time.Now()
		dto := Submission{
			BaseEntity: BaseEntity{
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
			ID:                database.Text("Text 1"),
			SessionID:         database.Text("Text 2"),
			AssessmentID:      database.Text("Text 3"),
			StudentID:         database.Text("Text 4"),
			AllocatedMarkerID: database.Text("Text 5"),
			MarkedBy:          database.Text("Text 6"),
			MarkedAt:          database.Timestamptz(now),
			GradingStatus:     database.Text("IN_PROGRESS"),
			MaxScore:          database.Int4(300),
			GradedScore:       database.Int4(100),
		}

		expected := domain.Submission{
			ID:                "Text 1",
			SessionID:         "Text 2",
			AssessmentID:      "Text 3",
			StudentID:         "Text 4",
			AllocatedMarkerID: "Text 5",
			MarkedBy:          "Text 6",
			MarkedAt:          &now,
			CreatedAt:         now,
			GradingStatus:     domain.GradingStatusInProgress,
			MaxScore:          300,
			GradedScore:       100,
		}

		// act
		actual := dto.ToEntity()

		// assert
		assert.Equal(t, expected, actual)
	})
}
