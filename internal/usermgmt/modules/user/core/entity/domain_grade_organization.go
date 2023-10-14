package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type GradeOrganzation interface {
	GradeID() field.String
	GradeValue() field.Int32
}

type DomainGradeOrganzation interface {
	GradeOrganzation
	valueobj.HasOrganizationID
}

type DomainGradeOrganzations []DomainGradeOrganzation

type EmptyDomainGradeOrganzation struct{}

func (EmptyDomainGradeOrganzation) GradeID() field.String {
	return field.NewUndefinedString()
}
func (EmptyDomainGradeOrganzation) GradeValue() field.Int32 {
	return field.NewUndefinedInt32()
}
func (EmptyDomainGradeOrganzation) OrganizationID() field.String {
	return field.NewNullString()
}
