package repositories

import (
	"context"
	"fmt"
	"strings"

	entities "github.com/manabie-com/backend/internal/eureka/entities/monitors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type StudyPlanMonitorRepo struct{}

// UCL upper control limit
// LCL lower control limit
type RetrieveFilter struct {
	StudyPlanMonitorType pgtype.Text
	IntervalTimeLCL      *pgtype.Text
	IntervalTimeULC      *pgtype.Text
}

func (r *StudyPlanMonitorRepo) QueueUpsertStudyPlanMonitor(b *pgx.Batch, item *entities.StudyPlanMonitor) {
	fieldNames := database.GetFieldNames(item)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT study_plan_monitor_pk DO UPDATE SET
		updated_at = $8`, item.TableName(), strings.Join(fieldNames, ","), placeHolders)
	scanFields := database.GetScanFields(item, fieldNames)
	b.Queue(query, scanFields...)
}

func (r *StudyPlanMonitorRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, studyPlanMonitorItems []*entities.StudyPlanMonitor) error {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanMonitorRepo.BulkUpsert")
	defer span.End()

	b := &pgx.Batch{}
	for _, item := range studyPlanMonitorItems {
		r.QueueUpsertStudyPlanMonitor(b, item)
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}
func (r *StudyPlanMonitorRepo) RetrieveByFilter(ctx context.Context, db database.QueryExecer, filter *RetrieveFilter) ([]*entities.StudyPlanMonitor, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanMonitorRepo.RetrieveByFilter")
	defer span.End()

	retrieveStmtPlt := `
	SELECT %s FROM %s 
	WHERE ($1::TEXT IS NULL OR "type" = $1::TEXT)
	AND ($2::interval IS NULL OR updated_at >= ( now() - $2::interval))
	AND ($3::interval IS NULL OR updated_at <= ( now() - $3::interval))
	AND deleted_at IS NULL;`
	var e entities.StudyPlanMonitor
	selectFields := database.GetFieldNames(&e)

	var items entities.StudyPlanMonitors
	err := database.Select(ctx, db, fmt.Sprintf(retrieveStmtPlt, strings.Join(selectFields, ","), e.TableName()), &filter.StudyPlanMonitorType, filter.IntervalTimeLCL, filter.IntervalTimeULC).ScanAll(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *StudyPlanMonitorRepo) MarkItemsAutoUpserted(ctx context.Context, db database.QueryExecer, studyPlanMonitorIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanMonitorRepo.MarkItemAutoUpserted")
	defer span.End()
	e := &entities.StudyPlanMonitor{}

	query := fmt.Sprintf("UPDATE %s SET auto_upserted_at = now(), updated_at = now() WHERE study_plan_monitor_id = ANY($1::TEXT[]) AND deleted_at IS NULL", e.TableName())
	_, err := db.Exec(ctx, query, &studyPlanMonitorIDs)
	if err != nil {
		return err
	}

	return nil
}

func (r *StudyPlanMonitorRepo) SoftDelete(ctx context.Context, db database.QueryExecer, studyPlanMonitorIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanMonitorRepo.SoftDelete")
	defer span.End()
	e := &entities.StudyPlanMonitor{}

	query := fmt.Sprintf("UPDATE %s SET deleted_at = now() WHERE study_plan_monitor_id = ANY($1::_TEXT)AND deleted_at IS NULL", e.TableName())
	_, err := db.Exec(ctx, query, &studyPlanMonitorIDs)
	if err != nil {
		return err
	}

	return nil
}

func (r *StudyPlanMonitorRepo) SoftDeleteTypeStudyPlan(ctx context.Context, db database.QueryExecer, filter *RetrieveFilter) error {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanMonitorRepo.SoftDeleteTypeStudyPlan")
	defer span.End()
	cmd := `
	WITH TMP AS( 
		SELECT spm.study_plan_monitor_id
		FROM student_study_plans ssp JOIN study_plans sp  
		USING(study_plan_id) 
		JOIN study_plan_monitors as spm
		USING(student_id,course_id)
		WHERE ssp.deleted_at IS NULL
		AND sp.deleted_at IS NULL
		AND spm.deleted_at IS NULL
		AND ($1::TEXT IS NULL OR "type" = $1::TEXT)
		AND ($2::interval IS NULL OR spm.updated_at >= ( now() - $2::interval))
		AND ($3::interval IS NULL OR spm.updated_at <= ( now() - $3::interval)) 
	)
	UPDATE study_plan_monitors  SET deleted_at = now() 
	WHERE study_plan_monitor_id IN (SELECT * FROM TMP)
	`
	// query := fmt.Sprintf("UPDATE %s SET deleted_at = now() WHERE study_plan_monitor_id = ANY($1::_TEXT)AND deleted_at IS NULL", e.TableName())
	_, err := db.Exec(ctx, cmd, &filter.StudyPlanMonitorType, filter.IntervalTimeLCL, filter.IntervalTimeULC)
	if err != nil {
		return err
	}

	return nil
}
