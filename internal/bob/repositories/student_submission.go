package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

// StudentSubmissionRepo work with student_submissions table
type StudentSubmissionRepo struct{}

// Create creates StudentSubmission with generated ID
func (r *StudentSubmissionRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.StudentSubmission) error {
	now := timeutil.Now()
	err := multierr.Combine(
		e.ID.Set(idutil.ULID(now)),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	)
	if err != nil {
		return err
	}

	cmdTag, err := database.Insert(ctx, e, db.Exec)
	if err == nil && cmdTag.RowsAffected() != 1 {
		return ErrUnAffected
	}

	return err
}

// StudentSubmissionFilter used in StudentSubmissionRepo.List method
type StudentSubmissionFilter struct {
	StudentIDs pgtype.TextArray
	TopicIDs   pgtype.TextArray
	Order      string
	Limit      int
	OffsetID   pgtype.Text
}

// List returns result with order DESC by default
func (r *StudentSubmissionRepo) List(ctx context.Context, db database.QueryExecer, filter *StudentSubmissionFilter) ([]*entities.StudentSubmission, error) {
	args := []interface{}{&filter.StudentIDs}
	stmt := "SELECT %s FROM %s WHERE student_id = ANY($1)"

	args = append(args, &filter.TopicIDs)
	stmt += fmt.Sprintf(" AND topic_id = ANY($%d)", len(args))

	if filter.OffsetID.Status == pgtype.Present {
		args = append(args, &filter.OffsetID)
		stmt += fmt.Sprintf(" AND student_submission_id > $%d", len(args))
	}

	if filter.Order == "ASC" {
		stmt += " ORDER BY student_submission_id ASC"
	} else {
		stmt += " ORDER BY student_submission_id DESC"
	}

	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		stmt += fmt.Sprintf(" LIMIT $%d", len(args))
	}

	e := &entities.StudentSubmission{}
	fieldNames := database.GetFieldNames(e)

	results := make([]*entities.StudentSubmission, 0, filter.Limit)
	rows, err := db.Query(ctx,
		fmt.Sprintf(stmt, strings.Join(fieldNames, ","), e.TableName()),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		submission := entities.StudentSubmission{}
		if err := rows.Scan(database.GetScanFields(&submission, fieldNames)...); err != nil {
			return nil, err
		}

		results = append(results, &submission)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *StudentSubmissionRepo) CountSubmissions(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray, topicIDs *pgtype.TextArray) (int, error) {
	args := []interface{}{&studentIDs}
	sub := fmt.Sprintf("SELECT student_id, topic_id FROM %s WHERE student_id = ANY($1)", (&entities.StudentSubmission{}).TableName())

	if topicIDs != nil {
		args = append(args, topicIDs)
		sub += " AND topic_id = ANY($2)"
	}
	sub += " GROUP BY student_id, topic_id"

	query := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS sub", sub)

	var count int
	if err := db.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

type SubmissionScore struct {
	Submission *entities.StudentSubmission
	GivenScore *pgtype.Numeric
	TotalScore *pgtype.Numeric
}

func (r *StudentSubmissionRepo) ListLatestScore(ctx context.Context, db database.QueryExecer, filter *StudentSubmissionFilter) ([]*SubmissionScore, error) {
	args := []interface{}{&filter.StudentIDs}
	stmt := `SELECT student_submissions.%s, given_score, total_score
		FROM %s
		INNER JOIN %s ON %[2]s.student_submission_id = %[3]s.student_submission_id
		WHERE student_id = ANY($1)`

	args = append(args, &filter.TopicIDs)
	stmt += " AND topic_id = ANY($2)"
	stmt += " ORDER BY student_submission_scores.student_submission_score_id DESC"

	e := &entities.StudentSubmission{}
	fields := database.GetFieldNames(e)

	rows, err := db.Query(ctx,
		fmt.Sprintf(stmt, strings.Join(fields, ",student_submissions."), e.TableName(), (&entities.StudentSubmissionScore{}).TableName()),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*SubmissionScore
	for rows.Next() {
		ss := &SubmissionScore{
			Submission: new(entities.StudentSubmission),
			GivenScore: new(pgtype.Numeric),
			TotalScore: new(pgtype.Numeric),
		}
		sc := database.GetScanFields(ss.Submission, fields)
		sc = append(sc, ss.GivenScore, ss.TotalScore)
		if err := rows.Scan(sc...); err != nil {
			return nil, err
		}

		results = append(results, ss)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *StudentSubmissionRepo) FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.StudentSubmission, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubmissionRepo.FindByID")
	defer span.End()

	e := new(entities.StudentSubmission)
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE student_submission_id = $1", strings.Join(fields, ","), e.TableName())

	row := db.QueryRow(ctx, selectStmt, &id)
	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, err
	}

	return e, nil
}

// StudentSubmissionScoreRepo work with student_submission_scores table
type StudentSubmissionScoreRepo struct{}

// Create creates StudentSubmissionScore
func (r *StudentSubmissionScoreRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.StudentSubmissionScore) error {
	now := timeutil.Now()
	err := multierr.Combine(
		e.ID.Set(idutil.ULID(now)),
		e.CreatedAt.Set(now),
	)
	if err != nil {
		return err
	}

	cmdTag, err := database.Insert(ctx, e, db.Exec)
	if err == nil && cmdTag.RowsAffected() != 1 {
		return ErrUnAffected
	}

	return err
}

// StudentSubmissionScoreFilter used in StudentSubmissionScoreRepo.List method
type StudentSubmissionScoreFilter struct {
	SubmissionIDs pgtype.TextArray
	Order         string
	Limit         int
	OffsetID      pgtype.Text
}

// List submissions scores
func (r *StudentSubmissionScoreRepo) List(ctx context.Context, db database.QueryExecer, filter *StudentSubmissionScoreFilter) ([]*entities.StudentSubmissionScore, error) {
	args := []interface{}{&filter.SubmissionIDs}
	stmt := "SELECT %s FROM %s WHERE student_submission_id = ANY($1)"

	if filter.OffsetID.Status == pgtype.Present {
		args = append(args, &filter.OffsetID)
		stmt += fmt.Sprintf(" AND student_submission_score_id > $%d", len(args))
	}

	if filter.Order == "ASC" {
		stmt += " ORDER BY student_submission_score_id ASC"
	} else {
		stmt += " ORDER BY student_submission_score_id DESC"
	}

	args = append(args, filter.Limit)
	stmt += fmt.Sprintf(" LIMIT $%d", len(args))

	e := &entities.StudentSubmissionScore{}
	fieldNames := database.GetFieldNames(e)

	results := make([]*entities.StudentSubmissionScore, 0, filter.Limit)
	rows, err := db.Query(ctx,
		fmt.Sprintf(stmt, strings.Join(fieldNames, ","), e.TableName()),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		score := entities.StudentSubmissionScore{}
		if err := rows.Scan(database.GetScanFields(&score, fieldNames)...); err != nil {
			return nil, err
		}

		results = append(results, &score)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
