package grpc

import (
	"context"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/transport"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/usecase"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"
)

type CourseService struct {
	UpsertCoursesUsecase usecase.UpsertCourse
	ListCoursesUsecase   usecase.ListCourse
}

func NewCourseService(
	coursesUsecase *usecase.CourseUsecase,
) *CourseService {
	return &CourseService{
		UpsertCoursesUsecase: coursesUsecase,
		ListCoursesUsecase:   coursesUsecase,
	}
}

func (b *CourseService) UpsertCourses(ctx context.Context, req *pb.UpsertCoursesRequest) (*pb.UpsertCoursesResponse, error) {
	resourcePath, err := getResourcePath(ctx)
	if err != nil {
		return &pb.UpsertCoursesResponse{}, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}
	courses := make([]domain.Course, len(req.Courses))
	courseIDs := make([]string, len(req.Courses))

	for i, v := range req.Courses {
		course := transformCourseFromPb(v, int(resourcePath))
		if err := validateCourse(course); err != nil {
			return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
		}
		courses[i] = course
		courseIDs[i] = course.ID
	}

	if err := b.UpsertCoursesUsecase.UpsertCourses(ctx, courses); err != nil {
		return &pb.UpsertCoursesResponse{}, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}
	return &pb.UpsertCoursesResponse{
		CourseIds: courseIDs,
	}, nil
}

func getResourcePath(ctx context.Context) (int64, error) {
	resourcePath, err := strconv.ParseInt(golibs.ResourcePathFromCtx(ctx), 10, 32)
	if err != nil {
		return 0, errors.NewConversionError("resource path is invalid", nil)
	}
	return resourcePath, nil
}

func transformCourseFromPb(req *pb.UpsertCoursesRequest_Course, resourcePath int) domain.Course {
	if req.CourseId == "" {
		req.CourseId = idutil.ULIDNow()
	}

	return domain.NewCourse(req.CourseId, req.Name, req.Icon, 1, resourcePath, req.LocationIds, req.CourseTypeId, req.BookId, req.GetTeachingMethod().String(), req.SubjectIds)
}

func validateCourse(course domain.Course) error {
	if strings.TrimSpace(course.Name) == "" {
		return errors.NewValidationError("name cannot be empty", nil)
	}
	return nil
}

func (b *CourseService) ListCoursesByIds(ctx context.Context, req *pb.ListCoursesByIdsRequest) (*pb.ListCoursesByIdsResponse, error) {
	if len(req.Ids) == 0 {
		return &pb.ListCoursesByIdsResponse{}, nil
	}
	courses, err := b.ListCoursesUsecase.ListCourses(ctx, req.Ids)
	if err != nil {
		return &pb.ListCoursesByIdsResponse{}, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	coursesPb := transformCoursesPbFromCourses(courses)
	return &pb.ListCoursesByIdsResponse{
		Courses: coursesPb,
	}, nil
}

func transformCoursesPbFromCourses(courses []*domain.Course) []*pb.ListCoursesByIdsResponse_Course {
	coursePbs := make([]*pb.ListCoursesByIdsResponse_Course, len(courses))
	for i, course := range courses {
		coursePbs[i] = &pb.ListCoursesByIdsResponse_Course{
			Id:      course.ID,
			Name:    course.Name,
			IconUrl: course.Icon,
			BookId:  course.BookID,
		}
	}
	return coursePbs
}
