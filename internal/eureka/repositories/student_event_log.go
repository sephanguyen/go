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
	"github.com/pkg/errors"
)

type StudentEventLogRepo struct{}

type StudentEventLogs []*entities.StudentEventLog

// Add append new ShuffledQuizSet
func (u *StudentEventLogs) Add() database.Entity {
	e := &entities.StudentEventLog{}
	*u = append(*u, e)

	return e
}

func (r *StudentEventLogRepo) RetrieveStudentEventLogsByStudyPlanItemIDs(
	ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray,
) ([]*entities.StudentEventLog, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEventLogRepo.RetrieveStudentEventLogsByStudyPlanItemIDs")
	defer span.End()
	e := &entities.StudentEventLog{}
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE study_plan_item_id = ANY($1::_TEXT) ORDER BY created_at`, strings.Join(database.GetFieldNames(e), ","), e.TableName())

	ss := StudentEventLogs{}
	if err := database.Select(ctx, db, query, &studyPlanItemIDs).ScanAll(&ss); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return ss, nil
}

func (r *StudentEventLogRepo) RetrieveStudentEventLogsByStudyPlanIdentities(
	ctx context.Context, db database.QueryExecer, studyPlanItemIdentities []*StudyPlanItemIdentity,
) ([]*entities.StudentEventLog, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEventLogRepo.RetrieveStudentEventLogsByStudyPlanIdentities")
	defer span.End()

	args := make([]interface{}, 0, 3*len(studyPlanItemIdentities))
	for _, studyPlanItemIdentity := range studyPlanItemIdentities {
		args = append(args, studyPlanItemIdentity.StudentID, studyPlanItemIdentity.StudyPlanID, studyPlanItemIdentity.LearningMaterialID)
	}

	var placeHolders string

	for i := 0; i < len(studyPlanItemIdentities); i++ {
		placeHolders += fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
		if i != len(studyPlanItemIdentities)-1 {
			placeHolders += ", "
		}
	}
	placeHolders = "(" + placeHolders + ")"

	e := &entities.StudentEventLog{}
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE (student_id, study_plan_id, learning_material_id) IN %s ORDER BY created_at`, strings.Join(database.GetFieldNames(e), ","), e.TableName(), placeHolders)

	ss := StudentEventLogs{}
	if err := database.Select(ctx, db, query, args...).ScanAll(&ss); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return ss, nil
}

func (r *StudentEventLogRepo) queueCreatedQuery(b *pgx.Batch, s *entities.StudentEventLog) {
	fieldNames := []string{"student_id", "event_id", "event_type", "payload", "created_at"}

	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf("INSERT INTO student_event_logs (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT event_id_un DO NOTHING;", strings.Join(fieldNames, ","), placeHolders)

	b.Queue(query, database.GetScanFields(s, fieldNames)...)
}

func (r *StudentEventLogRepo) Create(ctx context.Context, db database.QueryExecer, ss []*entities.StudentEventLog) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentEventLogRepo.Create")
	defer span.End()

	b := &pgx.Batch{}

	for _, s := range ss {
		r.queueCreatedQuery(b, s)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(ss); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}

	return nil
}

type QuestionSubmissionResult struct {
	QuestionID string
	Correct    bool
}

func (r *StudentEventLogRepo) LogsQuestionSubmitionByLO(ctx context.Context, db database.QueryExecer, studentID string, loIDs pgtype.TextArray) (map[string][]*QuestionSubmissionResult, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEventLogRepo.LogsQuestionSubmitionByLO")
	defer span.End()

	const query = `
SELECT DISTINCT ON (question_id) payload->>'question_id' AS question_id,
                                 payload->>'correctness' AS correctness,
                                 payload->>'lo_id' AS lo_id
FROM student_event_logs
WHERE student_id=$1::TEXT
  AND event_type='quiz_answer_selected'
  AND payload->>'lo_id'=ANY($2::TEXT[])
ORDER BY question_id,
         created_at DESC
  `

	rows, err := db.Query(ctx, query, &studentID, &loIDs)
	if err != nil {
		return nil, fmt.Errorf("LogsQuestionSubmitionByLO.Query %w", err)
	}
	defer rows.Close()

	results := make(map[string][]*QuestionSubmissionResult)
	for rows.Next() {
		s := &QuestionSubmissionResult{}
		var loId, questionID pgtype.Text
		var correct pgtype.Text
		if err := rows.Scan(&questionID, &correct, &loId); err != nil {
			return nil, fmt.Errorf("LogsQuestionSubmitionByLO.Scan %w", err)
		}
		s.QuestionID = questionID.String
		s.Correct = (correct.String == "true")

		results[loId.String] = append(results[loId.String], s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("LogsQuestionSubmitionByLO.Err %w", err)
	}

	return results, nil
}

func (r *StudentEventLogRepo) DeleteByStudyPlanIdentities(ctx context.Context, db database.QueryExecer, args StudyPlanItemIdentity) (int64, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEventLogRepo.DeleteByStudyPlanIdentities")
	defer span.End()
	e := entities.StudentEventLog{}
	query := fmt.Sprintf(`UPDATE %s SET deleted_at = now() WHERE learning_material_id = $1::TEXT AND student_id = $2::TEXT AND study_plan_id = $3::TEXT AND deleted_at IS NULL`, e.TableName())
	cmdTag, err := db.Exec(ctx, query, args.LearningMaterialID, args.StudentID, args.StudyPlanID)
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}
	return cmdTag.RowsAffected(), nil
}
func (r *StudentEventLogRepo) Retrieve(ctx context.Context, db database.QueryExecer, studentID, sessionID string, from, to *pgtype.Timestamptz) ([]*entities.StudentEventLog, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEventLogRepo.Retrieve")
	defer span.End()

	fields := database.GetFieldNames(&entities.StudentEventLog{})
	args := []interface{}{studentID}
	query := fmt.Sprintf("SELECT %s FROM student_event_logs WHERE student_id = $1", strings.Join(fields, ","))
	if sessionID != "" {
		args = append(args, sessionID)
		query += fmt.Sprintf(" AND payload->>'session_id' = $%d", len(args))
	}
	if from != nil {
		args = append(args, from)
		query += fmt.Sprintf(" AND created_at >= $%d", len(args))
	}
	if to != nil {
		args = append(args, to)
		query += fmt.Sprintf(" AND created_at <= $%d", len(args))
	}
	query += " ORDER BY created_at ASC"
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var ss []*entities.StudentEventLog
	for rows.Next() {
		s := new(entities.StudentEventLog)
		if err := rows.Scan(database.GetScanFields(s, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		ss = append(ss, s)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return ss, nil
}
