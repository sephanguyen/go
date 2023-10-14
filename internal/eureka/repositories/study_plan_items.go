package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	eureka_db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.opencensus.io/trace"
)

type StudyPlanItemRepo struct {
}

const bulkUpsertStudyPlanItemStmtTpl = `INSERT INTO
	%s (%s)
VALUES %s ON CONFLICT ON CONSTRAINT study_plan_items_pk DO UPDATE
SET
	available_from =excluded.available_from,
	available_to =excluded.available_to,
	start_date =excluded.start_date,
	end_date =excluded.end_date,
	updated_at = NOW(),
	content_structure =excluded.content_structure,
	display_order =excluded.display_order,
	status =excluded.status,
	school_date = COALESCE(excluded.school_date, study_plan_items.school_date)
	`

func (r *StudyPlanItemRepo) BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) error {
	err := eureka_db.BulkUpsert(ctx, db, bulkUpsertStudyPlanItemStmtTpl, items)
	if err != nil {
		return fmt.Errorf("eureka_db.BulkUpsertStudyPlanItem error: %s", err.Error())
	}
	return nil
}

const queueCopyStudyPlanItemStmt = `INSERT
	INTO
	study_plan_items(
		study_plan_item_id,
		study_plan_id,
		available_from,
		available_to,
		start_date,
		end_date,
		updated_at,
		created_at,
		deleted_at,
		completed_at,
		content_structure,
		display_order,
		copy_study_plan_item_id,
		status
	)
SELECT
	generate_ulid() AS student_plan_item_id,
	$1::TEXT AS study_plan_id,
	available_from,
	available_to,
	start_date,
	end_date,
	now() AS created_at,
	now() AS updated_at,
	deleted_at,
	completed_at,
	content_structure,
	display_order,
	study_plan_item_id AS copy_study_plan_item_id,
	status
FROM
	study_plan_items spi
WHERE
	spi.study_plan_id = $2`

func (r *StudyPlanItemRepo) QueueCopyStudyPlanItem(b *pgx.Batch, originalStudyPlanID pgtype.Text, copiedStudyPlanID pgtype.Text) {
	b.Queue(queueCopyStudyPlanItemStmt, &copiedStudyPlanID, &originalStudyPlanID)
}

