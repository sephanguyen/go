package usermanagement

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/eibanam"
	"github.com/manabie-com/backend/internal/enigma/dto"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bob_pbv1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	com_pbv1 "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	fatima_pbv1 "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
)

func (s *suite) stepAValidJPREPSignatureInItsHeader(req interface{}) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	sig, err := s.generateSignature(s.JPREPKey, string(data))
	if err != nil {
		return nil
	}
	s.JPREPSignature = sig
	return nil
}

func (s *suite) systemSyncsCourseWhichBelongTo(courseType string) error {
	courseId := rand.Intn(999) + 1
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Courses: []dto.Course{
				{
					ActionKind: dto.ActionKindUpserted,
					CourseID:   courseId,
					CourseName: "course-name-with-actionKind-upsert",
				},
			},
		},
	}

	switch courseType {
	case "kid":
		for i := range request.Payload.Courses {
			(&request.Payload.Courses[i]).CourseStudentDivID = dto.CourseIDKid
		}
		break
	default:
		return errors.New("this arg is not supported for testing")
	}

	if err := s.attachValidJPREPSignature(request); err != nil {
		return err
	}

	if err := s.performMasterRegistrationRequest(request); err != nil {
		return errors.Wrap(err, "makeJPREPHTTPRequest")
	}

	s.RequestStack.Push(request)

	return nil
}

type course struct {
	CourseID     string `graphql:"course_id"`
	Name         string `graphql:"name"`
	Country      string `graphql:"country"`
	Subject      string `graphql:"subject"`
	Icon         string `graphql:"icon"`
	Grade        int32  `graphql:"grade"`
	DisplayOrder int32  `graphql:"display_order"`
	Status       string `graphql:"status"`
}

func (s *suite) getCoursesByHasura(ctx context.Context, courseId int) ([]*course, error) {
	// Pre-setup for hasura query using admin secret
	if err := trackTableForHasuraQuery(bobHasuraAdminUrl+"/v1/query", "courses"); err != nil {
		return nil, errors.Wrap(err, "trackTableForHasuraQuery()")
	}
	if err := createSelectPermissionForHasuraQuery(bobHasuraAdminUrl+"/v1/query", "courses"); err != nil {
		return nil, errors.Wrap(err, "createSelectPermissionForHasuraQuery()")
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
					status
				}
		}
		`
	if err := addQueryToAllowListForHasuraQuery(bobHasuraAdminUrl+"/v1/query", query); err != nil {
		return nil, errors.Wrap(err, "addQueryToAllowListForHasuraQuery()")
	}

	// Query newly created teacher from hasura
	var profileQuery struct {
		Course []*course `graphql:"courses(where: {course_id: {_eq: $course_id}})"`
	}

	variables := map[string]interface{}{
		"course_id": graphql.String(toJprepCourseID(courseId)),
	}

	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithToken(s, ctx)

	err := queryHasura(ctx, &profileQuery, variables, bobHasuraAdminUrl+"/v1/graphql")
	if err != nil {
		return nil, errors.Wrap(err, "queryHasura")
	}

	return profileQuery.Course, nil
}

func (s *suite) schoolAdminSeesCourseOnCMS() error {
	// Setup context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return eibanam.TryUntilSuccess(ctx, 100*time.Millisecond, func(ctx context.Context) (bool, error) {
		// Pre-setup for hasura query using admin secret
		if err := trackTableForHasuraQuery(bobHasuraAdminUrl+"/v1/query", "courses"); err != nil {
			return false, errors.Wrap(err, "trackTableForHasuraQuery()")
		}
		if err := createSelectPermissionForHasuraQuery(bobHasuraAdminUrl+"/v1/query", "courses"); err != nil {
			return false, errors.Wrap(err, "createSelectPermissionForHasuraQuery()")
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
					status
				}
		}
		`

		if err := addQueryToAllowListForHasuraQuery(bobHasuraAdminUrl+"/v1/query", query); err != nil {
			return false, errors.Wrap(err, "addQueryToAllowListForHasuraQuery()")
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
				Status       string `graphql:"status"`
			} `graphql:"courses(where: {course_id: {_eq: $course_id}})"`
		}

		req, err := s.RequestStack.Peek()
		if err != nil {
			return false, errors.Wrap(err, "Peek()")
		}
		upsertCourseReq := req.(*dto.MasterRegistrationRequest)
		requestedToCreateCourse := upsertCourseReq.Payload.Courses[0]

		variables := map[string]interface{}{
			"course_id": graphql.String(toJprepCourseID(requestedToCreateCourse.CourseID)),
		}

		s.UserGroupInContext = constant.UserGroupSchoolAdmin
		ctx = contextWithToken(s, ctx)

		err = queryHasura(ctx, &profileQuery, variables, bobHasuraAdminUrl+"/v1/graphql")
		if err != nil {
			return false, errors.Wrap(err, "queryHasura")
		}

		if len(profileQuery.Course) != 1 {
			return true, errors.New("responded course list is empty")
		}

		createdCourse := profileQuery.Course[0]

		switch {
		case createdCourse.CourseID != toJprepCourseID(requestedToCreateCourse.CourseID):
			return false, fmt.Errorf(`expect created course has "id": %v but actual is %v`, toJprepCourseID(requestedToCreateCourse.CourseID), createdCourse.CourseID)
		case createdCourse.Name != requestedToCreateCourse.CourseName:
			return false, fmt.Errorf(`expect created course has "name": %v but actual is %v`, requestedToCreateCourse.CourseName, createdCourse.Name)
		}

		if requestedToCreateCourse.CourseStudentDivID == dto.CourseIDKid {
			if createdCourse.Status != com_pbv1.CourseStatus_COURSE_STATUS_ACTIVE.String() {
				return false, fmt.Errorf(`expect created course has "status": %v but actual is %v`, com_pbv1.CourseStatus_COURSE_STATUS_ACTIVE, createdCourse.Status)
			}
		}

		return false, nil
	})
}

