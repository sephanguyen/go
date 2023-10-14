package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (urs *UserReaderService) RetrieveStudentAssociatedToParentAccount(ctx context.Context, in *pb.RetrieveStudentAssociatedToParentAccountRequest) (*pb.RetrieveStudentAssociatedToParentAccountResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	parentID := pgtype.Text{}
	if err := parentID.Set(userID); err != nil {
		return nil, fmt.Errorf("failed to set parentID: %w", err)
	}

	users, err := urs.StudentRepo.GetStudentsByParentID(ctx, urs.DB, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get students by parentID: %w", err)
	}

	resp := make([]*cpb.BasicProfile, 0, len(users))
	for _, user := range users {
		resp = append(resp, &cpb.BasicProfile{
			UserId:     user.ID.String,
			Name:       user.FullName.String,
			Avatar:     user.Avatar.String,
			FacebookId: user.FacebookID.String,
			GivenName:  user.GivenName.String,
			Group:      cpb.UserGroup(cpb.UserGroup_value[user.Group.String]),
			CreatedAt:  timestamppb.New(user.CreatedAt.Time),
			Country:    cpb.Country(cpb.Country_value[user.Country.String]),
			FirstName:  user.FirstName.String,
			LastName:   user.LastName.String,
		})
	}

	return &pb.RetrieveStudentAssociatedToParentAccountResponse{
		Profiles: resp,
	}, nil
}
