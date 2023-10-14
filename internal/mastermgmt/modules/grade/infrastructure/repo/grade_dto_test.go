package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/domain"

	"github.com/stretchr/testify/assert"
)

func TestToGradeEntity(t *testing.T) {
	t.Parallel()
	now := time.Now()
	t.Run("convert to entity", func(t *testing.T) {
		// arrange
		grade := &domain.Grade{
			ID:                "grade-id-1",
			Name:              "grade-name-1",
			PartnerInternalID: "PID 1",
			IsArchived:        false,
			Sequence:          1,
			Remarks:           "remarks x",
			CreatedAt:         now,
			UpdatedAt:         now,
			DeletedAt:         &now,
			ResourcePath:      "rp",
		}
		dto := &Grade{
			ID:                database.Text("grade-id-1"),
			Name:              database.Text("grade-name-1"),
			PartnerInternalID: database.Text("PID 1"),
			IsArchived:        database.Bool(false),
			Sequence:          database.Int4(1),
			Remarks:           database.Text("remarks x"),
			CreatedAt:         database.Timestamptz(now),
			UpdatedAt:         database.Timestamptz(now),
			DeletedAt:         database.Timestamptz(now),
			ResourcePath:      database.Text("rp"),
		}

		// act
		res := dto.ToGradeEntity()

		// assert
		assert.Equal(t, grade, res)
	})
}
