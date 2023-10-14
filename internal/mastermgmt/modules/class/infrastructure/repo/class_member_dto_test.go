package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestNewClassMemberFromEntity(t *testing.T) {
	now := time.Now()
	t.Run("success", func(t *testing.T) {
		classEntity := &domain.ClassMember{
			ClassID:       "class-id",
			ClassMemberID: "class-member-1",
			UserID:        "user-1",
			UpdatedAt:     now,
			CreatedAt:     now,
			StartDate:     now,
			EndDate:       now,
		}
		expectedClass := &ClassMember{
			ClassID:       database.Text("class-id"),
			ClassMemberID: database.Text("class-member-1"),
			UserID:        database.Text("user-1"),
			CreatedAt:     database.Timestamptz(now),
			UpdatedAt:     database.Timestamptz(now),
			DeletedAt:     pgtype.Timestamptz{Status: pgtype.Null},
			StartDate:     database.Timestamptz(now),
			EndDate:       database.Timestamptz(now),
		}
		gotClass, err := NewClassMemberFromEntity(classEntity)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedClass, gotClass)
	})

}
