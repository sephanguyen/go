package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	eureka_db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type AssignmentRepo struct {
}

const bulkUpsetAssignmentStmtTpl = `
INSERT INTO %s (%s) 
VALUES %s ON CONFLICT ON CONSTRAINT assignments_pk DO UPDATE 
SET
	content = excluded.content::jsonb, 
	attachment = excluded.attachment, 
	settings = excluded.settings::jsonb, 
	check_list = excluded.check_list::jsonb, 
	created_at = excluded.created_at, 
	updated_at = excluded.updated_at, 
	deleted_at = excluded.deleted_at, 
	max_grade = excluded.max_grade, 
	status = excluded.status, 
	instruction = excluded.instruction, 
	type = excluded.type, 
	name = excluded.name, 
	is_required_grade = excluded.is_required_grade, 
	original_topic = excluded.original_topic,
	topic_id = excluded.topic_id`

// GetAssignmentSetting selects and parses AssignmentSetting
func (r *AssignmentRepo) GetAssignmentSetting(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.AssignmentSetting, error) {
	stmt := `SELECT settings FROM assignments
		WHERE assignment_id = $1`

	var v pgtype.JSONB
	if err := database.Select(ctx, db, stmt, &id).ScanFields(&v); err != nil {
		return nil, err
	}

	setting := &entities.AssignmentSetting{}
	if err := v.AssignTo(&setting); err != nil {
		return nil, err
	}

	return setting, nil
}

func (r *AssignmentRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, assignments []*entities.Assignment) error {
	err := eureka_db.BulkUpsert(ctx, db, bulkUpsetAssignmentStmtTpl, assignments)
	if err != nil {
		return fmt.Errorf("eureka_db.BulkUpsertAssignment error: %s", err.Error())
	}
	return nil
}

// IsStudentAssigned check if student was assigned to a specific assignment. Errors returned are unexpected error,
// for most of the case, check the boolean
func (r *AssignmentRepo) IsStudentAssigned(ctx context.Context, db database.QueryExecer, studyPlanItemID, assignmentID, studentID pgtype.Text) (bool, error) {
	stmt := `SELECT COUNT(*) FROM
	individual_study_plans_view 
	WHERE student_id = $2 and learning_material_id = $1`

	var count pgtype.Int8
	err := database.Select(ctx, db, stmt, &assignmentID, &studentID).ScanFields(&count)
	if err != nil {
		return false, err
	}

	return count.Int > 0, nil
}

// IsStudentAssigned check if student was assigned to a specific assignment. Errors returned are unexpected error,
// for most of the case, check the boolean
func (r *AssignmentRepo) IsStudentAssignedV2(ctx context.Context, db database.QueryExecer, studyPlanID, studentID pgtype.Text) (bool, error) {
	query := `SELECT EXISTS 
		(
			SELECT 1 
			FROM study_plans 
			JOIN course_students
			USING(course_id)
			WHERE study_plan_id = $1 AND student_id = $2)
			`

	var isStudentAssigned pgtype.Bool
	err := database.Select(ctx, db, query, studyPlanID, studentID).ScanFields(&isStudentAssigned)
	if err != nil {
		return false, err
	}

	return isStudentAssigned.Bool, nil
}

// SoftDelete for assignment
func (r *AssignmentRepo) SoftDelete(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error {
	query := `UPDATE assignments SET deleted_at = NOW() WHERE assignment_id = ANY($1) AND deleted_at IS NULL`
	cmd, err := db.Exec(ctx, query, &ids)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return errors.New("cannot delete assignments")
	}
	return nil
}