func (s *suite) getCoursesByListCoursesApi(ctx context.Context, courseIds ...string) (*bob_pbv1.ListCoursesResponse, error) {
	req := &bob_pbv1.ListCoursesRequest{
		Paging: nil,
		Filter: &com_pbv1.CommonFilter{
			Ids: courseIds,
		},
	}

	s.UserGroupInContext = constant.UserGroupTeacher
	ctx = contextWithTokenForGrpcCall(s, ctx)

	resp, err := bob_pbv1.NewCourseReaderServiceClient(s.bobConn).ListCourses(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "ListCourses()")
	}
	return resp, err
}

func (s *suite) teacherTheCourseOnTeacherApp(result string) error {
	reqs, err := s.RequestStack.PeekMulti(2)
	if err != nil {
		return errors.Wrap(err, "PeekMulti()")
	}
	teacherID := reqs[0].(*dto.UserRegistrationRequest).Payload.Staffs[0].StaffID
	err = s.saveCredential(teacherID, constant.UserGroupTeacher, constants.JPREPSchool)
	if err != nil {
		return errors.Wrap(err, "saveCredential()")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = eibanam.TryUntilSuccess(ctx, 100*time.Millisecond, func(ctx context.Context) (bool, error) {
		switch result {
		case "sees":
			break
		default:
			return false, errors.New("this arg is not supported for testing")
		}

		upsertCourseReq := reqs[1].(*dto.MasterRegistrationRequest)
		requestedToCreateCourse := upsertCourseReq.Payload.Courses[0]

		req := &bob_pbv1.ListCoursesRequest{
			Paging: nil,
			Filter: &com_pbv1.CommonFilter{
				Ids: []string{toJprepCourseID(requestedToCreateCourse.CourseID)},
			},
		}

		s.UserGroupInContext = constant.UserGroupTeacher
		ctx = contextWithTokenForGrpcCall(s, ctx)

		resp, err := bob_pbv1.NewCourseReaderServiceClient(s.bobConn).ListCourses(ctx, req)
		if err != nil {
			return false, errors.Wrap(err, "ListCourses()")
		}

		if len(resp.Items) < 1 {
			return true, errors.New("responded course list is empty")
		}
		createdCourse := resp.Items[0]

		switch {
		case createdCourse.Info.Id != toJprepCourseID(requestedToCreateCourse.CourseID):
			return false, fmt.Errorf(`expect created course has "id": %v but actual is %v`, toJprepCourseID(requestedToCreateCourse.CourseID), createdCourse.Info.Id)
		case createdCourse.Info.Name != requestedToCreateCourse.CourseName:
			return false, fmt.Errorf(`expect created course has "name": %v but actual is %v`, requestedToCreateCourse.CourseName, createdCourse.Info.Name)
		}

		return false, nil
	})
	return err
}

func (s *suite) systemHasSyncedCourseAndClassFromPartner() error {
	return s.systemHasSyncedCourseAndClassesFromPartner(1)
}

func (s *suite) systemSyncsCourseWithEditedCourseName() error {
	reqs, err := s.RequestStack.PeekMulti(2)
	if err != nil {
		return err
	}
	course := reqs[0].(*dto.MasterRegistrationRequest).Payload.Courses[0]
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Courses: []dto.Course{
				{
					ActionKind:         dto.ActionKindUpserted,
					CourseID:           course.CourseID,
					CourseName:         course.CourseName + "-edited",
					CourseStudentDivID: dto.CourseIDKid,
				},
			},
		},
	}
	if err := s.attachValidJPREPSignature(request); err != nil {
		return err
	}

	if err := s.performMasterRegistrationRequest(request); err != nil {
		return errors.Wrap(err, "makeJPREPHTTPRequest")
	}

	s.RequestStack.Push(request)

	return nil
}

