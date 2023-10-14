package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestUserGRPCService_GetTeachers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	teacherRepo := new(mock_repositories.MockTeacherRepo)

	tcs := []struct {
		name     string
		req      *lpb.GetTeachersRequest
		res      *lpb.GetTeachersResponse
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "get successfully",
			req: &lpb.GetTeachersRequest{
				TeacherIds: []string{"user-id-1", "user-id-2", "user-id-3"},
			},
			res: &lpb.GetTeachersResponse{
				Teachers: []*lpb.GetTeachersResponse_TeacherInfo{
					{
						Id:        "user-id-1",
						Name:      "name 1",
						CreatedAt: timestamppb.New(now),
						UpdatedAt: timestamppb.New(now),
					},
					{
						Id:        "user-id-2",
						Name:      "name 2",
						CreatedAt: timestamppb.New(now),
						UpdatedAt: timestamppb.New(now),
					},
					{
						Id:        "user-id-3",
						Name:      "name 3",
						CreatedAt: timestamppb.New(now),
						UpdatedAt: timestamppb.New(now),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				teacherRepo.
					On("ListByIDs", ctx, db, []string{"user-id-1", "user-id-2", "user-id-3"}).
					Return(domain.Teachers{
						{
							ID:        "user-id-1",
							Name:      "name 1",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        "user-id-2",
							Name:      "name 2",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        "user-id-3",
							Name:      "name 3",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).Once()
			},
		},
		{
			name: "could not found some teachers",
			req: &lpb.GetTeachersRequest{
				TeacherIds: []string{"user-id-1", "user-id-2", "user-id-3"},
			},
			res: &lpb.GetTeachersResponse{
				Teachers: []*lpb.GetTeachersResponse_TeacherInfo{
					{
						Id:        "user-id-1",
						Name:      "name 1",
						CreatedAt: timestamppb.New(now),
						UpdatedAt: timestamppb.New(now),
					},
					{
						Id:        "user-id-3",
						Name:      "name 3",
						CreatedAt: timestamppb.New(now),
						UpdatedAt: timestamppb.New(now),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				teacherRepo.
					On("ListByIDs", ctx, db, []string{"user-id-1", "user-id-2", "user-id-3"}).
					Return(domain.Teachers{
						{
							ID:        "user-id-1",
							Name:      "name 1",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        "user-id-3",
							Name:      "name 3",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).Once()
			},
		},
		{
			name: "return invalid teacher info",
			req: &lpb.GetTeachersRequest{
				TeacherIds: []string{"user-id-1", "user-id-2", "user-id-3"},
			},
			res: &lpb.GetTeachersResponse{
				Teachers: []*lpb.GetTeachersResponse_TeacherInfo{
					{
						Id:        "user-id-1",
						Name:      "name 1",
						CreatedAt: timestamppb.New(now),
						UpdatedAt: timestamppb.New(now),
					},
					{
						Id:        "user-id-2",
						Name:      "name 2",
						CreatedAt: timestamppb.New(now),
						UpdatedAt: timestamppb.New(now),
					},
					{
						Id:        "user-id-3",
						Name:      "name 3",
						CreatedAt: timestamppb.New(now),
						UpdatedAt: timestamppb.New(now),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				teacherRepo.
					On("ListByIDs", ctx, db, []string{"user-id-1", "user-id-2", "user-id-3"}).
					Return(domain.Teachers{
						{
							ID:        "user-id-1",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        "user-id-2",
							Name:      "name 2",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        "user-id-3",
							Name:      "name 3",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			srv := NewUserGRPCService(db, wrapperConnection, teacherRepo, nil, nil)
			actual, err := srv.GetTeachers(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, tc.res, actual)
			}
			mock.AssertExpectationsForObjects(
				t,
				db,
				teacherRepo,
				mockUnleashClient,
			)
		})
	}
}

func TestUserGRPCService_GetUserGroup(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	userRepo := new(mock_repositories.MockUserRepo)
	tcs := []struct {
		name     string
		req      *lpb.GetUserGroupRequest
		res      *lpb.GetUserGroupResponse
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "get successfully",
			req: &lpb.GetUserGroupRequest{
				UserId: "user-id-1",
			},
			res: &lpb.GetUserGroupResponse{
				UserGroup: "user-gr-1",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				userRepo.On("GetUserGroupByUserID", ctx, db, "user-id-1").
					Return("user-gr-1", nil).Once()
			},
		},
		{
			name: "could not found user",
			req: &lpb.GetUserGroupRequest{
				UserId: "user-id-1",
			},
			res: &lpb.GetUserGroupResponse{
				UserGroup: "user-gr-1",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				userRepo.On("GetUserGroupByUserID", ctx, db, "user-id-1").
					Return("", fmt.Errorf("could not found")).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			srv := NewUserGRPCService(db, wrapperConnection, nil, userRepo, nil)
			res, err := srv.GetUserGroup(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.res, res)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}

func TestUserGRPCService_GetTeachersSameGrantedLocation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	userRepo := new(mock_repositories.MockUserRepo)
	userBasicInfoRepo := new(mock_repositories.MockUserBasicInfoRepo)
	keyword := "123"
	locationID := "1223"
	limit := 10
	page := 0
	tcs := []struct {
		name     string
		req      *lpb.GetTeachersSameGrantedLocationRequest
		res      *lpb.GetTeachersSameGrantedLocationResponse
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "get successfully",
			req: &lpb.GetTeachersSameGrantedLocationRequest{
				Keyword:      keyword,
				LocationId:   locationID,
				IsAllTeacher: true,
				Paging: &cpb.Paging{
					Limit: uint32(limit),
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: int64(limit * page),
					},
				},
			},
			res: &lpb.GetTeachersSameGrantedLocationResponse{
				Teachers: []*lpb.GetTeachersSameGrantedLocationResponse_TeacherInfo{
					{
						Id:   "UserID",
						Name: "Name",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				userBasicInfoRepo.On("GetTeachersSameGrantedLocation", ctx, db, domain.UserBasicInfoQuery{
					KeyWord:    keyword,
					LocationID: "",
					Offset:     limit * page,
					Limit:      limit,
				}).
					Return(domain.UsersBasicInfo{
						&domain.UserBasicInfo{
							UserID:   "UserID",
							FullName: "Name",
						},
					}, 1, nil).Once()
			},
		},
		{
			name: "not found",
			req: &lpb.GetTeachersSameGrantedLocationRequest{
				Keyword:      "123",
				LocationId:   "1223",
				IsAllTeacher: false,
				Paging: &cpb.Paging{
					Limit: uint32(limit),
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: int64(limit * page),
					},
				},
			},
			res: &lpb.GetTeachersSameGrantedLocationResponse{
				Teachers: []*lpb.GetTeachersSameGrantedLocationResponse_TeacherInfo{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				userBasicInfoRepo.On("GetTeachersSameGrantedLocation", ctx, db, domain.UserBasicInfoQuery{
					KeyWord:    keyword,
					LocationID: "1223",
					Offset:     limit * page,
					Limit:      limit,
				}).
					Return(domain.UsersBasicInfo{}, 0, nil).Once()
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			srv := NewUserGRPCService(db, wrapperConnection, nil, userRepo, userBasicInfoRepo)
			_, err := srv.GetTeachersSameGrantedLocation(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}

func TestUserGRPCService_GetStudentsManyReferenceByNameOrEmail(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	userRepo := new(mock_repositories.MockUserRepo)
	keyword := "Id"
	limit := uint32(30)
	offset := uint32(0)
	tcs := []struct {
		name     string
		req      *lpb.GetStudentsManyReferenceByNameOrEmailRequest
		res      *lpb.GetStudentsManyReferenceByNameOrEmailResponse
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "get successfully",
			req: &lpb.GetStudentsManyReferenceByNameOrEmailRequest{
				Keyword: keyword,
				Limit:   limit,
				Offset:  offset,
			},
			res: &lpb.GetStudentsManyReferenceByNameOrEmailResponse{
				Students: []*lpb.GetStudentsManyReferenceByNameOrEmailResponse_StudentInfo{},
			},
			setup: func(ctx context.Context) {
				userRepo.On("GetStudentsManyReferenceByNameOrEmail", ctx, db, keyword, limit, offset).
					Return(domain.Students{
						&domain.Student{
							ID:    "Id",
							Name:  "Name",
							Email: "Email",
						},
					}, nil).Once()
			},
			hasError: false,
		},
		{
			name: "not found",
			req: &lpb.GetStudentsManyReferenceByNameOrEmailRequest{
				Keyword: keyword,
				Limit:   limit,
				Offset:  offset,
			},
			res: &lpb.GetStudentsManyReferenceByNameOrEmailResponse{
				Students: []*lpb.GetStudentsManyReferenceByNameOrEmailResponse_StudentInfo{},
			},
			setup: func(ctx context.Context) {
				userRepo.On("GetStudentsManyReferenceByNameOrEmail", ctx, db, keyword, limit, offset).
					Return(domain.Students{
						&domain.Student{}}, nil).Once()
			},
			hasError: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			srv := NewUserGRPCService(db, wrapperConnection, nil, userRepo, nil)
			_, err := srv.GetStudentsManyReferenceByNameOrEmail(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}
