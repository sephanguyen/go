package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
)

type AutoCreateFlagActivityLogRepoImpl struct {
}

func (r *AutoCreateFlagActivityLogRepoImpl) InsertFlagLog(ctx context.Context, db database.QueryExecer, flagLogData *entity.AutoCreateFlagActivityLog) (*entity.AutoCreateFlagActivityLog, error) {
	ctx, span := interceptors.StartSpan(ctx, "AutoCreateFlagActivityLogRepoImpl.InsertTimeSheet")
	defer span.End()

	if err := flagLogData.PreInsert(); err != nil {
		return nil, fmt.Errorf("PreInsert auto create log data failed, err: %w", err)
	}
	cmdTag, err := database.Insert(ctx, flagLogData, db.Exec)
	if err != nil {
		return nil, fmt.Errorf("err insert auto create log data: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return nil, fmt.Errorf("err insert auto create log data: %d RowsAffected", cmdTag.RowsAffected())
	}

	return flagLogData, nil
}

func (r *AutoCreateFlagActivityLogRepoImpl) GetAutoCreateFlagActivityLogByStaffIDs(ctx context.Context, db database.QueryExecer, lessonStartDate time.Time, teacherIDs []string) ([]*entity.AutoCreateFlagActivityLog, error) {
	ctx, span := interceptors.StartSpan(ctx, "AutoCreateFlagActivityLogRepoImpl.GetAutoCreateFlagActivityLogByStaffIDs")
	defer span.End()

	autoCreateFlagActivityLog := &entity.AutoCreateFlagActivityLog{}
	listAutoCreateFlagActivityLog := &entity.AutoCreateFlagActivityLogs{}

	fields, _ := autoCreateFlagActivityLog.FieldMap()
	orderBy := "staff_id,change_time DESC"
	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE staff_id = ANY($1::_TEXT) AND change_time <= $2 AND deleted_at IS NULL ORDER BY %s`,
		strings.Join(fields, ","), autoCreateFlagActivityLog.TableName(), orderBy)
	if err := database.Select(ctx, db, stmt, &teacherIDs, &lessonStartDate).ScanAll(listAutoCreateFlagActivityLog); err != nil {
		return nil, err
	}
	return *listAutoCreateFlagActivityLog, nil
}

func (r *AutoCreateFlagActivityLogRepoImpl) SoftDeleteFlagLogsAfterTime(ctx context.Context, db database.QueryExecer, staffID string, startTime time.Time) error {
	ctx, span := interceptors.StartSpan(ctx, "AutoCreateFlagActivityLogRepoImpl.SoftDeleteFlagLogsAfterTime")
	defer span.End()

	autoCreateFlagActivityLog := &entity.AutoCreateFlagActivityLog{}

	// soft delete flag logs after time
	stmt := fmt.Sprintf(`
		UPDATE %s SET deleted_at = NOW()
		WHERE staff_id = $1 AND change_time >= $2 AND deleted_at IS NULL;
	`, autoCreateFlagActivityLog.TableName())

	_, err := db.Exec(ctx, stmt, &staffID, &startTime)
	if err != nil {
		return fmt.Errorf("err delete SoftDeleteFlagLogsAfterTime: %w", err)
	}

	return nil
}
