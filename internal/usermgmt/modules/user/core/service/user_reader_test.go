package service

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func generateEnUser() *entity.LegacyUser {
	rand.Seed(time.Now().UnixNano())
	e := new(entity.LegacyUser)
	err := multierr.Combine(
		e.ID.Set(fmt.Sprintf("user-id %d", rand.Int())),
		e.Avatar.Set(fmt.Sprintf("http://avatar-%d", rand.Int())),
		e.Group.Set(entity.UserGroupAdmin),
		e.FullName.Set(fmt.Sprintf("user %d", rand.Int())),
		e.Country.Set("COUNTRY_VN"),
		e.PhoneNumber.Set(fmt.Sprintf("+849%d", rand.Int())),
		e.Email.Set(fmt.Sprintf("valid-%d@email.com", rand.Int())),
		e.DeviceToken.Set(fmt.Sprintf("random device %d", rand.Int())),
		e.CreatedAt.Set(time.Now()),
		e.UpdatedAt.Set(time.Now()),
		e.FirstName.Set(fmt.Sprintf("first_name: %d", rand.Int())),
		e.LastName.Set(fmt.Sprintf("last_name: %d", rand.Int())),
	)
	if err != nil {
		fmt.Println("generateEnUser: %w", err)
	}
	return e
}