func (r *StudyPlanItemRepo) BulkCopy(
	ctx context.Context, db database.QueryExecer,
	originalStudyPlanIDs pgtype.TextArray, newStudyPlanIDs pgtype.TextArray,
) error {
	b := &pgx.Batch{}
	if len(originalStudyPlanIDs.Elements) != len(newStudyPlanIDs.Elements) {
		return fmt.Errorf("original study plan ids and new study plan ids not match")
	}
	for i, originalStudyPlanID := range originalStudyPlanIDs.Elements {
		r.QueueCopyStudyPlanItem(b, originalStudyPlanID, newStudyPlanIDs.Elements[i])
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

func (r *StudyPlanItemRepo) FindByStudyPlanID(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text) ([]*entities.StudyPlanItem, error) {
	studyPlanItem := &entities.StudyPlanItem{}
	fieldNames := database.GetFieldNames(studyPlanItem)
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE study_plan_id = $1`, strings.Join(fieldNames, ", "), studyPlanItem.TableName())

	var studyPlanItems entities.StudyPlanItems
	err := database.Select(ctx, db, query, &studyPlanID).ScanAll(&studyPlanItems)
	if err != nil {
		return nil, err
	}
	return studyPlanItems, nil
}

func (r *StudyPlanItemRepo) FindByStudyPlanIDAndTopicIDs(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text, topicIDs pgtype.TextArray) ([]*entities.StudyPlanItem, error) {
	studyPlanItem := &entities.StudyPlanItem{}
	fieldNames := database.GetFieldNames(studyPlanItem)
	query := fmt.Sprintf(`
	SELECT %s 
	FROM %s 
	WHERE study_plan_id = $1 
	AND content_structure ->> 'topic_id' = ANY($2)
	AND deleted_at IS NULL
	`, strings.Join(fieldNames, ", "), studyPlanItem.TableName())

	var studyPlanItems entities.StudyPlanItems
	err := database.Select(ctx, db, query, &studyPlanID, &topicIDs).ScanAll(&studyPlanItems)
	if err != nil {
		return nil, err
	}
	return studyPlanItems, nil
}

const markItemsCompletedStmt = `UPDATE
study_plan_items
SET
completed_at = NOW()
WHERE
study_plan_item_id = $1`

// MarkItemCompleted marks all items completed_at = NOW()
func (r *StudyPlanItemRepo) MarkItemCompleted(ctx context.Context, db database.QueryExecer,
	itemID pgtype.Text) error {
	_, err := db.Exec(ctx, markItemsCompletedStmt, &itemID)
	if err != nil {
		return err
	}

	return nil
}

func (r *StudyPlanItemRepo) UnMarkItemCompleted(ctx context.Context, db database.QueryExecer,
	itemID pgtype.Text) error {
	const unMarkItemCompletedStmt = `UPDATE
study_plan_items
SET
completed_at = NULL
WHERE
study_plan_item_id = $1`
	_, err := db.Exec(ctx, unMarkItemCompletedStmt, &itemID)
	if err != nil {
		return err
	}

	return nil
}

const updateByCopiedStudyPlanItemStmt = `
UPDATE
	study_plan_items
SET
	available_from = $1,
	start_date = $2,
	end_date = $3,
	available_to = $4,
	display_order = $5,
	updated_at = NOW(),
	deleted_at = NULL,
	status = $7
WHERE
	copy_study_plan_item_id = $6
RETURNING %s`

func (r *StudyPlanItemRepo) UpdateWithCopiedFromItem(ctx context.Context, db database.QueryExecer, studyPlanItems []*entities.StudyPlanItem) error {
	b := &pgx.Batch{}
	studyPlanItem := &entities.StudyPlanItem{}
	fieldNames := database.GetFieldNames(studyPlanItem)
	stmt := fmt.Sprintf(updateByCopiedStudyPlanItemStmt, strings.Join(fieldNames, ", "))
	for _, studyPlanItem := range studyPlanItems {
		b.Queue(stmt, &studyPlanItem.AvailableFrom, &studyPlanItem.StartDate,
			&studyPlanItem.EndDate, &studyPlanItem.AvailableTo, &studyPlanItem.DisplayOrder,
			&studyPlanItem.ID, &studyPlanItem.Status)
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

const findByIDStmt = `
SELECT study_plan_item_id, study_plan_id
FROM study_plan_items 
WHERE study_plan_item_id = ANY($1)
`

func (r *StudyPlanItemRepo) FindStudyPlanIDByItemIDs(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) (map[string]string, error) {
	rows, err := db.Query(ctx, findByIDStmt, studyPlanIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	studyPlanIDMap := make(map[string]string)

	for rows.Next() {
		var studyPlanItemID, studyPlanID string
		if err := rows.Scan(&studyPlanItemID, &studyPlanID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		studyPlanIDMap[studyPlanItemID] = studyPlanID
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Scan: %w", err)
	}

	return studyPlanIDMap, nil
}

func (r *StudyPlanItemRepo) FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.StudyPlanItem, error) {
	studyPlanItem := &entities.StudyPlanItem{}
	fieldNames := database.GetFieldNames(studyPlanItem)
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE study_plan_item_id = ANY($1) AND deleted_at IS NULL ORDER BY display_order ASC`, strings.Join(fieldNames, ", "), studyPlanItem.TableName())

	var studyPlanItems entities.StudyPlanItems
	err := database.Select(ctx, db, query, &ids).ScanAll(&studyPlanItems)
	if err != nil {
		return nil, err
	}
	return studyPlanItems, nil
}

// FindAndSortInBookByIDs return all study_plan_items sorted by display order from Learning Materials Tree
func (r *StudyPlanItemRepo) FindAndSortByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.StudyPlanItem, error) {
	studyPlanItem := &entities.StudyPlanItem{}
	fieldNames := database.GetFieldNames(studyPlanItem)
	const queryTmpl = `
WITH study_plan_items_tmp AS (
  SELECT spi.%s, spi.content_structure ->> 'book_id' AS book_id, 
                 spi.content_structure ->> 'chapter_id' AS chapter_id,
                 spi.content_structure ->> 'topic_id' AS topic_id,
                 COALESCE(NULLIF(spi.content_structure ->> 'assignment_id', ''), 
                                 spi.content_structure ->> 'lo_id') AS lm_id
  FROM %s spi
  WHERE spi.study_plan_item_id = ANY($1::TEXT[])
    AND spi.deleted_at IS NULL
) SELECT study_plan_items_tmp.%s
  FROM study_plan_items_tmp
  JOIN books b ON b.book_id = study_plan_items_tmp.book_id
  join books_chapters bc on bc.book_id = b.book_id
  JOIN chapters c ON c.chapter_id = bc.chapter_id and c.chapter_id = study_plan_items_tmp.chapter_id
  JOIN topics t ON t.chapter_id = c.chapter_id  and t.topic_id = study_plan_items_tmp.topic_id
  LEFT JOIN topics_assignments ta ON 
    ta.topic_id = study_plan_items_tmp.topic_id AND ta.assignment_id = study_plan_items_tmp.lm_id
  LEFT JOIN topics_learning_objectives tlo ON 
    tlo.topic_id = study_plan_items_tmp.topic_id AND tlo.lo_id = study_plan_items_tmp.lm_id
  WHERE b.deleted_at IS NULL
    AND c.deleted_at IS NULL
    AND t.deleted_at IS NULL
    AND (ta.deleted_at IS NULL OR tlo.deleted_at IS NULL)
  ORDER BY b.book_id,
           c.display_order,
           t.display_order,
           coalesce(ta.display_order, tlo.display_order)
`

	query := fmt.Sprintf(queryTmpl, strings.Join(fieldNames, ", spi."), studyPlanItem.TableName(), strings.Join(fieldNames, ", study_plan_items_tmp."))
	var studyPlanItems entities.StudyPlanItems
	err := database.Select(ctx, db, query, &ids).ScanAll(&studyPlanItems)
	if err != nil {
		return nil, err
	}
	return studyPlanItems, nil
}

