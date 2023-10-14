package repo

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type UserAccessPath struct {
	UserID     pgtype.Text
	LocationID pgtype.Text
	AccessPath pgtype.Text
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
	DeletedAt  pgtype.Timestamptz
}

func (u *UserAccessPath) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"location_id",
			"access_path",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&u.UserID,
			&u.LocationID,
			&u.AccessPath,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.DeletedAt,
		}
}

func (u *UserAccessPath) TableName() string {
	return "user_access_paths"
}
func (u *UserAccessPath) PreUpsert() error {
	now := time.Now()

	if err := multierr.Combine(
		u.CreatedAt.Set(now),
		u.UpdatedAt.Set(now),
	); err != nil {
		return err
	}

	return nil
}

func NewUserAccessPath(userPath []*domain.UserAccessPath) ([]*UserAccessPath, error) {
	entities := make([]*UserAccessPath, 0, len(userPath))
	for _, u := range userPath {
		entity := &UserAccessPath{}
		database.AllNullEntity(entity)
		if err := multierr.Combine(
			entity.UserID.Set(u.UserID),
			entity.LocationID.Set(u.LocationID),
		); err != nil {
			return nil, fmt.Errorf("could not map user access path: %w", err)
		}
		entities = append(entities, entity)
	}
	return entities, nil
}