func (s *suite) schoolAdminSeesEditedCourseNameOnCMS() error {
	// Setup context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := eibanam.TryUntilSuccess(ctx, 100*time.Millisecond, func(ctx context.Context) (bool, error) {
		req, err := s.RequestStack.Peek()
		if err != nil {
			return false, errors.Wrap(err, "Peek()")
		}
		upsertCourseReq := req.(*dto.MasterRegistrationRequest)
		requestedToCreateCourse := upsertCourseReq.Payload.Courses[0]

		courses, err := s.getCoursesByHasura(ctx, requestedToCreateCourse.CourseID)
		if len(courses) < 1 {
			return true, errors.New("responded course list is empty")
		}

		createdCourse := courses[0]
		//fmt.Println(fmt.Sprintf("%+v", createdCourse))

		switch {
		case createdCourse.CourseID != toJprepCourseID(requestedToCreateCourse.CourseID):
			return false, fmt.Errorf(`expect created course has "id": %v but actual is %v`, toJprepCourseID(requestedToCreateCourse.CourseID), createdCourse.CourseID)
		case createdCourse.Name != requestedToCreateCourse.CourseName:
			return false, fmt.Errorf(`expect created course has "name": %v but actual is %v`, requestedToCreateCourse.CourseName, createdCourse.Name)
		}

		if requestedToCreateCourse.CourseStudentDivID == dto.CourseIDKid {
			if createdCourse.Status != com_pbv1.CourseStatus_COURSE_STATUS_ACTIVE.String() {
				return false, fmt.Errorf(`expect created course has "status": %v but actual is %v`, com_pbv1.CourseStatus_COURSE_STATUS_ACTIVE, createdCourse.Status)
			}
		}

		return false, nil
	})

	return err
}

