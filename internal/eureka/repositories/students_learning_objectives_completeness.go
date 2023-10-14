package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type StudentsLearningObjectivesCompletenessRepo struct{}

var queryFindLos = `
SELECT DISTINCT
	sloc.student_id,
	sloc.lo_id, 
	sloc.preset_study_plan_weekly_id,
	sloc.first_attempt_score, 
	sloc.is_finished_quiz,
	sloc.is_finished_video,
	sloc.is_finished_study_guide,
	sloc.first_quiz_correctness,
	sloc.finished_quiz_at,
	coalesce((select max(gelrs.graded_point::float4 / gelrs.total_point) *100 from get_exam_lo_returned_scores() gelrs where sloc.student_id = gelrs.student_id AND sloc.lo_id = gelrs.learning_material_id) , sloc.highest_quiz_score)::float4 as highest_quiz_score,
	sloc.updated_at,
	sloc.created_at
FROM %s sloc 
LEFT JOIN %s lo ON sloc.lo_id = lo.lo_id
LEFT JOIN get_exam_lo_returned_scores() gelrs ON sloc.student_id = gelrs.student_id AND sloc.lo_id = gelrs.learning_material_id
WHERE sloc.student_id = $1
AND sloc.lo_id = ANY($2) 
AND lo.deleted_at IS NULL`