func (r *AssignmentRepo) RetrieveAssignments(ctx context.Context, db database.QueryExecer, assignmentIDs pgtype.TextArray) ([]*entities.Assignment, error) {
	assignment := &entities.Assignment{}
	fieldNames := database.GetFieldNames(assignment)
	query := `
		SELECT
			%s
		FROM
			%s
		WHERE
			assignment_id = ANY($1)
		AND deleted_at IS NULL
		`
	var assignments entities.Assignments
	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fieldNames, ","), assignment.TableName()), &assignmentIDs).ScanAll(&assignments)
	if err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *AssignmentRepo) RetrieveAssignmentsByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.Assignment, error) {
	assignment := &entities.Assignment{}
	fieldNames := database.GetFieldNames(assignment)
	query := `
		SELECT
			%s
		FROM
			%s
		WHERE
			topic_id = ANY($1)
			AND deleted_at IS NULL
		`
	var assignments entities.Assignments
	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fieldNames, ","), assignment.TableName()), &topicIDs).ScanAll(&assignments)
	if err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *AssignmentRepo) QueueDuplicateAssignment(b *pgx.Batch, copiedFromTopicID pgtype.Text, newTopicID pgtype.Text) {
	assignment := &entities.Assignment{}
	var assignmentContent pgtype.JSONB
	content := entities.AssignmentContent{
		TopicID: newTopicID.String,
	}
	_ = assignmentContent.Set(content)

	fieldNames := database.GetFieldNames(assignment)
	selectFields := golibs.Replace(fieldNames, []string{"assignment_id", "created_at", "updated_at", "content", "topic_id"},
		[]string{"generate_ulid()", "NOW()", "NOW()", "$1", "$3"})

	query := fmt.Sprintf(`
		INSERT INTO 
		%s (%s)
		SELECT
		%s
		FROM
		%s
		WHERE
		"content" ->>'topic_id' = $2
		AND deleted_at IS NULL
	`, assignment.TableName(), strings.Join(fieldNames, ", "), strings.Join(selectFields, ", "), assignment.TableName())

	b.Queue(query, &assignmentContent, &copiedFromTopicID, &newTopicID)
}