func (s *suite) teacherSeesEditedCourseNameOnTeacherApp() error {
	reqs, err := s.RequestStack.PeekMulti(4)
	if err != nil {
		return errors.Wrap(err, "PeekMulti()")
	}
	teacherID := reqs[0].(*dto.UserRegistrationRequest).Payload.Staffs[0].StaffID
	err = s.saveCredential(teacherID, constant.UserGroupTeacher, constants.JPREPSchool)
	if err != nil {
		return errors.Wrap(err, "saveCredential()")
	}

	requestedToCreateCourse := reqs[3].(*dto.MasterRegistrationRequest).Payload.Courses[0]

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	err = eibanam.TryUntilSuccess(ctx, 100*time.Millisecond, func(ctx context.Context) (bool, error) {
		s.UserGroupInContext = constant.UserGroupTeacher
		ctx = contextWithTokenForGrpcCall(s, ctx)

		resp, err := s.getCoursesByListCoursesApi(ctx, toJprepCourseID(requestedToCreateCourse.CourseID))
		if err != nil {
			return false, errors.Wrap(err, "ListCourses()")
		}

		if len(resp.Items) < 1 {
			return true, errors.New("responded course list is empty")
		}
		createdCourse := resp.Items[0]

		switch {
		case createdCourse.Info.Id != toJprepCourseID(requestedToCreateCourse.CourseID):
			return false, fmt.Errorf(`expect created course has "id": %v but actual is %v`, toJprepCourseID(requestedToCreateCourse.CourseID), createdCourse.Info.Id)
		case createdCourse.Info.Name != requestedToCreateCourse.CourseName:
			return false, fmt.Errorf(`expect created course has "name": %v but actual is %v`, requestedToCreateCourse.CourseName, createdCourse.Info.Name)
		}

		return false, nil
	})
	return err
}

func (s *suite) studentSeesEditedCourseNameOnLearnerApp() error {
	reqs, err := s.RequestStack.PeekMulti(4)
	if err != nil {
		return errors.Wrap(err, "PeekMulti()")
	}
	studentID := reqs[2].(*dto.UserRegistrationRequest).Payload.Students[0].StudentID
	err = s.saveCredential(studentID, constant.UserGroupStudent, s.getSchoolId())
	if err != nil {
		return err
	}

	requestedToCreateCourse := reqs[3].(*dto.MasterRegistrationRequest).Payload.Courses[0]

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	err = eibanam.TryUntilSuccess(ctx, 50*time.Millisecond, func(ctx context.Context) (bool, error) {
		s.UserGroupInContext = constant.UserGroupStudent
		ctx = contextWithTokenForGrpcCall(s, ctx)

		retrieveAccessibilityResp, err := fatima_pbv1.NewAccessibilityReadServiceClient(s.fatimaConn).
			RetrieveAccessibility(ctx, &fatima_pbv1.RetrieveAccessibilityRequest{})
		if err != nil {
			return false, errors.Wrap(err, "RetrieveAccessibility()")
		}

		if len(retrieveAccessibilityResp.Courses) < 1 {
			return true, errors.New("responded course list is empty")
		}

		accessibilityCourseIDs := []string{}
		for key := range retrieveAccessibilityResp.Courses {
			accessibilityCourseIDs = append(accessibilityCourseIDs, key)
		}

		retrieveCoursesResp, err := bob_pbv1.NewCourseReaderServiceClient(s.bobConn).
			ListCourses(ctx, &bob_pbv1.ListCoursesRequest{
				Paging: &com_pbv1.Paging{Limit: 100},
				Filter: &com_pbv1.CommonFilter{
					Ids: []string{toJprepCourseID(requestedToCreateCourse.CourseID)},
				},
			})

		if len(retrieveCoursesResp.Items) < 1 {
			return true, errors.New("responded course list is empty")
		}
		createdCourse := retrieveCoursesResp.Items[0]

		switch {
		case createdCourse.Info.Id != toJprepCourseID(requestedToCreateCourse.CourseID):
			return false, fmt.Errorf(`expect created course has "id": %v but actual is %v`, toJprepCourseID(requestedToCreateCourse.CourseID), createdCourse.Info.Id)
		case createdCourse.Info.Name != requestedToCreateCourse.CourseName:
			return false, fmt.Errorf(`expect created course has "name": %v but actual is %v`, requestedToCreateCourse.CourseName, createdCourse.Info.Name)
		}

		return false, nil
	})

	return err
}
