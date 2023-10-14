package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type Grade interface {
	GradeID() field.String
	Name() field.String
	IsArchived() field.Boolean
	Sequence() field.Int32
}

type DomainGrade interface {
	Grade
	valueobj.HasPartnerInternalID
	valueobj.HasOrganizationID
}

type GradeWillBeDelegated struct {
	Grade
	valueobj.HasPartnerInternalID
	valueobj.HasOrganizationID
}

type NullDomainGrade struct{}

func (e NullDomainGrade) GradeID() field.String {
	return field.NewUndefinedString()
}
func (e NullDomainGrade) Name() field.String {
	return field.NewNullString()
}
func (e NullDomainGrade) IsArchived() field.Boolean {
	return field.NewNullBoolean()
}
func (e NullDomainGrade) PartnerInternalID() field.String {
	return field.NewNullString()
}
func (e NullDomainGrade) Sequence() field.Int32 {
	return field.NewNullInt32()
}
func (e NullDomainGrade) OrganizationID() field.String {
	return field.NewNullString()
}
