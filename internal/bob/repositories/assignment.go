package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type AssignmentRepo struct{}

type AssignmentWithTopic struct {
	Topic       *entities_bob.Topic
	Assignment  *entities_bob.Assignment
	User        *entities_bob.User
	CompletedAt *pgtype.Timestamptz
}

type AssignmentPagination struct {
	Assignments []*entities_bob.Assignment
	Total       pgtype.Int8
}

func (r *AssignmentRepo) FindAssignmentByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities_bob.Assignment, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.FindAssignmentById")
	defer span.End()
	t := &entities_bob.Assignment{}
	fields := database.GetFieldNames(t)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE assignment_id = ANY ($1) ", strings.Join(fields, ","), t.TableName())
	rows, err := db.Query(ctx, query, &ids)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pp []*entities_bob.Assignment
	for rows.Next() {
		p := new(entities_bob.Assignment)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (r *AssignmentRepo) FindClassAssignment(ctx context.Context, db database.QueryExecer, classID pgtype.Int4, isPassed bool, limit int, page int) (*AssignmentPagination, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.FindClassAssignment")
	defer span.End()

	t := &entities_bob.Assignment{}
	fields := database.GetFieldNames(t)
	args := []interface{}{&classID}
	query := fmt.Sprintf("SELECT %s, COUNT(*) OVER() AS total FROM %s WHERE class_id = $1 AND deleted_at IS NULL AND ", strings.Join(fields, ","), t.TableName())
	filter := "end_date > NOW()"
	if isPassed {
		filter = "end_date < NOW()"
	}
	orderBy := " ORDER BY end_date, assignment_id ASC"

	query += filter + orderBy

	if page > 0 {
		if limit == 0 {
			limit = 10
		}
		args = append(args, limit, limit*(page-1))
		query += " LIMIT $2 OFFSET $3"
	}

	rows, err := db.Query(ctx, query, args...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pp []*entities_bob.Assignment
	var total pgtype.Int8
	for rows.Next() {
		p := new(entities_bob.Assignment)
		scanFields := append(database.GetScanFields(p, fields), &total)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return &AssignmentPagination{
		Assignments: pp,
		Total:       total,
	}, nil
}

func (r *AssignmentRepo) queueAssignment(b *pgx.Batch, t *entities_bob.Assignment) {
	fieldNames := database.GetFieldNames(t)
	// placeHolders :=  database.GeneratePlaceholders(len(fieldNames))
	placeHolders := "$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12"

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT assignments_pk DO NOTHING", t.TableName(), strings.Join(fieldNames, ","), placeHolders)

	b.Queue(query, database.GetScanFields(t, fieldNames)...)
}

func (r *AssignmentRepo) queueStudentAssignment(b *pgx.Batch, t *entities_bob.StudentAssignment) {
	fieldNames := database.GetFieldNames(t)
	placeHolders := "$1, $2, $3, $4, $5, $6"

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT students_assignments_pk DO UPDATE SET assignment_status = $3, updated_at = $4", t.TableName(), strings.Join(fieldNames, ","), placeHolders)
	b.Queue(query, database.GetScanFields(t, fieldNames)...)
}

func (r *AssignmentRepo) ExecQueueAssignment(ctx context.Context, db database.QueryExecer, assignments []*entities_bob.Assignment) error {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.ExecQueueAssignment")
	defer span.End()
	now := time.Now()

	b := &pgx.Batch{}

	for _, assignment := range assignments {
		assignment.CreatedAt.Set(now)
		assignment.UpdatedAt.Set(now)
		assignment.DeletedAt.Set(nil)
		r.queueAssignment(b, assignment)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(assignments); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("assignment not inserted")
		}
	}
	return nil
}

func (r *AssignmentRepo) ExecQueueStudentAssignment(ctx context.Context, db database.QueryExecer, studentAssignments []*entities_bob.StudentAssignment) error {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.ExecQueueStudentAssignment")
	defer span.End()

	now := time.Now()

	b := &pgx.Batch{}

	for _, studentAssignment := range studentAssignments {
		studentAssignment.CreatedAt.Set(now)
		studentAssignment.UpdatedAt.Set(now)
		r.queueStudentAssignment(b, studentAssignment)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(studentAssignments); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("student assignment not inserted")
		}
	}

	return nil
}

func (r *AssignmentRepo) DeleteAssignment(ctx context.Context, db database.QueryExecer, assignmentIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.DeleteAssignment")
	defer span.End()
	var deletedAt pgtype.Timestamptz
	deletedAt.Set(time.Now())

	a := &entities_bob.Assignment{}
	query := fmt.Sprintf("UPDATE %s SET deleted_at = $1 WHERE assignment_id = ANY($2)", a.TableName())
	_, err := db.Exec(ctx, query, &deletedAt, &assignmentIDs)
	if err != nil {
		return err
	}
	return nil
}

func (r *AssignmentRepo) RetrieveStudentAssignment(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, isActive bool, userGroup pgtype.Text) ([]Topic, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.RetrieveStudentAssignment")
	defer span.End()

	args := []interface{}{&studentID}
	query := `SELECT topics.topic_id, topics.name, start_date, end_date, topics.total_los, stc.total_finished_los, u.user_id, u.name, u.user_group, u.avatar
				FROM public.assignments asm
				JOIN public.student_assignments stasm
					ON asm.assignment_id= stasm.assignment_id
				LEFT JOIN public.students_topics_completeness stc
					ON asm.topic_id=stc.topic_id
					AND stasm.student_id=stc.student_id
				LEFT JOIN public.topics ON (topics.topic_id=asm.topic_id)
				LEFT JOIN users u ON (asm.assigned_by = u.user_id)
				WHERE stasm.student_id=$1 AND asm.deleted_at IS NULL AND topics.deleted_at IS NULL`

	if isActive {
		query += ` AND stasm.assignment_status='STUDENT_ASSIGNMENT_STATUS_ACTIVE' AND asm.end_date >= NOW() `
	} else {
		query += ` AND (stasm.assignment_status='STUDENT_ASSIGNMENT_STATUS_ACTIVE' OR stasm.assignment_status='STUDENT_ASSIGNMENT_STATUS_COMPLETED')`
	}
	if from != nil {
		args = append(args, from)
		query += fmt.Sprintf(" AND asm.start_date >= $%d", len(args))
	}
	if to != nil {
		args = append(args, to)
		query += fmt.Sprintf(" AND asm.start_date <= $%d", len(args))
	}

	query += " ORDER BY asm.end_date, asm.assignment_id asc"
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "QueryEx")
	}
	defer rows.Close()

	var pp []Topic
	for rows.Next() {
		var p Topic
		if err := rows.Scan(&p.Topic.ID, &p.Topic.Name, &p.StartDate, &p.EndDate, &p.TotalLOs,
			&p.TotalFinishedLOs, &p.AssignedBy.ID, &p.AssignedBy.LastName, &p.AssignedBy.Group, &p.AssignedBy.Avatar); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return pp, nil
}

