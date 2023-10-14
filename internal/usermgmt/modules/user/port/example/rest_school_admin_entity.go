package example

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type RESTCreateSchoolAdminRequest struct {
	entity.NullDomainSchoolAdmin

	SchoolAdminProfile RESTSchoolAdmin `json:"schoolProfile"`
}

type RESTSchoolAdmin struct {
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	SchoolID int32  `json:"schoolID"`
}

func (request *RESTCreateSchoolAdminRequest) Email() field.String {
	return field.NewString(request.SchoolAdminProfile.Email)
}

func (request *RESTCreateSchoolAdminRequest) FullName() field.String {
	return field.NewString(request.SchoolAdminProfile.FullName)
}

func (request *RESTCreateSchoolAdminRequest) Avatar() field.String {
	return field.NewNullString()
}

func (request *RESTCreateSchoolAdminRequest) Group() field.String {
	return field.NewNullString()
}

func (request *RESTCreateSchoolAdminRequest) SchoolID() field.Int32 {
	return field.NewInt32(request.SchoolAdminProfile.SchoolID)
}
