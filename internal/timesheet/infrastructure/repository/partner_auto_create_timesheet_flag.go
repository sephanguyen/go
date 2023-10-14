package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
)

type PartnerAutoCreateTimesheetFlagRepoImpl struct {
}

func (r *PartnerAutoCreateTimesheetFlagRepoImpl) GetPartnerAutoCreateDefaultValue(ctx context.Context, db database.QueryExecer) (*entity.PartnerAutoCreateTimesheetFlag, error) {
	ctx, span := interceptors.StartSpan(ctx, "PartnerAutoCreateTimesheetFlagRepoImpl.GetPartnerAutoCreateDefaultValue")
	defer span.End()

	partnerAutoCreateFlagE := &entity.PartnerAutoCreateTimesheetFlag{}
	fields, _ := partnerAutoCreateFlagE.FieldMap()

	stmt := fmt.Sprintf(`SELECT %s FROM %s
	WHERE deleted_at IS NULL
	limit 1`, strings.Join(fields, ", "), partnerAutoCreateFlagE.TableName())

	if err := database.Select(ctx, db, stmt).ScanOne(partnerAutoCreateFlagE); err != nil {
		return nil, err
	}
	return partnerAutoCreateFlagE, nil
}
