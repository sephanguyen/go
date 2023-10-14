package eureka

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var organizationID string

func init() {
	bootstrap.RegisterJob("eureka_migrate_max_score_submission", runUpdateMaxScoreSubmissionForLO).
		StringVar(&organizationID, "tenantID", "", "specify tenantID").
		Desc("eureka update_student_event_logs study_plan_item_identity")
}

type loScore struct {
	StudentID          pgtype.Text
	StudyPlanID        pgtype.Text
	LearningMaterialID pgtype.Text
	SubmissionID       pgtype.Text
	GradedPoint        pgtype.Int2
	TotalPoint         pgtype.Int2
}

func findLOScore(ctx context.Context, db *database.DBTrace, submissionID pgtype.Text) ([]*loScore, error) {
	query := `SELECT student_id, study_plan_id, learning_material_id, submission_id, graded_point, total_point
	FROM lo_graded_score_v2()
	WHERE ($1::text IS NULL OR submission_id > $1)
	ORDER BY submission_id ASC
	LIMIT 1000`
	var loScores []*loScore
	rows, err := db.Query(ctx, query, &submissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find lo submission: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		score := new(loScore)
		err := rows.Scan(&score.StudentID, &score.StudyPlanID, &score.LearningMaterialID, &score.SubmissionID, &score.GradedPoint, &score.TotalPoint)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lo submission: %w", err)
		}
		loScores = append(loScores, score)
	}
	return loScores, nil
}

func findTaskAssignmentScore(ctx context.Context, db *database.DBTrace, submissionID pgtype.Text) ([]*loScore, error) {
	query := `SELECT student_id, study_plan_id, learning_material_id, student_submission_id, graded_point, total_point
	FROM task_assignment_graded_score_v2()
	WHERE ($1::text IS NULL OR student_submission_id > $1)
	ORDER BY student_submission_id ASC
	LIMIT 1000`
	var loScores []*loScore
	rows, err := db.Query(ctx, query, &submissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task assignment submission: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		score := new(loScore)
		err := rows.Scan(&score.StudentID, &score.StudyPlanID, &score.LearningMaterialID, &score.SubmissionID, &score.GradedPoint, &score.TotalPoint)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task assignment submission: %w", err)
		}
		loScores = append(loScores, score)
	}
	return loScores, nil
}

func findAssignmentScore(ctx context.Context, db *database.DBTrace, submissionID pgtype.Text) ([]*loScore, error) {
	query := `SELECT student_id, study_plan_id, learning_material_id, student_submission_id, graded_point, total_point
	FROM assignment_graded_score_v2()
	WHERE ($1::text IS NULL OR student_submission_id > $1)
	ORDER BY student_submission_id ASC
	LIMIT 1000`
	var loScores []*loScore
	rows, err := db.Query(ctx, query, &submissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find assignment submission: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		score := new(loScore)
		err := rows.Scan(&score.StudentID, &score.StudyPlanID, &score.LearningMaterialID, &score.SubmissionID, &score.GradedPoint, &score.TotalPoint)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment submission: %w", err)
		}
		loScores = append(loScores, score)
	}
	return loScores, nil
}

func findGradedScore(ctx context.Context, db *database.DBTrace, submissionID pgtype.Text) ([]*loScore, error) {
	query := `SELECT student_id, study_plan_id, learning_material_id, submission_id, graded_point, total_point
	FROM fc_graded_score_v2()
	WHERE ($1::text IS NULL OR submission_id > $1)
	ORDER BY submission_id ASC
	LIMIT 1000`
	var loScores []*loScore
	rows, err := db.Query(ctx, query, &submissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find graded submission: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		score := new(loScore)
		err := rows.Scan(&score.StudentID, &score.StudyPlanID, &score.LearningMaterialID, &score.SubmissionID, &score.GradedPoint, &score.TotalPoint)
		if err != nil {
			return nil, fmt.Errorf("failed to scan graded submission: %w", err)
		}
		loScores = append(loScores, score)
	}
	return loScores, nil
}

func findExamLOSubmission(ctx context.Context, db *database.DBTrace, submissionID pgtype.Text) ([]*loScore, error) {
	query := `SELECT student_id, study_plan_id, learning_material_id, submission_id, graded_point, total_point
	FROM exam_lo_graded_score_v2()
	WHERE ($1::text IS NULL OR submission_id > $1)
	ORDER BY submission_id ASC
	LIMIT 1000`
	var loScores []*loScore
	rows, err := db.Query(ctx, query, &submissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find submission: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		score := new(loScore)
		err := rows.Scan(&score.StudentID, &score.StudyPlanID, &score.LearningMaterialID, &score.SubmissionID, &score.GradedPoint, &score.TotalPoint)
		if err != nil {
			return nil, fmt.Errorf("failed to scan submission: %w", err)
		}
		loScores = append(loScores, score)
	}
	return loScores, nil
}

