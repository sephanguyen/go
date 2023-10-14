package mock_postgres

import (
	"context"

	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockMasterDataCourseClient struct {
	mock.Mock
}

func (r *MockMasterDataCourseClient) UpsertCourses(ctx context.Context, in *mpb.UpsertCoursesRequest, opts ...grpc.CallOption) (*mpb.UpsertCoursesResponse, error) {
	args := r.Called(ctx, in)
	return args.Get(0).(*mpb.UpsertCoursesResponse), args.Error(1)
}

func (r *MockMasterDataCourseClient) ExportCourses(ctx context.Context, in *mpb.ExportCoursesRequest, opts ...grpc.CallOption) (*mpb.ExportCoursesResponse, error) {
	args := r.Called(ctx, in)
	return args.Get(0).(*mpb.ExportCoursesResponse), args.Error(1)
}

func (r *MockMasterDataCourseClient) GetCoursesByIDs(ctx context.Context, in *mpb.GetCoursesByIDsRequest, opts ...grpc.CallOption) (*mpb.GetCoursesByIDsResponse, error) {
	args := r.Called(ctx, in)
	return args.Get(0).(*mpb.GetCoursesByIDsResponse), args.Error(1)
}

func (r *MockMasterDataCourseClient) ImportCourses(ctx context.Context, in *mpb.ImportCoursesRequest, opts ...grpc.CallOption) (*mpb.ImportCoursesResponse, error) {
	args := r.Called(ctx, in)
	return args.Get(0).(*mpb.ImportCoursesResponse), args.Error(1)
}
