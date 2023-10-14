package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	eureka_db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type LoStudyPlanItemRepo struct{}

const BulkUpsertLoStudyPlanItem = `
INSERT INTO %s (%s)
VALUES %s ON CONFLICT ON CONSTRAINT lo_study_plan_items_pk DO UPDATE 
SET updated_at = excluded.updated_at`

func (r *LoStudyPlanItemRepo) BulkInsert(ctx context.Context, db database.QueryExecer, loStudyPlanItems []*entities.LoStudyPlanItem) error {
	err := eureka_db.BulkUpsert(ctx, db, BulkUpsertLoStudyPlanItem, loStudyPlanItems)
	if err != nil {
		return fmt.Errorf("eureka_db.BulkUpsertLoStudyPlanItem error: %s", err.Error())
	}
	return nil
}

func (r *LoStudyPlanItemRepo) CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error {
	query := `INSERT INTO lo_study_plan_items 
		SELECT lospi.lo_id, spi.study_plan_item_id, lospi.created_at, lospi.updated_at, lospi.deleted_at from study_plan_items spi 
		JOIN lo_study_plan_items lospi on lospi.study_plan_item_id = spi.copy_study_plan_item_id 
		WHERE spi.study_plan_id = ANY($1) AND lospi.deleted_at is NULL`
	_, err := db.Exec(ctx, query, &studyPlanIDs)
	if err != nil {
		return err
	}
	return nil
}

func (r *LoStudyPlanItemRepo) FindByStudyPlanItemIDs(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*entities.LoStudyPlanItem, error) {
	var loStudyPlanItems entities.LoStudyPlanItems
	e := &entities.LoStudyPlanItem{}
	fieldNames := database.GetFieldNames(e)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE study_plan_item_id = ANY($1)", strings.Join(fieldNames, ", "), e.TableName())
	err := database.Select(ctx, db, query, &studyPlanItemIDs).ScanAll(&loStudyPlanItems)
	if err != nil {
		return nil, err
	}
	return loStudyPlanItems, nil
}

func (r *LoStudyPlanItemRepo) UpdateCompleted(ctx context.Context, db database.QueryExecer, studyPlanItemID pgtype.Text, loID pgtype.Text) error {
	s := &entities.StudyPlanItem{}
	lo := &entities.LoStudyPlanItem{}
	stmt := fmt.Sprintf(`UPDATE %s
	SET completed_at = $1 
	WHERE study_plan_item_id = (
		SELECT study_plan_item_id FROM %s WHERE study_plan_item_id = $2 AND lo_id = $3
	);`, s.TableName(), lo.TableName())

	completedAt := database.Timestamptz(time.Now())
	_, err := db.Exec(ctx, stmt, completedAt, studyPlanItemID, loID)
	if err != nil {
		return err
	}

	return nil
}

const queueCopyItemStmt = `
INSERT INTO
	lo_study_plan_items
SELECT
	$1::text AS lo_id,
	spi.study_plan_item_id,
	NOW(),
	NOW(),
	NULL
FROM
	study_plan_items spi
WHERE
	spi.copy_study_plan_item_id = $2 AND spi.deleted_at IS NULL
ON CONFLICT DO NOTHING;
`

func (r *LoStudyPlanItemRepo) BulkCopy(ctx context.Context, db database.QueryExecer, items []*entities.LoStudyPlanItem) error {
	queueFn := func(b *pgx.Batch, item *entities.LoStudyPlanItem) {
		b.Queue(queueCopyItemStmt, &item.LoID, &item.StudyPlanItemID)
	}

	b := &pgx.Batch{}
	for _, item := range items {
		queueFn(b, item)
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

func (r *LoStudyPlanItemRepo) DeleteLoStudyPlanItemsByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error {
	e := &entities.LoStudyPlanItem{}
	deleteLoStudyPlanItemsByLoIDsStmtTpl := "UPDATE %s SET deleted_at = now() WHERE lo_id = ANY($1::_TEXT)"
	_, err := db.Exec(ctx, fmt.Sprintf(deleteLoStudyPlanItemsByLoIDsStmtTpl, e.TableName()), loIDs)
	return err
}

func (r *LoStudyPlanItemRepo) DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error {
	lospi := &entities.LoStudyPlanItem{}
	spi := &entities.StudyPlanItem{}

	deleteStmtTpl := `
WITH deleted_lospi AS (
	UPDATE %s 
    SET deleted_at = NOW() 
    WHERE lo_id = ANY($1::_TEXT) 
    AND deleted_at IS NULL
    RETURNING study_plan_item_id 
)
UPDATE %s spi
SET deleted_at = NOW()
FROM deleted_lospi
WHERE spi.study_plan_item_id = deleted_lospi.study_plan_item_id
AND deleted_at IS NULL
  `
	_, err := db.Exec(ctx, fmt.Sprintf(deleteStmtTpl, lospi.TableName(), spi.TableName()), &loIDs)
	return err
}

func (r *LoStudyPlanItemRepo) BulkUpsertByStudyPlanItem(ctx context.Context, db database.QueryExecer, loStudyPlanItems []*entities.LoStudyPlanItem) error {
	err := eureka_db.BulkUpsert(ctx, db, BulkUpsertLoStudyPlanItem, loStudyPlanItems)
	if err != nil {
		return fmt.Errorf("eureka_db.BulkUpsertLoByStudyPlanItem error: %s", err.Error())
	}
	return nil
}
