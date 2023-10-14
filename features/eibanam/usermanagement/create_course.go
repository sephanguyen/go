package usermanagement

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	yasuo_pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	bob_pbv1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	com_pbv1 "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	fatima_pbv1 "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
)

type createCourseFeature struct{}

func (f *createCourseFeature) addCourseReq(s *suite) (*yasuo_pb.UpsertCoursesRequest, error) {
	id := idutil.ULIDNow()
	country := bob_pb.COUNTRY_VN
	grade, err := i18n.ConvertIntGradeToString(country, 1)
	if err != nil {
		return nil, errors.Wrap(err, "ConvertIntGradeToString")
	}

	req := &yasuo_pb.UpsertCoursesRequest{
		Courses: []*yasuo_pb.UpsertCoursesRequest_Course{
			{
				Id:           id,
				Name:         id,
				Country:      country,
				Subject:      bob_pb.SUBJECT_MATHS,
				Grade:        grade,
				DisplayOrder: 1,
				SchoolId:     int32(s.getSchoolId()),
				Icon:         "https://example-url.com",
			},
		},
	}
	return req, nil
}

func (s *suite) schoolAdminIsOnTheCoursePage() error {
	return nil
}

func (s *suite) schoolAdminCreatesANewCourse() error {
	// Setup context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithTokenForGrpcCall(s, ctx)

	req, err := new(createCourseFeature).addCourseReq(s)
	if err != nil {
		return errors.Wrap(err, "addCourseReq()")
	}
	resp, err := yasuo_pb.NewCourseServiceClient(s.yasuoConn).UpsertCourses(ctx, req)
	if err != nil {
		return errors.Wrap(err, "UpsertCourses()")
	}
	s.ResponseStack.Push(resp)
	s.RequestStack.Push(req)

	return nil
}

func (s *suite) schoolAdminSeesTheNewCourseOnCMS() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithToken(s, ctx)

	// Pre-setup for hasura query using admin secret
	if err := trackTableForHasuraQuery(bobHasuraAdminUrl+"/v1/query", "courses"); err != nil {
		return errors.Wrap(err, "trackTableForHasuraQuery()")
	}
	if err := createSelectPermissionForHasuraQuery(bobHasuraAdminUrl+"/v1/query", "courses"); err != nil {
		return errors.Wrap(err, "createSelectPermissionForHasuraQuery()")
	}

	query :=
		`
		query ($course_id: String!) {
			courses(where: {course_id: {_eq: $course_id}}) {
					course_id
					name
					country
					subject
					icon
					grade
					display_order
				}
		}
		`

	if err := addQueryToAllowListForHasuraQuery(bobHasuraAdminUrl+"/v1/query", query); err != nil {
		return errors.Wrap(err, "addQueryToAllowListForHasuraQuery()")
	}

	// Query newly created teacher from hasura
	var profileQuery struct {
		Course []struct {
			CourseID     string `graphql:"course_id"`
			Name         string `graphql:"name"`
			Country      string `graphql:"country"`
			Subject      string `graphql:"subject"`
			Icon         string `graphql:"icon"`
			Grade        int32  `graphql:"grade"`
			DisplayOrder int32  `graphql:"display_order"`
		} `graphql:"courses(where: {course_id: {_eq: $course_id}})"`
	}

	upsertCourseReq := s.RequestStack.Requests[0].(*yasuo_pb.UpsertCoursesRequest)
	//upsertCourseResp := s.ResponseStack.Responses[0].(*yasuo_pb.UpsertCoursesResponse)
	requestedToCreateCourse := upsertCourseReq.Courses[0]

	variables := map[string]interface{}{
		"course_id": graphql.String(requestedToCreateCourse.Id),
	}
	err := queryHasura(ctx, &profileQuery, variables, bobHasuraAdminUrl+"/v1/graphql")
	if err != nil {
		return errors.Wrap(err, "queryHasura")
	}

	if len(profileQuery.Course) != 1 {
		return errors.New("failed to query course")
	}

	createdCourse := profileQuery.Course[0]

	requestedToCreateCourseGrade, err := i18n.ConvertStringGradeToInt(requestedToCreateCourse.Country, requestedToCreateCourse.Grade)
	if err != nil {
		return errors.Wrap(err, "ConvertStringGradeToInt()")
	}

	switch {
	case createdCourse.Name != requestedToCreateCourse.Name:
		return fmt.Errorf(`expect created course has "name": %v but actual is %v`, requestedToCreateCourse.Name, createdCourse.Name)
	case createdCourse.Icon != requestedToCreateCourse.Icon:
		return fmt.Errorf(`expect created course has "icon": %v but actual is %v`, requestedToCreateCourse.Icon, createdCourse.Icon)
	case createdCourse.DisplayOrder != requestedToCreateCourse.DisplayOrder:
		return fmt.Errorf(`expect created course has "display_order": %v but actual is %v`, requestedToCreateCourse.DisplayOrder, createdCourse.DisplayOrder)
	case createdCourse.Subject != requestedToCreateCourse.Subject.String():
		return fmt.Errorf(`expect created course has "subject": %v but actual is %v`, requestedToCreateCourse.Subject.String(), createdCourse.Subject)
	case createdCourse.Country != requestedToCreateCourse.Country.String():
		return fmt.Errorf(`expect created course has "name": %v but actual is %v`, requestedToCreateCourse.Name, createdCourse.Name)
	case createdCourse.Grade != int32(requestedToCreateCourseGrade):
		return fmt.Errorf(`expect created course has "grade": %v but actual is %v`, requestedToCreateCourseGrade, createdCourse.Grade)
	}

	return nil
}

