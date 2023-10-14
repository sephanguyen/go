package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type Parent struct {
	RandomParent
}

type RandomParent struct {
	entity.NullDomainParent
	UserID             field.String
	ExternalUserIDAttr field.String
	EmailAttr          field.String
	UserNameAttr       field.String
	FirstNameAttr      field.String
	LastNameAttr       field.String
}

func (m *Parent) UserID() field.String {
	return m.RandomParent.UserID
}

func (m *Parent) Email() field.String {
	return m.RandomParent.EmailAttr
}

func (m *Parent) UserName() field.String {
	return m.RandomParent.UserNameAttr
}

func (m *Parent) FirstName() field.String {
	return m.RandomParent.FirstNameAttr
}

func (m *Parent) LastName() field.String {
	return m.RandomParent.LastNameAttr
}

func (m *Parent) ExternalUserID() field.String {
	return m.RandomParent.ExternalUserIDAttr
}
