package entities

import "github.com/jackc/pgtype"

type UserAccessPaths struct {
	UserID       pgtype.Text
	LocationID   pgtype.Text
	AccessPath   pgtype.Text
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (e *UserAccessPaths) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"location_id",
			"access_path",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.UserID,
			&e.LocationID,
			&e.AccessPath,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *UserAccessPaths) TableName() string {
	return "user_access_paths"
}
