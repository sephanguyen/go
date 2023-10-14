package repo

import (
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"

	"github.com/jackc/pgtype"
)

type ZoomAccount struct {
	ID        pgtype.Text `sql:"zoom_id,pk"`
	Email     pgtype.Text
	UserName  pgtype.Text `sql:"user_name"`
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

// FieldMap returns field in users table
func (z *ZoomAccount) FieldMap() ([]string, []interface{}) {
	return []string{
			"zoom_id",
			"email",
			"user_name",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&z.ID,
			&z.Email,
			&z.UserName,
			&z.UpdatedAt,
			&z.CreatedAt,
			&z.DeletedAt,
		}
}

func (z *ZoomAccount) TableName() string {
	return "zoom_account"
}

func (z *ZoomAccount) ToZoomAccountEntity() *domain.ZoomAccount {
	return &domain.ZoomAccount{
		ID:       z.ID.String,
		Email:    z.Email.String,
		UserName: z.UserName.String,
	}
}
