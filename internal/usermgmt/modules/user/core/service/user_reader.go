package service

import (
	"context"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserReaderService struct {
	DB       database.Ext
	UserRepo interface {
		SearchProfile(ctx context.Context, db database.QueryExecer, filter *repository.SearchProfileFilter) ([]*entity.LegacyUser, error)
		Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*entity.LegacyUser, error)
	}

	StudentRepo interface {
		GetStudentsByParentID(ctx context.Context, db database.QueryExecer, parentID pgtype.Text) ([]*entity.LegacyUser, error)
	}

	UserGroupV2Repo interface {
		FindAndMapUserGroupAndRolesByUserID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (map[entity.UserGroupV2][]*entity.Role, error)
	}
	OrganizationRepo interface {
		Find(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.Organization, error)
	}
}

// SearchBasicProfile return basic profiles
func (urs *UserReaderService) SearchBasicProfile(ctx context.Context, req *pb.SearchBasicProfileRequest) (*pb.SearchBasicProfileResponse, error) {
	filter := &repository.SearchProfileFilter{
		StudentIDs:    database.TextArray(req.UserIds),
		LocationIDs:   database.TextArray(req.LocationIds),
		OffsetInteger: uint(req.Paging.GetOffsetInteger()),
		Limit:         uint(req.Paging.GetLimit()),
	}
	if err := multierr.Combine(
		filter.StudentName.Set(nil),
	); err != nil {
		return nil, err
	}

	if req.SearchText != nil && req.SearchText.Value != "" {
		filter.StudentName = database.Text("%" + req.SearchText.Value + "%")
	}

	users, err := urs.UserRepo.SearchProfile(ctx, urs.DB, filter)
	if err != nil {
		return nil, err
	}
	data := make([]*cpb.BasicProfile, 0, len(users))

	for _, e := range users {
		data = append(data, &cpb.BasicProfile{
			UserId:     e.ID.String,
			Name:       e.GetName(),
			Avatar:     e.Avatar.String,
			Group:      cpb.UserGroup(cpb.UserGroup_value[e.Group.String]),
			FacebookId: e.FacebookID.String,
			Country:    cpb.Country(cpb.Country_value[e.Country.String]),
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

	return &pb.SearchBasicProfileResponse{Profiles: data, NextPage: next}, nil
}

func (urs *UserReaderService) GetBasicProfile(ctx context.Context, req *pb.GetBasicProfileRequest) (*pb.GetBasicProfileResponse, error) {
	userIDs := req.UserIds
	if len(userIDs) == 0 {
		// if userIDs in request empty, get userID from ctx
		userIDs = []string{interceptors.UserIDFromContext(ctx)}
	}

	users, err := urs.UserRepo.Retrieve(ctx, urs.DB, database.TextArray(userIDs))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(users) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user not found")
	}

	organization, err := urs.OrganizationRepo.Find(ctx, urs.DB, database.Text(golibs.ResourcePathFromCtx(ctx)))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	profiles := make([]*pb.BasicProfile, 0, len(users))
	for _, user := range users {
		userGroupV2, err := urs.GetUserGroupV2(ctx, user.ID.String)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		profile, err := toUserBasicProfile(user, organization, userGroupV2)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		profiles = append(profiles, profile)
	}

	return &pb.GetBasicProfileResponse{
		Profiles: profiles,
	}, nil
}

func toUserBasicProfile(user *entity.LegacyUser, organization *entity.Organization, userGroupV2 []*pb.BasicProfile_UserGroup) (*pb.BasicProfile, error) {
	schoolID, err := strconv.Atoi(organization.OrganizationID.String)
	if err != nil {
		return nil, err
	}

	profile := &pb.BasicProfile{
		UserId:    user.ID.String,
		Name:      user.FullName.String,
		Email:     user.Email.String,
		Avatar:    user.Avatar.String,
		UserGroup: user.Group.String,
		Country:   cpb.Country(cpb.Country_value[user.Country.String]),
		School: &pb.BasicProfile_School{
			SchoolId:   int64(schoolID),
			SchoolName: organization.Name.String,
		},
		UserGroupV2:   userGroupV2,
		CreatedAt:     timestamppb.New(user.CreatedAt.Time),
		LastLoginDate: timestamppb.New(user.LastLoginDate.Time),
		FirstName:     user.FirstName.String,
		LastName:      user.LastName.String,
	}

	return profile, nil
}

func (urs *UserReaderService) GetUserGroupV2(ctx context.Context, userID string) ([]*pb.BasicProfile_UserGroup, error) {
	userGroupAndRole, err := urs.UserGroupV2Repo.FindAndMapUserGroupAndRolesByUserID(ctx, urs.DB, database.Text(userID))
	if err != nil {
		return nil, err
	}

	userGroupV2 := []*pb.BasicProfile_UserGroup{}
	for userGroup, roleEntities := range userGroupAndRole {
		roles := []*pb.BasicProfile_Role{}
		for _, role := range roleEntities {
			roles = append(roles, &pb.BasicProfile_Role{
				Role:      role.RoleName.String,
				CreatedAt: &timestamppb.Timestamp{Seconds: role.CreatedAt.Time.Unix()},
				RoleId:    role.RoleID.String,
			})
		}
		user := &pb.BasicProfile_UserGroup{
			UserGroup:   userGroup.UserGroupName.String,
			Roles:       roles,
			UserGroupId: userGroup.UserGroupID.String,
		}
		userGroupV2 = append(userGroupV2, user)
	}

	return userGroupV2, nil
}