func (r *AssignmentRepo) DuplicateAssignment(ctx context.Context, db database.QueryExecer, copiedFromTopicIDs pgtype.TextArray, newTopicIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.DuplicateAssignment")
	defer span.End()
	b := &pgx.Batch{}
	for i, copiedTopicID := range copiedFromTopicIDs.Elements {
		newTopicID := newTopicIDs.Elements[i]
		r.QueueDuplicateAssignment(b, copiedTopicID, newTopicID)
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

func (r *AssignmentRepo) UpdateDisplayOrders(ctx context.Context, db database.QueryExecer, mDisplayOrder map[pgtype.Text]pgtype.Int4) error {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.UpdateDisplayOrders")
	defer span.End()

	queueFn := func(b *pgx.Batch, assignmentID pgtype.Text, displayOrder pgtype.Int4) {
		query := `UPDATE assignments 
		SET display_order = $1, updated_at = NOW()
		WHERE assignment_id = $2 AND deleted_at IS NULL`
		b.Queue(query, &displayOrder, &assignmentID)
	}

	var d pgtype.Timestamptz
	if err := d.Set(time.Now()); err != nil {
		return err
	}

	b := &pgx.Batch{}
	for assignmentID, displayOrder := range mDisplayOrder {
		queueFn(b, assignmentID, displayOrder)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(mDisplayOrder); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("display_order not changed")
		}
	}

	return nil
}

type CalculateHighestScoreResponse struct {
	StudyPlanItemID pgtype.Text
	Percentage      pgtype.Float4
}

func (r *AssignmentRepo) CalculateHigestScore(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*CalculateHighestScoreResponse, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.CalculateHigestScore")
	defer span.End()

	query := `
		SELECT study_plan_item_id, max(grade/max_grade) * 100 AS percentage
		FROM assignments a
		JOIN student_submissions ss ON a.assignment_id = ss.assignment_id
		JOIN student_submission_grades ssg ON ss.student_submission_id = ssg.student_submission_id
			WHERE a.deleted_at IS NULL
				AND ss.deleted_at IS NULL
				AND ssg.deleted_at IS NULL
				AND ssg.status = 'SUBMISSION_STATUS_RETURNED'
				AND ss.study_plan_item_id = ANY($1::TEXT[])
				AND a.type != 'ASSIGNMENT_TYPE_TASK'
				AND a.max_grade > 0
				and ssg.grade != -1
		GROUP BY ss.study_plan_item_id
	`

	var res []*CalculateHighestScoreResponse
	rows, err := db.Query(ctx, query, &studyPlanItemIDs)
	if err != nil {
		return nil, fmt.Errorf("AssignmentRepo.CalculateHigestScore.Query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var studyPlanItemID pgtype.Text
		var percentage float32
		if err := rows.Scan(&studyPlanItemID, &percentage); err != nil {
			return nil, fmt.Errorf("AssignmentRepo.CalculateHigestScore.Scan: %w", err)
		}

		res = append(res, &CalculateHighestScoreResponse{
			StudyPlanItemID: studyPlanItemID,
			Percentage:      database.Float4(percentage),
		})
	}

	return res, nil
}

func (r *AssignmentRepo) CalculateTaskAssignmentHighestScore(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*CalculateHighestScoreResponse, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.CalculateTaskAssignmentHighestScore")
	defer span.End()

	// TODO(giahuy): probably add index on table student_submissions to avoid seq scan
	const query = `
SELECT study_plan_item_id,
  max((correct_score/total_score)*100)::FLOAT4 AS grade
FROM student_submissions ss
JOIN assignments a ON ss.assignment_id = a.assignment_id
WHERE ss.study_plan_item_id = ANY($1::TEXT[])
  AND ss.deleted_at IS NULL
  AND a.deleted_at IS NULL
  AND a.type = 'ASSIGNMENT_TYPE_TASK'
  AND ss.total_score IS NOT NULL
  AND ss.total_score != 0
GROUP BY study_plan_item_id
  `

	var res []*CalculateHighestScoreResponse
	rows, err := db.Query(ctx, query, &studyPlanItemIDs)
	if err != nil {
		return nil, fmt.Errorf("AssignmentRepo.CalculateTaskAssignmentHighestScore.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var studyPlanItemID pgtype.Text
		var percentage pgtype.Float4
		if err := rows.Scan(&studyPlanItemID, &percentage); err != nil {
			return nil, fmt.Errorf("AssignmentRepo.CalculateTaskAssignmentHighestScore.Scan: %w", err)
		}

		res = append(res, &CalculateHighestScoreResponse{
			StudyPlanItemID: studyPlanItemID,
			Percentage:      percentage,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("AssignmentRepo.CalculateTaskAssignmentHighestScore.Err() %w", err)
	}

	return res, nil
}

func (r *AssignmentRepo) RetrieveByIntervalTime(ctx context.Context, db database.QueryExecer, intervalTime pgtype.Text) ([]*entities.Assignment, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseStudentRepo.RetrieveByIntervalTime")
	defer span.End()
	stmt := "SELECT %s FROM %s WHERE deleted_at IS NULL AND updated_at >= ( now() - $1::interval)"
	var e entities.Assignment
	selectFields := database.GetFieldNames(&e)

	query := fmt.Sprintf(stmt, strings.Join(selectFields, ", "), e.TableName())

	var items entities.Assignments
	err := database.Select(ctx, db, query, &intervalTime).ScanAll(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *AssignmentRepo) RetrieveBookAssignmentByIntervalTime(ctx context.Context, db database.QueryExecer, intervalTime pgtype.Text) ([]*entities.BookAssignment, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.RetrieveBookAssignmentByIntervalTime")
	defer span.End()
	stmt := `
	SELECT assignments.%s, book_id, books_chapters.chapter_id, topics.topic_id FROM assignments
	JOIN topics ON topics.topic_id = assignments.content->>'topic_id'
	JOIN books_chapters USING (chapter_id)
	WHERE topics.deleted_at IS NULL
      AND assignments.deleted_at IS NULL
      AND books_chapters.deleted_at IS NULL
	  AND assignments.deleted_at IS NULL AND assignments.updated_at >= ( now() - $1::interval)`
	lo := &entities.Assignment{}
	fields := database.GetFieldNames(lo)
	stmtSelect := fmt.Sprintf(stmt, strings.Join(fields, ", assignments."))
	rows, err := db.Query(ctx, stmtSelect, intervalTime)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	bAssignments := make([]*entities.BookAssignment, 0)

	for rows.Next() {
		assignmentTemp := entities.Assignment{}
		var (
			bookID, chapterID, topicID pgtype.Text
		)
		scanFields := database.GetScanFields(&assignmentTemp, fields)
		scanFields = append(scanFields, &bookID, &chapterID, &topicID)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		bAssignment := entities.BookAssignment{
			Assignment: assignmentTemp,
			BookID:     bookID,
			ChapterID:  chapterID,
			TopicID:    topicID,
		}
		bAssignments = append(bAssignments, &bAssignment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row.Err: %w", err)
	}

	return bAssignments, nil
}
