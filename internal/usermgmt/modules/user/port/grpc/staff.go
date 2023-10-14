package grpc

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

type UpdateStaffRequest struct {
	Body *pb.UpdateStaffRequest

	entity.EmptyUser
}

func (req UpdateStaffRequest) UserID() field.String {
	return field.NewString(req.Body.Staff.StaffId)
}

func (req UpdateStaffRequest) Email() field.String {
	return field.NewString(req.Body.Staff.Email)
}

func (req UpdateStaffRequest) Birthday() field.Date {
	return field.NewDate(req.Body.Staff.Birthday.AsTime())
}

func (req UpdateStaffRequest) Gender() field.String {
	return field.NewString(req.Body.Staff.Gender.String())
}
