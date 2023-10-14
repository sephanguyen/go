package services

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_noti_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetBasicProfile(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepo := new(mock_repositories.MockUserRepo)
	e := generateEnUser()
	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         ctx,
			req:         &pb.GetBasicProfileRequest{UserIds: []string{"1"}},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.User{e}, nil)
			},
		},
		{
			name:        "fail due to user repo Retrieve return error",
			ctx:         ctx,
			req:         &pb.GetBasicProfileRequest{UserIds: []string{"1"}},
			expectedErr: toStatusError(pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "request user IDs empty",
			ctx:         ctx,
			req:         &pb.GetBasicProfileRequest{UserIds: []string{}},
			expectedErr: status.Error(codes.InvalidArgument, "missing userIds"),
			setup:       func(ctx context.Context) {},
		},
	}

	s := &UserService{
		UserRepo: userRepo,
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.GetBasicProfileRequest)
			_, err := s.GetBasicProfile(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestUpdateUserProfile(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepo := new(mock_repositories.MockUserRepo)
	jsm := new(mock_nats.JetStreamManagement)

	s := &UserService{
		UserRepo: userRepo,
		JSM:      jsm,
	}

	p := generateUserProfile()
	emptyIdP := generateUserProfile()
	emptyIdP.Id = ""
	testCases := []TestCase{
		{
			name:        "happy case admin",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &pb.UpdateUserProfileRequest{Profile: p},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupAdmin, nil)
				userRepo.On("UpdateProfile", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectUserDeviceTokenUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "happy case student",
			ctx:         interceptors.ContextWithUserID(ctx, p.Id),
			req:         &pb.UpdateUserProfileRequest{Profile: p},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupStudent, nil)
				userRepo.On("UpdateProfile", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectUserDeviceTokenUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "empty user id",
			ctx:         interceptors.ContextWithUserID(ctx, "empty"),
			req:         &pb.UpdateUserProfileRequest{Profile: emptyIdP},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupStudent, nil)
				userRepo.On("UpdateProfile", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectUserDeviceTokenUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "student update other profile",
			ctx:         interceptors.ContextWithUserID(ctx, "studentId"),
			req:         &pb.UpdateUserProfileRequest{Profile: p},
			expectedErr: status.Error(codes.PermissionDenied, "user can only update own profile"),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupStudent, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.UpdateUserProfileRequest)
			_, err := s.UpdateUserProfile(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestUpdateUserDeviceToken(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockDB.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("Commit", mock.Anything).Return(nil)

	userRepo := new(mock_repositories.MockUserRepo)
	userDeviceTokenRepo := new(mock_noti_repositories.MockUserDeviceTokenRepo)
	jsm := new(mock_nats.JetStreamManagement)

	s := &UserService{
		DB:                  mockDB,
		UserRepo:            userRepo,
		JSM:                 jsm,
		UserDeviceTokenRepo: userDeviceTokenRepo,
	}

	userModel := &entities.User{}
	database.AllNullEntity(userModel)
	userModel.ID.Set("id")
	userModel.GivenName.Set("given-name")
	userModel.LastName.Set("last-name")
	userModel.DeviceToken.Set("device")
	userModel.AllowNotification.Set(true)
	userModels := []*entities.User{userModel}

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "userid"),
			req:         &pb.UpdateUserDeviceTokenRequest{UserId: "id", DeviceToken: "device", AllowNotification: true},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{"id"}), mock.Anything).Once().Return(userModels, nil)
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupAdmin, nil)
				// userRepo.On("StoreDeviceToken", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				userDeviceTokenRepo.On("UpsertUserDeviceToken", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectUserDeviceTokenUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "empty user id",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &pb.UpdateUserDeviceTokenRequest{UserId: "", DeviceToken: "device", AllowNotification: true},
			expectedErr: status.Error(codes.InvalidArgument, "userID or device token empty"),
			setup: func(ctx context.Context) {
				userRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{"id"}), mock.Anything).Once().Return(userModels, nil)
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupAdmin, nil)
				// userRepo.On("StoreDeviceToken", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				userDeviceTokenRepo.On("UpsertUserDeviceToken", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectUserDeviceTokenUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "empty device",
			ctx:         interceptors.ContextWithUserID(ctx, "empty device"),
			req:         &pb.UpdateUserDeviceTokenRequest{UserId: "id", DeviceToken: "", AllowNotification: true},
			expectedErr: status.Error(codes.InvalidArgument, "userID or device token empty"),
			setup: func(ctx context.Context) {
				userRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{"id"})).Once().Return(userModels, nil)
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupAdmin, nil)
				// userRepo.On("StoreDeviceToken", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				userDeviceTokenRepo.On("UpsertUserDeviceToken", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectUserDeviceTokenUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "error update",
			ctx:         interceptors.ContextWithUserID(ctx, "error update"),
			req:         &pb.UpdateUserDeviceTokenRequest{UserId: "id", DeviceToken: "device", AllowNotification: true},
			expectedErr: fmt.Errorf("s.UserDeviceTokenRepo.UpsertUserDeviceToken: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				userRepo.On("Retrieve", ctx, mock.Anything, database.TextArray([]string{"id"}), mock.Anything).Once().Return(userModels, nil)
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupAdmin, nil)
				// userRepo.On("StoreDeviceToken", ctx, tx, mock.Anything).Once().Return(pgx.ErrNoRows)
				userDeviceTokenRepo.On("UpsertUserDeviceToken", ctx, tx, mock.Anything).Once().Return(pgx.ErrNoRows)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectUserDeviceTokenUpdated, mock.Anything).Once().Return("", nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.UpdateUserDeviceTokenRequest)
			_, err := s.UpdateUserDeviceToken(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestToUserEntity(t *testing.T) {
	t.Parallel()
	profile1 := generateUserProfile()
	profile2 := generateUserProfile()
	profile2.CreatedAt = nil
	profile2.UpdatedAt = nil
	e1 := toUserEntity(profile1)
	e2 := toUserEntity(profile2)
	assert.True(t, isEqualUserEnAndPb(profile1, e1))
	assert.True(t, isEqualUserEnAndPb(profile2, e2))
}

func TestToUserProfilePb(t *testing.T) {
	t.Parallel()
	e := generateEnUser()
	p := toUserProfilePb(e)
	assert.True(t, isEqualUserEnAndPb(p, e))
}

func isEqualUserEnAndPb(p *pb.UserProfile, e *entities.User) bool {
	if p.CreatedAt == nil {
		p.CreatedAt = &types.Timestamp{Seconds: e.UpdatedAt.Time.Unix()}
	}
	if p.UpdatedAt == nil {
		p.UpdatedAt = &types.Timestamp{Seconds: e.UpdatedAt.Time.Unix()}
	}

	updatedAt := &types.Timestamp{Seconds: e.UpdatedAt.Time.Unix()}
	createdAt := &types.Timestamp{Seconds: e.CreatedAt.Time.Unix()}

	return (e.GetName() == p.Name) &&
		(pb.Country(pb.Country_value[e.Country.String]) == p.Country) &&
		(e.PhoneNumber.String == p.PhoneNumber) &&
		(e.Email.String == p.Email) &&
		(e.Avatar.String == p.Avatar) &&
		(e.DeviceToken.String == p.DeviceToken) &&
		(e.Group.String == p.UserGroup) &&
		updatedAt.Equal(p.UpdatedAt) &&
		createdAt.Equal(p.CreatedAt)
}

func generateUserProfile() *pb.UserProfile {
	rand.Seed(time.Now().UnixNano())
	return &pb.UserProfile{
		Id:          fmt.Sprintf("%d", rand.Int()),
		Name:        fmt.Sprintf("user %d", rand.Int()),
		Country:     pb.COUNTRY_VN,
		PhoneNumber: fmt.Sprintf("+849%d", rand.Int()),
		Email:       fmt.Sprintf("valid-%d@email.com", rand.Int()),
		Avatar:      fmt.Sprintf("http://avatar-%d", rand.Int()),
		DeviceToken: fmt.Sprintf("random device %d", rand.Int()),
		UserGroup:   entities.UserGroupStudent,
		CreatedAt:   &types.Timestamp{Seconds: time.Now().Unix()},
		UpdatedAt:   &types.Timestamp{Seconds: time.Now().Unix()},
	}
}

func generateEnUser() *entities.User {
	rand.Seed(time.Now().UnixNano())
	e := new(entities.User)
	e.Avatar.Set(fmt.Sprintf("http://avatar-%d", rand.Int()))
	e.Group.Set(entities.UserGroupStudent)
	e.LastName.Set(fmt.Sprintf("user %d", rand.Int()))
	e.Country.Set("COUNTRY_VN")
	e.PhoneNumber.Set(fmt.Sprintf("+849%d", rand.Int()))
	e.Email.Set(fmt.Sprintf("valid-%d@email.com", rand.Int()))
	e.DeviceToken.Set(fmt.Sprintf("random device %d", rand.Int()))
	e.CreatedAt.Set(time.Now())
	e.UpdatedAt.Set(time.Now())
	return e
}

func TestGetCurrentUserProfile(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepo := new(mock_repositories.MockUserRepo)
	e := generateEnUser()
	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &pb.GetCurrentUserProfileRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.User{e}, nil)
			},
		},
		{
			name:        "error query",
			ctx:         interceptors.ContextWithUserID(ctx, "errQuery"),
			req:         &pb.GetCurrentUserProfileRequest{},
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupTeacher, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "cant find profile",
			ctx:         interceptors.ContextWithUserID(ctx, "cant find profile"),
			req:         &pb.GetCurrentUserProfileRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "user does not exist"),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupTeacher, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
	}

	s := &UserService{
		UserRepo: userRepo,
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.GetCurrentUserProfileRequest)
			_, err := s.GetCurrentUserProfile(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}
