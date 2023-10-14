package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type PartnerSyncDataLogSplitRepo struct{}

func (p *PartnerSyncDataLogSplitRepo) Create(ctx context.Context, db database.QueryExecer, log *entities.PartnerSyncDataLogSplit) error {
	ctx, span := interceptors.StartSpan(ctx, "PartnerSyncDataLogSplitRepo.Create")
	defer span.End()

	now := timeutil.Now()
	if err := multierr.Combine(
		log.CreatedAt.Set(now),
		log.UpdatedAt.Set(now),
	); err != nil {
		return err
	}

	if _, err := database.InsertIgnoreConflict(ctx, log, db.Exec); err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	return nil
}

func (p *PartnerSyncDataLogSplitRepo) UpdateLogStatus(ctx context.Context, db database.QueryExecer, log *entities.PartnerSyncDataLogSplit) error {
	ctx, span := interceptors.StartSpan(ctx, "PartnerSyncDataLogSplitRepo.UpdateLogStatus")
	defer span.End()

	updateLogStatusStmt := fmt.Sprintf(`UPDATE %s SET status = $1, updated_at = now() WHERE partner_sync_data_log_split_id = $2`, log.TableName())
	cmdTag, err := db.Exec(ctx, updateLogStatusStmt, log.Status, log.PartnerSyncDataLogSplitID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

func (p *PartnerSyncDataLogSplitRepo) UpdateLogsStatusAndRetryTime(ctx context.Context, db database.QueryExecer, logs []*entities.PartnerSyncDataLogSplit) error {
	ctx, span := interceptors.StartSpan(ctx, "PartnerSyncDataLogSplitRepo.UpdateLogStatusAndRetryTime")
	defer span.End()

	b := &pgx.Batch{}

	for _, log := range logs {
		query := fmt.Sprintf(`UPDATE %s SET status = $1, retry_times = $2, updated_at = now() WHERE partner_sync_data_log_split_id = $3`, log.TableName())
		b.Queue(query, &log.Status, &log.RetryTimes, &log.PartnerSyncDataLogSplitID)
	}

	results := db.SendBatch(ctx, b)
	defer results.Close()

	for i := 0; i < b.Len(); i++ {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("results.Exec: %w", err)
		}
	}

	return nil
}

func (p *PartnerSyncDataLogSplitRepo) GetLogsBySignature(ctx context.Context, db database.QueryExecer, signature pgtype.Text) ([]*entities.PartnerSyncDataLogSplit, error) {
	ctx, span := interceptors.StartSpan(ctx, "PartnerSyncDataLogSplitRepo.GetLogBySignature")
	defer span.End()

	query := `SELECT ls.partner_sync_data_log_split_id, ls.status, ls.updated_at 
		FROM partner_sync_data_log l INNER JOIN partner_sync_data_log_split ls ON l.partner_sync_data_log_id = ls.partner_sync_data_log_id
		WHERE l.signature = $1`

	rows, err := db.Query(ctx, query, &signature)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	partnerLogSplits := []*entities.PartnerSyncDataLogSplit{}
	for rows.Next() {
		var (
			id        pgtype.Text
			status    pgtype.Text
			updatedAt pgtype.Timestamp
		)

		if err := rows.Scan(&id, &status, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan fail: %w", err)
		}
		partnerLogSplit := &entities.PartnerSyncDataLogSplit{
			PartnerSyncDataLogSplitID: id,
			Status:                    status,
			UpdatedAt:                 pgtype.Timestamptz(updatedAt),
		}
		partnerLogSplits = append(partnerLogSplits, partnerLogSplit)
	}

	return partnerLogSplits, nil
}

func (p *PartnerSyncDataLogSplitRepo) GetLogsReportByDate(ctx context.Context, db database.QueryExecer, fromDate, toDate pgtype.Date) ([]*entities.PartnerSyncDataLogReport, error) {
	ctx, span := interceptors.StartSpan(ctx, "PartnerSyncDataLogSplitRepo.GetLogsReportByDate")
	defer span.End()

	query := `select count(*), status, CAST(created_at AS DATE)
	from partner_sync_data_log_split where DATE(created_at) >= $1 and DATE(created_at) <= $2
	group by CAST(created_at AS DATE), status
	order by created_at asc`

	rows, err := db.Query(ctx, query, &fromDate, &toDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reports := []*entities.PartnerSyncDataLogReport{}
	for rows.Next() {
		var (
			total     pgtype.Int8
			status    pgtype.Text
			createdAt pgtype.Date
		)

		if err := rows.Scan(&total, &status, &createdAt); err != nil {
			return nil, fmt.Errorf("scan fail: %w", err)
		}
		report := &entities.PartnerSyncDataLogReport{
			Total:     total,
			Status:    status,
			CreatedAt: createdAt,
		}
		reports = append(reports, report)
	}

	return reports, nil
}

func (p *PartnerSyncDataLogSplitRepo) GetLogsByDateRange(ctx context.Context, db database.QueryExecer, fromDate, toDate pgtype.Date) ([]*entities.PartnerSyncDataLogSplit, error) {
	ctx, span := interceptors.StartSpan(ctx, "PartnerSyncDataLogSplitRepo.GetLogsByDateRange")
	defer span.End()

	log := &entities.PartnerSyncDataLogSplit{}
	fields := database.GetFieldNames(log)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE DATE(created_at) >= $1 and DATE(created_at) <= $2", strings.Join(fields, ","), log.TableName())
	rows, err := db.Query(ctx, query, &fromDate, &toDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*entities.PartnerSyncDataLogSplit
	for rows.Next() {
		partnerLog := new(entities.PartnerSyncDataLogSplit)
		if err := rows.Scan(database.GetScanFields(partnerLog, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		logs = append(logs, partnerLog)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return logs, nil
}
