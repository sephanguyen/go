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
	bootstrap.RegisterJob("eureka_lo_submission_total_score", runUpdateSubmissionTotalScore).
		StringVar(&organizationID, "tenantID", "", "specify tenantID").
		Desc("eureka update lo submission total score")
}

type submissionScore struct {
	submissionID pgtype.Text
	totalScore   pgtype.Int4
}

func findNewTotalPoint(ctx context.Context, db *database.DBTrace, submissionID pgtype.Text) ([]*submissionScore, error) {
	query := `SELECT ls.submission_id, coalesce(
			(
				SELECT sum(point) 
				FROM quizzes q WHERE q.external_id = ANY(qs.quiz_external_ids) AND q.deleted_at IS NULL
			),total_point
		) AS new_total
	FROM lo_submission ls join shuffled_quiz_sets sqs 
	ON ls.shuffled_quiz_set_id = sqs.shuffled_quiz_set_id 
	JOIN quiz_sets qs on qs.quiz_set_id = sqs.original_quiz_set_id
	WHERE ($1::text IS NULL OR ls.submission_id > $1)
	order by ls.submission_id ASC
	LIMIT 1000;`
	var ss []*submissionScore
	rows, err := db.Query(ctx, query, &submissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find lo submission: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var totalScore int
		score := new(submissionScore)
		err := rows.Scan(&score.submissionID, &totalScore)
		_ = score.totalScore.Set(totalScore)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lo submission: %w", err)
		}
		ss = append(ss, score)
	}
	return ss, nil
}

func updateSubmissionNewPoint(ctx context.Context, db *database.DBTrace, ss []*submissionScore) error {
	query := `UPDATE lo_submission SET total_point = $1 WHERE submission_id = $2`

	b := &pgx.Batch{}
	for _, s := range ss {
		b.Queue(query, s.totalScore, s.submissionID)
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

func runUpdateSubmissionTotalScore(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	db := rsc.DBWith("eureka")

	start := time.Now()
	defer func() {
		fmt.Println("Migration complete for max score submission: ", time.Since(start))
	}()
	ctx = auth.InjectFakeJwtToken(ctx, organizationID)

	zapLogger := rsc.Logger()
	// migrate lo submission
	zapLogger.Info("====", zap.String("start migrate LO submission", time.Now().String()))
	var submissionID pgtype.Text
	_ = submissionID.Set(nil)

	for {
		submissions, err := findNewTotalPoint(ctx, db, submissionID)
		if err != nil {
			return fmt.Errorf("failed to find submission: %w", err)
		}
		zapLogger.Info("====", zap.Int32("fetched submission:", int32(len(submissions))))
		if len(submissions) == 0 {
			break
		}
		err = updateSubmissionNewPoint(ctx, db, submissions)
		if err != nil {
			return fmt.Errorf("failed to update max submission: %w", err)
		}
		submissionID = submissions[len(submissions)-1].submissionID
	}
	return nil
}
