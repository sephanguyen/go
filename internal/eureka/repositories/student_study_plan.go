package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

// StudentStudyPlanRepo works with study plan and study plan item
type StudentStudyPlanRepo struct {
}

const queueUpsertStudentStudyPlanStmtTpl = `INSERT INTO %s (%s)
VALUES (%s)
ON CONFLICT ON CONSTRAINT student_study_plans_pk DO UPDATE
SET
	updated_at = $4,
	deleted_at = NULL`

func (r *StudentStudyPlanRepo) QueueUpsertStudentStudyPlan(b *pgx.Batch, studyPlan *entities.StudentStudyPlan) {
	fieldNames := database.GetFieldNames(studyPlan)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := fmt.Sprintf(queueUpsertStudentStudyPlanStmtTpl,
		studyPlan.TableName(), strings.Join(fieldNames, ","), placeHolders)
	scanFields := database.GetScanFields(studyPlan, fieldNames)
	b.Queue(query, scanFields...)
}

func (r *StudentStudyPlanRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, studentStudyPlans []*entities.StudentStudyPlan) error {
	b := &pgx.Batch{}
	for _, studyPlan := range studentStudyPlans {
		r.QueueUpsertStudentStudyPlan(b, studyPlan)
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

type ListStudentAvailableContentsArgs struct {
	StudentID    pgtype.Text
	StudyPlanIDs pgtype.TextArray
	Offset       pgtype.Timestamptz
	BookID       pgtype.Text
	ChapterID    pgtype.Text
	TopicID      pgtype.Text
	CourseID     pgtype.Text
}

const listStudentAvailableContentsStmtTpl = `SELECT
	i.%s
FROM
	%s AS i
INNER JOIN student_study_plans s ON
	i.study_plan_id = s.study_plan_id
INNER JOIN study_plans sp ON
	sp.study_plan_id = s.study_plan_id
WHERE
	s.student_id = $1
	AND ($2::TEXT[] IS NULL
	OR s.study_plan_id = ANY($2))
	AND ($3::timestamptz IS NULL
	OR $3 BETWEEN i.available_from AND i.available_to)
	AND ($4::TEXT IS NULL
	OR content_structure ->> 'book_id' = $4::TEXT)
	AND ($5::TEXT IS NULL
	OR content_structure ->> 'chapter_id' = $5::TEXT)
	AND ($6::TEXT IS NULL
	OR content_structure ->> 'topic_id' = $6::TEXT)
	AND ($7::TEXT IS NULL
		OR sp.course_id = $7::TEXT)
	AND sp.status = 'STUDY_PLAN_STATUS_ACTIVE'
	AND i.status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE'
	AND s.deleted_at IS NULL
	AND i.deleted_at IS NULL
	AND sp.deleted_at IS NULL
	ORDER BY content_structure ->> 'book_id'`

func (r *StudentStudyPlanRepo) ListStudentAvailableContents(ctx context.Context, db database.QueryExecer, q *ListStudentAvailableContentsArgs) ([]*entities.StudyPlanItem, error) {
	var e entities.StudyPlanItem
	selectFields := database.GetFieldNames(&e)

	query := fmt.Sprintf(listStudentAvailableContentsStmtTpl,
		strings.Join(selectFields, ", i."), e.TableName())

	var items entities.StudyPlanItems
	if err := database.Select(ctx, db, query, &q.StudentID, &q.StudyPlanIDs, &q.Offset,
		&q.BookID, &q.ChapterID, &q.TopicID, &q.CourseID).ScanAll(&items); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return items, nil
}

func (r *StudentStudyPlanRepo) GetBookIDsBelongsToStudentStudyPlan(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, bookIDs pgtype.TextArray) ([]string, error) {
	var studentIDReq pgtype.Text
	if err := studentIDReq.Set(studentID); err != nil {
		return nil, fmt.Errorf("pgtype.Set: %w", err)
	}

	var e entities.StudyPlanItem
	query := fmt.Sprintf(`
		SELECT DISTINCT
			i.content_structure ->> 'book_id' as book_id
		FROM
			%s AS i
		INNER JOIN student_study_plans s ON
			i.study_plan_id = s.study_plan_id
		INNER JOIN study_plans sp ON
			sp.study_plan_id = s.study_plan_id
		WHERE
			s.deleted_at IS NULL
			AND s.student_id = $1
			AND ($2::TEXT[] IS NULL OR i.content_structure ->> 'book_id' = ANY($2::TEXT[]))
			AND i.deleted_at IS NULL
			AND sp.deleted_at IS NULL;
	`,
		e.TableName(),
	)

	var items entities.Books
	if err := database.Select(ctx, db, query, &studentIDReq, &bookIDs).ScanAll(&items); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	bookIDsRes := make([]string, 0)
	for _, item := range items {
		bookIDsRes = append(bookIDsRes, item.ID.String)
	}

	return bookIDsRes, nil
}

type ListStudyPlansArgs struct {
	StudentID pgtype.Text
	CourseID  pgtype.Text
	SchoolID  pgtype.Int4
	Limit     uint32
	Offset    pgtype.Text
}

const listStudyPlansStmtTpl = `SELECT
	i.%s
FROM
	%s AS i
INNER JOIN student_study_plans s ON
	i.study_plan_id = s.study_plan_id
WHERE
	s.student_id = $1
	AND ($2::TEXT IS NULL
	OR i.course_id = $2)
	AND ($3::int IS NULL
	OR i.school_id = $3)
	AND ($4::TEXT IS NULL
	OR i.study_plan_id < $4)
	AND i.deleted_at IS NULL
	AND s.deleted_at IS NULL
ORDER BY
	i.study_plan_id DESC
LIMIT $5`

func (r *StudentStudyPlanRepo) ListStudyPlans(ctx context.Context, db database.QueryExecer, q *ListStudyPlansArgs) ([]*entities.StudyPlan, error) {
	var e entities.StudyPlan
	selectFields := database.GetFieldNames(&e)

	query := fmt.Sprintf(listStudyPlansStmtTpl,
		strings.Join(selectFields, ", i."), e.TableName())

	var items entities.StudyPlans
	if err := database.Select(ctx, db, query, &q.StudentID, &q.CourseID, &q.SchoolID, &q.Offset, &q.Limit).ScanAll(&items); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return items, nil
}

type ListStudyPlanItemsArgs struct {
	StudentID        pgtype.Text
	Limit            uint32
	Now              pgtype.Timestamptz
	CourseIDs        pgtype.TextArray
	StudyPlanID      pgtype.Text
	IncludeCompleted bool

	// fields used for pagination query
	Offset          pgtype.Timestamptz
	StudyPlanItemID pgtype.Text
	DisplayOrder    pgtype.Int4
}

const listStudyPlanItemsStmtTpl = `SELECT
	i.%s
FROM
	%s AS i
INNER JOIN student_study_plans s ON
	i.study_plan_id = s.study_plan_id
JOIN study_plans sp ON
	sp.study_plan_id = s.study_plan_id
WHERE
	s.student_id = $1
	AND (($2::int IS NULL AND $4::text IS NULL) OR (i.display_order, i.study_plan_item_id) > ($2::int, $4::text))
	AND i.available_from IS NOT NULL
	AND ((i.available_to IS NULL AND i.available_from <= $3) OR ($3 BETWEEN i.available_from AND i.available_to))
	AND ($5::text IS NULL OR sp.study_plan_id = $5::text)
	AND ($6::_text IS NULL OR sp.course_id = ANY($6::_text))
	AND s.deleted_at IS NULL
	AND i.deleted_at IS NULL
	AND sp.deleted_at IS NULL
ORDER BY
	i.display_order ASC,
	i.study_plan_item_id ASC
LIMIT $7`

func (r *StudentStudyPlanRepo) ListStudyPlanItems(ctx context.Context, db database.QueryExecer, q *ListStudyPlanItemsArgs) ([]*entities.StudyPlanItem, error) {
	var e entities.StudyPlanItem
	selectFields := database.GetFieldNames(&e)

	query := fmt.Sprintf(listStudyPlanItemsStmtTpl,
		strings.Join(selectFields, ", i."), e.TableName())
	var items entities.StudyPlanItems
	if err := database.Select(ctx, db, query, &q.StudentID, &q.DisplayOrder, &q.Now, &q.StudyPlanItemID, &q.StudyPlanID, &q.CourseIDs, &q.Limit).ScanAll(&items); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return items, nil
}

const listActiveStudyPlanItemsStmtTpl = `
SELECT
	i.%s
FROM
	%s AS i
INNER JOIN student_study_plans s ON
	i.study_plan_id = s.study_plan_id
JOIN study_plans sp ON
	sp.study_plan_id = s.study_plan_id
JOIN books b ON b.book_id = i.content_structure ->> 'book_id'
JOIN chapters c ON c.chapter_id = i.content_structure ->> 'chapter_id'
JOIN topics t ON t.topic_id = i.content_structure ->> 'topic_id'
WHERE
	s.student_id = $1
	AND i.start_date < NOW()
	AND (($2::timestamp IS NULL OR $7::integer IS NULL OR $8::text IS NULL) OR ((i.start_date, i.display_order, i.study_plan_item_id) > ($2, $7, $8)))
	AND (($3 BETWEEN i.available_from AND i.available_to) OR (i.available_from <= $3 AND i.available_to IS NULL))
	AND (i.end_date IS NULL OR i.end_date >= $3)
	AND ($4::TEXT[] IS NULL OR sp.course_id = ANY($4))
	AND ($5::TEXT IS NULL OR sp.study_plan_id = $5)
	AND ($6 = TRUE OR i.completed_at IS NULL)
	AND sp.status = 'STUDY_PLAN_STATUS_ACTIVE'
	AND i.status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE'
	AND s.deleted_at IS NULL
	AND i.deleted_at IS NULL
	AND sp.deleted_at IS NULL
  AND b.deleted_at IS NULL
  AND c.deleted_at IS NULL
  AND t.deleted_at IS NULL
ORDER BY
	i.start_date ASC,
	i.display_order ASC,
	i.study_plan_item_id ASC
LIMIT $9
`

func (r *StudentStudyPlanRepo) ListActiveStudyPlanItems(ctx context.Context, db database.QueryExecer, q *ListStudyPlanItemsArgs) ([]*entities.StudyPlanItem, error) {
	var e entities.StudyPlanItem
	selectFields := database.GetFieldNames(&e)

	query := fmt.Sprintf(listActiveStudyPlanItemsStmtTpl,
		strings.Join(selectFields, ", i."), e.TableName())

	var items entities.StudyPlanItems
	if err := database.Select(
		ctx, db, query,
		&q.StudentID, &q.Offset, &q.Now, &q.CourseIDs,
		&q.StudyPlanID, &q.IncludeCompleted, &q.DisplayOrder, &q.StudyPlanItemID, &q.Limit).
		ScanAll(&items); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return items, nil
}

const listUpcomingStudyPlanItemsStmtTpl = `
SELECT
	i.%s
FROM
  	%s AS i
INNER JOIN student_study_plans s ON i.study_plan_id = s.study_plan_id
JOIN study_plans sp ON sp.study_plan_id = s.study_plan_id
JOIN books b ON b.book_id = i.content_structure ->> 'book_id'
JOIN chapters c ON c.chapter_id = i.content_structure ->> 'chapter_id'
JOIN topics t ON t.topic_id = i.content_structure ->> 'topic_id'
WHERE
	s.student_id = $1
	AND (COALESCE(i.start_date, $4::timestamptz + '100 year'::interval), study_plan_item_id) > (COALESCE($2::timestamptz, $4::timestamptz + '100 year'::interval), $3)
	AND (i.end_date IS NULL OR i.end_date >= $4)
	AND (($4 BETWEEN i.available_from AND i.available_to) OR (i.available_from <= $4 AND available_to IS NULL))
	AND ($5::text[] IS NULL OR sp.course_id = ANY ($5))
	AND ($6::text IS NULL OR sp.study_plan_id = $6)
	AND ($7 = true OR i.completed_at IS NULL)
	AND sp.status = 'STUDY_PLAN_STATUS_ACTIVE'
	AND i.status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE'
	AND s.deleted_at IS NULL
	AND i.deleted_at IS NULL
	AND sp.deleted_at IS NULL
  AND b.deleted_at IS NULL
  AND c.deleted_at IS NULL
  AND t.deleted_at IS NULL
ORDER BY
	i.start_date ASC,
	i.study_plan_item_id ASC
LIMIT $8
`

func (r *StudentStudyPlanRepo) ListUpcomingStudyPlanItems(ctx context.Context, db database.QueryExecer, q *ListStudyPlanItemsArgs) ([]*entities.StudyPlanItem, error) {
	var e entities.StudyPlanItem
	selectFields := database.GetFieldNames(&e)

	query := fmt.Sprintf(listUpcomingStudyPlanItemsStmtTpl, strings.Join(selectFields, ", i."), e.TableName())

	var items entities.StudyPlanItems
	if err := database.Select(
		ctx, db, query,
		&q.StudentID, &q.Offset, &q.StudyPlanItemID, &q.Now,
		&q.CourseIDs, &q.StudyPlanID, &q.IncludeCompleted, &q.Limit).
		ScanAll(&items); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return items, nil
}

const listCompletedStudyPlanItemsStmtTpl = `SELECT
	i.%s
FROM
	%s AS i
INNER JOIN student_study_plans s ON
	i.study_plan_id = s.study_plan_id
JOIN study_plans sp ON
	sp.study_plan_id = s.study_plan_id
JOIN books b ON b.book_id = i.content_structure ->> 'book_id'
JOIN chapters c ON c.chapter_id = i.content_structure ->> 'chapter_id'
JOIN topics t ON t.topic_id = i.content_structure ->> 'topic_id'
WHERE
	s.student_id = $1
	AND 
	(
		($4::int IS NULL OR $6::text IS NULL) OR 
		(
			(i.start_date < $2) OR 
			(i.start_date IS NULL AND $2 IS NULL AND i.display_order > $4) OR
			(i.start_date IS NULL AND $2 IS NULL AND i.study_plan_item_id < $6) OR
			(i.start_date = $2 AND i.display_order > $4) OR
			(i.start_date = $2 AND i.display_order = $4 AND i.study_plan_item_id < $6)
		)
	)
	AND (($3 BETWEEN i.available_from AND i.available_to) OR (i.available_from <= $3 AND i.available_to IS NULL))
	AND ($5::text[] IS NULL OR sp.course_id = ANY ($5))
	AND sp.status = 'STUDY_PLAN_STATUS_ACTIVE'
	AND i.status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE'
	AND i.completed_at IS NOT NULL
	AND s.deleted_at IS NULL
	AND i.deleted_at IS NULL
	AND sp.deleted_at IS NULL
  AND b.deleted_at IS NULL
  AND c.deleted_at IS NULL
  AND t.deleted_at IS NULL
ORDER BY
	i.start_date DESC,
	i.display_order ASC,
	i.study_plan_item_id DESC
LIMIT $7`

func (r *StudentStudyPlanRepo) ListCompletedStudyPlanItems(ctx context.Context, db database.QueryExecer, q *ListStudyPlanItemsArgs) ([]*entities.StudyPlanItem, error) {
	var e entities.StudyPlanItem
	selectFields := database.GetFieldNames(&e)

	query := fmt.Sprintf(listCompletedStudyPlanItemsStmtTpl,
		strings.Join(selectFields, ", i."), e.TableName())

	var items entities.StudyPlanItems
	if err := database.Select(ctx, db, query,
		&q.StudentID, &q.Offset, &q.Now, &q.DisplayOrder,
		&q.CourseIDs, &q.StudyPlanItemID, &q.Limit).
		ScanAll(&items); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return items, nil
}

const listOverdueStudyPlanItemsStmtTpl = `SELECT
	i.%s
FROM
	%s AS i
INNER JOIN student_study_plans s ON
	i.study_plan_id = s.study_plan_id
JOIN study_plans sp ON
	sp.study_plan_id = s.study_plan_id
JOIN books b ON b.book_id = i.content_structure ->> 'book_id'
JOIN chapters c ON c.chapter_id = i.content_structure ->> 'chapter_id'
JOIN topics t ON t.topic_id = i.content_structure ->> 'topic_id'
WHERE
	s.student_id = $1
	AND (i.start_date, i.study_plan_item_id) < ($2, $4)
	AND i.end_date < $3::timestamptz
	AND (($3 BETWEEN i.available_from AND i.available_to) OR (i.available_from <= $3 AND i.available_to IS NULL))
	AND ($5::text[] IS NULL OR sp.course_id = ANY ($5))
	AND sp.status = 'STUDY_PLAN_STATUS_ACTIVE'
	AND i.status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE'
	AND i.completed_at IS NULL
	AND s.deleted_at IS NULL
	AND i.deleted_at IS NULL
	AND sp.deleted_at IS NULL
  AND b.deleted_at IS NULL
  AND c.deleted_at IS NULL
  AND t.deleted_at IS NULL
ORDER BY
	i.start_date DESC,
	i.study_plan_item_id DESC
LIMIT $6`

func (r *StudentStudyPlanRepo) ListOverdueStudyPlanItems(ctx context.Context, db database.QueryExecer, q *ListStudyPlanItemsArgs) ([]*entities.StudyPlanItem, error) {
	var e entities.StudyPlanItem
	selectFields := database.GetFieldNames(&e)

	query := fmt.Sprintf(listOverdueStudyPlanItemsStmtTpl,
		strings.Join(selectFields, ", i."),
		e.TableName())

	var items entities.StudyPlanItems
	if err := database.Select(ctx, db, query, &q.StudentID, &q.Offset, &q.Now, &q.StudyPlanItemID, &q.CourseIDs, &q.Limit).ScanAll(&items); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return items, nil
}

const checkStudentAssignedItemStmt = `SELECT
	count(*)
FROM
	student_study_plans ssp
JOIN study_plan_items ssi ON
	ssp.study_plan_id = ssi.study_plan_id
	AND ssi.study_plan_item_id = ANY($1)
WHERE
	ssp.student_id = $2`

// IsStudentAssignedItem check if student assigned item
func (r *StudentStudyPlanRepo) IsStudentAssignedItem(ctx context.Context, db database.QueryExecer,
	studentID pgtype.Text, itemIDs pgtype.TextArray) (bool, error) {
	var count pgtype.Int8

	err := database.Select(ctx, db, checkStudentAssignedItemStmt,
		&itemIDs, &studentID).ScanFields(&count)
	if err != nil {
		return false, err
	}

	return count.Int == int64(len(itemIDs.Elements)), nil
}

func (r *StudentStudyPlanRepo) CountAssignedStudyPlanItems(ctx context.Context, db database.QueryExecer, studentID, studyPlanID pgtype.Text, now pgtype.Timestamptz) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM
			student_study_plans
		INNER JOIN
			study_plan_items ON study_plan_items.study_plan_id = student_study_plans.study_plan_id
		WHERE
			student_study_plans.student_id = $1
			AND student_study_plans.study_plan_id = $2
			AND ((study_plan_items.start_date <= $3) OR ($3 BETWEEN study_plan_items.available_from AND study_plan_items.available_to))
			AND study_plan_items.deleted_at IS NULL
			AND student_study_plans.deleted_at IS NULL
	`
	var count pgtype.Int8
	if err := database.Select(ctx, db, query, &studentID, &studyPlanID).ScanFields(&count); err != nil {
		return 0, err
	}

	return int(count.Int), nil
}

func (r *StudentStudyPlanRepo) CountStudentStudyPlanItems(ctx context.Context, db database.QueryExecer,
	studentID, studyPlanID pgtype.Text, now pgtype.Timestamptz, onlyCompleted pgtype.Bool) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM
			student_study_plans
		INNER JOIN
			study_plan_items ON study_plan_items.study_plan_id = student_study_plans.study_plan_id
		WHERE
			student_study_plans.student_id = $1
			AND student_study_plans.study_plan_id = $2
			AND ((study_plan_items.available_from <= $3 AND study_plan_items.available_to IS NULL) 
				OR ($3 BETWEEN study_plan_items.available_from AND study_plan_items.available_to))
			AND ($4 = FALSE OR study_plan_items.completed_at IS NOT NULL)
			AND study_plan_items.deleted_at IS NULL
			AND student_study_plans.deleted_at IS NULL
	`
	var count pgtype.Int8
	if err := database.Select(ctx, db, query, &studentID, &studyPlanID, &now, &onlyCompleted).ScanFields(&count); err != nil {
		return 0, err
	}

	return int(count.Int), nil
}

func (r *StudentStudyPlanRepo) FindStudentStudyPlanWithCourseIDs(ctx context.Context, db database.QueryExecer, studentIDs, courseIDs []string) ([]string, error) {
	query := `SELECT ssp.study_plan_id FROM student_study_plans ssp JOIN study_plans sp ON ssp.study_plan_id = sp.study_plan_id WHERE (ssp.student_id, sp.course_id) IN (%s)`
	inCondition, args := database.CompositeKeysPlaceHolders(len(studentIDs), func(i int) []interface{} {
		return []interface{}{studentIDs[i], courseIDs[i]}
	})

	rows, err := db.Query(ctx, fmt.Sprintf(query, inCondition), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var studyPlanIDs []string
	for rows.Next() {
		var studyPlanID string
		if err := rows.Scan(&studyPlanID); err != nil {
			return nil, fmt.Errorf("rows.Err: %w", err)
		}
		studyPlanIDs = append(studyPlanIDs, studyPlanID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return studyPlanIDs, nil
}

const softDeleteStmt = `UPDATE 
	student_study_plans ssp 
	SET deleted_at = NOW() 
	WHERE deleted_at IS NULL 
	AND student_id = ANY($1)`

func (r *StudentStudyPlanRepo) SoftDeleteByStudentID(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) error {
	_, err := db.Exec(ctx, softDeleteStmt, studentIDs)
	if err != nil {
		return err
	}
	return nil
}

const findByStudentIDs = `SELECT study_plan_id
FROM student_study_plans ssp
WHERE student_id = ANY($1)
`

func (r *StudentStudyPlanRepo) FindByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]string, error) {
	rows, err := db.Query(ctx, findByStudentIDs, &studentIDs)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var studyPlanIDs []string
	for rows.Next() {
		var studyPlanID string
		if err := rows.Scan(&studyPlanID); err != nil {
			return nil, fmt.Errorf("rows.Err: %w", err)
		}
		studyPlanIDs = append(studyPlanIDs, studyPlanID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return studyPlanIDs, nil
}

const softDeleteCourseStudentStmt = `UPDATE 
	student_study_plans ssp 
SET deleted_at = NOW() 
WHERE deleted_at IS NULL 
	AND study_plan_id = ANY($1)`

func (r *StudentStudyPlanRepo) SoftDelete(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error {
	_, err := db.Exec(ctx, softDeleteCourseStudentStmt, &studyPlanIDs)
	if err != nil {
		return err
	}
	return nil
}

const findAllStudentStudyPlanStmt = `SELECT sp.%s 
FROM study_plans sp JOIN student_study_plans ssp 
ON sp.study_plan_id = ssp.study_plan_id 
WHERE sp.master_study_plan_id = ANY($1) AND student_id = $2;`

func (r *StudentStudyPlanRepo) FindAllStudentStudyPlan(ctx context.Context, db database.QueryExecer, masterStudentStudyPlanIDs pgtype.TextArray, studentID pgtype.Text) ([]*entities.StudyPlan, error) {
	var studyPlans entities.StudyPlans
	studyPlan := entities.StudyPlan{}
	fieldNames := database.GetFieldNames(&studyPlan)
	query := fmt.Sprintf(findAllStudentStudyPlanStmt, strings.Join(fieldNames, ", sp."))
	err := database.Select(ctx, db, query, masterStudentStudyPlanIDs, &studentID).ScanAll(&studyPlans)
	if err != nil {
		return nil, err
	}
	return studyPlans, nil
}

func (r *StudentStudyPlanRepo) DeleteStudentStudyPlans(ctx context.Context, db database.QueryExecer, studyPlanIds pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentStudyPlanRepo.DeleteStudentStudyPlans")
	defer span.End()

	query := `
		UPDATE student_study_plans AS ssp SET deleted_at = now() WHERE ssp.study_plan_id = ANY($1)
	`
	if _, err := db.Exec(ctx, query, &studyPlanIds); err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}
	return nil
}

const findRetrieveByStudentCourseStmt = `SELECT ssp.%s 
FROM student_study_plans ssp JOIN study_plans sp  
USING(study_plan_id) 
WHERE ssp.deleted_at IS NULL
AND sp.deleted_at IS NULL
AND ssp.student_id = ANY($1::_TEXT) 
AND sp.course_id = ANY($2::_TEXT);`

func (r *StudentStudyPlanRepo) RetrieveByStudentCourse(ctx context.Context, db database.QueryExecer, studentIDs, courseIDs pgtype.TextArray) ([]*entities.StudentStudyPlan, error) {
	var studentStudyPlans entities.StudentStudyPlans
	studentStudyPlan := entities.StudentStudyPlan{}
	fieldNames := database.GetFieldNames(&studentStudyPlan)
	query := fmt.Sprintf(findRetrieveByStudentCourseStmt, strings.Join(fieldNames, ", ssp."))
	err := database.Select(ctx, db, query, &studentIDs, &courseIDs).ScanAll(&studentStudyPlans)
	if err != nil {
		return nil, err
	}
	return studentStudyPlans, nil
}

func (r *StudentStudyPlanRepo) GetByStudyPlanStudentAndLO(ctx context.Context, db database.QueryExecer, studyPlanIDs, studentIDs, loIDs pgtype.TextArray) ([]*entities.StudentStudyPlan, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentStudyPlanRepo.GetByStudyPlanStudentAndLO")
	defer span.End()

	var studentStudyPlan entities.StudentStudyPlan
	var studentStudyPlans entities.StudentStudyPlans

	fields, _ := studentStudyPlan.FieldMap()

	stmt := fmt.Sprintf(`
    SELECT ssp.%s FROM %s ssp
	INNER JOIN individual_study_plan isp 
	ON isp.student_id = ssp.student_id
	AND isp.study_plan_id = ssp.master_study_plan_id
    WHERE isp.learning_material_id = ANY($1::_TEXT)
	AND ssp.master_study_plan_id = ANY($2::_TEXT)
	AND ssp.student_id = ANY($3::_TEXT)
	AND isp.available_from IS NOT NULL
	AND isp.available_to IS NOT NULL
	AND isp.deleted_at IS NULL`, strings.Join(fields, ", ssp."), studentStudyPlan.TableName())

	err := database.Select(ctx, db, stmt, &loIDs, &studyPlanIDs, &studentIDs).ScanAll(&studentStudyPlans)

	if err != nil {
		return nil, err
	}
	return studentStudyPlans, nil
}
