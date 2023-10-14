package grpc

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/transport"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs"
	mock_usecase "github.com/manabie-com/backend/mock/eureka/v2/modules/course/usecase"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	ctx          context.Context
	name         string
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestCourseService_UpsertCourses(t *testing.T) {
	t.Parallel()

	ctx := golibs.ResourcePathToCtx(context.Background(), "-113")

	courseUsecase := &mock_usecase.MockCourseUsecase{}

	courseSvc := &CourseService{
		UpsertCoursesUsecase: courseUsecase,
		ListCoursesUsecase:   courseUsecase,
	}

	testCases := map[string]TestCase{
		"missing name": {
			req: &pb.UpsertCoursesRequest{
				Courses: []*pb.UpsertCoursesRequest_Course{{
					CourseId: "id",
					Name:     "",
				}},
			},
			expectedErr: errors.NewGrpcError(errors.NewValidationError("name cannot be empty", nil), transport.GrpcErrorMap),
		},
		"happy case": {
			req: &pb.UpsertCoursesRequest{
				Courses: []*pb.UpsertCoursesRequest_Course{{
					CourseId: "id",
					Name:     "name",
				}},
			},
			setup: func(ctx context.Context) {
				courseUsecase.On("UpsertCourses", ctx, mock.Anything).
					Once().
					Run(func(args mock.Arguments) {
						courses := args[1].([]domain.Course)
						assert.Equal(t, "id", courses[0].ID)
						assert.Equal(t, "name", courses[0].Name)
					}).
					Return(nil)
			},
			expectedResp: &pb.UpsertCoursesResponse{
				CourseIds: []string{"id"},
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(ctx)
			}

			resp, err := courseSvc.UpsertCourses(ctx, testCase.req.(*pb.UpsertCoursesRequest))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedResp.(*pb.UpsertCoursesResponse), resp)
			}
		})
	}
}

func TestCourseService_ListCourses(t *testing.T) {
	t.Parallel()

	ctx := golibs.ResourcePathToCtx(context.Background(), "-113")

	courseUsecase := &mock_usecase.MockCourseUsecase{}

	courseSvc := &CourseService{
		UpsertCoursesUsecase: courseUsecase,
		ListCoursesUsecase:   courseUsecase,
	}

	testCases := map[string]TestCase{
		"missing ids": {
			req: &pb.ListCoursesByIdsRequest{
				Ids: []string{},
			},
			setup: func(ctx context.Context) {
				courseUsecase.On("ListCourses", ctx, mock.Anything).
					Once().
					Run(func(args mock.Arguments) {
						courseIds := args[1].([]string)
						assert.Empty(t, courseIds)
					}).
					Return(nil, nil)
			},
			expectedResp: &pb.ListCoursesByIdsResponse{},
		},
		"happy case": {
			req: &pb.ListCoursesByIdsRequest{
				Ids: []string{},
			},
			setup: func(ctx context.Context) {
				courses := []domain.Course{
					{
						ID: "course-id-1",
					},
					{
						ID: "course-id-2",
					},
				}

				courseUsecase.On("ListCourses", ctx, mock.Anything).
					Once().
					Run(func(args mock.Arguments) {
						courseIds := args[1].([]string)
						assert.Equal(t, courses[0].ID, courseIds[0])
						assert.Equal(t, courses[1].ID, courseIds[1])
					}).
					Return(courses, nil)
			},
			expectedResp: &pb.ListCoursesByIdsResponse{},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(ctx)
			}

			resp, err := courseSvc.ListCoursesByIds(ctx, testCase.req.(*pb.ListCoursesByIdsRequest))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedResp.(*pb.ListCoursesByIdsResponse), resp)
			}
		})
	}
}
