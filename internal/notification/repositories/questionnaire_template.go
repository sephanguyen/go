package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type QuestionnaireTemplateRepo struct{}

func (repo *QuestionnaireTemplateRepo) Upsert(ctx context.Context, db database.QueryExecer, questionnaireTemplate *entities.QuestionnaireTemplate) error {
	now := time.Now()
	err := multierr.Combine(
		questionnaireTemplate.CreatedAt.Set(now),
		questionnaireTemplate.UpdatedAt.Set(now),
		questionnaireTemplate.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	fields := database.GetFieldNames(questionnaireTemplate)
	values := database.GetScanFields(questionnaireTemplate, fields)
	placeHolders := database.GeneratePlaceholders(len(fields))
	tableName := questionnaireTemplate.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s as qt (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT pk__questionnaire_templates 
		DO UPDATE SET 
			name=EXCLUDED.name,
			resubmit_allowed=EXCLUDED.resubmit_allowed,
			expiration_date=EXCLUDED.expiration_date,
			type=EXCLUDED.type,
			updated_at=EXCLUDED.updated_at
		WHERE qt.deleted_at IS NULL;
	`, tableName, strings.Join(fields, ", "), placeHolders)

	cmd, err := db.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("QuestionnaireTemplateRepo.Upsert: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("QuestionnaireTemplateRepo.Upsert: Questionnaire template is not inserted")
	}

	return nil
}

type CheckTemplateNameFilter struct {
	TemplateID pgtype.Text
	Name       pgtype.Text
	Type       pgtype.Text
}

func NewCheckTemplateNameFilter() *CheckTemplateNameFilter {
	filter := &CheckTemplateNameFilter{}
	_ = filter.TemplateID.Set(nil)
	_ = filter.Name.Set(nil)
	_ = filter.Type.Set(nil)
	return filter
}

func (repo *QuestionnaireTemplateRepo) CheckIsExistNameAndType(ctx context.Context, db database.QueryExecer, filter *CheckTemplateNameFilter) (bool, error) {
	query := `
		SELECT COUNT(*) FROM questionnaire_templates qt 
		WHERE qt.questionnaire_template_id != $1 AND qt.name = $2 AND qt.TYPE = $3 AND qt.deleted_at IS NULL
	`
	row := db.QueryRow(ctx, query, filter.TemplateID, filter.Name, filter.Type)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("QuestionnaireTemplateRepo.CheckIsExistNameAndType: %w", err)
	}
	if count >= 1 {
		return true, nil
	}
	return false, nil
}
