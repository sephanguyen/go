// BEGIN: 8f7e2d3b8c5d
package dto_test

import (
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/repository/postgres/dto"
	"github.com/stretchr/testify/assert"
)

func TestStudyPlanItemDto_ToStudyPlanItemEntity(t *testing.T) {
	now := time.Now()
	var pgTime, nilTime pgtype.Timestamp
	_ = pgTime.Set(now)
	_ = nilTime.Set(nil)
	studyPlanItemDto := &dto.StudyPlanItemDto{
		StudyPlanItemID: pgtype.Text{String: "id", Status: pgtype.Present},
		StudyPlanID:     pgtype.Text{String: "plan_id", Status: pgtype.Present},
		LmListID:        pgtype.Text{String: "lm_list_id", Status: pgtype.Present},
		Name:            pgtype.Text{String: "name", Status: pgtype.Present},
		StartDate:       pgTime,
		EndDate:         pgTime,
		DisplayOrder:    pgtype.Int4{Int: 1, Status: pgtype.Present},
		Status:          pgtype.Text{String: "status", Status: pgtype.Present},
		CreatedAt:       pgTime,
		UpdatedAt:       pgTime,
		DeletedAt:       nilTime,
	}

	expected := &dto.StudyPlanItemDto{
		StudyPlanItemID: pgtype.Text{String: "id", Status: pgtype.Present},
		StudyPlanID:     pgtype.Text{String: "plan_id", Status: pgtype.Present},
		LmListID:        pgtype.Text{String: "lm_list_id", Status: pgtype.Present},
		Name:            pgtype.Text{String: "name", Status: pgtype.Present},
		StartDate:       pgTime,
		EndDate:         pgTime,
		DisplayOrder:    pgtype.Int4{Int: 1, Status: pgtype.Present},
		Status:          pgtype.Text{String: "status", Status: pgtype.Present},
		CreatedAt:       pgTime,
		UpdatedAt:       pgTime,
		DeletedAt:       nilTime,
	}

	studyPlanItemEntity := domain.StudyPlanItem{
		StudyPlanItemID: "id",
		StudyPlanID:     "plan_id",
		LmListID:        "lm_list_id",
		Name:            "name",
		StartDate:       now,
		EndDate:         now,
		DisplayOrder:    1,
		Status:          "status",
		CreatedAt:       now,
		UpdatedAt:       now,
		DeletedAt:       nil,
	}

	result, err := studyPlanItemDto.FromEntity(studyPlanItemEntity)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

// END: 8f7e2d3b8c5d
