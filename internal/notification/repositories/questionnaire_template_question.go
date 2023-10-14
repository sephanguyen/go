package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type QuestionnaireTemplateQuestionRepo struct{}

func (repo *QuestionnaireTemplateQuestionRepo) queueForceUpsert(b *pgx.Batch, item *entities.QuestionnaireTemplateQuestion) error {
	now := time.Now()
	err := multierr.Combine(
		item.CreatedAt.Set(now),
		item.UpdatedAt.Set(now),
		item.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	if item.QuestionnaireTemplateQuestionID.String == "" {
		_ = item.QuestionnaireTemplateQuestionID.Set(idutil.ULIDNow())
	}

	fieldNames := database.GetFieldNames(item)
	values := database.GetScanFields(item, fieldNames)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	tableName := item.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s as qnq (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT pk__questionnaire_template_questions
		DO UPDATE SET
			questionnaire_template_id=EXCLUDED.questionnaire_template_id,
			order_index = EXCLUDED.order_index,
			type = EXCLUDED.type,
			title = EXCLUDED.title,
			choices = EXCLUDED.choices,
			is_required = EXCLUDED.is_required,
			updated_at = EXCLUDED.updated_at,
			deleted_at = NULL;
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)

	return nil
}

func (repo *QuestionnaireTemplateQuestionRepo) BulkForceUpsert(ctx context.Context, db database.QueryExecer, items entities.QuestionnaireTemplateQuestions) error {
	b := &pgx.Batch{}
	for _, item := range items {
		err := repo.queueForceUpsert(b, item)
		if err != nil {
			return fmt.Errorf("repo.queueForceUpsert: %w", err)
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

func (repo *QuestionnaireTemplateQuestionRepo) SoftDelete(ctx context.Context, db database.QueryExecer, questionnaireTemplateID []string) error {
	pgIDs := database.TextArray(questionnaireTemplateID)

	query := `
		UPDATE questionnaire_template_questions as qtq
		SET deleted_at = now(), 
			updated_at = now() 
		WHERE questionnaire_template_id = ANY($1) 
		AND qtq.deleted_at IS NULL
	`

	_, err := db.Exec(ctx, query, &pgIDs)
	if err != nil {
		return fmt.Errorf("repo.SoftDelete: %w", err)
	}

	return nil
}