const softDeleteByStudyPlanIDs = `
UPDATE study_plan_items SET deleted_at = NOW() 
WHERE study_plan_id = ANY($1)
`

func (r *StudyPlanItemRepo) SoftDeleteWithStudyPlanIDs(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error {
	_, err := db.Exec(ctx, softDeleteByStudyPlanIDs, &studyPlanIDs)
	if err != nil {
		return err
	}
	return nil
}

// CountStudentInStudyPlanItem count number of student of every study plan item which include completed item or not
func (r *StudyPlanItemRepo) CountStudentInStudyPlanItem(ctx context.Context, db database.QueryExecer, masterStudentStudyPlanID pgtype.Text, onlyCompleted pgtype.Bool) (map[string]int, []string, error) {
	query := `
	SELECT spi.copy_study_plan_item_id AS root_study_plan_item_id, COUNT(*) AS count_student 
	FROM study_plan_items spi 
		INNER JOIN study_plans sp ON spi.study_plan_id = sp.study_plan_id 
	WHERE sp.master_study_plan_id = $1
	AND ($2 = FALSE OR spi.completed_at IS NOT NULL)
	GROUP BY spi.copy_study_plan_item_id
	ORDER BY spi.copy_study_plan_item_id ASC;`

	studyPlanItemCountStudent := make(map[string]int)
	sortedStudyPlanItem := make([]string, 0)
	rows, err := db.Query(ctx, query, masterStudentStudyPlanID, onlyCompleted)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var studyPlanItemID pgtype.Text
		var countStudent pgtype.Int8
		err := rows.Scan(&studyPlanItemID, &countStudent)
		if err != nil {
			return nil, nil, err
		}
		studyPlanItemCountStudent[studyPlanItemID.String] = int(countStudent.Int)
		sortedStudyPlanItem = append(sortedStudyPlanItem, studyPlanItemID.String)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return studyPlanItemCountStudent, sortedStudyPlanItem, nil
}

const retrieveChildStudyPlanItemStmtPlt = `SELECT 
	ssp.student_id, spi.%s
FROM study_plan_items spi JOIN study_plans sp USING(study_plan_id) JOIN student_study_plans ssp USING(study_plan_id)
WHERE spi.copy_study_plan_item_id = $1
    AND ($2::_text IS NULL OR ssp.student_id = ANY($2::_text))
    AND spi.deleted_at IS NULL
    AND sp.deleted_at IS NULL
    AND ssp.deleted_at IS NULL
ORDER BY ssp.student_id ASC`

func (r *StudyPlanItemRepo) RetrieveChildStudyPlanItem(ctx context.Context, db database.Ext, studyPlanItemID pgtype.Text, userIDs pgtype.TextArray) (map[string]*entities.StudyPlanItem, error) {
	userStudyPlanItem := make(map[string]*entities.StudyPlanItem, len(userIDs.Elements))
	item := &entities.StudyPlanItem{}
	fields, _ := item.FieldMap()
	rows, err := db.Query(ctx, fmt.Sprintf(retrieveChildStudyPlanItemStmtPlt, strings.Join(fields, ", spi.")), studyPlanItemID, userIDs)
	if err != nil {
		return userStudyPlanItem, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID pgtype.Text
		studyPlanItem := &entities.StudyPlanItem{}

		allvalues := make([]interface{}, 0, len(fields)+1)
		allvalues = append(allvalues, &userID)
		allvalues = append(allvalues, database.GetScanFields(studyPlanItem, fields)...)
		err := rows.Scan(allvalues...)
		if err != nil {
			return userStudyPlanItem, err
		}
		if studyPlanItem.ID.Status != pgtype.Null {
			userStudyPlanItem[userID.String] = studyPlanItem
		}
	}
	return userStudyPlanItem, nil
}

func (r *StudyPlanItemRepo) RetrieveBookIDByStudyPlanID(ctx context.Context, db database.Ext, studyPlanID pgtype.Text) (string, error) {
	query := `SELECT book_id
		FROM study_plans
		WHERE study_plan_id = $1 AND deleted_at IS NULL`

	var bookID pgtype.Text
	if err := db.QueryRow(ctx, query, &studyPlanID).Scan(&bookID); err != nil {
		return "", err
	}

	return bookID.String, nil
}

type CountStudentStudyPlanItemsInClassFilter struct {
	ClassID         pgtype.Text
	StudyPlanItemID pgtype.Text
	IsCompleted     pgtype.Bool
}

const countStudentStudyPlanItemsInClassStmtPlt = `
	SELECT COUNT(*) 
	FROM study_plan_items spi 
		JOIN student_study_plans ssp USING(study_plan_id)
		JOIN class_students cs USING(student_id)
	WHERE cs.class_id = $1::text
	AND spi.copy_study_plan_item_id = $2::text 
	AND ($3 IS FALSE OR spi.completed_at IS NOT NULL)
	AND spi.deleted_at IS NULL
	AND ssp.deleted_at IS NULL
	AND cs.deleted_at IS NULL`

func (r *StudyPlanItemRepo) CountStudentStudyPlanItemsInClass(ctx context.Context, db database.Ext, filter *CountStudentStudyPlanItemsInClassFilter) (int, error) {
	var total pgtype.Int8
	if err := db.QueryRow(ctx, countStudentStudyPlanItemsInClassStmtPlt, &filter.ClassID, &filter.StudyPlanItemID, &filter.IsCompleted).Scan(&total); err != nil {
		return 0, err
	}
	return int(total.Int), nil
}

func (r *StudyPlanItemRepo) RetrieveStudyPlanContentStructuresByBooks(ctx context.Context, db database.QueryExecer, books pgtype.TextArray) (map[string][]entities.ContentStructure, error) {
	if len(books.Elements) == 0 {
		return nil, nil
	}

	ctx, span := interceptors.StartSpan(ctx, "StudyPlanItemRepo.RetrieveStudyPlanContentStructuresByBooks")
	defer span.End()

	cond := fmt.Sprintf("spi.content_structure_flatten LIKE 'book::%s%%'", books.Elements[0].String)
	for i := 1; i < len(books.Elements); i++ {
		cond += fmt.Sprintf(" OR spi.content_structure_flatten LIKE 'book::%s%%'", books.Elements[i].String)
	}

	b := &entities.StudyPlanItem{}
	query := fmt.Sprintf(`
		SELECT spi.study_plan_id, spi.content_structure->>'book_id', spi.content_structure->>'course_id'
		FROM %s spi
		INNER JOIN study_plans sp
		ON spi.study_plan_id = sp.study_plan_id
		WHERE sp.deleted_at IS NULL
		AND spi.copy_study_plan_item_id IS NULL
		AND (%s)
		GROUP BY spi.study_plan_id, spi.content_structure->>'book_id', spi.content_structure->>'course_id'
	`, b.TableName(), cond)

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %v", err)
	}
	defer rows.Close()

	m := make(map[string][]entities.ContentStructure)
	for rows.Next() {
		var studyPlanID, bookID, courseID string
		if err := rows.Scan(&studyPlanID, &bookID, &courseID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %v", err)
		}
		m[studyPlanID] = append(m[studyPlanID], entities.ContentStructure{
			BookID:   bookID,
			CourseID: courseID,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %v", err)
	}

	return m, nil
}

const copyItemForCopiedStudyPlansStmt = `INSERT
	INTO
	study_plan_items
SELECT
	generate_ulid() AS student_plan_item_id,
	sp.study_plan_id,
	available_from,
	start_date,
	end_date,
	spi.deleted_at,
	available_to,
	NOW(),
	NOW(),
	study_plan_item_id AS copy_study_plan_item_id,
	content_structure,
	completed_at,
	display_order,
	content_structure_flatten
FROM
	study_plan_items spi
INNER JOIN
	study_plans sp ON sp.master_study_plan_id = spi.study_plan_id
WHERE
	spi.study_plan_id = $1
AND
	study_plan_item_id = $2
ON CONFLICT (study_plan_id, content_structure_flatten)
DO UPDATE SET
	display_order = EXCLUDED.display_order
`

func (r *StudyPlanItemRepo) CopyItemsForCopiedStudyPlans(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) error {
	queueFn := func(b *pgx.Batch, item *entities.StudyPlanItem) {
		b.Queue(copyItemForCopiedStudyPlansStmt, &item.StudyPlanID, &item.ID)
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

const syncStudyPlanItemStmtTpl = `INSERT INTO
	study_plan_items (%s)
VALUES (%s)
ON CONFLICT (study_plan_id, content_structure_flatten)
DO UPDATE SET
	updated_at = $7,
	display_order = $12
RETURNING study_plan_item_id`

// BulkSync inserts items into study_plan_items table, updates study plan item's display order
// on duplicate (study_plan_id, content_structure_flatten) fields, and returns the items
// that are inserted.
func (r *StudyPlanItemRepo) BulkSync(
	ctx context.Context,
	db database.QueryExecer,
	items []*entities.StudyPlanItem,
) (
	insertItems []*entities.StudyPlanItem,
	err error,
) {
	queueFn := func(b *pgx.Batch, item *entities.StudyPlanItem) {
		fieldNames := database.GetFieldNames(item)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))
		query := fmt.Sprintf(syncStudyPlanItemStmtTpl, strings.Join(fieldNames, ","), placeHolders)
		scanFields := database.GetScanFields(item, fieldNames)
		b.Queue(query, scanFields...)
	}

	b := &pgx.Batch{}
	for _, item := range items {
		queueFn(b, item)
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		var returnedID pgtype.Text
		if serr := result.QueryRow().Scan(&returnedID); serr != nil {
			err = fmt.Errorf("result.QueryRow.Scan: %w", serr)
			return
		}

		// In case of insert, the item id in DB should be the same with passed item id.
		// In case of update, the item id in DB should be different with passed item id.
		if returnedID.String == items[i].ID.String {
			insertItems = append(insertItems, items[i])
		} else {
			// for update case, set the item id to match with the id in the DB.
			items[i].ID = returnedID
		}
	}
	return
}

func (r *StudyPlanItemRepo) DeleteStudyPlanItemsByStudyPlans(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanItemRepo.DeleteStudyPlanItemsByStudyPlans")
	defer span.End()

	query := `
		UPDATE study_plan_items SET deleted_at = now() WHERE study_plan_id = ANY($1)
	`
	if _, err := db.Exec(ctx, query, &studyPlanIDs); err != nil {
		return err
	}
	return nil
}

func (r *StudyPlanItemRepo) DeleteStudyPlanItemsByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanItemRepo.DeleteStudyPlanItemsByStudyPlans")
	defer span.End()

	query := `
		UPDATE study_plan_items SET deleted_at = now() WHERE study_plan_item_id = ANY(SELECT study_plan_item_id FROM lo_study_plan_items WHERE lo_id = ANY ($1::_TEXT))
	`
	if _, err := db.Exec(ctx, query, &loIDs); err != nil {
		return err
	}
	return nil
}

func (r *StudyPlanItemRepo) UpdateCompletedAtByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, completedAt pgtype.Timestamptz) error {
	studyPlanItem := entities.StudyPlanItem{}

	stmt := fmt.Sprintf(`UPDATE %s
	SET completed_at = $2, updated_at = NOW()
	WHERE study_plan_item_id = $1 AND deleted_at IS NULL`, studyPlanItem.TableName())
	cmd, err := db.Exec(ctx, stmt, &id, &completedAt)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found any study plan item to update: %w", pgx.ErrNoRows)
	}

	return nil
}

