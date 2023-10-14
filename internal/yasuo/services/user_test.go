package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/auth"
	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	usermgmt_entities "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	mock_bobpb "github.com/manabie-com/backend/mock/bob/protobuf"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_multitenant "github.com/manabie-com/backend/mock/golibs/auth/multitenant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_usermgmt_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	ppb "github.com/manabie-com/backend/pkg/genproto/bob"
	pby "github.com/manabie-com/backend/pkg/genproto/yasuo"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pby_v1 "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/go-pg/pg"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockSchoolRepo struct {
	getFn func(ctx context.Context, db database.QueryExecer, schoolIDs []int32) (map[int32]*entities.School, error)
}

func (m *mockSchoolRepo) Get(ctx context.Context, db database.QueryExecer, schoolIDs []int32) (map[int32]*entities.School, error) {
	return m.getFn(ctx, db, schoolIDs)
}

type mockUserModifierService struct {
	handleCreateUserFn func(ctx context.Context, userProfiles []*pby_v1.CreateUserProfile, userGroup, org string, schoolID int64) ([]*entities.User, error)
}

func (m *mockUserModifierService) HandleCreateUser(ctx context.Context, userProfiles []*pby_v1.CreateUserProfile, userGroup, org string, schoolID int64) ([]*entities.User, error) {
	return m.handleCreateUserFn(ctx, userProfiles, userGroup, org, schoolID)
}

func TestCreateUser(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = auth.InjectFakeJwtToken(ctx, "1")
	mockErr := fmt.Errorf("mock error")

	s := &UserService{}
	req := &pby.CreateUserRequest{
		UserGroup: pby.USER_GROUP_STUDENT,
		SchoolId:  1,
		Users: []*pby.CreateUserProfile{
			{
				Name:        "someName",
				GivenName:   "someGivenName",
				Country:     ppb.COUNTRY_VN,
				PhoneNumber: "0123456789",
				Email:       "someEmail@gmail.com",
				Avatar:      "some/avatar/url.jpg",
				Grade:       7,
			},
		},
	}

	t.Run("CreateUser school admin success", func(t *testing.T) {
		s.UserModifierService = &mockUserModifierService{
			handleCreateUserFn: func(ctx context.Context, userProfiles []*pby_v1.CreateUserProfile, userGroup, org string, schoolID int64) ([]*entities.User, error) {
				return []*entities.User{
					{
						ID:           pgtype.Text{String: "someUserID"},
						GivenName:    pgtype.Text{String: "someGivenName"},
						LastName:     pgtype.Text{String: "someName"},
						Avatar:       pgtype.Text{String: "some/avatar/url.jpg"},
						Email:        pgtype.Text{String: "someEmail@gmail.com"},
						Country:      pgtype.Text{String: constant.CountryVN},
						PhoneNumber:  pgtype.Text{String: "0123456789"},
						Group:        pgtype.Text{String: constant.UserGroupSchoolAdmin},
						ResourcePath: pgtype.Text{String: "1"},
					},
				}, nil
			},
		}

		resp, err := s.CreateUser(ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.EqualValues(t, resp.Users[0].Name, req.Users[0].GivenName+req.Users[0].Name)
		assert.EqualValues(t, resp.Users[0].Avatar, req.Users[0].Avatar)
		assert.EqualValues(t, resp.Users[0].Country, req.Users[0].Country)
		assert.EqualValues(t, resp.Users[0].Email, req.Users[0].Email)
		assert.EqualValues(t, resp.Users[0].PhoneNumber, req.Users[0].PhoneNumber)
	})
	t.Run("CreateUser success", func(t *testing.T) {
		s.UserModifierService = &mockUserModifierService{
			handleCreateUserFn: func(ctx context.Context, userProfiles []*pby_v1.CreateUserProfile, userGroup, org string, schoolID int64) ([]*entities.User, error) {
				return []*entities.User{
					{
						ID:           pgtype.Text{String: "someUserID"},
						GivenName:    pgtype.Text{String: "someGivenName"},
						LastName:     pgtype.Text{String: "someName"},
						Avatar:       pgtype.Text{String: "some/avatar/url.jpg"},
						Email:        pgtype.Text{String: "someEmail@gmail.com"},
						Country:      pgtype.Text{String: constant.CountryVN},
						PhoneNumber:  pgtype.Text{String: "0123456789"},
						Group:        pgtype.Text{String: constant.UserGroupStudent},
						ResourcePath: pgtype.Text{String: "1"},
					},
				}, nil
			},
		}

		resp, err := s.CreateUser(ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.EqualValues(t, resp.Users[0].Name, req.Users[0].GivenName+req.Users[0].Name)
		assert.EqualValues(t, resp.Users[0].Avatar, req.Users[0].Avatar)
		assert.EqualValues(t, resp.Users[0].Country, req.Users[0].Country)
		assert.EqualValues(t, resp.Users[0].Email, req.Users[0].Email)
		assert.EqualValues(t, resp.Users[0].PhoneNumber, req.Users[0].PhoneNumber)
	})

	t.Run("CreateUser failed due to UserModifierService.HandleCreateUser return error", func(t *testing.T) {
		s.UserModifierService = &mockUserModifierService{
			handleCreateUserFn: func(ctx context.Context, userProfiles []*pby_v1.CreateUserProfile, userGroup, org string, schoolID int64) ([]*entities.User, error) {
				return nil, mockErr
			},
		}

		resp, err := s.CreateUser(ctx, req)
		assert.NotNil(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, mockErr)
	})
}

func TestUserModifierService_CreateUserInIdentityPlatform(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = auth.InjectFakeJwtToken(ctx, "1")

	tenantManager := new(mock_multitenant.TenantManager)
	s := &UserModifierService{
		TenantManager: tenantManager,
	}

	t.Run("get tenant client error", func(t *testing.T) {
		tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(nil, errors.New("mock err"))

		err := s.CreateUserInIdentityPlatform(ctx, "test-tenant-id", []*entities.User{{}}, 1)
		assert.NotNil(t, err)
	})

	t.Run("failed to import users", func(t *testing.T) {
		tenantClient := &mock_multitenant.TenantClient{}
		tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
		tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(nil, errors.New("mock err"))

		err := s.CreateUserInIdentityPlatform(ctx, "test-tenant-id", []*entities.User{{}}, 1)
		assert.NotNil(t, err)
	})

	t.Run("failed to import users", func(t *testing.T) {
		tenantClient := &mock_multitenant.TenantClient{}
		tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
		tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{UsersFailedToImport: internal_auth_user.UsersFailedToImport{{
			User: &entities.User{Email: database.Text("email@example.com")},
			Err:  "user is existed",
		}}}, nil)

		err := s.CreateUserInIdentityPlatform(ctx, "test-tenant-id", []*entities.User{{}}, 1)
		assert.NotNil(t, err)
	})

	t.Run("create user successfully", func(t *testing.T) {
		tenantClient := &mock_multitenant.TenantClient{}
		tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
		tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)

		err := s.CreateUserInIdentityPlatform(ctx, "test-tenant-id", []*entities.User{{}}, 1)
		assert.Nil(t, err)
	})
}

