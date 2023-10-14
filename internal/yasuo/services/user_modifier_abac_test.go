package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestABACUpdateUser(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userId := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userId)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupSchoolAdmin)

	useRepo := new(mock_repositories.MockUserRepo)
	schoolAdminRepo := new(mock_repositories.MockSchoolAdminRepo)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)

	db := new(mock_database.Ext)

	s := UserModifierServiceABAC{
		&UserModifierService{
			DB:              db,
			UserRepo:        useRepo,
			SchoolAdminRepo: schoolAdminRepo,
			TeacherRepo:     teacherRepo,
			StudentRepo:     studentRepo,
		},
	}

	testCases := []TestCase{
		{
			name: "invalid params",
			ctx:  ctx,
			req: &pb.UpdateUserProfileRequest{
				Id:    "",
				Name:  "name",
				Grade: 1,
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid params"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "user group current user",
			ctx:  ctx,
			req: &pb.UpdateUserProfileRequest{
				Id:    "id",
				Name:  "name",
				Grade: 1,
			},
			expectedErr: fmt.Errorf("s.UserRepo.UserGroup: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				useRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return("", pgx.ErrNoRows)
			},
		},
		{
			name: "school admin not found",
			ctx:  ctx,
			req: &pb.UpdateUserProfileRequest{
				Id:   "id",
				Name: "name",

				Grade: 1,
			},
			expectedErr: fmt.Errorf("s.SchoolAdminRepo.Get: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				useRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupSchoolAdmin, nil)
				schoolAdminRepo.On("Get", ctx, mock.Anything, database.Text(userId)).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "teacher different school",
			ctx:  ctx,
			req: &pb.UpdateUserProfileRequest{
				Id:    "id",
				Name:  "name",
				Grade: 1,
			},
			expectedErr: status.Error(codes.PermissionDenied, "school staff only update their teacher"),
			setup: func(ctx context.Context) {
				useRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupSchoolAdmin, nil)
				schoolAdminRepo.On("Get", ctx, mock.Anything, database.Text(userId)).Once().Return(&entities.SchoolAdmin{SchoolAdminID: database.Text(userId), SchoolID: database.Int4(1)}, nil)
				useRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupTeacher, nil)
				teacherRepo.On("FindByID", ctx, mock.Anything, mock.Anything).Once().Return(&entities.Teacher{
					ID:        database.Text("1"),
					SchoolIDs: database.Int4Array([]int32{2}),
				}, nil)
			},
		},
		{
			name: "teacher not found",
			ctx:  ctx,
			req: &pb.UpdateUserProfileRequest{
				Id:    "id",
				Name:  "name",
				Grade: 1,
			},
			expectedErr: fmt.Errorf("s.TeacherRepo.FindByID: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				useRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupSchoolAdmin, nil)
				schoolAdminRepo.On("Get", ctx, mock.Anything, database.Text(userId)).Once().Return(&entities.SchoolAdmin{SchoolAdminID: database.Text(userId), SchoolID: database.Int4(1)}, nil)
				useRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupTeacher, nil)
				teacherRepo.On("FindByID", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "student not found",
			ctx:  ctx,
			req: &pb.UpdateUserProfileRequest{
				Id:   "id",
				Name: "name",

				Grade: 1,
			},
			expectedErr: fmt.Errorf("s.StudentRepo.Find: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				useRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupSchoolAdmin, nil)
				schoolAdminRepo.On("Get", ctx, mock.Anything, database.Text(userId)).Once().Return(&entities.SchoolAdmin{SchoolAdminID: database.Text(userId), SchoolID: database.Int4(1)}, nil)
				useRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupStudent, nil)
				studentRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "student different school",
			ctx:  ctx,
			req: &pb.UpdateUserProfileRequest{
				Id:   "id",
				Name: "name",

				Grade: 1,
			},
			expectedErr: status.Error(codes.PermissionDenied, "school staff only update their student"),
			setup: func(ctx context.Context) {
				useRepo.On("UserGroup", ctx, mock.Anything, database.Text(userId)).Once().Return(entities.UserGroupSchoolAdmin, nil)
				schoolAdminRepo.On("Get", ctx, mock.Anything, database.Text(userId)).Once().Return(&entities.SchoolAdmin{SchoolAdminID: database.Text(userId), SchoolID: database.Int4(1)}, nil)
				useRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupStudent, nil)
				studentRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(&entities.Student{
					ID:       database.Text("id"),
					SchoolID: database.Int4(2),
				}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Log("Test case: " + testCase.name)
		testCase.setup(testCase.ctx)

		_, err := s.UpdateUserProfile(testCase.ctx, testCase.req.(*pb.UpdateUserProfileRequest))

		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
