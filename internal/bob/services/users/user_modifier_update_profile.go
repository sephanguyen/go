package users

import (
	"context"

	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/grpc/metadata"
)

// UpdateUserProfile updates a user's profile
func (s *UserModifierService) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UpdateUserProfileResponse, error) {
	headers, ok := metadata.FromIncomingContext(ctx)
	var pkg, token, version string
	if ok {
		pkg = headers["pkg"][0]
		token = headers["token"][0]
		version = headers["version"][0]
	}

	resp, err := s.UserMgmtModifierSvc.UpdateUserProfile(metadata.AppendToOutgoingContext(ctx, "pkg", pkg, "version", version, "token", token), &upb.UpdateUserProfileRequest{Profile: &upb.UpdateUserProfileRequest_UserProfile{
		Id:          req.Profile.Id,
		Name:        req.Profile.Name,
		Country:     req.Profile.Country,
		PhoneNumber: req.Profile.PhoneNumber,
		Email:       req.Profile.Email,
		Avatar:      req.Profile.Avatar,
		DeviceToken: req.Profile.DeviceToken,
		Group:       req.Profile.UserGroup,
	}})
	if err != nil {
		return nil, err
	}

	return &pb.UpdateUserProfileResponse{
		Successful: resp.Successful,
	}, nil
}