func TestGetBasicProfile(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = interceptors.ContextWithUserID(ctx, "someUserID")
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"pkg": []string{"somePkg"}, "token": []string{"someToken"}, "version": []string{"1"}})
	teacherRepo := &mock_repositories.MockTeacherRepo{}
	schoolAdminRepo := &mock_repositories.MockSchoolAdminRepo{}
	userRepo := &mock_repositories.MockUserRepo{}
	userServiceClient := &mock_bobpb.UserServiceClient{}
	userGroupV2Repo := &mock_usermgmt_repositories.MockUserGroupV2Repo{}
	mockErr := fmt.Errorf("mock error")

	s := &UserService{
		TeacherRepo:     teacherRepo,
		SchoolAdminRepo: schoolAdminRepo,
		UserController:  userServiceClient,
		UserRepo:        userRepo,
		UserGroupV2Repo: userGroupV2Repo,
	}

	mapUserGroupAnRole := make(map[string][]*usermgmt_entities.Role)

	t.Run("student GetBasicProfile success", func(t *testing.T) {
		userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.User{Group: pgtype.Text{String: constant.UserGroupStudent}}, nil)
		userGroupV2Repo.On("FindUserGroupAndRoleByUserID", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(mapUserGroupAnRole, nil)
		mockSchool := entities.School{ID: pgtype.Int4{Int: 1}, Name: database.Text("school id 1")}
		s.SchoolRepo = &mockSchoolRepo{
			getFn: func(ctx context.Context, db database.QueryExecer, schoolIDs []int32) (map[int32]*entities.School, error) {
				return map[int32]*entities.School{
					1: &mockSchool,
				}, nil
			},
		}

		resp, err := s.GetBasicProfile(ctx, &pby.GetBasicProfileRequest{})
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.EqualValues(t, resp.User.Id, "someUserID")
	})
	t.Run("teacher GetBasicProfile success", func(t *testing.T) {
		userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.User{Group: pgtype.Text{String: constant.UserGroupTeacher}}, nil)
		teacherRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.Teacher{SchoolIDs: pgtype.Int4Array{Elements: []pgtype.Int4{{Int: 1}}}}, nil)
		userGroupV2Repo.On("FindUserGroupAndRoleByUserID", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(mapUserGroupAnRole, nil)
		mockSchool := entities.School{ID: pgtype.Int4{Int: 1}, Name: database.Text("school id 1")}
		s.SchoolRepo = &mockSchoolRepo{
			getFn: func(ctx context.Context, db database.QueryExecer, schoolIDs []int32) (map[int32]*entities.School, error) {
				return map[int32]*entities.School{
					1: &mockSchool,
				}, nil
			},
		}

		resp, err := s.GetBasicProfile(ctx, &pby.GetBasicProfileRequest{})
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.EqualValues(t, resp.User.Id, "someUserID")
		assert.EqualValues(t, resp.User.Schools[0].SchoolId, mockSchool.ID.Int)
		assert.EqualValues(t, resp.User.Schools[0].SchoolName, mockSchool.Name.String)
		assert.EqualValues(t, resp.User.SchoolIds[0], 1)
	})
	t.Run("school admin GetBasicProfile success", func(t *testing.T) {
		userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.User{Group: pgtype.Text{String: constant.UserGroupSchoolAdmin}}, nil)
		schoolAdminRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.SchoolAdmin{SchoolID: pgtype.Int4{Int: 1}}, nil)
		userGroupV2Repo.On("FindUserGroupAndRoleByUserID", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(mapUserGroupAnRole, nil)
		mockSchool := entities.School{ID: pgtype.Int4{Int: 1}, Name: database.Text("school id 1")}
		s.SchoolRepo = &mockSchoolRepo{
			getFn: func(ctx context.Context, db database.QueryExecer, schoolIDs []int32) (map[int32]*entities.School, error) {
				return map[int32]*entities.School{
					1: &mockSchool,
				}, nil
			},
		}

		resp, err := s.GetBasicProfile(ctx, &pby.GetBasicProfileRequest{})
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.EqualValues(t, resp.User.Id, "someUserID")
		assert.EqualValues(t, resp.User.Schools[0].SchoolId, mockSchool.ID.Int)
		assert.EqualValues(t, resp.User.Schools[0].SchoolName, mockSchool.Name.String)
		assert.EqualValues(t, resp.User.SchoolIds[0], 1)
	})
	t.Run("user GetBasicProfile with user group data success", func(t *testing.T) {
		mapUserGroupAnRole["userGroupMock"] = []*usermgmt_entities.Role{
			{
				RoleID:    database.Text(idutil.ULIDNow()),
				RoleName:  database.Text("role mock"),
				CreatedAt: database.Timestamptz(time.Now()),
			},
		}
		userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.User{Group: pgtype.Text{String: constant.UserGroupSchoolAdmin}}, nil)
		schoolAdminRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.SchoolAdmin{SchoolID: pgtype.Int4{Int: 1}}, nil)
		userGroupV2Repo.On("FindUserGroupAndRoleByUserID", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(mapUserGroupAnRole, nil)
		mockSchool := entities.School{ID: pgtype.Int4{Int: 1}, Name: database.Text("school id 1")}
		s.SchoolRepo = &mockSchoolRepo{
			getFn: func(ctx context.Context, db database.QueryExecer, schoolIDs []int32) (map[int32]*entities.School, error) {
				return map[int32]*entities.School{
					1: &mockSchool,
				}, nil
			},
		}
		resp, err := s.GetBasicProfile(ctx, &pby.GetBasicProfileRequest{})
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.EqualValues(t, resp.User.Id, "someUserID")
		assert.EqualValues(t, resp.User.Schools[0].SchoolId, mockSchool.ID.Int)
		assert.EqualValues(t, resp.User.Schools[0].SchoolName, mockSchool.Name.String)
		assert.EqualValues(t, resp.User.SchoolIds[0], 1)
		assert.EqualValues(t, len(resp.User.UserGroupV2[0].Roles), 1)
	})
	t.Run("users GetBasicProfile failed due to UserController.GetCurrentUserProfile return error", func(t *testing.T) {
		userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(nil, mockErr)

		resp, err := s.GetBasicProfile(ctx, &pby.GetBasicProfileRequest{})
		assert.NotNil(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, mockErr)
	})
	t.Run("teacher GetBasicProfile failed due to TeacherRepo.FindByID return no rows error", func(t *testing.T) {
		userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.User{Group: pgtype.Text{String: constant.UserGroupTeacher}}, nil)
		teacherRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(nil, pg.ErrNoRows)

		resp, err := s.GetBasicProfile(ctx, &pby.GetBasicProfileRequest{})
		assert.NotNil(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, status.Error(codes.NotFound, "cannot find teacher"))
	})
	t.Run("teacher GetBasicProfile failed due to TeacherRepo.FindByID return error", func(t *testing.T) {
		userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.User{Group: pgtype.Text{String: constant.UserGroupTeacher}}, nil)
		teacherRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(nil, mockErr)

		resp, err := s.GetBasicProfile(ctx, &pby.GetBasicProfileRequest{})
		assert.NotNil(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, mockErr)
	})
	t.Run("school admin GetBasicProfile failed due to SchoolAdminRepo.Get return no rows error", func(t *testing.T) {
		userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.User{Group: pgtype.Text{String: constant.UserGroupSchoolAdmin}}, nil)
		schoolAdminRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(nil, pg.ErrNoRows)

		resp, err := s.GetBasicProfile(ctx, &pby.GetBasicProfileRequest{})
		assert.NotNil(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, status.Error(codes.NotFound, "cannot find school admin"))
	})
	t.Run("school admin GetBasicProfile failed due to SchoolAdminRepo.Get return error", func(t *testing.T) {
		userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.User{Group: pgtype.Text{String: constant.UserGroupSchoolAdmin}}, nil)
		schoolAdminRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(nil, mockErr)

		resp, err := s.GetBasicProfile(ctx, &pby.GetBasicProfileRequest{})
		assert.NotNil(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, mockErr)
	})
	t.Run("teacher GetBasicProfile failed due to SchoolRepo.Get return error", func(t *testing.T) {
		userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.User{Group: pgtype.Text{String: constant.UserGroupTeacher}}, nil)
		teacherRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.Teacher{SchoolIDs: pgtype.Int4Array{Elements: []pgtype.Int4{{Int: 1}}}}, nil)

		s.SchoolRepo = &mockSchoolRepo{
			getFn: func(ctx context.Context, db database.QueryExecer, schoolIDs []int32) (map[int32]*entities.School, error) {
				return nil, mockErr
			},
		}

		resp, err := s.GetBasicProfile(ctx, &pby.GetBasicProfileRequest{})
		assert.NotNil(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, mockErr)
	})
	t.Run("school admin GetBasicProfile failed due to SchoolRepo.Get return error", func(t *testing.T) {
		userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.User{Group: pgtype.Text{String: constant.UserGroupSchoolAdmin}}, nil)
		schoolAdminRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.SchoolAdmin{SchoolID: pgtype.Int4{Int: 1}}, nil)

		s.SchoolRepo = &mockSchoolRepo{
			getFn: func(ctx context.Context, db database.QueryExecer, schoolIDs []int32) (map[int32]*entities.School, error) {
				return nil, mockErr
			},
		}

		resp, err := s.GetBasicProfile(ctx, &pby.GetBasicProfileRequest{})
		assert.NotNil(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, mockErr)
	})
	t.Run("teacher GetBasicProfile failed due to SchoolRepo.Get return empty", func(t *testing.T) {
		userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.User{Group: pgtype.Text{String: constant.UserGroupTeacher}}, nil)
		teacherRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.Teacher{SchoolIDs: pgtype.Int4Array{Elements: []pgtype.Int4{{Int: 1}}}}, nil)

		s.SchoolRepo = &mockSchoolRepo{
			getFn: func(ctx context.Context, db database.QueryExecer, schoolIDs []int32) (map[int32]*entities.School, error) {
				return map[int32]*entities.School{}, nil
			},
		}

		resp, err := s.GetBasicProfile(ctx, &pby.GetBasicProfileRequest{})
		assert.NotNil(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, status.Error(codes.NotFound, "cannot find schools"))
	})
	t.Run("school admin GetBasicProfile failed due to SchoolRepo.Get return empty", func(t *testing.T) {
		userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.User{Group: pgtype.Text{String: constant.UserGroupSchoolAdmin}}, nil)
		schoolAdminRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.SchoolAdmin{SchoolID: pgtype.Int4{Int: 1}}, nil)

		s.SchoolRepo = &mockSchoolRepo{
			getFn: func(ctx context.Context, db database.QueryExecer, schoolIDs []int32) (map[int32]*entities.School, error) {
				return map[int32]*entities.School{}, nil
			},
		}

		resp, err := s.GetBasicProfile(ctx, &pby.GetBasicProfileRequest{})
		assert.NotNil(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, status.Error(codes.NotFound, "cannot find school"))
	})
	t.Run("user GetBasicProfile failed due to UserGroupV2.FindUserGroupAndRoleByUserID return error", func(t *testing.T) {
		mapUserGroupAnRole["userGroupMock"] = []*usermgmt_entities.Role{
			{
				RoleID:    database.Text(idutil.ULIDNow()),
				RoleName:  database.Text("role mock"),
				CreatedAt: database.Timestamptz(time.Now()),
			},
		}
		userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.User{Group: pgtype.Text{String: constant.UserGroupSchoolAdmin}}, nil)
		schoolAdminRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(&entities.SchoolAdmin{SchoolID: pgtype.Int4{Int: 1}}, nil)
		userGroupV2Repo.On("FindUserGroupAndRoleByUserID", mock.AnythingOfType("*context.valueCtx"), mock.Anything, mock.Anything).Once().
			Return(nil, fmt.Errorf("failed to find user group and role"))
		mockSchool := entities.School{ID: pgtype.Int4{Int: 1}, Name: database.Text("school id 1")}
		s.SchoolRepo = &mockSchoolRepo{
			getFn: func(ctx context.Context, db database.QueryExecer, schoolIDs []int32) (map[int32]*entities.School, error) {
				return map[int32]*entities.School{
					1: &mockSchool,
				}, nil
			},
		}
		resp, err := s.GetBasicProfile(ctx, &pby.GetBasicProfileRequest{})
		assert.NotNil(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, status.Error(codes.Internal, "getUserGroupV2: failed to find user group and role"))
	})
}

