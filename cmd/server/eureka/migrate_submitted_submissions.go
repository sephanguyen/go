package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

func init() {
	bootstrap.RegisterJob("eureka_submitted_submissions", runUpdateSubmittedSubmissions).
		StringVar(&organizationID, "tenantID", "", "specify tenantID").
		Desc("eureka update is_submitted for flashcard submissions, lo submissions")
}

type submittedSubmission struct {
	shuffleQuizSetID   pgtype.Text
	studyPlanID        pgtype.Text
	learningMaterialID pgtype.Text
	studentID          pgtype.Text
}

func findNewSubmittedSubmission(ctx context.Context, db *database.DBTrace, shuffleQuizSetID pgtype.Text, isFlashcard bool) ([]*submittedSubmission, error) {
	query := `select sqs.shuffled_quiz_set_id,sqs.study_plan_id, sqs.learning_material_id , sqs.student_id from shuffled_quiz_sets sqs
	join flash_card_submission ls on sqs.shuffled_quiz_set_id = ls.shuffled_quiz_set_id
	where sqs.session_id = any(select payload ->> 'session_id' from student_event_logs where 
	student_id = ls.student_id and study_plan_id = ls.study_plan_id and learning_material_id = ls.learning_material_id and
	event_type = 'learning_objective' and payload ->> 'event' = 'completed') and ($1::text IS NULL OR sqs.shuffled_quiz_set_id > $1)
	order by shuffled_quiz_set_id ASC
	LIMIT 1000;`
	if !isFlashcard {
		query = `select sqs.shuffled_quiz_set_id,sqs.study_plan_id, sqs.learning_material_id , sqs.student_id from shuffled_quiz_sets sqs
		join lo_submission ls on sqs.shuffled_quiz_set_id = ls.shuffled_quiz_set_id
		where sqs.session_id = any(select payload ->> 'session_id' from student_event_logs where 
		student_id = ls.student_id and study_plan_id = ls.study_plan_id and learning_material_id = ls.learning_material_id and
		event_type = 'learning_objective' and payload ->> 'event' = 'completed') and ($1::text IS NULL OR sqs.shuffled_quiz_set_id > $1)
		order by shuffled_quiz_set_id ASC
		LIMIT 1000;`
	}

	var ss []*submittedSubmission
	rows, err := db.Query(ctx, query, &shuffleQuizSetID)
	if err != nil {
		return nil, fmt.Errorf("failed to find submission: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		submission := new(submittedSubmission)
		err := rows.Scan(&submission.shuffleQuizSetID, &submission.studyPlanID, &submission.learningMaterialID, &submission.studentID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan submission: %w", err)
		}

		ss = append(ss, submission)
	}
	return ss, nil
}

func updateSubmittedSubmissions(ctx context.Context, db *database.DBTrace, ss []*submittedSubmission, isFlashcard bool) error {
	query := `update flash_card_submission set is_submitted = true, updated_at = now()    
	where shuffled_quiz_set_id = $1 
		and learning_material_id = $2
		and study_plan_id = $3
		and student_id = $4`

	if !isFlashcard {
		query = `update lo_submission set is_submitted = true, updated_at = now()    
		where shuffled_quiz_set_id = $1 
			and learning_material_id = $2
			and study_plan_id = $3
			and student_id = $4`
	}

	b := &pgx.Batch{}
	for _, s := range ss {
		b.Queue(query, s.shuffleQuizSetID, s.learningMaterialID, s.studyPlanID, s.studentID)
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

func runUpdateSubmittedSubmissions(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	db := rsc.DBWith("eureka")

	start := time.Now()
	defer func() {
		fmt.Println("Migration complete for max score submission: ", time.Since(start))
	}()
	ctx = auth.InjectFakeJwtToken(ctx, organizationID)

	zapLogger := rsc.Logger()
	zapLogger.Info("====", zap.String("start update submitted submission", time.Now().String()))
	var shuffleQuizSetID pgtype.Text

	// migrate flashcard submission
	zapLogger.Info("====", zap.String("start update submitted flash card submission", time.Now().String()))
	_ = shuffleQuizSetID.Set(nil)
	for {
		submissions, err := findNewSubmittedSubmission(ctx, db, shuffleQuizSetID, true)
		if err != nil {
			return fmt.Errorf("failed to find submission: %w", err)
		}
		zapLogger.Info("====", zap.Int32("fetched submission:", int32(len(submissions))))
		if len(submissions) == 0 {
			break
		}
		err = updateSubmittedSubmissions(ctx, db, submissions, true)
		if err != nil {
			return fmt.Errorf("failed to update max submission: %w", err)
		}
		shuffleQuizSetID = submissions[len(submissions)-1].shuffleQuizSetID
	}

	zapLogger.Info("====", zap.String("update submitted flashcard submission successfully", time.Now().String()))

	// migrate lo submission
	zapLogger.Info("====", zap.String("start update submitted learning objective submission", time.Now().String()))
	_ = shuffleQuizSetID.Set(nil)
	for {
		submissions, err := findNewSubmittedSubmission(ctx, db, shuffleQuizSetID, false)
		if err != nil {
			return fmt.Errorf("failed to find submission: %w", err)
		}
		zapLogger.Info("====", zap.Int32("fetched submission:", int32(len(submissions))))
		if len(submissions) == 0 {
			break
		}
		err = updateSubmittedSubmissions(ctx, db, submissions, false)
		if err != nil {
			return fmt.Errorf("failed to update submitted submission: %w", err)
		}
		shuffleQuizSetID = submissions[len(submissions)-1].shuffleQuizSetID
	}

	zapLogger.Info("====", zap.String("update submitted lo submission successfully", time.Now().String()))

	return nil
}
