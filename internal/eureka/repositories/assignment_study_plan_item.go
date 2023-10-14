package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	eureka_db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type AssignmentStudyPlanItemRepo struct {
}

const bulkUpsertAssignmentStudyPlanItemStmtTpl = `
INSERT INTO %s (%s) 
VALUES %s ON CONFLICT ON CONSTRAINT assignment_study_plan_items_pk DO UPDATE
SET updated_at = excluded.updated_at`

func (r *AssignmentStudyPlanItemRepo) BulkInsert(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error {
	err := eureka_db.BulkUpsert(ctx, db, bulkUpsertAssignmentStudyPlanItemStmtTpl, assignmentStudyPlanItems)
	if err != nil {
		return fmt.Errorf("eureka_db.BulkUpsertAssignmentStudyPlanItem error: %s", err.Error())
	}
	return nil
}

func (r *AssignmentStudyPlanItemRepo) CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error {
	query := `INSERT INTO assignment_study_plan_items 
		SELECT astpi.assignment_id, spi.study_plan_item_id, astpi.created_at, astpi.updated_at, astpi.deleted_at from study_plan_items spi 
		JOIN assignment_study_plan_items astpi on astpi.study_plan_item_id = spi.copy_study_plan_item_id 
		WHERE spi.study_plan_id = ANY($1) AND astpi.deleted_at is NULL`
	_, err := db.Exec(ctx, query, &studyPlanIDs)
	if err != nil {
		return err
	}
	return nil
}

func (r *AssignmentStudyPlanItemRepo) FindByStudyPlanItemIDs(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*entities.AssignmentStudyPlanItem, error) {
	var assignmentStudyPlanItems entities.AssignmentStudyPlanItems
	e := &entities.AssignmentStudyPlanItem{}
	fieldNames := database.GetFieldNames(e)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE study_plan_item_id = ANY($1)", strings.Join(fieldNames, ", "), e.TableName())
	err := database.Select(ctx, db, query, &studyPlanItemIDs).ScanAll(&assignmentStudyPlanItems)
	if err != nil {
		return nil, err
	}
	return assignmentStudyPlanItems, nil
}

func (r *AssignmentStudyPlanItemRepo) EditAssignmentTime(
	ctx context.Context,
	db database.QueryExecer,
	studentID pgtype.Text,
	studyPlanItemIDs pgtype.TextArray,
	startDate, endDate pgtype.Timestamptz,
) error {
	query := `
		UPDATE study_plan_items spi
		SET start_date = $1, end_date = $2
		WHERE study_plan_item_id IN (
			SELECT study_plan_item_id
			FROM study_plan_items spi
			INNER JOIN student_study_plans ssp ON ssp.study_plan_id = spi.study_plan_id
			WHERE ssp.student_id = $3
			AND spi.study_plan_item_id = ANY($4)
			AND ($1::timestamptz IS NULL OR spi.available_from <= $1::timestamptz)
			AND ($2::timestamptz IS NULL OR spi.available_to > $2::timestamptz OR spi.available_to IS NULL)
		)
	`
	cTag, err := db.Exec(ctx, query, &startDate, &endDate, &studentID, &studyPlanItemIDs)
	if err != nil {
		return err
	}
	// if length of the rows effected is less than length of study plan items, maybe one or more invalid time so can't update
	if cTag.RowsAffected() != int64(len(studyPlanItemIDs.Elements)) {
		return fmt.Errorf("cannot update all study plan items")
	}
	return nil
}

func (r *AssignmentStudyPlanItemRepo) BulkEditAssignmentTime(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, ens []*entities.StudyPlanItem) error {
	b := &pgx.Batch{}
	for _, item := range ens {
		query := `
			UPDATE study_plan_items spi
			SET start_date = $1, end_date = $2, updated_at = NOW()
			WHERE study_plan_item_id IN (
				SELECT study_plan_item_id
				FROM study_plan_items spi
				INNER JOIN student_study_plans ssp ON ssp.study_plan_id = spi.study_plan_id
				WHERE ssp.student_id = $3
				AND spi.study_plan_item_id = ANY($4)
				AND ($1::timestamptz IS NULL OR spi.available_from <= $1::timestamptz)
				AND ($2::timestamptz IS NULL OR spi.available_to > $2::timestamptz OR spi.available_to IS NULL)
			)
		`
		b.Queue(query, &item.StartDate, &item.EndDate, &studentID, database.TextArray([]string{item.ID.String}))
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	affectedRows := 0
	for i := 0; i < b.Len(); i++ {
		ctag, err := result.Exec()
		if err != nil {
			return fmt.Errorf("BulkEditAssignmentTime.Exec: %w", err)
		}
		if ctag.RowsAffected() == 1 {
			affectedRows++
		}
	}

	if affectedRows != len(ens) {
		return fmt.Errorf("cannot update all study plan items")
	}
	return nil
}

func (r *AssignmentStudyPlanItemRepo) CountAssignment(ctx context.Context, db database.QueryExecer, assignmentIDs pgtype.TextArray) (int, error) {
	var counter pgtype.Int8
	e := &entities.Assignment{}
	query := `SELECT COUNT (*) FROM %s WHERE assignment_id = ANY($1::_TEXT) AND deleted_at IS NULL`
	if err := db.QueryRow(ctx, fmt.Sprintf(query, e.TableName()), assignmentIDs).Scan(&counter); err != nil {
		return int(counter.Int), err
	}
	return int(counter.Int), nil
}

func (r *AssignmentStudyPlanItemRepo) BulkCopy(ctx context.Context, db database.QueryExecer, items []*entities.AssignmentStudyPlanItem) error {
	const queueCopyItemStmt = `
INSERT INTO
	assignment_study_plan_items
SELECT
	$1::text AS assignment_id,
	spi.study_plan_item_id,
	NOW(),
	NOW(),
	NULL
FROM
	study_plan_items spi
WHERE
	spi.copy_study_plan_item_id = $2
ON CONFLICT DO NOTHING;
`

	queueFn := func(b *pgx.Batch, item *entities.AssignmentStudyPlanItem) {
		b.Queue(queueCopyItemStmt, &item.AssignmentID, &item.StudyPlanItemID)
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

func (r *AssignmentStudyPlanItemRepo) SoftDeleteByAssigmentIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (pgtype.TextArray, error) {
	const query = `
		UPDATE assignment_study_plan_items aspi
			SET deleted_at = NOW()
		WHERE assignment_id = ANY($1::TEXT[]) AND deleted_at IS NULL
		RETURNING aspi.study_plan_item_id
	`

	var res pgtype.TextArray
	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return res, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var studyPlanIDs []string
	for rows.Next() {
		var id pgtype.Text
		if err := rows.Scan(&id); err != nil {
			return res, fmt.Errorf("db.Scan: %w", err)
		}
		studyPlanIDs = append(studyPlanIDs, id.String)
	}
	if err := rows.Err(); err != nil {
		return res, fmt.Errorf("db.Error: %w", err)
	}

	res = database.TextArray(studyPlanIDs)
	return res, nil
}

func (r *AssignmentStudyPlanItemRepo) BulkUpsertByStudyPlanItem(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentStudyPlanItemRepo.QueueUpsertByStudyPlanItem")
	defer span.End()
	err := eureka_db.BulkUpsert(ctx, db, bulkUpsertAssignmentStudyPlanItemStmtTpl, assignmentStudyPlanItems)
	if err != nil {
		return fmt.Errorf("eureka_db.BulkUpsertByStudyPlanItem error: %s", err.Error())
	}
	return nil
}
