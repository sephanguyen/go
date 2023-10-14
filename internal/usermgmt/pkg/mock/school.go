package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type School struct {
	RandomSchool
}

type RandomSchool struct {
	entity.DefaultDomainSchool
	SchoolID          field.String
	SchoolLevelID     field.String
	PartnerInternalID field.String
	IsArchived        field.Boolean
}

func (s School) SchoolID() field.String {
	return s.RandomSchool.SchoolID
}
func (s School) SchoolLevelID() field.String {
	return s.RandomSchool.SchoolLevelID
}
func (s School) PartnerInternalID() field.String {
	return s.RandomSchool.PartnerInternalID
}
func (s School) IsArchived() field.Boolean {
	return s.RandomSchool.IsArchived
}
