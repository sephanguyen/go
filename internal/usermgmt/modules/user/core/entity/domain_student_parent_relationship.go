package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type StudentParentRelationship interface {
	Relationship() field.String
}

type DomainStudentParentRelationship interface {
	StudentParentRelationship

	valueobj.HasStudentID
	valueobj.HasParentID
}

type DomainStudentParentRelationships []DomainStudentParentRelationship

type DomainStudentParentRelationshipWillBeDelegated struct {
	StudentParentRelationship

	valueobj.HasStudentID
	valueobj.HasParentID
}

type NullDomainStudentParentRelationship struct{}

func (p NullDomainStudentParentRelationship) StudentID() field.String {
	return field.NewNullString()
}

func (p NullDomainStudentParentRelationship) ParentID() field.String {
	return field.NewNullString()
}

func (p NullDomainStudentParentRelationship) Relationship() field.String {
	return field.NewNullString()
}
