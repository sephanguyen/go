package dto_test

import (
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/repository/postgres/dto"
	"github.com/stretchr/testify/assert"
)

func TestLmListDto_ToLmListEntity(t *testing.T) {
	now := time.Now()
	var pgTime, nilTime pgtype.Timestamp
	_ = pgTime.Set(now)
	_ = nilTime.Set(nil)

	var pgArray pgtype.TextArray
	_ = pgArray.Set([]string{"lm_id_1", "lm_id_2"})

	lmList := domain.LmList{
		LmListID:  "lm_list_id",
		LmIDs:     []string{"lm_id_1", "lm_id_2"},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
	}

	expected := &dto.LmListDto{
		LmListID:  pgtype.Text{String: "lm_list_id", Status: pgtype.Present},
		LmIDs:     pgArray,
		CreatedAt: pgTime,
		UpdatedAt: pgTime,
		DeletedAt: nilTime,
	}

	actual, err := (&dto.LmListDto{}).ToLmListEntity(lmList)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
