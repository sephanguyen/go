package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type StudyPlanRepo struct{}

func (a *StudyPlanRepo) Upsert(ctx context.Context, db database.Ext, now time.Time, studyPlan domain.StudyPlan) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanRepo.Upsert")
	defer span.End()

	var returnID string

	studyPlanDto := dto.StudyPlan{}
	if err := studyPlanDto.FromEntity(now, studyPlan); err != nil {
		return "", errors.NewConversionError("StudyPlanDto.FromEntity", err)
	}

	fields, values := studyPlanDto.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))

	stmt := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES(%s)
		ON CONFLICT ON CONSTRAINT lms_study_plans_pkey DO
		UPDATE SET deleted_at = NULL, updated_at = EXCLUDED.updated_at, name = EXCLUDED.name
		RETURNING study_plan_id;`,
		studyPlanDto.TableName(),
		strings.Join(fields, ","),
		placeHolders,
	)

	if err := db.QueryRow(ctx, stmt, values...).Scan(&returnID); err != nil {
		return "", errors.NewDBError("db.QueryRow", err)
	}

	return returnID, nil
}
