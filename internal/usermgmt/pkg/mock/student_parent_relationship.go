package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type StudentParentRelationship struct {
	RandomStudentParentRelationship
}

type RandomStudentParentRelationship struct {
	entity.NullDomainStudentParentRelationship

	StudentIDAttr    field.String
	ParentIDAttr     field.String
	RelationshipAttr field.String
}

func (m *StudentParentRelationship) StudentID() field.String {
	return m.StudentIDAttr
}

func (m *StudentParentRelationship) ParentID() field.String {
	return m.ParentIDAttr
}

func (m *StudentParentRelationship) Relationship() field.String {
	return m.RelationshipAttr
}