func (r *StudyPlanItemRepo) SoftDeleteByStudyPlanItemIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanItemRepo.SoftDeleteByStudyPlanItemIDs")
	defer span.End()

	query := `
		UPDATE study_plan_items SET deleted_at = NOW() WHERE study_plan_item_id = ANY($1::TEXT[]) AND deleted_at IS NULL
	`

	if _, err := db.Exec(ctx, query, &ids); err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (r *StudyPlanItemRepo) UpdateSchoolDate(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, studentID pgtype.Text, schoolDate pgtype.Timestamptz) error {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanItemRepo.UpdateSchoolDate")
	defer span.End()

	query := `
		UPDATE study_plan_items SET school_date = $1, updated_at = NOW() WHERE study_plan_item_id IN (
			WITH temp AS (
				SELECT  csp.study_plan_id
				FROM course_students AS cs,
						course_study_plans AS csp
				WHERE cs.student_id = $2::TEXT
					AND cs.course_id = csp.course_id
						AND cs.deleted_at IS NULL
						AND csp.deleted_at IS NULL
			)
			SELECT DISTINCT(spi.study_plan_item_id)
			FROM study_plan_items AS spi, temp
			WHERE
				spi.study_plan_item_id = ANY($3::TEXT[])
				AND spi.study_plan_id = temp.study_plan_id
					AND spi.deleted_at IS NULL
			UNION
			SELECT  DISTINCT(spi.study_plan_item_id)
			FROM study_plan_items AS spi, student_study_plans AS ssp
			WHERE ssp.student_id = $4::TEXT
				AND spi.study_plan_item_id = ANY($5::TEXT[])
				AND ssp.study_plan_id = spi.study_plan_id
						AND ssp.deleted_at IS NULL
						AND spi.deleted_at IS NULL
		)
	`

	if _, err := db.Exec(ctx, query, &schoolDate, &studentID, &ids, &studentID, &ids); err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (r *StudyPlanItemRepo) UpdateSchoolDateByStudyPlanItemIdentity(ctx context.Context, db database.QueryExecer, lmID, studyPlanID pgtype.Text, studentIDs pgtype.TextArray, schoolDate pgtype.Timestamptz) error {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanItemRepo.UpdateSchoolDateByStudyPlanItemIdentity")
	defer span.End()

	query := `
		UPDATE study_plan_items SET school_date = $1, updated_at = NOW() where 
		study_plan_item_id = any(select spi.study_plan_item_id from study_plan_items spi join student_study_plans ssp 
			on spi.study_plan_id = ssp.study_plan_id
			where ssp.master_study_plan_id = $2::TEXT 
			and (spi.content_structure ->> 'lo_id' = $3::TEXT or spi.content_structure ->> 'assignment_id' = $3::TEXT )
			and (ssp.student_id is null or ssp.student_id = any($4::TEXT[])))
		`

	if _, err := db.Exec(ctx, query, &schoolDate, &studyPlanID, &lmID, &studentIDs); err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (r *StudyPlanItemRepo) BulkUpdateSchoolDate(ctx context.Context, db database.QueryExecer, studyPlanItemIds pgtype.TextArray, schoolDate pgtype.Timestamptz) error {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanItemRepo.UpdateSchoolDateByStudyPlanItemIdentity")
	defer span.End()

	query := `
		UPDATE study_plan_items SET school_date = $1, updated_at = NOW() where 
		study_plan_item_id = ANY($2::TEXT[]);
		`

	if _, err := db.Exec(ctx, query, &schoolDate, &studyPlanItemIds); err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (r *StudyPlanItemRepo) UpdateStatus(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, studentID pgtype.Text, status pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanItemRepo.UpdateStatus")
	defer span.End()

	query := `
		UPDATE study_plan_items SET status = $1, updated_at = NOW() WHERE study_plan_item_id IN (
			WITH temp AS (
				SELECT  csp.study_plan_id
				FROM course_students AS cs,
						course_study_plans AS csp
				WHERE cs.student_id = $2::TEXT
					AND cs.course_id = csp.course_id
						AND cs.deleted_at IS NULL
						AND csp.deleted_at IS NULL
			)
			SELECT DISTINCT(spi.study_plan_item_id)
			FROM study_plan_items AS spi, temp
			WHERE
				spi.study_plan_item_id = ANY($3::TEXT[])
				AND spi.study_plan_id = temp.study_plan_id
					AND spi.deleted_at IS NULL
			UNION
			SELECT  DISTINCT(spi.study_plan_item_id)
			FROM study_plan_items AS spi, student_study_plans AS ssp
			WHERE ssp.student_id = $4::TEXT
				AND spi.study_plan_item_id = ANY($5::TEXT[])
				AND ssp.study_plan_id = spi.study_plan_id
						AND ssp.deleted_at IS NULL
						AND spi.deleted_at IS NULL
		)
	`

	if _, err := db.Exec(ctx, query, &status, &studentID, &ids, &studentID, &ids); err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

type StudyPlanItemArgs struct {
	StudyPlanIDs        pgtype.TextArray
	TopicIDs            pgtype.TextArray
	AssignmentIDs       pgtype.TextArray
	LoIDs               pgtype.TextArray
	AvailableDateFilter bool
}

type FilterStudyPlanItemArgs struct {
	StudyPlanID  pgtype.Text
	TopicID      pgtype.Text
	AssignmentID pgtype.Text
	LoID         pgtype.Text
}

func (r *StudyPlanItemRepo) FindWithFilter(ctx context.Context, db database.QueryExecer, filter *StudyPlanItemArgs) ([]*entities.StudyPlanItem, error) {
	studyPlanItem := &entities.StudyPlanItem{}
	fieldNames := database.GetFieldNames(studyPlanItem)

	whereStm := `
		(study_plan_id = ANY($1) OR $1::TEXT[] IS NULL) 
		AND (content_structure ->> 'topic_id' = ANY($2) OR $2::TEXT[] IS NULL)
		AND (
			(content_structure ->> 'lo_id' = ANY($3) OR content_structure ->> 'assignment_id' = ANY($4))
			OR 	($3::TEXT[] IS NULL AND $4::TEXT[] IS NULL)
		)
		AND deleted_at IS NULL
	`

	if filter.AvailableDateFilter {
		whereStm += " AND available_from <= now() AND now() <= available_to"
	}

	query := fmt.Sprintf(`
	SELECT %s 
	FROM %s
	WHERE %s 	
	`, strings.Join(fieldNames, ", "), studyPlanItem.TableName(), whereStm)

	var studyPlanItems entities.StudyPlanItems
	err := database.Select(
		ctx,
		db,
		query,
		&filter.StudyPlanIDs,
		&filter.TopicIDs,
		&filter.LoIDs,
		&filter.AssignmentIDs,
	).ScanAll(&studyPlanItems)
	if err != nil {
		return nil, err
	}
	return studyPlanItems, nil
}

func (r *StudyPlanItemRepo) FindWithFilterV2(ctx context.Context, db database.QueryExecer, filter *FilterStudyPlanItemArgs) ([]*entities.StudyPlanItem, error) {
	studyPlanItem := &entities.StudyPlanItem{}
	fieldNames := database.GetFieldNames(studyPlanItem)
	query := fmt.Sprintf(`SELECT %s FROM study_plan_items 
	WHERE study_plan_id = $1 AND content_structure ->> 'topic_id' = $2
	AND (content_structure ->> 'lo_id' = $3::TEXT OR $3::TEXT IS NULL)
	AND (content_structure ->> 'assignment_id' = $4::TEXT OR $4::TEXT IS NULL) `, strings.Join(fieldNames, ", "))

	var studyPlanItems entities.StudyPlanItems
	err := database.Select(
		ctx,
		db,
		query,
		&filter.StudyPlanID,
		&filter.TopicID,
		&filter.LoID,
		&filter.AssignmentID,
	).ScanAll(&studyPlanItems)
	if err != nil {
		return studyPlanItems, err
	}
	return studyPlanItems, nil
}

func (r *StudyPlanItemRepo) FetchByStudyProgressRequest(ctx context.Context, db database.QueryExecer, courseID pgtype.Text, bookID pgtype.Text, studentID pgtype.Text) ([]*entities.StudyPlanItem, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanItemRepo.FetchByStudyProgressRequest")
	defer span.End()

	query := `
		SELECT  spi.study_plan_item_id, spi.content_structure, spi.completed_at 
		FROM student_study_plans ssp
		JOIN course_students cs ON cs.student_id = ssp.student_id
		JOIN study_plan_items spi ON spi.study_plan_id = ssp.study_plan_id
		JOIN study_plans sp ON sp.study_plan_id = spi.study_plan_id
		WHERE
				cs .deleted_at IS NULL
				AND ssp.deleted_at IS NULL
				AND spi.deleted_at IS NULL
				AND cs.course_id = $1
				AND ssp.student_id = $2
				AND spi.content_structure ->> 'book_id' = $3
				AND spi.content_structure ->> 'course_id' = $1
				AND spi.status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE'
				AND sp.status = 'STUDY_PLAN_STATUS_ACTIVE'
				AND spi.available_from <= now() AND now() <= spi.available_to
		GROUP BY  spi.study_plan_item_id, spi.content_structure;
	`

	studyPlanItems := make([]*entities.StudyPlanItem, 0)
	rows, err := db.Query(ctx, query, &courseID, &studentID, &bookID)
	if err != nil {
		return nil, fmt.Errorf("StudyPlanItemRepo.FetchByStudyProgressRequest.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		studyPlanItem := new(entities.StudyPlanItem)
		if err := rows.Scan(&studyPlanItem.ID, &studyPlanItem.ContentStructure, &studyPlanItem.CompletedAt); err != nil {
			return nil, fmt.Errorf("StudyPlanItemRepo.FetchByStudyProgressRequest.Scan: %w", err)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("StudyPlanItemRepo.FetchByStudyProgressRequest.Err: %w", err)
		}
		studyPlanItems = append(studyPlanItems, studyPlanItem)
	}

	return studyPlanItems, nil
}

func (r *StudyPlanItemRepo) RetrieveByBookContent(ctx context.Context, db database.QueryExecer, bookIDs, loIDs, assignmentIDs pgtype.TextArray) ([]*entities.StudyPlanItem, error) {
	ctx, span := trace.StartSpan(ctx, "StudyPlanItemRepo.RetrieveCombineStudent")
	defer span.End()

	spi := &entities.StudyPlanItem{}
	spis := entities.StudyPlanItems{}
	fields := database.GetFieldNames(spi)
	selectStmt := fmt.Sprintf(`SELECT spi.%s FROM study_plan_items spi JOIN study_plans sp USING(study_plan_id)
	LEFT JOIN lo_study_plan_items USING (study_plan_item_id)
	LEFT JOIN assignment_study_plan_items USING (study_plan_item_id)
	LEFT JOIN student_study_plans ssp USING (study_plan_id)
	WHERE book_id= ANY($1::_TEXT)
	AND spi.deleted_at IS NULL
	AND sp.deleted_at IS NULL
	AND( lo_id= ANY($2::_TEXT)
		OR assignment_id = ANY($3::_TEXT))`, strings.Join(fields, ", spi."))

	err := database.Select(ctx, db, selectStmt, &bookIDs, &loIDs, &assignmentIDs).ScanAll(&spis)
	if err != nil {
		return nil, err
	}
	return spis, nil
}

func (r *StudyPlanItemRepo) BulkUpdateStartEndDate(ctx context.Context, db database.QueryExecer, studyPlanItemIds pgtype.TextArray, updateFields sspb.UpdateStudyPlanItemsStartEndDateFields, startDate, endDate pgtype.Timestamptz) (int64, error) {
	prepareQuery := `UPDATE study_plan_items SET updated_at = now(), %s WHERE study_plan_item_id = ANY($1::TEXT[]) AND %s`
	var err error
	var cmdTag pgconn.CommandTag
	switch updateFields {
	case sspb.UpdateStudyPlanItemsStartEndDateFields_ALL:
		query := fmt.Sprintf(prepareQuery, `start_date = $2, end_date = $3`, `available_from <= $2::timestamptz
					AND (available_to > $3::timestamptz OR available_to IS NULL)`)
		cmdTag, err = db.Exec(ctx, query, studyPlanItemIds, startDate, endDate)
	case sspb.UpdateStudyPlanItemsStartEndDateFields_START_DATE:
		query := fmt.Sprintf(prepareQuery, `start_date = $2`, `available_from <= $2::timestamptz`)
		cmdTag, err = db.Exec(ctx, query, studyPlanItemIds, startDate)
	default:
		query := fmt.Sprintf(prepareQuery, `end_date = $2`, `(available_to > $2::timestamptz OR available_to IS NULL)`)
		cmdTag, err = db.Exec(ctx, query, studyPlanItemIds, endDate)
	}
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}
	return cmdTag.RowsAffected(), nil
}

func (r *StudyPlanItemRepo) ListSPItemByIdentity(ctx context.Context, db database.QueryExecer, studyPlanItemIdentities []StudyPlanItemIdentity) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanItemRepo.ListSPItemByIdentity")
	defer span.End()

	query := `SELECT spi.study_plan_item_id
FROM student_study_plans ssp
JOIN LATERAL (
    SELECT *
    FROM study_plan_items
    WHERE study_plan_id = ssp.study_plan_id
      AND coalesce(NULLIF(content_structure ->> 'lo_id',''), content_structure ->> 'assignment_id') = $2
) spi ON TRUE
WHERE (ssp.master_study_plan_id = $1 OR (ssp.master_study_plan_id IS NULL AND ssp.study_plan_id = $1))
        AND ($3::TEXT IS NULL OR ssp.student_id = $3)`

	b := &pgx.Batch{}
	for _, identity := range studyPlanItemIdentities {
		b.Queue(query, identity.StudyPlanID, identity.LearningMaterialID, identity.StudentID)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	studyPlanItemIDs := make([]string, 0)
	for i := 0; i < b.Len(); i++ {
		var studyPlanItemID string
		err := result.QueryRow().Scan(&studyPlanItemID)
		if err != nil {
			return nil, fmt.Errorf("result.QueryRow().Scan err: %s", err)
		}
		studyPlanItemIDs = append(studyPlanItemIDs, studyPlanItemID)
	}

	return studyPlanItemIDs, nil
}

func (r *StudyPlanItemRepo) UpdateStudyPlanItemsStatus(ctx context.Context, db database.QueryExecer, studyPlanItemIds pgtype.TextArray, spiStatus pgtype.Text) (int64, error) {
	prepareQuery := `UPDATE study_plan_items SET status = $1::TEXT, updated_at = now() WHERE study_plan_item_id = ANY($2::TEXT[]) AND deleted_at IS NULL`
	cmdTag, err := db.Exec(ctx, prepareQuery, spiStatus, studyPlanItemIds)
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}
	return cmdTag.RowsAffected(), nil
}

func (r *StudyPlanItemRepo) UpdateCompletedAtToNullByStudyPlanItemIdentity(ctx context.Context, db database.QueryExecer, args StudyPlanItemIdentity) (int64, error) {
	studyPlanItem := entities.StudyPlanItem{}
	stmt := fmt.Sprintf(`UPDATE %s
	SET completed_at = NULL, updated_at = NOW()
	WHERE study_plan_item_id = (
		SELECT DISTINCT study_plan_item_id	
		FROM shuffled_quiz_sets sqs 
		WHERE sqs.learning_material_id = $1
			AND sqs.student_id = $2
			AND sqs.study_plan_id = $3
			AND deleted_at IS NULL
	) AND deleted_at IS NULL`,
		studyPlanItem.TableName())
	cmd, err := db.Exec(ctx, stmt, args.LearningMaterialID, args.StudentID, args.StudyPlanID)
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}
	return cmd.RowsAffected(), nil
}

type FindLearningMaterialByStudyPlanID struct {
	LearningMaterialID pgtype.Text
	StudyPlanItemID    pgtype.Text
}

func (r *StudyPlanItemRepo) FindLearningMaterialByStudyPlanID(ctx context.Context, db database.QueryExecer, id pgtype.Text) ([]*FindLearningMaterialByStudyPlanID, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanItemRepo.FindLearningMaterialByStudyPlanID")
	defer span.End()

	e := entities.StudyPlanItem{}
	stmt := fmt.Sprintf(`
	SELECT
	COALESCE(NULLIF(content_structure ->> 'lo_id', ''), content_structure->>'assignment_id', '') AS learning_material_id, study_plan_item_id 
		FROM %s spi
	WHERE study_plan_id  = $1
	AND deleted_at IS NULL`, e.TableName())

	rows, err := db.Query(ctx, stmt, id)
	if err != nil {
		return nil, fmt.Errorf("db.Exec: %w", err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("StudyPlanItemRepo.FindLearningMaterialByStudyPlanID.Err: %w", err)
	}

	result := []*FindLearningMaterialByStudyPlanID{}
	for rows.Next() {
		i := new(FindLearningMaterialByStudyPlanID)
		if err := rows.Scan(&i.LearningMaterialID, &i.StudyPlanItemID); err != nil {
			return nil, fmt.Errorf("StudyPlanItemRepo.FindLearningMaterialByStudyPlanID.Scan: %w", err)
		}

		result = append(result, i)
	}

	return result, nil
}
