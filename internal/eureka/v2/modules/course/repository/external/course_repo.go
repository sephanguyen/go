package external

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/domain"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CourseRepo struct {
	CourseClient mpb.MasterDataCourseServiceClient
}

func (c *CourseRepo) Upsert(ctx context.Context, courses []domain.Course) ([]domain.Course, error) {
	bobCoursesRequest := convertDomainToUpsertCoursesRequest(courses)

	mdctx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, fmt.Errorf("CourseService.UpsertCourses.GetOutgoingContext: %w", err).Error())
	}
	bobResponse, err := c.CourseClient.UpsertCourses(mdctx, bobCoursesRequest)
	if err != nil {
		return nil, err
	}

	if bobResponse.Successful {
		return courses, nil
	}
	return nil, nil
}

func convertDomainToUpsertCoursesRequest(courses []domain.Course) *mpb.UpsertCoursesRequest {
	bobCoursesRequest := make([]*mpb.UpsertCoursesRequest_Course, len(courses))
	for index, courseRequest := range courses {
		bobCoursesRequest[index] = convertCourseRequestToBobCourseRequest(courseRequest)
	}
	return &mpb.UpsertCoursesRequest{Courses: bobCoursesRequest}
}

func convertCourseRequestToBobCourseRequest(course domain.Course) *mpb.UpsertCoursesRequest_Course {
	req := mpb.UpsertCoursesRequest_Course{
		Id:             course.ID,
		Name:           course.Name,
		DisplayOrder:   int32(course.DisplayOrder),
		Icon:           course.Icon,
		TeachingMethod: convertTeachingMethod(course.TeachingMethod),
		CourseType:     course.CourseTypeID,
		LocationIds:    course.LocationIDs,
		SubjectIds:     course.SubjectIDs,
	}
	return &req
}

func convertTeachingMethod(method string) mpb.CourseTeachingMethod {
	switch method {
	case "COURSE_TEACHING_METHOD_INDIVIDUAL":
		return mpb.CourseTeachingMethod_COURSE_TEACHING_METHOD_INDIVIDUAL
	case "COURSE_TEACHING_METHOD_GROUP":
		return mpb.CourseTeachingMethod_COURSE_TEACHING_METHOD_GROUP
	default:
		return mpb.CourseTeachingMethod_COURSE_TEACHING_METHOD_NONE
	}
}