func updateMaxSubmission(ctx context.Context, db *database.DBTrace, loSubmissionScores []*loScore) error {
	query := `INSERT INTO max_score_submission (max_score, updated_at, total_score, max_percentage, study_plan_id, learning_material_id, student_id, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, now())
	ON CONFLICT (study_plan_id, learning_material_id, student_id)
	DO UPDATE SET max_score = $1,
	updated_at = $2,
	total_score = $3,
	max_percentage = $4
	WHERE max_score_submission.max_percentage < $4;`

	queueFn := func(b *pgx.Batch, loSubmissionScore *loScore) error {
		var gradedPoint, totalPoint pgtype.Int4
		err := multierr.Combine(
			gradedPoint.Set(int32(loSubmissionScore.GradedPoint.Int)),
			totalPoint.Set(int32(loSubmissionScore.TotalPoint.Int)),
		)
		if err != nil {
			return fmt.Errorf("failed to set pgtype: %w", err)
		}
		// handle totalscore > grade score which is invalid. So we set to 100%
		if loSubmissionScore.TotalPoint.Int < loSubmissionScore.GradedPoint.Int {
			loSubmissionScore.TotalPoint.Int = loSubmissionScore.GradedPoint.Int
		}
		// handle total score = 0 which is invalid. So we set to 1
		if loSubmissionScore.TotalPoint.Int == 0 {
			loSubmissionScore.TotalPoint.Int = 1
		}
		maxPercent := math.Round(float64(loSubmissionScore.GradedPoint.Int/loSubmissionScore.TotalPoint.Int)) * 100
		b.Queue(query, gradedPoint,
			time.Now(), totalPoint, maxPercent,
			loSubmissionScore.StudyPlanID, loSubmissionScore.LearningMaterialID, loSubmissionScore.StudentID)
		return nil
	}
	b := &pgx.Batch{}
	for _, loSubmission := range loSubmissionScores {
		s := loSubmission
		err := queueFn(b, s)
		if err != nil {
			return fmt.Errorf("queueFn: %w", err)
		}
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

func migrateSubmission(ctx context.Context, db *database.DBTrace, zapLogger *zap.Logger,
	findData func(ctx context.Context, db *database.DBTrace, submissionID pgtype.Text) ([]*loScore, error)) error {
	var submissionID pgtype.Text
	_ = submissionID.Set(nil)

	for {
		submissions, err := findData(ctx, db, submissionID)
		if err != nil {
			return fmt.Errorf("failed to find submission: %w", err)
		}
		zapLogger.Info("====", zap.Int32("fetched submission:", int32(len(submissions))))
		if len(submissions) == 0 {
			break
		}
		err = updateMaxSubmission(ctx, db, submissions)
		if err != nil {
			return fmt.Errorf("failed to update max submission: %w", err)
		}
		submissionID = submissions[len(submissions)-1].SubmissionID
	}
	return nil
}

func runUpdateMaxScoreSubmissionForLO(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	db := rsc.DBWith("eureka")

	start := time.Now()
	defer func() {
		fmt.Println("Migration complete for max score submission: ", time.Since(start))
	}()
	ctx = auth.InjectFakeJwtToken(ctx, organizationID)

	zapLogger := rsc.Logger()
	// migrate lo submission
	zapLogger.Info("====", zap.String("start migrate LO submission", time.Now().String()))
	err := migrateSubmission(ctx, db, zapLogger, findLOScore)
	if err != nil {
		return fmt.Errorf("failed to migrate lo submission: %w", err)
	}
	// migrate task assignment submission
	zapLogger.Info("====", zap.String("start migrate task assignment submission", time.Now().String()))
	err = migrateSubmission(ctx, db, zapLogger, findTaskAssignmentScore)
	if err != nil {
		return fmt.Errorf("failed to migrate task LO submission: %w", err)
	}
	// migrate assignment submission
	zapLogger.Info("====", zap.String("start migrate assignment submission", time.Now().String()))
	err = migrateSubmission(ctx, db, zapLogger, findAssignmentScore)
	if err != nil {
		return fmt.Errorf("failed to migrate assignment submission: %w", err)
	}
	// migrate exam lo submission
	zapLogger.Info("====", zap.String("start migrate exam LO submission", time.Now().String()))
	err = migrateSubmission(ctx, db, zapLogger, findExamLOSubmission)
	if err != nil {
		return fmt.Errorf("failed to migrate exam lo submission: %w", err)
	}

	// migrate exam lo submission
	zapLogger.Info("====", zap.String("start migrate graded score submission", time.Now().String()))
	err = migrateSubmission(ctx, db, zapLogger, findGradedScore)
	if err != nil {
		return fmt.Errorf("failed to migrate graded score submission: %w", err)
	}
	return nil
}
