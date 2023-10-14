package grpc

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

type UserProfile struct {
	Profile *pb.UserProfile

	entity.EmptyUser
}

func NewUserProfileWithID(id string) entity.User {
	return &UserProfile{Profile: &pb.UserProfile{UserId: id}}
}

func NewUserProfile(profile *pb.UserProfile) entity.User {
	return &UserProfile{
		Profile: profile,
	}
}

func (c *UserProfile) UserID() field.String {
	return field.NewString(c.Profile.GetUserId())
}

func (c *UserProfile) ExternalUserID() field.String {
	return field.NewString(c.Profile.GetExternalUserId())
}

func (c *UserProfile) Email() field.String {
	return field.NewString(c.Profile.GetEmail())
}

func (c *UserProfile) UserName() field.String {
	return field.NewString(c.Profile.GetUsername())
}

func (c *UserProfile) FirstName() field.String {
	return field.NewString(c.Profile.GetFirstName())
}

func (c *UserProfile) LastName() field.String {
	return field.NewString(c.Profile.GetLastName())
}
