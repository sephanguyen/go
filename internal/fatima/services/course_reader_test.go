package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/fatima/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	ubp "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mockUserMgmtUserReader struct {
	searchbasicFn func(ctx context.Context, in *ubp.SearchBasicProfileRequest, opts ...grpc.CallOption) (*ubp.SearchBasicProfileResponse, error)
}

func (m *mockUserMgmtUserReader) SearchBasicProfile(ctx context.Context, in *ubp.SearchBasicProfileRequest, opts ...grpc.CallOption) (*ubp.SearchBasicProfileResponse, error) {
	return m.searchbasicFn(ctx, in, opts...)
}

type TestCase struct {
	ctx          context.Context
	name         string
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestCourseReaderService_ListStudentByCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	spapRepo := new(mock_repositories.MockStudentPackageAccessPathRepo)
	mockDB := &mock_database.Ext{}
	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	locationIDs := []string{constants.ManabieOrgLocation}

	t.Run("failed query", func(t *testing.T) {
		s := &CourseReaderService{
			DB:                           mockDB,
			StudentPackageAccessPathRepo: spapRepo,
		}
		testCase := TestCase{
			ctx: ctx,
			req: &pb.ListStudentByCourseRequest{
				CourseId:    "course-id",
				Paging:      &cpb.Paging{Limit: 0, Offset: &cpb.Paging_OffsetInteger{}},
				LocationIds: locationIDs,
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("s.StudentPackageAccessPathRepo.GetByCourseIDAndLocationIDs: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				spapRepo.On("GetByCourseIDAndLocationIDs", ctx, mockDB, database.Text("course-id"), database.TextArray(locationIDs)).Once().Return(nil, pgx.ErrNoRows)
			},
		}
		testCase.setup(testCase.ctx)

		resp, err := s.ListStudentByCourse(testCase.ctx, testCase.req.(*pb.ListStudentByCourseRequest))
		if testCase.expectedErr == nil {
			assert.NoError(t, err, "expecting no error")
		} else {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error(), "unexpected error message")
		}

		if testCase.expectedResp == nil {
			assert.Nil(t, testCase.expectedResp, resp)
		} else {
			assert.Equal(t, testCase.expectedResp, resp)
		}
	})
	t.Run("failed call usermgmt `GetByCourseIDAndLocationIDs`", func(t *testing.T) {
		userMgmtService := &mockUserMgmtUserReader{searchbasicFn: func(ctx context.Context, in *ubp.SearchBasicProfileRequest, opts ...grpc.CallOption) (*ubp.SearchBasicProfileResponse, error) {
			return nil, fmt.Errorf("usermgmt failed")
		}}
		s := &CourseReaderService{
			DB:                           mockDB,
			UserMgmtUserReader:           userMgmtService,
			StudentPackageAccessPathRepo: spapRepo,
		}
		testCase := TestCase{
			ctx: ctx,
			req: &pb.ListStudentByCourseRequest{
				CourseId:    "course-id",
				Paging:      &cpb.Paging{Limit: 0},
				LocationIds: locationIDs,
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("s.UserMgmtUserReader.SearchBasicProfile: usermgmt failed"),
			setup: func(ctx context.Context) {
				spapRepo.On("GetByCourseIDAndLocationIDs", ctx, mockDB, database.Text("course-id"), database.TextArray(locationIDs)).Once().Return([]*entities.StudentPackageAccessPath{}, nil)
			},
		}
		testCase.setup(testCase.ctx)
		resp, err := s.ListStudentByCourse(testCase.ctx, testCase.req.(*pb.ListStudentByCourseRequest))
		if testCase.expectedErr == nil {
			assert.NoError(t, err, "expecting no error")
		} else {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error(), "unexpected error message")
		}

		if testCase.expectedResp == nil {
			assert.Nil(t, testCase.expectedResp, resp)
		} else {
			assert.Equal(t, testCase.expectedResp, resp)
		}
	})
	t.Run("happy case`", func(t *testing.T) {
		userMgmtService := &mockUserMgmtUserReader{searchbasicFn: func(ctx context.Context, in *ubp.SearchBasicProfileRequest, opts ...grpc.CallOption) (*ubp.SearchBasicProfileResponse, error) {
			return &ubp.SearchBasicProfileResponse{Profiles: []*cpb.BasicProfile{{UserId: "usr1"}}, NextPage: &cpb.Paging{}}, nil
		}}
		s := &CourseReaderService{
			DB:                           mockDB,
			UserMgmtUserReader:           userMgmtService,
			StudentPackageAccessPathRepo: spapRepo,
		}
		testCase := TestCase{
			ctx: ctx,
			req: &pb.ListStudentByCourseRequest{
				CourseId:    "course-id",
				Paging:      &cpb.Paging{Limit: 0},
				LocationIds: locationIDs,
			},
			expectedResp: &pb.ListStudentByCourseResponse{Profiles: []*cpb.BasicProfile{{UserId: "usr1"}}, NextPage: &cpb.Paging{}},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				spapRepo.On("GetByCourseIDAndLocationIDs", ctx, mockDB, database.Text("course-id"), database.TextArray(locationIDs)).Once().
					Return([]*entities.StudentPackageAccessPath{
						{
							StudentID: database.Text("student-id"),
						},
					}, nil)
			},
		}
		testCase.setup(testCase.ctx)
		resp, err := s.ListStudentByCourse(testCase.ctx, testCase.req.(*pb.ListStudentByCourseRequest))
		if testCase.expectedErr == nil {
			assert.NoError(t, err, "expecting no error")
		} else {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error(), "unexpected error message")
		}
		if testCase.expectedResp == nil {
			assert.Nil(t, testCase.expectedResp, resp)
		} else {
			assert.Equal(t, testCase.expectedResp, resp)
		}
	})
}