func TestUserService_SyncStudent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	userRepo := &mock_repositories.MockUserRepo{}

	s := &UserService{
		DBPgx:       db,
		StudentRepo: studentRepo,
		UserRepo:    userRepo,
	}

	t.Run("err insert student", func(t *testing.T) {
		studentID := idutil.ULIDNow()
		req := &npb.EventUserRegistration_Student{
			ActionKind:  npb.ActionKind_ACTION_KIND_UPSERTED,
			StudentId:   studentID,
			StudentDivs: []int64{1},
			LastName:    "Student",
			GivenName:   "Name",
			Packages:    nil,
		}

		studentRepo.On("Find", ctx, db, database.Text(req.StudentId)).Once().
			Return(nil, pgx.ErrNoRows)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		studentRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Student")).Once().
			Return(pgx.ErrTxClosed)

		err := s.SyncStudent(ctx, []*npb.EventUserRegistration_Student{req})
		assert.EqualError(t, err, fmt.Sprintf("s.UpsertStudent studentID %s: s.StudentRepo.Create: tx is closed", studentID))
	})

	t.Run("success insert student", func(t *testing.T) {
		req := &npb.EventUserRegistration_Student{
			ActionKind:  npb.ActionKind_ACTION_KIND_UPSERTED,
			StudentId:   idutil.ULIDNow(),
			StudentDivs: []int64{1},
			LastName:    "Student",
			GivenName:   "Name",
			Packages:    nil,
		}

		studentRepo.On("Find", ctx, db, database.Text(req.StudentId)).Once().
			Return(nil, pgx.ErrNoRows)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		studentRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Student")).Once().
			Run(func(args mock.Arguments) {
				s := args[2].(*entities.Student)

				assert.Equal(t, req.StudentId, s.ID.String)
				assert.Equal(t, req.GivenName+" "+req.LastName, s.GetName())
				assert.Equal(t, "COUNTRY_JP", s.Country.String)
			}).Return(nil)

		err := s.SyncStudent(ctx, []*npb.EventUserRegistration_Student{req})
		assert.Nil(t, err)
	})

	t.Run("err update student record", func(t *testing.T) {
		studentID := idutil.ULIDNow()
		req := &npb.EventUserRegistration_Student{
			ActionKind:  npb.ActionKind_ACTION_KIND_UPSERTED,
			StudentId:   studentID,
			StudentDivs: []int64{1},
			LastName:    "Student",
			GivenName:   "Name",
			Packages:    nil,
		}

		studentRepo.On("Find", ctx, db, database.Text(req.StudentId)).Once().
			Return(&entities.Student{}, nil)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		studentRepo.On("Update", ctx, tx, mock.AnythingOfType("*entities.Student")).Once().
			Return(pgx.ErrTxClosed)

		err := s.SyncStudent(ctx, []*npb.EventUserRegistration_Student{req})
		assert.EqualError(t, err, fmt.Sprintf("s.UpsertStudent studentID %s: s.StudentRepo.Update: tx is closed", studentID))
	})

	t.Run("err find user record", func(t *testing.T) {
		studentID := idutil.ULIDNow()

		req := &npb.EventUserRegistration_Student{
			ActionKind:  npb.ActionKind_ACTION_KIND_UPSERTED,
			StudentId:   studentID,
			StudentDivs: []int64{1},
			LastName:    "Student",
			GivenName:   "Name",
			Packages:    nil,
		}

		studentRepo.On("Find", ctx, db, database.Text(req.StudentId)).Once().
			Return(&entities.Student{
				ID: database.Text(studentID),
			}, nil)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		studentRepo.On("Update", ctx, tx, mock.AnythingOfType("*entities.Student")).Once().
			Return(nil)

		userRepo.On("FindByIDUnscope", ctx, tx, database.Text(studentID)).Once().
			Return(nil, pgx.ErrNoRows)

		err := s.SyncStudent(ctx, []*npb.EventUserRegistration_Student{req})
		assert.EqualError(t, err, fmt.Sprintf("s.UpsertStudent studentID %s: err FindUser: no rows in result set", studentID))
	})

	t.Run("err update user record", func(t *testing.T) {
		studentID := idutil.ULIDNow()

		req := &npb.EventUserRegistration_Student{
			ActionKind:  npb.ActionKind_ACTION_KIND_UPSERTED,
			StudentId:   studentID,
			StudentDivs: []int64{1},
			LastName:    "Student",
			GivenName:   "Name",
			Packages:    nil,
		}

		studentRepo.On("Find", ctx, db, database.Text(req.StudentId)).Once().
			Return(&entities.Student{
				ID: database.Text(studentID),
			}, nil)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		studentRepo.On("Update", ctx, tx, mock.AnythingOfType("*entities.Student")).Once().
			Return(nil)

		userRepo.On("FindByIDUnscope", ctx, tx, database.Text(studentID)).Once().
			Return(&entities.User{}, nil)

		userRepo.On("Update", ctx, tx, mock.AnythingOfType("*entities.User")).Once().
			Return(pgx.ErrTxClosed)

		err := s.SyncStudent(ctx, []*npb.EventUserRegistration_Student{req})
		assert.EqualError(t, err, fmt.Sprintf("s.UpsertStudent studentID %s: err Update: tx is closed", studentID))
	})

	t.Run("success update user record", func(t *testing.T) {
		studentID := idutil.ULIDNow()

		req := &npb.EventUserRegistration_Student{
			ActionKind:  npb.ActionKind_ACTION_KIND_UPSERTED,
			StudentId:   studentID,
			StudentDivs: []int64{1},
			LastName:    "Student",
			GivenName:   "Name",
			Packages:    nil,
		}

		studentRepo.On("Find", ctx, db, database.Text(req.StudentId)).Once().
			Return(&entities.Student{
				ID: database.Text(studentID),
			}, nil)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		studentRepo.On("Update", ctx, tx, mock.AnythingOfType("*entities.Student")).Once().
			Return(nil)

		userRepo.On("FindByIDUnscope", ctx, tx, database.Text(studentID)).Once().
			Return(&entities.User{}, nil)

		userRepo.On("Update", ctx, tx, mock.AnythingOfType("*entities.User")).Once().
			Return(nil)

		err := s.SyncStudent(ctx, []*npb.EventUserRegistration_Student{req})
		assert.Nil(t, err)
	})

	t.Run("err delete student records", func(t *testing.T) {
		studentID := idutil.ULIDNow()

		req := &npb.EventUserRegistration_Student{
			ActionKind:  npb.ActionKind_ACTION_KIND_DELETED,
			StudentId:   studentID,
			StudentDivs: []int64{1},
			LastName:    "Student",
			GivenName:   "Name",
			Packages:    nil,
		}

		studentRepo.On("Find", ctx, db, database.Text(req.StudentId)).Once().
			Return(&entities.Student{
				ID: database.Text(studentID),
			}, nil)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		studentRepo.On("SoftDelete", ctx, tx, database.TextArray([]string{studentID})).Once().
			Return(pgx.ErrTxClosed)

		err := s.SyncStudent(ctx, []*npb.EventUserRegistration_Student{req})
		assert.EqualError(t, err, fmt.Sprintf("s.DeleteStudent studentIDs %v: s.StudentRepo.SoftDelete: tx is closed", []string{studentID}))
	})

	t.Run("err delete user records", func(t *testing.T) {
		studentID := idutil.ULIDNow()

		req := &npb.EventUserRegistration_Student{
			ActionKind:  npb.ActionKind_ACTION_KIND_DELETED,
			StudentId:   studentID,
			StudentDivs: []int64{1},
			LastName:    "Student",
			GivenName:   "Name",
			Packages:    nil,
		}

		studentRepo.On("Find", ctx, db, database.Text(req.StudentId)).Once().
			Return(&entities.Student{
				ID: database.Text(studentID),
			}, nil)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		studentRepo.On("SoftDelete", ctx, tx, database.TextArray([]string{studentID})).Once().
			Return(nil)

		userRepo.On("SoftDelete", ctx, tx, database.TextArray([]string{studentID})).Once().
			Return(pgx.ErrTxClosed)

		err := s.SyncStudent(ctx, []*npb.EventUserRegistration_Student{req})
		assert.EqualError(t, err, fmt.Sprintf("s.DeleteStudent studentIDs %v: s.UserRepo.SoftDelete: tx is closed", []string{studentID}))
	})

	t.Run("success delete students", func(t *testing.T) {
		studentID := idutil.ULIDNow()

		req := &npb.EventUserRegistration_Student{
			ActionKind:  npb.ActionKind_ACTION_KIND_DELETED,
			StudentId:   studentID,
			StudentDivs: []int64{1},
			LastName:    "Student",
			GivenName:   "Name",
			Packages:    nil,
		}

		studentRepo.On("Find", ctx, db, database.Text(req.StudentId)).Once().
			Return(&entities.Student{
				ID: database.Text(studentID),
			}, nil)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		studentRepo.On("SoftDelete", ctx, tx, database.TextArray([]string{studentID})).Once().
			Return(nil)

		userRepo.On("SoftDelete", ctx, tx, database.TextArray([]string{studentID})).Once().
			Return(nil)

		err := s.SyncStudent(ctx, []*npb.EventUserRegistration_Student{req})
		assert.Nil(t, err)
	})
}

