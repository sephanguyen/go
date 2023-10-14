package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type ClassDoAccount struct {
	ClassDoID     pgtype.Text
	ClassDoEmail  pgtype.Text
	ClassDoAPIKey pgtype.Text
	CreatedAt     pgtype.Timestamptz
	DeletedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
}

func NewClassDoAccountFromDomain(c *domain.ClassDoAccount) (*ClassDoAccount, error) {
	dto := &ClassDoAccount{}
	database.AllNullEntity(dto)

	if err := multierr.Combine(
		dto.ClassDoID.Set(c.ClassDoID),
		dto.ClassDoEmail.Set(c.ClassDoEmail),
		dto.ClassDoAPIKey.Set(c.ClassDoAPIKey),
		dto.CreatedAt.Set(c.CreatedAt),
		dto.UpdatedAt.Set(c.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("failed to set from classdo domain to classdo dto: %w", err)
	}

	if c.DeletedAt != nil {
		if err := dto.DeletedAt.Set(c.DeletedAt); err != nil {
			return nil, fmt.Errorf("failed to set deleted_at in classdo account dto: %w", err)
		}
	}

	return dto, nil
}

func (c *ClassDoAccount) FieldMap() ([]string, []interface{}) {
	return []string{
			"classdo_id",
			"classdo_email",
			"classdo_api_key",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&c.ClassDoID,
			&c.ClassDoEmail,
			&c.ClassDoAPIKey,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
		}
}

func (c *ClassDoAccount) TableName() string {
	return "classdo_account"
}

func (c *ClassDoAccount) ToClassDoAccountDomain(secretKey string) *domain.ClassDoAccount {
	account := &domain.ClassDoAccount{
		ClassDoID:     c.ClassDoID.String,
		ClassDoEmail:  c.ClassDoEmail.String,
		ClassDoAPIKey: c.ClassDoAPIKey.String,
	}
	account.DecryptAPIKey(secretKey)
	return account
}
