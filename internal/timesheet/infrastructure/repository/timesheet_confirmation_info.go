package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"

	"github.com/jackc/pgtype"
)

type TimesheetConfirmationInfoRepoImpl struct {
}

func (r *TimesheetConfirmationInfoRepoImpl) InsertConfirmationInfo(ctx context.Context, db database.QueryExecer, confirmationInfoE *entity.TimesheetConfirmationInfo) (*entity.TimesheetConfirmationInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetConfirmationInfoRepoImpl.InsertConfirmationInfo")
	defer span.End()

	if err := confirmationInfoE.PreInsert(); err != nil {
		return nil, fmt.Errorf("PreInsert Confirmation Info failed, err: %w", err)
	}

	cmdTag, err := database.Insert(ctx, confirmationInfoE, db.Exec)
	if err != nil {
		return nil, fmt.Errorf("err insert Confirmation Info: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return nil, fmt.Errorf("err insert Confirmation Info: %d RowsAffected", cmdTag.RowsAffected())
	}

	return confirmationInfoE, nil
}

func (r *TimesheetConfirmationInfoRepoImpl) GetConfirmationInfoByPeriodAndLocation(ctx context.Context, db database.QueryExecer, periodID, locationID pgtype.Text) (*entity.TimesheetConfirmationInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetConfirmationInfoRepoImpl.GetConfirmationInfoByPeriodAndLocation")
	defer span.End()

	tsConfirmationInfo := &entity.TimesheetConfirmationInfo{}
	fields, _ := tsConfirmationInfo.FieldMap()

	stmt := fmt.Sprintf(`SELECT %s FROM %s
	WHERE deleted_at IS NULL
	AND period_id = $1
	AND location_id = $2`, strings.Join(fields, ", "), tsConfirmationInfo.TableName())

	if err := database.Select(ctx, db, stmt, periodID, locationID).ScanOne(tsConfirmationInfo); err != nil {
		return nil, err
	}

	return tsConfirmationInfo, nil
}