func (s *suite) teacherSeesTheNewCourseOnTeacherApp() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.UserGroupInContext = constant.UserGroupTeacher
	ctx = contextWithTokenForGrpcCall(s, ctx)

	upsertCourseReq := s.RequestStack.Requests[0].(*yasuo_pb.UpsertCoursesRequest)

	req := &bob_pbv1.ListCoursesRequest{
		Paging: nil,
		Filter: &com_pbv1.CommonFilter{
			Ids: []string{upsertCourseReq.Courses[0].Id},
		},
	}
	resp, err := bob_pbv1.NewCourseReaderServiceClient(s.bobConn).ListCourses(ctx, req)
	if err != nil {
		return errors.Wrap(err, "ListCourses()")
	}

	if len(resp.Items) != 1 {
		return errors.New("responded course list is empty")
	}

	requestedToCreateCourse := upsertCourseReq.Courses[0]
	requestedToCreateCourseGrade, err := i18n.ConvertStringGradeToInt(requestedToCreateCourse.Country, requestedToCreateCourse.Grade)
	if err != nil {
		return errors.Wrap(err, "ConvertStringGradeToInt()")
	}
	createdCourse := resp.Items[0]

	switch {
	case createdCourse.Info.Name != requestedToCreateCourse.Name:
		return fmt.Errorf(`expect created course has "name": %v but actual is %v`, requestedToCreateCourse.Name, createdCourse.Info.Name)
	case createdCourse.Info.IconUrl != requestedToCreateCourse.Icon:
		return fmt.Errorf(`expect created course has "icon": %v but actual is %v`, requestedToCreateCourse.Icon, createdCourse.Info.IconUrl)
	case createdCourse.Info.DisplayOrder != requestedToCreateCourse.DisplayOrder:
		return fmt.Errorf(`expect created course has "display_order": %v but actual is %v`, requestedToCreateCourse.DisplayOrder, createdCourse.Info.DisplayOrder)
	case createdCourse.Info.Subject.String() != requestedToCreateCourse.Subject.String():
		return fmt.Errorf(`expect created course has "subject": %v but actual is %v`, requestedToCreateCourse.Subject.String(), createdCourse.Info.Subject)
	case createdCourse.Info.Country.String() != requestedToCreateCourse.Country.String():
		return fmt.Errorf(`expect created course has "name": %v but actual is %v`, requestedToCreateCourse.Name, createdCourse.Info.Name)
	case createdCourse.Info.Grade != int32(requestedToCreateCourseGrade):
		return fmt.Errorf(`expect created course has "grade": %v but actual is %v`, requestedToCreateCourseGrade, createdCourse.Info.Grade)
	}

	return nil
}

func (s *suite) studentCanNotSeeTheNewCourseOnLearnerApp() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.UserGroupInContext = constant.UserGroupStudent
	ctx = contextWithTokenForGrpcCall(s, ctx)

	req := &fatima_pbv1.RetrieveAccessibilityRequest{}
	resp, err := fatima_pbv1.NewAccessibilityReadServiceClient(s.fatimaConn).RetrieveAccessibility(ctx, req)
	if err != nil {
		return errors.Wrap(err, "RetrieveAccessibility()")
	}

	if len(resp.Courses) > 0 {
		return errors.New("expected can not see any course")
	}

	return nil
}
