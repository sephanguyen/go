package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

type StudentLearningTimeDailyRepo struct{}

func (r *StudentLearningTimeDailyRepo) Retrieve(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...QueryEnhancer) ([]*entities.StudentLearningTimeDaily, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentLearningTimeDailyRepo.Retrieve")
	defer span.End()

	fields := database.GetFieldNames(&entities.StudentLearningTimeDaily{})
	args := []interface{}{&studentID}
	query := fmt.Sprintf("SELECT %s FROM student_learning_time_by_daily WHERE student_id = $1::TEXT", strings.Join(fields, ","))
	if from != nil {
		args = append(args, from)
		query += fmt.Sprintf(" AND day >= $%d", len(args))
	}
	if to != nil {
		args = append(args, to)
		query += fmt.Sprintf(" AND day <= $%d", len(args))
	}
	query += " ORDER BY day ASC"
	for _, ehc := range queryEnhancers {
		ehc(&query)
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var ss []*entities.StudentLearningTimeDaily
	for rows.Next() {
		s := new(entities.StudentLearningTimeDaily)
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

type StudentLearningTimeDailyV2 struct {
	StudentID               pgtype.Text
	LearningTime            pgtype.Int4
	Day                     pgtype.Timestamptz
	Sessions                pgtype.Text
	AssignmentLearningTime  pgtype.Int4
	AssignmentSubmissionIDs pgtype.Text
}

func (r *StudentLearningTimeDailyRepo) RetrieveV2(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...QueryEnhancer) ([]*StudentLearningTimeDailyV2, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentLearningTimeDailyRepo.RetrieveV2")
	defer span.End()

	args := []interface{}{
		studentID,
	}
	query := `
	SELECT
		(learning_time_by_minutes)*60 as learning_time,
		student_id,
		day,
		assignment_duration*60 as assignment_learning_time,
		submit_learning_material_id::text[] as assignment_submission_ids
	FROM public.calculate_learning_time($1::TEXT)
	WHERE 1=1`
	if from != nil {
		args = append(args, from)
		query += fmt.Sprintf(" AND day >= $%d", len(args))
	}
	if to != nil {
		args = append(args, to)
		query += fmt.Sprintf(" AND day <= $%d", len(args))
	}
	query += " ORDER BY day ASC"
	for _, ehc := range queryEnhancers {
		ehc(&query)
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var ss []*StudentLearningTimeDailyV2
	for rows.Next() {
		s := new(StudentLearningTimeDailyV2)
		if err := rows.Scan(&s.LearningTime, &s.StudentID, &s.Day, &s.AssignmentLearningTime, &s.AssignmentSubmissionIDs); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		ss = append(ss, s)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return ss, nil
}

func (r *StudentLearningTimeDailyRepo) Upsert(ctx context.Context, db database.QueryExecer, s *entities.StudentLearningTimeDaily) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentLearningTimeDailyRepo.Upsert")
	defer span.End()

	now := timeutil.Now()
	s.CreatedAt.Set(now)
	s.UpdatedAt.Set(now)

	fieldNames := database.GetFieldNamesExcepts(s, []string{"learning_time_id", "assignment_learning_time", "assignment_submission_ids"})
	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT student_learning_time_by_daily_un
		DO UPDATE SET
			learning_time = student_learning_time_by_daily.learning_time + $2,
			sessions = $4,
			updated_at = $6;
		`,
		s.TableName(),
		strings.Join(fieldNames, ","),
		database.GeneratePlaceholders(len(fieldNames)),
	)
	args := database.GetScanFields(s, fieldNames)
	if _, err := db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	return nil
}

func (r *StudentLearningTimeDailyRepo) UpsertTaskAssignment(ctx context.Context, db database.QueryExecer, studentLearningTimeDaily *entities.StudentLearningTimeDaily) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentLearningTimeDailyRepo.UpsertTaskAssignment")
	defer span.End()

	const upsertTaskAssignmentLearningTimeTmpl = `
INSERT INTO %s
AS sltbd (%s) 
VALUES (%s)
ON CONFLICT ON CONSTRAINT student_learning_time_by_daily_un 
DO UPDATE SET
  learning_time = sltbd.learning_time + $2, 
  updated_at = $6,
  assignment_learning_time = sltbd.assignment_learning_time + $7,
  assignment_submission_ids = $8::_TEXT;
`

	now := timeutil.Now()
	studentLearningTimeDaily.Sessions.Set(nil)
	studentLearningTimeDaily.CreatedAt.Set(now)
	studentLearningTimeDaily.UpdatedAt.Set(now)

	// skip ID field
	fieldNames := database.GetFieldNames(studentLearningTimeDaily)[1:]
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	args := database.GetScanFields(studentLearningTimeDaily, fieldNames)

	query := fmt.Sprintf(upsertTaskAssignmentLearningTimeTmpl,
		studentLearningTimeDaily.TableName(),
		strings.Join(fieldNames, ","),
		placeHolders)

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("StudentLearningTimeDailyRepo.UpsertTaskAssignment %w", err)
	}

	return nil
}
