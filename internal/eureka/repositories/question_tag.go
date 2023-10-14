package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	dbeureka "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type QuestionTagRepo struct{}

const upsertQuestionTagQuery = `INSERT INTO %s (%s) VALUES %s 
ON CONFLICT ON CONSTRAINT question_tag_id_pk DO 
UPDATE SET name = excluded.name, question_tag_type_id = excluded.question_tag_type_id, updated_at = NOW();`

func (l *QuestionTagRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, e []*entities.QuestionTag) error {
	ctx, span := interceptors.StartSpan(ctx, "QuestionTagRepo.Upsert")
	defer span.End()

	err := dbeureka.BulkUpsert(ctx, db, upsertQuestionTagQuery, e)
	if err != nil {
		return fmt.Errorf("QuestionTagRepo.BulkUpsert error: %s", err.Error())
	}
	return nil
}

type GetPointPerTagBySubmissionIDData struct {
	QuestionTagID   string
	QuestionTagName string
	GradedPoint     int
	TotalPoint      int
}

func (l *QuestionTagRepo) GetPointPerTagBySubmissionID(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) ([]GetPointPerTagBySubmissionIDData, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionTagRepo.GetPointPerTagBySubmissionID")
	defer span.End()

	query := `SELECT qt.question_tag_id,
					   qt.name,
					   SUM(coalesce(elss.point, elsa.point)) AS graded_point,
					   SUM(q.point)                          AS total_point
				FROM exam_lo_submission_answer elsa
						 LEFT JOIN exam_lo_submission_score elss USING (submission_id, quiz_id)
						 JOIN quizzes q ON q.external_id = elsa.quiz_id
						 JOIN question_tag qt ON qt.question_tag_id = ANY (q.question_tag_ids)
				WHERE elsa.submission_id = $1 AND q.deleted_at IS NULL
				GROUP BY qt.question_tag_id`

	rows, err := db.Query(ctx, query, submissionID)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %s", err)
	}
	defer rows.Close()

	result := make([]GetPointPerTagBySubmissionIDData, 0)
	for rows.Next() {
		var data GetPointPerTagBySubmissionIDData
		err = rows.Scan(&data.QuestionTagID, &data.QuestionTagName, &data.GradedPoint, &data.TotalPoint)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %s", err)
		}
		result = append(result, data)
	}
	return result, nil
}