func (r *AssignmentRepo) FindStudentAssignmentWithStudyPlan(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.FindStudentAssignmentWithStudyPlan")
	defer span.End()
	query := `SELECT a.assignment_id FROM student_assignments sa JOIN assignments a ON (sa.assignment_id=a.assignment_id) WHERE student_id = $1
	AND preset_study_plan_id NOTNULL AND sa.deleted_at IS NULL`
	rows, err := db.Query(ctx, query, &studentID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignmentIDs []string
	for rows.Next() {
		var assignmendID string
		if err := rows.Scan(&assignmendID); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		assignmentIDs = append(assignmentIDs, assignmendID)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return assignmentIDs, nil
}

func (r *AssignmentRepo) FindStudentOverdueAssignment(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*AssignmentWithTopic, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.FindStudentOverdueAssignment")
	defer span.End()
	a := &entities_bob.Assignment{}
	assignmentFields := database.GetFieldNames(a)
	t := &entities_bob.Topic{}
	topicFields := database.GetFieldNames(t)
	query := fmt.Sprintf(`SELECT a.%s, t.%s, u.user_id, u.name, u.user_group, u.avatar FROM assignments a
		JOIN student_assignments sta ON (a.assignment_id = sta.assignment_id)
		JOIN topics t ON (a.topic_id = t.topic_id)
		JOIN users u ON (u.user_id = a.assigned_by)
		WHERE sta.student_id = $1 AND a.deleted_at IS NULL AND end_date < NOW() AND sta.assignment_status='STUDENT_ASSIGNMENT_STATUS_ACTIVE' AND t.deleted_at IS NULL
		ORDER BY a.end_date desc`, strings.Join(assignmentFields, ", a."), strings.Join(topicFields, ", t."))

	rows, err := db.Query(ctx, query, &studentID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pp []*AssignmentWithTopic
	for rows.Next() {
		a := new(entities_bob.Assignment)
		topic := new(entities_bob.Topic)
		user := new(entities_bob.User)
		fields := database.GetScanFields(a, assignmentFields)
		fields = append(fields, database.GetScanFields(topic, topicFields)...)
		fields = append(fields, &user.ID, &user.LastName, &user.Group, &user.Avatar)
		if err := rows.Scan(fields...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		p := &AssignmentWithTopic{
			Assignment: a,
			Topic:      topic,
			User:       user,
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (r *AssignmentRepo) FindStudentCompletedAssignmentWeeklies(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz) ([]*AssignmentWithTopic, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.FindStudentCompletedAssignmentWeeklies")
	defer span.End()
	a := &entities_bob.Assignment{}
	assignmentFields := database.GetFieldNames(a)
	t := &entities_bob.Topic{}
	topicFields := database.GetFieldNames(t)
	args := []interface{}{&studentID}
	query := fmt.Sprintf(`SELECT a.%s, t.%s, u.user_id, u.name, u.user_group, u.avatar, sta.completed_at FROM assignments a
		JOIN student_assignments sta ON (a.assignment_id = sta.assignment_id)
		JOIN topics t ON (a.topic_id = t.topic_id)
		JOIN users u ON (u.user_id = a.assigned_by)
		WHERE sta.student_id = $1 AND a.deleted_at IS NULL AND sta.assignment_status='STUDENT_ASSIGNMENT_STATUS_COMPLETED'`, strings.Join(assignmentFields, ", a."), strings.Join(topicFields, ", t."))

	if from != nil {
		args = append(args, from)
		query += fmt.Sprintf(" AND sta.completed_at >= $%d", len(args))
	}
	if to != nil {
		args = append(args, to)
		query += fmt.Sprintf(" AND sta.completed_at <= $%d", len(args))
	}
	query += " ORDER BY sta.completed_at desc"
	rows, err := db.Query(ctx, query, args...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pp []*AssignmentWithTopic
	for rows.Next() {
		a := new(entities_bob.Assignment)
		topic := new(entities_bob.Topic)
		user := new(entities_bob.User)
		var completedAt pgtype.Timestamptz
		fields := database.GetScanFields(a, assignmentFields)
		fields = append(fields, database.GetScanFields(topic, topicFields)...)
		fields = append(fields, &user.ID, &user.LastName, &user.Group, &user.Avatar)
		fields = append(fields, &completedAt)
		if err := rows.Scan(fields...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		p := &AssignmentWithTopic{
			Assignment:  a,
			Topic:       topic,
			User:        user,
			CompletedAt: &completedAt,
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (r *AssignmentRepo) RetrieveStudentAssignmentByTopic(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, topicIDs pgtype.TextArray) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.RetrieveStudentAssignmentIDs")
	defer span.End()

	args := []interface{}{&studentID, &topicIDs}
	query := `SELECT asm.assignment_id
				FROM public.assignments asm
				JOIN public.student_assignments stasm
					ON asm.assignment_id= stasm.assignment_id
				WHERE stasm.student_id=$1 AND asm.deleted_at IS NULL AND asm.topic_id = ANY($2)
					AND NOT stasm.assignment_status='STUDENT_ASSIGNMENT_STATUS_REMOVED_FROM_CLASS'`

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "QueryEx")
	}
	defer rows.Close()

	var assignmentIDs []string
	for rows.Next() {
		var assignmentID string
		if err := rows.Scan(&assignmentID); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		assignmentIDs = append(assignmentIDs, assignmentID)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return assignmentIDs, nil
}

func (r *AssignmentRepo) FindByTopicID(ctx context.Context, db database.QueryExecer, topicID pgtype.Text) (*entities_bob.Assignment, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.FindByTopicID")
	defer span.End()

	e := new(entities_bob.Assignment)
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE topic_id = $1", strings.Join(fields, ","), e.TableName())

	row := db.QueryRow(ctx, selectStmt, &topicID)
	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, err
	}

	return e, nil
}

func (r *AssignmentRepo) RetrieveByTopicIDs(ctx context.Context, db database.QueryExecer, topicID pgtype.TextArray) ([]*entities_bob.Assignment, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.FindByTopicID")
	defer span.End()

	e := new(entities_bob.Assignment)
	fields := database.GetFieldNames(e)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE topic_id = ANY($1) AND deleted_at IS NULL", strings.Join(fields, ","), e.TableName())

	var result []*entities_bob.Assignment
	rows, err := db.Query(ctx, stmt, &topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var e entities_bob.Assignment
		if err := rows.Scan(&e); err != nil {
			return nil, err
		}
		result = append(result, &e)
	}
	return result, nil
}
