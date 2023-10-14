package users

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func generateEnUser() *entities_bob.User {
	rand.Seed(time.Now().UnixNano())
	e := new(entities_bob.User)
	e.Avatar.Set(fmt.Sprintf("http://avatar-%d", rand.Int()))
	e.Group.Set(entities_bob.UserGroupAdmin)
	e.LastName.Set(fmt.Sprintf("user %d", rand.Int()))
	e.Country.Set("COUNTRY_VN")
	e.PhoneNumber.Set(fmt.Sprintf("+849%d", rand.Int()))
	e.Email.Set(fmt.Sprintf("valid-%d@email.com", rand.Int()))
	e.DeviceToken.Set(fmt.Sprintf("random device %d", rand.Int()))
	e.CreatedAt.Set(time.Now())
	e.UpdatedAt.Set(time.Now())
	return e
}

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
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
			req: &bpb.SearchBasicProfileRequest{SearchText: &wrapperspb.StringValue{
				Value: "abc",
			}, UserIds: []string{e.UserID.String}, Paging: &cpb.Paging{}},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("SearchProfile", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities_bob.User{e}, nil)
			},
		},
		{
			name:        "error query",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &bpb.SearchBasicProfileRequest{UserIds: []string{}, Paging: &cpb.Paging{}},
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				userRepo.On("SearchProfile", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities_bob.User{}, pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*bpb.SearchBasicProfileRequest)
			_, err := s.SearchBasicProfile(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestUserReaderService_RetrieveBasicProfile(t *testing.T) {
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
			name:        "happy case",
			ctx:         ctx,
			req:         &bpb.RetrieveBasicProfileRequest{UserIds: []string{e.UserID.String}},
			expectedErr: nil,
			expectedResp: []*cpb.BasicProfile{{
				Name:    e.GivenName.String,
				Group:   cpb.UserGroup(cpb.UserGroup_value[e.Group.String]),
				Country: cpb.Country(cpb.Country_value[e.Country.String]),
			}},
			setup: func(ctx context.Context) {
				userRepo.
					On("Retrieve", ctx, db, database.TextArray([]string{e.UserID.String}), mock.Anything).
					Return([]*entities_bob.User{e}, nil).
					Once()
			},
		},
		{
			name:        "fail due to user repo Retrieve return error",
			ctx:         ctx,
			req:         &bpb.RetrieveBasicProfileRequest{UserIds: []string{e.UserID.String}},
			expectedErr: services.ToStatusError(pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				userRepo.
					On("Retrieve", ctx, db, database.TextArray([]string{e.UserID.String}), mock.Anything).
					Return(nil, pgx.ErrNoRows).
					Once()
			},
		},
		{
			name:        "request user IDs empty",
			ctx:         ctx,
			req:         &bpb.RetrieveBasicProfileRequest{UserIds: []string{}},
			expectedErr: status.Error(codes.InvalidArgument, "missing userIds"),
			setup:       func(ctx context.Context) {},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*bpb.RetrieveBasicProfileRequest)
			resp, err := s.RetrieveBasicProfile(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				expectedResp := testCase.expectedResp.([]*cpb.BasicProfile)
				assert.Equal(t, len(expectedResp), len(resp.Profiles))
				assert.Equal(t, expectedResp[0].GivenName, resp.Profiles[0].GivenName)
				assert.Equal(t, expectedResp[0].Group, resp.Profiles[0].Group)
				assert.Equal(t, expectedResp[0].Country, resp.Profiles[0].Country)
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, db, userRepo)
		})
	}
}
