package users

import (
	"context"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserReaderService struct {
	DB       database.Ext
	UserRepo interface {
		SearchProfile(ctx context.Context, db database.QueryExecer, filter *repositories.SearchProfileFilter) ([]*entities.User, error)
		Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*entities.User, error)
	}
}

// SearchBasicProfile return basic profiles
func (urs *UserReaderService) SearchBasicProfile(ctx context.Context, req *bpb.SearchBasicProfileRequest) (*bpb.SearchBasicProfileResponse, error) {
	filter := &repositories.SearchProfileFilter{
		StudentIDs:    database.TextArray(req.UserIds),
		OffsetInteger: uint(req.Paging.GetOffsetInteger()),
		Limit:         uint(req.Paging.GetLimit()),
	}
	if err := multierr.Combine(
		filter.StudentName.Set(nil),
	); err != nil {
		return nil, err
	}

	if req.SearchText != nil && req.SearchText.Value != "" {
		filter.StudentName = database.Text(req.SearchText.Value + "%")
	}

	users, err := urs.UserRepo.SearchProfile(ctx, urs.DB, filter)
	if err != nil {
		return nil, err
	}
	data := make([]*cpb.BasicProfile, 0, len(users))

	for _, e := range users {
		data = append(data, &cpb.BasicProfile{
			UserId:           e.ID.String,
			Name:             e.GetName(),
			Avatar:           e.Avatar.String,
			Group:            cpb.UserGroup(cpb.UserGroup_value[e.Group.String]),
			FacebookId:       e.FacebookID.String,
			FullNamePhonetic: e.GetFullNamePhonetic(),
		})
	}
	next := new(cpb.Paging)
	if len(data) == int(req.Paging.Limit) {
		next = &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: req.Paging.GetOffsetInteger() + int64(req.Paging.Limit),
			},
		}
	}

	return &bpb.SearchBasicProfileResponse{Profiles: data, NextPage: next}, nil
}

// TODO: implement
func (*UserReaderService) GetCurrentUserProfile(context.Context, *bpb.GetCurrentUserProfileRequest) (*bpb.GetCurrentUserProfileResponse, error) {
	return nil, nil
}

// TODO: implement
func (*UserReaderService) RetrieveTeacherProfiles(context.Context, *bpb.RetrieveTeacherProfilesRequest) (*bpb.RetrieveTeacherProfilesResponse, error) {
	return nil, nil
}

func (urs *UserReaderService) RetrieveBasicProfile(ctx context.Context, req *bpb.RetrieveBasicProfileRequest) (*bpb.RetrieveBasicProfileResponse, error) {
	if len(req.UserIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing userIds")
	}

	users, err := urs.UserRepo.Retrieve(ctx, urs.DB, database.TextArray(req.UserIds))
	if err != nil {
		return nil, services.ToStatusError(err)
	}

	profiles := make([]*cpb.BasicProfile, 0, len(users))
	for _, e := range users {
		profiles = append(profiles, &cpb.BasicProfile{
			UserId:     e.ID.String,
			Name:       e.GetName(),
			Avatar:     e.Avatar.String,
			FacebookId: e.FacebookID.String,
			Group:      cpb.UserGroup(cpb.UserGroup_value[e.Group.String]),
			Country:    cpb.Country(cpb.Country_value[e.Country.String]),
		})
	}

	return &bpb.RetrieveBasicProfileResponse{
		Profiles: profiles,
	}, nil
}

// TODO: implement

func (*UserReaderService) CheckProfile(context.Context, *bpb.CheckProfileRequest) (*bpb.CheckProfileResponse, error) {
	return nil, nil
}
