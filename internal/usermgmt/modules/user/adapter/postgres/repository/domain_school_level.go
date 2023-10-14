package repository

import (
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type DomainSchoolLevelRepo struct{}

type SchoolLevelAttribute struct {
	ID             field.String
	Name           field.String
	Sequence       field.Int64
	IsArchived     field.Boolean
	OrganizationID field.String
}

type SchoolLevel struct {
	SchoolLevelAttribute

	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Time
}

func (sh *SchoolLevel) SchoolLevelID() field.String {
	return sh.SchoolLevelAttribute.ID
}
func (sh *SchoolLevel) Name() field.String {
	return sh.SchoolLevelAttribute.Name
}
func (sh *SchoolLevel) Sequence() field.Int64 {
	return sh.SchoolLevelAttribute.Sequence
}
func (sh *SchoolLevel) IsArchived() field.Boolean {
	return sh.SchoolLevelAttribute.IsArchived
}
func (sh *SchoolLevel) OrganizationID() field.String {
	return sh.SchoolLevelAttribute.OrganizationID
}

func (*SchoolLevel) TableName() string {
	return "school_level"
}

func (sh *SchoolLevel) FieldMap() ([]string, []interface{}) {
	return []string{
			"school_level_id",
			"school_level_name",
			"sequence",
			"is_archived",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&sh.SchoolLevelAttribute.ID,
			&sh.SchoolLevelAttribute.Name,
			&sh.SchoolLevelAttribute.Sequence,
			&sh.SchoolLevelAttribute.IsArchived,
			&sh.CreatedAt,
			&sh.UpdatedAt,
			&sh.DeletedAt,
			&sh.SchoolLevelAttribute.OrganizationID,
		}
}
