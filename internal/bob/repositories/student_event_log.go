package repositories

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type StudentEventLogRepo struct{}

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
			return errors.Wrap(err, "batchResults.Exec")
		}
	}

	return nil
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

func (r *StudentEventLogRepo) LogsQuestionSubmitionByLO(ctx context.Context, db database.QueryExecer, studentID string, loIDs pgtype.TextArray) (map[string][]*pb.SubmissionResult, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEventLogRepo.LogsQuestionSubmitionByLO")
	defer span.End()

	query := "SELECT DISTINCT ON (question_id) payload->>'question_id' AS question_id, payload->>'correctness' AS correctness, payload->>'lo_id' AS lo_id FROM student_event_logs WHERE student_id=$1 AND event_type='quiz_answer_selected' AND payload->>'lo_id'=ANY($2) ORDER BY question_id, created_at desc"

	rows, err := db.Query(ctx, query, &studentID, &loIDs)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	resp := make(map[string][]*pb.SubmissionResult)
	for rows.Next() {
		s := new(pb.SubmissionResult)
		var loId, questionID pgtype.Text
		var correct pgtype.Text
		if err := rows.Scan(&questionID, &correct, &loId); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		s.QuestionId = questionID.String
		s.Correct = correct.String == "true"

		resp[loId.String] = append(resp[loId.String], s)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return resp, nil
}

type SubmissionOrderType int

const (
	SubmissionOrderTypeFirst SubmissionOrderType = iota + 1
	SubmissionOrderTypeLast
)

type SubmissionPagination struct {
	Submissions []*Submission
	Total       pgtype.Int8
}

type Submission struct {
	Question       *entities.Question
	Correct        bool
	SelectedAnswer int
}

func (r *StudentEventLogRepo) RetrieveStudentSubmissionsByLO(ctx context.Context, db database.QueryExecer, studentID, loID pgtype.Text, o SubmissionOrderType, limit, offset int) (*SubmissionPagination, error) {
	var orderBy string

	if o == SubmissionOrderTypeFirst {
		orderBy = "created_at ASC"
	} else {
		orderBy = "created_at DESC"
	}

	// getSessionIDQuery gets the oldest/latest session id that student does the LO.
	getSessionIDQuery := fmt.Sprintf(`SELECT payload->>'session_id' AS session_id, MIN(created_at) AS created_at
		FROM student_event_logs
		WHERE student_id = $1 AND event_type = 'quiz_answer_selected' AND payload->>'lo_id' = $2
		GROUP BY payload->>'session_id'
		ORDER BY %s LIMIT 1`, orderBy)

	var sessionID, createdAt pgtype.Text
	if err := db.QueryRow(ctx, getSessionIDQuery, studentID, loID).Scan(&sessionID, &createdAt); err != nil {
		if err != pgx.ErrNoRows {
			return nil, err
		}
		// set session id to Present to let below query can run successfully
		sessionID = pgtype.Text{Status: pgtype.Present}
	}

	q := &entities.Question{}
	questionFields := database.GetFieldNames(q)
	query := fmt.Sprintf(`SELECT questions.%s, sub.correctness, sub.lo_id, sub.selected_answer_index, COUNT(*) OVER() AS total
		FROM quizsets
		INNER JOIN questions ON questions.question_id = quizsets.question_id
		LEFT JOIN (
			SELECT payload->>'question_id' AS question_id, payload->>'correctness' AS correctness,
				payload->>'lo_id' AS lo_id, payload->>'selected_answer_index' AS selected_answer_index
			FROM student_event_logs
			WHERE student_id = $1 AND payload->>'session_id' = $2
		) sub ON sub.question_id = questions.question_id
		WHERE quizsets.lo_id = $3
		ORDER BY quizsets.display_order ASC
		LIMIT $4 OFFSET $5`, strings.Join(questionFields, ",questions."))

	rows, err := db.Query(ctx, query, &studentID, &sessionID, &loID, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	sp := new(SubmissionPagination)
	for rows.Next() {
		var (
			e                                 = new(entities.Question)
			loID, correctness, selectedAnswer pgtype.Text
			total                             pgtype.Int8
		)

		scanFields := database.GetScanFields(e, questionFields)
		scanFields = append(scanFields, &correctness, &loID, &selectedAnswer, &total)

		if err := rows.Scan(scanFields...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		sub := &Submission{
			Question: e,
			Correct:  correctness.String == "true",
		}
		if selectedAnswer.Status == pgtype.Null {
			sub.SelectedAnswer = -1
		} else {
			selected, _ := strconv.Atoi(selectedAnswer.String)
			sub.SelectedAnswer = selected
		}

		sp.Submissions = append(sp.Submissions, sub)
		sp.Total = total
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return sp, nil
}

func (r *StudentEventLogRepo) RetrieveBySessions(ctx context.Context, db database.QueryExecer, sessionIDs pgtype.TextArray) (map[string][]*entities.StudentEventLog, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEventLogRepo.RetrieveBySessions")
	defer span.End()

	fields := database.GetFieldNames(&entities.StudentEventLog{})
	args := []interface{}{&sessionIDs}
	query := fmt.Sprintf(`SELECT %s FROM student_event_logs WHERE event_type = 'learning_objective' AND payload->>'session_id' = ANY($1)
		ORDER BY created_at ASC`, strings.Join(fields, ", "))

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	resp := make(map[string][]*entities.StudentEventLog) // student id => array of logs
	for rows.Next() {
		s := new(entities.StudentEventLog)
		if err := rows.Scan(database.GetScanFields(s, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		resp[s.StudentID.String] = append(resp[s.StudentID.String], s)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return resp, nil
}

func (r *StudentEventLogRepo) GetSubmitAnswerEventLog(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, quizID pgtype.Text) (*entities.StudentEventLog, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEventLogRepo.GetSubmitAnswerEventLog")
	defer span.End()

	e := &entities.StudentEventLog{}
	fields, values := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1 AND event_type = 'quiz_answer_selected' AND payload->>'question_id' = $2", strings.Join(fields, ","), e.TableName())
	err := db.QueryRow(ctx, query, &studentID, &quizID).Scan(values...)
	if err != nil {
		return nil, errors.Wrap(err, "db.QueryRowEx")
	}

	return e, nil
}

func (r *StudentEventLogRepo) RetrieveLOEvents(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz) ([]*entities.StudentEventLog, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEventLogRepo.RetrieveCompletedLOLogs")
	defer span.End()

	fields := database.GetFieldNames(&entities.StudentEventLog{})
	args := []interface{}{&studentID}
	query := fmt.Sprintf("SELECT %s FROM student_event_logs WHERE student_id = $1 AND event_type = 'learning_objective'", strings.Join(fields, ","))
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

func (r *StudentEventLogRepo) RetrieveStudentEventLogsByStudyPlanItemIDs(
	ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray,
) ([]*entities.StudentEventLog, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEventLogRepo.RetrieveStudentEventLogsByStudyPlanItemIDs")
	defer span.End()

	fields := database.GetFieldNames(&entities.StudentEventLog{})
	query := fmt.Sprintf(`SELECT %s 
			FROM student_event_logs 
			WHERE payload->>'study_plan_item_id' = ANY($1) AND payload->>'study_plan_item_id' IS NOT NULL
			ORDER BY created_at`, strings.Join(fields, ","))
	rows, err := db.Query(ctx, query, studyPlanItemIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var ss []*entities.StudentEventLog
	for rows.Next() {
		s := new(entities.StudentEventLog)
		if err := rows.Scan(database.GetScanFields(s, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		ss = append(ss, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return ss, nil
}

type SubmissionResult struct {
	QuestionID string
	SessionID  string
	TimeSpent  int64 // in seconds
	Correct    bool
	CreatedAt  time.Time
}

func (r *StudentEventLogRepo) RetrieveAllSubmitionsOfStudent(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray, loIDs *pgtype.TextArray) (map[string]map[string][]*SubmissionResult, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEventLogRepo.LogsQuestionSubmitionByLO")
	defer span.End()

	args := []interface{}{&studentIDs}
	query := `SELECT student_id, payload->>'question_id' AS question_id, payload->>'correctness' AS correctness,
		payload->>'lo_id' AS lo_id, payload->>'session_id' AS session_id, payload->>'time_spent' as time_spent, created_at
		FROM student_event_logs WHERE student_id = ANY($1) AND event_type = 'quiz_answer_selected'`

	if loIDs != nil {
		args = append(args, &loIDs)
		query += " AND payload->>'lo_id' = ANY($2)"
	}
	query += " ORDER BY created_at ASC"

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	resp := make(map[string]map[string][]*SubmissionResult) // student id => lo id => array of submissions
	for rows.Next() {
		var studentID, questionID, correct, loID, sessionID pgtype.Text
		var createdAt pgtype.Timestamptz
		var timeSpent pgtype.Text
		if err := rows.Scan(&studentID, &questionID, &correct, &loID, &sessionID, &timeSpent, &createdAt); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		if resp[studentID.String] == nil {
			resp[studentID.String] = make(map[string][]*SubmissionResult)
		}

		ts, _ := strconv.ParseInt(timeSpent.String, 10, 64)
		resp[studentID.String][loID.String] = append(resp[studentID.String][loID.String], &SubmissionResult{
			QuestionID: questionID.String,
			SessionID:  sessionID.String,
			Correct:    correct.String == "true",
			TimeSpent:  ts,
			CreatedAt:  createdAt.Time,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return resp, nil
}