func TestUserReaderService_SearchBasicProfile(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	userRepo := new(mock_repositories.MockUserRepo)
	e := generateEnUser()
	s := &UserReaderService{
		UserRepo: userRepo,
		DB:       db,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &pb.SearchBasicProfileRequest{
				SearchText: &wrapperspb.StringValue{
					Value: "abc",
				},
				UserIds:     []string{e.ID.String},
				Paging:      &cpb.Paging{Limit: 1},
				LocationIds: []string{constants.ManabieOrgLocation},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("SearchProfile", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entity.LegacyUser{e}, nil)
			},
		},
		{
			name: "happy case without search text",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &pb.SearchBasicProfileRequest{
				SearchText:  nil,
				UserIds:     []string{e.ID.String},
				Paging:      &cpb.Paging{Limit: 1},
				LocationIds: []string{constants.ManabieOrgLocation},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("SearchProfile", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entity.LegacyUser{e}, nil)
			},
		},
		{
			name:        "error query",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &pb.SearchBasicProfileRequest{UserIds: []string{}, Paging: &cpb.Paging{}},
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				userRepo.On("SearchProfile", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entity.LegacyUser{}, pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.SearchBasicProfileRequest)
			_, err := s.SearchBasicProfile(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestGetBasicProfile(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	userRepo := new(mock_repositories.MockUserRepo)
	userGroupV2Rep := new(mock_repositories.MockUserGroupV2Repo)
	organizationRepo := new(mock_repositories.MockOrganizationRepo)

	user := generateEnUser()
	mapUserGroupAndRole := map[entity.UserGroupV2][]*entity.Role{}
	mockUserGroup := entity.UserGroupV2{
		UserGroupID:   database.Text(idutil.ULIDNow()),
		UserGroupName: database.Text(fmt.Sprintf("user group %s", constant.RoleTeacher)),
	}
	mockRole := entity.Role{
		RoleName:  database.Text(constant.RoleTeacher),
		CreatedAt: database.Timestamptz(time.Now()),
		RoleID:    database.Text(idutil.ULIDNow()),
	}
	mapUserGroupAndRole[mockUserGroup] = []*entity.Role{&mockRole}
	organization := &entity.Organization{
		OrganizationID: database.Text(fmt.Sprint(constants.ManabieSchool)),
		Name:           database.Text("Manabie"),
	}
	userReaderService := &UserReaderService{
		DB:               db,
		UserRepo:         userRepo,
		UserGroupV2Repo:  userGroupV2Rep,
		OrganizationRepo: organizationRepo,
	}
	schoolID, err := strconv.Atoi(organization.OrganizationID.String)
	if err != nil {
		assert.NoError(t, err)
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, user.ID.String),
			req: &pb.GetBasicProfileRequest{
				UserIds: []string{user.ID.String},
			},
			setup: func(ctx context.Context) {
				userRepo.On("Retrieve", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.LegacyUser{user}, nil)
				organizationRepo.On("Find", ctx, db, mock.Anything).Once().Return(organization, nil)
				userGroupV2Rep.On("FindAndMapUserGroupAndRolesByUserID", ctx, db, mock.Anything).Once().Return(mapUserGroupAndRole, nil)
			},
			expectedErr: nil,
			expectedResp: pb.GetBasicProfileResponse{
				Profiles: []*pb.BasicProfile{
					{
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
						UserGroupV2: []*pb.BasicProfile_UserGroup{
							{
								UserGroup:   mockUserGroup.UserGroupName.String,
								UserGroupId: mockUserGroup.UserGroupID.String,
								Roles: []*pb.BasicProfile_Role{
									{
										Role:      mockRole.RoleName.String,
										RoleId:    mockRole.RoleID.String,
										CreatedAt: &timestamppb.Timestamp{Seconds: mockRole.CreatedAt.Time.Unix()},
									},
								},
							},
						},
						CreatedAt:     timestamppb.New(user.CreatedAt.Time),
						LastLoginDate: timestamppb.New(user.LastLoginDate.Time),
						FirstName:     user.FirstName.String,
						LastName:      user.LastName.String},
				},
			},
		},
		{
			name: "success: get basic profile without user id in request",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req:  &pb.GetBasicProfileRequest{},
			setup: func(ctx context.Context) {
				userRepo.On("Retrieve", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.LegacyUser{user}, nil)
				organizationRepo.On("Find", ctx, db, mock.Anything).Once().Return(organization, nil)
				userGroupV2Rep.On("FindAndMapUserGroupAndRolesByUserID", ctx, db, mock.Anything).Once().Return(mapUserGroupAndRole, nil)
			},
			expectedErr: nil,
			expectedResp: pb.GetBasicProfileResponse{
				Profiles: []*pb.BasicProfile{
					{
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
						UserGroupV2: []*pb.BasicProfile_UserGroup{
							{
								UserGroup:   mockUserGroup.UserGroupName.String,
								UserGroupId: mockUserGroup.UserGroupID.String,
								Roles: []*pb.BasicProfile_Role{
									{
										Role:      mockRole.RoleName.String,
										RoleId:    mockRole.RoleID.String,
										CreatedAt: &timestamppb.Timestamp{Seconds: mockRole.CreatedAt.Time.Unix()},
									},
								},
							},
						},
						CreatedAt:     timestamppb.New(user.CreatedAt.Time),
						LastLoginDate: timestamppb.New(user.LastLoginDate.Time),
						FirstName:     user.FirstName.String,
						LastName:      user.LastName.String},
				},
			},
		},
		{
			name: "fail: user repo retrieve return error",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &pb.GetBasicProfileRequest{
				UserIds: []string{user.ID.String},
			},
			setup: func(ctx context.Context) {
				userRepo.On("Retrieve", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.LegacyUser{user}, pgx.ErrNoRows)
			},
			expectedErr: status.Error(codes.Internal, pgx.ErrNoRows.Error()),
		},
		{
			name: "fail: user not found",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &pb.GetBasicProfileRequest{
				UserIds: []string{user.ID.String},
			},
			setup: func(ctx context.Context) {
				userRepo.On("Retrieve", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, "user not found"),
		},
		{
			name: "fail: organizationRepo.Find return error",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &pb.GetBasicProfileRequest{
				UserIds: []string{user.ID.String},
			},
			setup: func(ctx context.Context) {
				userRepo.On("Retrieve", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.LegacyUser{user}, nil)
				organizationRepo.On("Find", ctx, db, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr: status.Error(codes.Internal, pgx.ErrNoRows.Error()),
		},
		{
			name: "fail: userGroupV2Rep.FindAndMapUserGroupAndRolesByUserID return error",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &pb.GetBasicProfileRequest{
				UserIds: []string{user.ID.String},
			},
			setup: func(ctx context.Context) {
				userRepo.On("Retrieve", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.LegacyUser{user}, nil)
				organizationRepo.On("Find", ctx, db, mock.Anything).Once().Return(organization, nil)
				userGroupV2Rep.On("FindAndMapUserGroupAndRolesByUserID", ctx, db, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
			expectedErr: status.Error(codes.Internal, pgx.ErrTxClosed.Error()),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			testCase.setup(testCase.ctx)

			req := testCase.req.(*pb.GetBasicProfileRequest)
			profile, err := userReaderService.GetBasicProfile(testCase.ctx, req)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, *profile)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}
