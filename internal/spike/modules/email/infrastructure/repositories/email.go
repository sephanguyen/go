package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/model"

	"go.uber.org/multierr"
)

type EmailRepo struct{}

func (repo *EmailRepo) UpsertEmail(ctx context.Context, db database.QueryExecer, email *model.Email) error {
	ctx, span := interceptors.StartSpan(ctx, "EmailRepo.UpsertEmail")
	defer span.End()
	now := time.Now()
	err := multierr.Combine(
		email.CreatedAt.Set(now),
		email.UpdatedAt.Set(now),
		email.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	fields := database.GetFieldNames(email)
	values := database.GetScanFields(email, fields)
	placeHolders := database.GeneratePlaceholders(len(fields))
	tableName := email.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s as e (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT pk__emails 
		DO UPDATE SET 
			sg_message_id = EXCLUDED.sg_message_id,
			subject = EXCLUDED.subject,
			content = EXCLUDED.content,
			email_from = EXCLUDED.email_from,
			status = EXCLUDED.status,
			email_recipients = EXCLUDED.email_recipients,
			updated_at = EXCLUDED.updated_at
		WHERE e.deleted_at IS NULL;
	`, tableName, strings.Join(fields, ", "), placeHolders)

	cmd, err := db.Exec(ctx, query, values...)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("can not upsert email")
	}

	return nil
}

func (repo *EmailRepo) UpdateEmail(ctx context.Context, db database.QueryExecer, emailID string, attributes map[string]interface{}) error {
	ctx, span := interceptors.StartSpan(ctx, "EmailRepo.UpsertEmail")
	defer span.End()

	// cover case empty
	if len(attributes) == 0 {
		return nil
	}

	attributes["updated_at"] = time.Now()

	paramIndex := 1
	query := fmt.Sprintf(`UPDATE %v SET `, (&model.Email{}).TableName())
	params := make([]interface{}, 0)
	for field, value := range attributes {
		query += fmt.Sprintf(`%v = $%v`, field, paramIndex)
		if paramIndex < len(attributes) {
			query += `, `
		}
		paramIndex++
		params = append(params, value)
	}

	query += fmt.Sprintf(" WHERE email_id = $%v", paramIndex)
	params = append(params, database.Text(emailID))

	cmd, err := db.Exec(ctx, query, params...)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}