func (r *StudentsLearningObjectivesCompletenessRepo) Find(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, loIds pgtype.TextArray) (map[pgtype.Text]*entities.StudentsLearningObjectivesCompleteness, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentsLearningObjectivesCompletenessRepo.Find")
	defer span.End()

	lo := &entities.StudentsLearningObjectivesCompleteness{}
	loE := &entities.LearningObjective{}
	fields := database.GetFieldNames(lo)
	query := fmt.Sprintf(queryFindLos, lo.TableName(), loE.TableName())
	rows, err := db.Query(ctx, query, &studentID, &loIds)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	pp := make(map[pgtype.Text]*entities.StudentsLearningObjectivesCompleteness)
	for rows.Next() {
		p := new(entities.StudentsLearningObjectivesCompleteness)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp[p.LoID] = p
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (r *StudentsLearningObjectivesCompletenessRepo) RetrieveFinishedLOs(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*entities.StudentsLearningObjectivesCompleteness, error) {
	sloc := &entities.StudentsLearningObjectivesCompleteness{}
	lo := &entities.LearningObjective{}
	fields := database.GetFieldNames(sloc)

	query := fmt.Sprintf("SELECT DISTINCT sloc.%s FROM %s sloc LEFT JOIN %s lo ON sloc.lo_id = lo.lo_id WHERE sloc.student_id = $1::TEXT AND sloc.is_finished_quiz IS TRUE AND lo.deleted_at IS NULL", strings.Join(fields, ", sloc."), sloc.TableName(), lo.TableName())
	rows, err := db.Query(ctx, query, &studentID)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var los []*entities.StudentsLearningObjectivesCompleteness
	for rows.Next() {
		lo := &entities.StudentsLearningObjectivesCompleteness{}
		if err := rows.Scan(database.GetScanFields(lo, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		los = append(los, lo)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return los, nil
}

var upsertHighestQuizScoreStmt string = `INSERT INTO %s (lo_id, student_id, highest_quiz_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT ON CONSTRAINT students_learning_objectives_completeness_pk
		DO UPDATE SET updated_at = excluded.updated_at, highest_quiz_score = excluded.highest_quiz_score
		WHERE %s.highest_quiz_score IS NULL OR %s.highest_quiz_score < excluded.highest_quiz_score;
	`

var upsertFirstQuizCorrectnessStmt string = `INSERT INTO %s (lo_id, student_id, first_quiz_correctness, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT ON CONSTRAINT students_learning_objectives_completeness_pk
		DO UPDATE SET updated_at = excluded.updated_at, first_quiz_correctness = excluded.first_quiz_correctness
		WHERE %s.first_quiz_correctness IS NULL;`

// UpsertHighestQuizScore update highest quiz score
func (r *StudentsLearningObjectivesCompletenessRepo) UpsertHighestQuizScore(ctx context.Context, db database.QueryExecer, loID pgtype.Text, studentID pgtype.Text, newScore pgtype.Float4) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentsLearningObjectivesCompletenessRepo.UpdateHighestQuizScore")
	defer span.End()

	createdAt := time.Now()
	updatedAt := time.Now()
	sloc := &entities.StudentsLearningObjectivesCompleteness{}
	query := fmt.Sprintf(upsertHighestQuizScoreStmt, sloc.TableName(), sloc.TableName(), sloc.TableName())

	_, err := db.Exec(ctx, query, loID, studentID, newScore, createdAt, updatedAt)
	if err != nil {
		return err
	}
	return nil
}

// UpsertFirstQuizCompleteness insert new learning objective completeness
func (r *StudentsLearningObjectivesCompletenessRepo) UpsertFirstQuizCompleteness(ctx context.Context, db database.QueryExecer, loID pgtype.Text, studentID pgtype.Text, firstQuizScore pgtype.Float4) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentsLearningObjectivesCompletenessRepo.UpsertFirstQuizCompleteness")
	defer span.End()

	createdAt := time.Now()
	updatedAt := time.Now()
	sloc := &entities.StudentsLearningObjectivesCompleteness{}
	query := fmt.Sprintf(upsertFirstQuizCorrectnessStmt, sloc.TableName(), sloc.TableName())

	_, err := db.Exec(ctx, query, loID, studentID, firstQuizScore, createdAt, updatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *StudentsLearningObjectivesCompletenessRepo) UpsertLOCompleteness(ctx context.Context, db database.QueryExecer, ss []*entities.StudentsLearningObjectivesCompleteness) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentsLearningObjectivesCompletenessRepo.UpsertLOCompleteness")
	defer span.End()

	now := time.Now().UTC()

	queueFn := func(b *pgx.Batch, s *entities.StudentsLearningObjectivesCompleteness) {
		s.UpdatedAt.Set(now)
		s.CreatedAt.Set(now)

		fieldNames := []string{"student_id", "lo_id", "created_at", "updated_at"}
		if s.FirstAttemptScore.Status == pgtype.Present {
			fieldNames = append(fieldNames, "first_attempt_score")
		}

		var updateQuery string

		// only update 1 completeness type at 1 time
		switch {
		case s.IsFinishedQuiz.Status != pgtype.Undefined:
			fieldNames = append(fieldNames, "is_finished_quiz", "first_quiz_correctness", "finished_quiz_at")
			updateQuery = fmt.Sprintf(`is_finished_quiz = $%d, first_quiz_correctness = $%d, finished_quiz_at = $%d
				WHERE students_learning_objectives_completeness.is_finished_quiz = FALSE`, len(fieldNames)-2, len(fieldNames)-1, len(fieldNames))

		case s.HighestQuizScore.Status != pgtype.Undefined:
			fieldNames = append(fieldNames, "highest_quiz_score")
			updateQuery = fmt.Sprintf(`highest_quiz_score = $%d
				WHERE students_learning_objectives_completeness.highest_quiz_score IS NULL
				OR students_learning_objectives_completeness.highest_quiz_score < excluded.highest_quiz_score`, len(fieldNames))

		case s.IsFinishedVideo.Status != pgtype.Undefined:
			fieldNames = append(fieldNames, "is_finished_video")
			updateQuery = fmt.Sprintf("is_finished_video = $%d", len(fieldNames))

		case s.IsFinishedStudyGuide.Status != pgtype.Undefined:
			fieldNames = append(fieldNames, "is_finished_study_guide")
			updateQuery = fmt.Sprintf("is_finished_study_guide = $%d", len(fieldNames))

		case s.PresetStudyPlanWeeklyID.Status != pgtype.Undefined && s.PresetStudyPlanWeeklyID.Status != pgtype.Null:
			fieldNames = append(fieldNames, "preset_study_plan_weekly_id")
			updateQuery = fmt.Sprintf("preset_study_plan_weekly_id = $%d", len(fieldNames))
		}

		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", s.TableName(), strings.Join(fieldNames, ","), placeHolders)
		query += " ON CONFLICT ON CONSTRAINT students_learning_objectives_completeness_pk DO UPDATE SET updated_at = $4"
		if updateQuery != "" {
			query += fmt.Sprintf(", %s", updateQuery)
		}

		b.Queue(query, database.GetScanFields(s, fieldNames)...)
	}

	b := &pgx.Batch{}
	for _, s := range ss {
		queueFn(b, s)
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