func TestUserService_SyncTeacher(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	teacherRepo := &mock_repositories.MockTeacherRepo{}
	userRepo := &mock_repositories.MockUserRepo{}

	s := &UserService{
		DBPgx:       db,
		TeacherRepo: teacherRepo,
		UserRepo:    userRepo,
	}

	t.Run("err insert teacher", func(t *testing.T) {
		teacherID := idutil.ULIDNow()
		req := &npb.EventUserRegistration_Staff{
			ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			StaffId:    teacherID,
			Name:       "Teacher",
		}

		teacherRepo.On("FindRegardlessDeletion", ctx, db, database.Text(req.StaffId)).Once().
			Return(nil, pgx.ErrNoRows)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		teacherRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Teacher")).Once().
			Return(pgx.ErrTxClosed)

		err := s.SyncTeacher(ctx, []*npb.EventUserRegistration_Staff{req})
		assert.EqualError(t, err, fmt.Sprintf("s.UpsertTeacher teacherID %s: s.TeacherRepo.Create: tx is closed", teacherID))
	})

	t.Run("success insert teacher", func(t *testing.T) {
		req := &npb.EventUserRegistration_Staff{
			ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			StaffId:    idutil.ULIDNow(),
			Name:       "Teacher",
		}

		teacherRepo.On("FindRegardlessDeletion", ctx, db, database.Text(req.StaffId)).Once().
			Return(nil, pgx.ErrNoRows)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		teacherRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Teacher")).Once().
			Run(func(args mock.Arguments) {
				s := args[2].(*entities.Teacher)

				assert.Equal(t, req.StaffId, s.ID.String)
				assert.Equal(t, req.Name, s.GetName())
				assert.Equal(t, "COUNTRY_JP", s.Country.String)
			}).Return(nil)

		err := s.SyncTeacher(ctx, []*npb.EventUserRegistration_Staff{req})
		assert.Nil(t, err)
	})

	t.Run("err update teacher record", func(t *testing.T) {
		teacherID := idutil.ULIDNow()
		req := &npb.EventUserRegistration_Staff{
			ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			StaffId:    teacherID,
			Name:       "Name",
		}

		teacherRepo.On("FindRegardlessDeletion", ctx, db, database.Text(req.StaffId)).Once().
			Return(&entities.Teacher{}, nil)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		teacherRepo.On("Update", ctx, tx, mock.AnythingOfType("*entities.Teacher")).Once().
			Return(pgx.ErrTxClosed)

		err := s.SyncTeacher(ctx, []*npb.EventUserRegistration_Staff{req})
		assert.EqualError(t, err, fmt.Sprintf("s.UpsertTeacher teacherID %s: s.TeacherRepo.Update: tx is closed", teacherID))
	})

	t.Run("err find user record", func(t *testing.T) {
		teacherID := idutil.ULIDNow()

		req := &npb.EventUserRegistration_Staff{
			ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			StaffId:    teacherID,
			Name:       "Name",
		}

		teacherRepo.On("FindRegardlessDeletion", ctx, db, database.Text(req.StaffId)).Once().
			Return(&entities.Teacher{
				ID: database.Text(teacherID),
			}, nil)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		teacherRepo.On("Update", ctx, tx, mock.AnythingOfType("*entities.Teacher")).Once().
			Return(nil)

		userRepo.On("FindByIDUnscope", ctx, tx, database.Text(teacherID)).Once().
			Return(nil, pgx.ErrNoRows)

		err := s.SyncTeacher(ctx, []*npb.EventUserRegistration_Staff{req})
		assert.EqualError(t, err, fmt.Sprintf("s.UpsertTeacher teacherID %s: err FindUser: no rows in result set", teacherID))
	})

	t.Run("err update user record", func(t *testing.T) {
		teacherID := idutil.ULIDNow()

		req := &npb.EventUserRegistration_Staff{
			ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			StaffId:    teacherID,
			Name:       "Teacher",
		}

		teacherRepo.On("FindRegardlessDeletion", ctx, db, database.Text(req.StaffId)).Once().
			Return(&entities.Teacher{
				ID: database.Text(teacherID),
			}, nil)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		teacherRepo.On("Update", ctx, tx, mock.AnythingOfType("*entities.Teacher")).Once().
			Return(nil)

		userRepo.On("FindByIDUnscope", ctx, tx, database.Text(teacherID)).Once().
			Return(&entities.User{}, nil)

		userRepo.On("Update", ctx, tx, mock.AnythingOfType("*entities.User")).Once().
			Return(pgx.ErrTxClosed)

		err := s.SyncTeacher(ctx, []*npb.EventUserRegistration_Staff{req})
		assert.EqualError(t, err, fmt.Sprintf("s.UpsertTeacher teacherID %s: err Update: tx is closed", teacherID))
	})

	t.Run("success update user record", func(t *testing.T) {
		teacherID := idutil.ULIDNow()

		req := &npb.EventUserRegistration_Staff{
			ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			StaffId:    teacherID,
			Name:       "Teacher",
		}

		teacherRepo.On("FindRegardlessDeletion", ctx, db, database.Text(req.StaffId)).Once().
			Return(&entities.Teacher{
				ID: database.Text(teacherID),
			}, nil)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		teacherRepo.On("Update", ctx, tx, mock.AnythingOfType("*entities.Teacher")).Once().
			Return(nil)

		userRepo.On("FindByIDUnscope", ctx, tx, database.Text(teacherID)).Once().
			Return(&entities.User{}, nil)

		userRepo.On("Update", ctx, tx, mock.AnythingOfType("*entities.User")).Once().
			Return(nil)

		err := s.SyncTeacher(ctx, []*npb.EventUserRegistration_Staff{req})
		assert.Nil(t, err)
	})

	t.Run("err delete teacher records", func(t *testing.T) {
		teacherID := idutil.ULIDNow()

		req := &npb.EventUserRegistration_Staff{
			ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
			StaffId:    teacherID,
			Name:       "Teacher",
		}

		teacherRepo.On("FindRegardlessDeletion", ctx, db, database.Text(req.StaffId)).Once().
			Return(&entities.Teacher{
				ID: database.Text(teacherID),
			}, nil)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		teacherRepo.On("SoftDelete", ctx, tx, database.TextArray([]string{teacherID})).Once().
			Return(pgx.ErrTxClosed)

		err := s.SyncTeacher(ctx, []*npb.EventUserRegistration_Staff{req})
		assert.EqualError(t, err, fmt.Sprintf("s.DeleteTeacher teacherIDs %v: s.TeacherRepo.SoftDelete: tx is closed", []string{teacherID}))
	})

	t.Run("err delete user records", func(t *testing.T) {
		teacherID := idutil.ULIDNow()

		req := &npb.EventUserRegistration_Staff{
			ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
			StaffId:    teacherID,
			Name:       "Teacher",
		}

		teacherRepo.On("FindRegardlessDeletion", ctx, db, database.Text(req.StaffId)).Once().
			Return(&entities.Teacher{
				ID: database.Text(teacherID),
			}, nil)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		teacherRepo.On("SoftDelete", ctx, tx, database.TextArray([]string{teacherID})).Once().
			Return(nil)

		userRepo.On("SoftDelete", ctx, tx, database.TextArray([]string{teacherID})).Once().
			Return(pgx.ErrTxClosed)

		err := s.SyncTeacher(ctx, []*npb.EventUserRegistration_Staff{req})
		assert.EqualError(t, err, fmt.Sprintf("s.DeleteTeacher teacherIDs %v: s.UserRepo.SoftDelete: tx is closed", []string{teacherID}))
	})

	t.Run("success delete teachers", func(t *testing.T) {
		teacherID := idutil.ULIDNow()

		req := &npb.EventUserRegistration_Staff{
			ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
			StaffId:    teacherID,
			Name:       "Teacher",
		}

		teacherRepo.On("FindRegardlessDeletion", ctx, db, database.Text(req.StaffId)).Once().
			Return(&entities.Teacher{
				ID: database.Text(teacherID),
			}, nil)

		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		teacherRepo.On("SoftDelete", ctx, tx, database.TextArray([]string{teacherID})).Once().
			Return(nil)

		userRepo.On("SoftDelete", ctx, tx, database.TextArray([]string{teacherID})).Once().
			Return(nil)

		err := s.SyncTeacher(ctx, []*npb.EventUserRegistration_Staff{req})
		assert.Nil(t, err)
	})
}
