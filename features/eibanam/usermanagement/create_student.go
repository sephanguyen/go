package usermanagement

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	yasuo_pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	bob_pbv1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	common_pbv1 "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	eureka_pbv1 "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	fatima_pbv1 "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	yasuo_pbv1 "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/hasura/go-graphql-client"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) loginsCMS(role string) error {
	return s.signedInAsAccount(role)
}

func (s *suite) loginsTeacherApp(role string) error {
	return s.signedInAsAccount("teacher")
}

func (s *suite) loginsLearnerApp(role string) error {
	return s.signedInAsAccount(role)
}

func (s *suite) schoolAdminCreatesANewStudentWithStudentInfo() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithTokenForGrpcCall(s, ctx)

	// Make create student request
	req := reqWithOnlyStudentInfo()

	// Create new student using
	// yasuo v1 CreateStudent api
	resp, err :=
		yasuo_pbv1.
			NewUserModifierServiceClient(s.yasuoConn).
			CreateStudent(ctx, req)
	if err != nil {
		return err
	}
	s.ResponseStack.Push(resp)
	s.RequestStack.Push(req)

	return nil
}

func (s *suite) schoolAdminSeesNewlyCreatedStudentOnCMS() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithToken(s, ctx)

	// Pre-setup for hasura query using admin secret
	trackTableForHasuraQuery(bobHasuraAdminUrl+"/v1/query", "students", "users")
	createSelectPermissionForHasuraQuery(bobHasuraAdminUrl+"/v1/query", "students", "users")
	query := `
query ($userID: String!){
	students(where: {student_id: {_eq: $userID}}) {
	  school_id
	  current_grade
	}
	users(where: {user_id: {_eq: $userID}}) {
	  user_id
	  email
	  name
	  country
	  phone_number
	}
  }
`
	addQueryToAllowListForHasuraQuery(bobHasuraAdminUrl+"/v1/query", query)

	// Query newly created student from hasura
	var profileQuery struct {
		Students []struct {
			SchoolID     int32 `graphql:"school_id"`
			CurrentGrade int32 `graphql:"current_grade"`
		} `graphql:"students(where: {student_id: {_eq: $userID}})"`
		Users []struct {
			UserID      string `graphql:"user_id"`
			Email       string `graphql:"email"`
			Name        string `graphql:"name"`
			Country     string `graphql:"country"`
			PhoneNumber string `graphql:"phone_number"`
		} `graphql:"users(where: {user_id: {_eq: $userID}})"`
	}
	resp, err := s.ResponseStack.Peek()
	if err != nil {
		return err
	}
	userID := resp.(*yasuo_pbv1.CreateStudentResponse).StudentProfile.Student.UserProfile.UserId
	variables := map[string]interface{}{
		"userID": graphql.String(userID),
	}
	err = queryHasura(ctx, &profileQuery, variables, bobHasuraAdminUrl+"/v1/graphql")
	if err != nil {
		return fmt.Errorf("error query hasura: %w", err)
	}

	// Compare the newly created student info with the requested info for create
	req0, err := s.RequestStack.Peek()
	if err != nil {
		return err
	}
	req := req0.(*yasuo_pbv1.CreateStudentRequest)
	if userID != profileQuery.Users[0].UserID {
		return fmt.Errorf(`expected profile "user_id": %v but actual is %v`, userID, profileQuery.Users[0].UserID)
	}
	if req.StudentProfile.Email != profileQuery.Users[0].Email {
		return fmt.Errorf(`expected profile "email": %v but actual is %v`, req.StudentProfile.Email, profileQuery.Users[0].Email)
	}
	if req.StudentProfile.Name != profileQuery.Users[0].Name {
		return fmt.Errorf(`expected profile "name": %v but actual is %v`, req.StudentProfile.Name, profileQuery.Users[0].Name)
	}
	if req.StudentProfile.CountryCode.String() != profileQuery.Users[0].Country {
		return fmt.Errorf(`expected profile "country": %v but actual is %v`, req.StudentProfile.CountryCode.String(), profileQuery.Users[0].Country)
	}
	if req.StudentProfile.PhoneNumber != profileQuery.Users[0].PhoneNumber {
		return fmt.Errorf(`expected profile "phone_number": %v but actual is %v`, req.StudentProfile.PhoneNumber, profileQuery.Users[0].PhoneNumber)
	}
	if req.SchoolId != profileQuery.Students[0].SchoolID {
		return fmt.Errorf(`expected profile "school_id": %v but actual is %v`, req.SchoolId, profileQuery.Students[0].SchoolID)
	}
	if req.StudentProfile.Grade != profileQuery.Students[0].CurrentGrade {
		return fmt.Errorf(`expected profile "current_grade": %v but actual is %v`, req.StudentProfile.Grade, profileQuery.Students[0].CurrentGrade)
	}

	return nil
}

func (s *suite) studentLoginsLearnerAppSuccessfullyWithCredentialsWhichSchoolAdminGives() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Check if the newly created student can login into learner app successfully
	// by calling api to signin into firebase
	req0, err := s.RequestStack.Peek()
	if err != nil {
		return err
	}
	req := req0.(*yasuo_pbv1.CreateStudentRequest)
	token, err := loginFirebaseAccount(ctx, s.Config.FirebaseAPIKey, req.StudentProfile.Email, req.StudentProfile.Password)
	if err != nil {
		return err
	}

	// Exchange and store student token for using later later
	resp, err := s.ResponseStack.Peek()
	if err != nil {
		return err
	}
	userID := resp.(*yasuo_pbv1.CreateStudentResponse).StudentProfile.Student.UserProfile.UserId
	err = s.exchangeTokenAndUpdateUserGroupCredential(token, userID, constant.UserGroupStudent)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) schoolAdminCreatesANewStudentWithParentInfo() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithTokenForGrpcCall(s, ctx)

	// Make create student request
	req := reqWithOnlyStudentInfo()
	newParentProfiles := s.newParentProfiles()
	req.ParentProfiles =
		append(req.ParentProfiles, newParentProfiles...)

	// Create new student (together with new parent) using
	// yasuo v1 CreateStudent api
	resp, err :=
		yasuo_pbv1.
			NewUserModifierServiceClient(s.yasuoConn).
			CreateStudent(ctx, req)
	if err != nil {
		return err
	}
	s.ResponseStack.Push(resp)
	s.RequestStack.Push(req)

	return nil
}

func (s *suite) newParentLoginsLearnerAppSuccessfullyWithCredentialsWhichSchoolAdminGives() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Check if the newly created parent can login into learner app successfully
	// by calling api to signin into firebase
	resp, err := s.ResponseStack.Peek()
	if err != nil {
		return err
	}
	parentProfile := resp.(*yasuo_pbv1.CreateStudentResponse).ParentProfiles[0]
	token, err := loginFirebaseAccount(ctx, s.Config.FirebaseAPIKey, parentProfile.Parent.UserProfile.Email, parentProfile.ParentPassword)
	if err != nil {
		return err
	}

	// Exchange and store parent token for checking student stats later
	err = s.exchangeTokenAndUpdateUserGroupCredential(token, parentProfile.Parent.UserProfile.UserId, constant.UserGroupParent)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) parentSeesStudentsStatsOnLearnerApp(numStudent int) error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupParent
	ctx = contextWithTokenForGrpcCall(s, ctx)

	// Check if parent can see student stats by calling
	// RetrieveStudentAssociatedToParentAccount api
	retrieveStudentResp, err := bob_pbv1.NewStudentReaderServiceClient(s.bobConn).
		RetrieveStudentAssociatedToParentAccount(ctx, &bob_pbv1.RetrieveStudentAssociatedToParentAccountRequest{})
	if err != nil {
		return err
	}
	if len(retrieveStudentResp.Profiles) != numStudent {
		return fmt.Errorf(`expected %v profiles retrieved, but actual is %v`, numStudent, len(retrieveStudentResp.Profiles))
	}
	studentProfiles := make(map[string]*yasuo_pbv1.UserProfile)
	resps, err := s.ResponseStack.PeekMulti(numStudent)
	if err != nil {
		return err
	}
	for _, resp := range resps {
		studentProfile := resp.(*yasuo_pbv1.CreateStudentResponse).StudentProfile.Student.UserProfile
		studentProfiles[studentProfile.UserId] = studentProfile
	}
	for _, studentProfileRetrieved := range retrieveStudentResp.Profiles {
		studentProfile, ok := studentProfiles[studentProfileRetrieved.UserId]
		if !ok {
			return fmt.Errorf(`student profile retrieved does not match with any of student(s) created`)
		}
		if studentProfile.UserId != studentProfileRetrieved.UserId {
			return fmt.Errorf(`expected profile "user_id": %v but actual is %v`, studentProfile.UserId, studentProfileRetrieved.UserId)
		}
		if studentProfile.Name != studentProfileRetrieved.Name {
			return fmt.Errorf(`expected profile "name": %v but actual is %v`, studentProfile.Name, studentProfileRetrieved.Name)
		}
		if studentProfile.GivenName != studentProfileRetrieved.GivenName {
			return fmt.Errorf(`expected profile "given_name": %v but actual is %v`, studentProfile.GivenName, studentProfileRetrieved.GivenName)
		}
		if studentProfile.Avatar != studentProfileRetrieved.Avatar {
			return fmt.Errorf(`expected profile "avatar": %v but actual is %v`, studentProfile.Avatar, studentProfileRetrieved.Avatar)
		}
		if studentProfile.Group != studentProfileRetrieved.Group {
			return fmt.Errorf(`expected profile "user_group": %v but actual is %v`, studentProfile.Group, studentProfileRetrieved.Group)
		}
	}

	return nil
}

func (s *suite) schoolAdminHasCreatedAStudentWithParentInfo() error {
	return s.schoolAdminCreatesANewStudentWithParentInfo()
}

func (s *suite) schoolAdminCreatesANewStudentWithExistedParentInfo() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithTokenForGrpcCall(s, ctx)

	// Make create student request
	prevResp, err := s.ResponseStack.Peek()
	if err != nil {
		return err
	}

	prevRespParentProfiles := prevResp.(*yasuo_pbv1.CreateStudentResponse).ParentProfiles
	req := reqWithOnlyStudentInfo()
	req.ParentProfiles =
		append(req.ParentProfiles, &yasuo_pbv1.CreateStudentRequest_ParentProfile{
			Id: prevRespParentProfiles[0].Parent.UserProfile.UserId,
		})

	// Create new student with existed parent info using
	// yasuo v1 CreateStudent api
	resp, err :=
		yasuo_pbv1.
			NewUserModifierServiceClient(s.yasuoConn).
			CreateStudent(ctx, req)
	if err != nil {
		return err
	}
	s.ResponseStack.Push(resp)
	s.RequestStack.Push(req)

	return nil
}

func (s *suite) existedParentLoginsLearnerAppSuccessfullyWithHisExistedCredentials() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Check if the existed parent can login into learner app successfully
	// by calling api to signin into firebase
	var parentProfile *yasuo_pbv1.CreateStudentRequest_ParentProfile
	// Search parent profile through request stack
	// (ignore the top because we need to search for existed parent)
	i := 2
	for {
		reqs, err := s.RequestStack.PeekMulti(i)
		if err != nil {
			return err
		}
		req, ok := reqs[0].(*yasuo_pbv1.CreateStudentRequest)
		if ok {
			parentProfile = req.ParentProfiles[0]
			break
		}
		i++
	}
	token, err := loginFirebaseAccount(ctx, s.Config.FirebaseAPIKey, parentProfile.Email, parentProfile.Password)
	if err != nil {
		return err
	}

	resp, err := s.ResponseStack.Peek()
	if err != nil {
		return err
	}
	parentProfiles := resp.(*yasuo_pbv1.CreateStudentResponse).ParentProfiles
	parentID := parentProfiles[len(parentProfiles)-1].Parent.UserProfile.UserId
	// Exchange and store parent token for checking student stats later
	err = s.exchangeTokenAndUpdateUserGroupCredential(token, parentID, constant.UserGroupParent)
	if err != nil {
		return err
	}

	return nil
}

func reqWithOnlyStudentInfo() *yasuo_pbv1.CreateStudentRequest {
	randomId := idutil.ULIDNow()
	return &yasuo_pbv1.CreateStudentRequest{
		SchoolId: 1,
		StudentProfile: &yasuo_pbv1.CreateStudentRequest_StudentProfile{
			Email:             fmt.Sprintf("%v@example.com", randomId),
			Password:          fmt.Sprintf("password-%v", randomId),
			Name:              fmt.Sprintf("user-%v", randomId),
			CountryCode:       common_pbv1.Country_COUNTRY_VN,
			EnrollmentStatus:  common_pbv1.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			PhoneNumber:       fmt.Sprintf("phone-number-%v", randomId),
			StudentExternalId: fmt.Sprintf("student-external-id-%v", randomId),
			StudentNote:       fmt.Sprintf("some random student note %v", randomId),
			Grade:             5,
		},
	}
}

func (s *suite) newParentProfiles() []*yasuo_pbv1.CreateStudentRequest_ParentProfile {
	parentId := idutil.ULIDNow()
	profiles := []*yasuo_pbv1.CreateStudentRequest_ParentProfile{
		{
			Name:         fmt.Sprintf("user-%v", parentId),
			CountryCode:  common_pbv1.Country_COUNTRY_VN,
			PhoneNumber:  fmt.Sprintf("phone-number-%v", parentId),
			Email:        fmt.Sprintf("%v@example.com", parentId),
			Relationship: yasuo_pbv1.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
			Password:     fmt.Sprintf("password-%v", parentId),
		},
	}
	return profiles
}

func (s *suite) newStudentPackageProfiles(startTime, endTime time.Time) ([]*yasuo_pbv1.CreateStudentRequest_StudentPackageProfile, error) {
	randomCourseID := idutil.ULIDNow()
	if err := s.createACourse(randomCourseID); err != nil {
		return nil, err
	}
	profiles := []*yasuo_pbv1.CreateStudentRequest_StudentPackageProfile{
		{
			CourseId: randomCourseID,
			Start:    timestamppb.New(startTime),
			End:      timestamppb.New(endTime),
		},
	}
	return profiles, nil
}

func (s *suite) schoolAdminCreatesANewStudentWithCourseWhichHas(condition string) error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithTokenForGrpcCall(s, ctx)

	// Make create student request
	req := reqWithOnlyStudentInfo()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var newStudentPackageProfiles []*yasuo_pbv1.CreateStudentRequest_StudentPackageProfile
	var err error
	switch condition {
	case "start date <= current date <= end date":
		newStudentPackageProfiles, err = s.newStudentPackageProfiles(today.Add(-24*time.Hour), today.Add(24*time.Hour))
	case "start date > current date":
		newStudentPackageProfiles, err = s.newStudentPackageProfiles(today.Add(24*time.Hour), today.Add(48*time.Hour))
	case "end time < current date":
		newStudentPackageProfiles, err = s.newStudentPackageProfiles(today.Add(-48*time.Hour), today.Add(-24*time.Hour))
	}
	if err != nil {
		return err
	}
	req.StudentPackageProfiles = newStudentPackageProfiles

	// Create new student using
	// yasuo v1 CreateStudent api
	resp, err :=
		yasuo_pbv1.
			NewUserModifierServiceClient(s.yasuoConn).
			CreateStudent(ctx, req)
	if err != nil {
		return err
	}
	s.ResponseStack.Push(resp)
	s.RequestStack.Push(req)

	return nil
}

func (s *suite) studentTheCourseOnLearnerAppWhen(result, condition string) error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupStudent
	ctx = contextWithTokenForGrpcCall(s, ctx)
	reqs, err := s.RequestStack.PeekMulti(2)
	if err != nil {
		return err
	}
	courseRequested := reqs[0].(*yasuo_pb.UpsertCoursesRequest).Courses[0]

	// Check if student can see course on learner app by calling
	// RetrieveAccessibility api to get course IDs and
	// ListCourses api to get course info and compare with course in request
	retrieveAccessibilityResp, err := fatima_pbv1.NewAccessibilityReadServiceClient(s.fatimaConn).
		RetrieveAccessibility(contextWithToken(s, ctx), &fatima_pbv1.RetrieveAccessibilityRequest{})
	if err != nil {
		return err
	}
	courseIDs := []string{}
	for key := range retrieveAccessibilityResp.Courses {
		courseIDs = append(courseIDs, key)
	}
	var courseRetrieved *common_pbv1.Course
	if len(courseIDs) > 0 {
		retrieveCoursesResp, err := bob_pbv1.NewCourseReaderServiceClient(s.bobConn).
			ListCourses(contextWithToken(s, ctx), &bob_pbv1.ListCoursesRequest{
				Paging: &common_pbv1.Paging{Limit: 100},
				Filter: &common_pbv1.CommonFilter{
					Ids: courseIDs,
				},
			})
		if err != nil {
			return err
		}
		for _, course := range retrieveCoursesResp.Items {
			if courseRequested.Id == course.Info.Id {
				courseRetrieved = course
				break
			}
		}
	}

	switch result {
	case "sees":
		if courseRetrieved == nil {
			return fmt.Errorf(`student cannot see course but expected "sees"`)
		}
		if courseRequested.Name != courseRetrieved.Info.Name {
			return fmt.Errorf(`expected course "name": %v but actual is %v`, courseRequested.Name, courseRetrieved.Info.Name)
		}
		if courseRequested.SchoolId != courseRetrieved.Info.SchoolId {
			return fmt.Errorf(`expected course "school_id": %v but actual is %v`, courseRequested.SchoolId, courseRetrieved.Info.SchoolId)
		}
		if !strSliceEq(courseRequested.BookIds, courseRetrieved.BookIds) {
			return fmt.Errorf(`expected course "book_ids": %v but actual is %v`, courseRequested.BookIds, courseRetrieved.BookIds)
		}
		if courseRequested.DisplayOrder != courseRetrieved.Info.DisplayOrder {
			return fmt.Errorf(`expected course "display_order": %v but actual is %v`, courseRequested.DisplayOrder, courseRetrieved.Info.DisplayOrder)
		}
		if courseRequested.Icon != courseRetrieved.Info.IconUrl {
			return fmt.Errorf(`expected course "icon": %v but actual is %v`, courseRequested.Icon, courseRetrieved.Info.IconUrl)
		}
	case "does not see":
		if courseRetrieved != nil {
			return fmt.Errorf(`student still see course but expected "does not see"`)
		}
	}

	return nil
}

func (s *suite) teacherSeesNewlyCreatedStudentOnTeacherApp() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupTeacher
	ctx = contextWithTokenForGrpcCall(s, ctx)
	req, err := s.RequestStack.Peek()
	if err != nil {
		return err
	}
	courseID := req.(*yasuo_pbv1.CreateStudentRequest).StudentPackageProfiles[0].CourseId
	resp, err := s.ResponseStack.Peek()
	if err != nil {
		return err
	}
	studentProfile := resp.(*yasuo_pbv1.CreateStudentResponse).StudentProfile.Student.UserProfile

	// Check if teacher can see newly created student on teacher app
	// by calling ListStudentByCourse api
	retrieveStudentResp, err := eureka_pbv1.NewCourseReaderServiceClient(s.eurekaConn).
		ListStudentByCourse(contextWithToken(s, ctx), &eureka_pbv1.ListStudentByCourseRequest{
			CourseId: courseID,
			Paging:   &common_pbv1.Paging{Limit: 100},
		})
	if err != nil {
		return err
	}
	var studentProfileRetrieved *common_pbv1.BasicProfile
	for _, profile := range retrieveStudentResp.Profiles {
		if studentProfile.UserId == profile.UserId {
			studentProfileRetrieved = profile
			break
		}
	}
	if studentProfileRetrieved == nil {
		return fmt.Errorf(`student created does not match with any of student profile(s) retrieved`)
	}
	if studentProfile.Name != studentProfileRetrieved.Name {
		return fmt.Errorf(`expected profile "name": %v but actual is %v`, studentProfile.Name, studentProfileRetrieved.Name)
	}
	if studentProfile.GivenName != studentProfileRetrieved.GivenName {
		return fmt.Errorf(`expected profile "given_name": %v but actual is %v`, studentProfile.GivenName, studentProfileRetrieved.GivenName)
	}
	if studentProfile.Avatar != studentProfileRetrieved.Avatar {
		return fmt.Errorf(`expected profile "avatar": %v but actual is %v`, studentProfile.Avatar, studentProfileRetrieved.Avatar)
	}
	if studentProfile.Group != studentProfileRetrieved.Group {
		return fmt.Errorf(`expected profile "user_group": %v but actual is %v`, studentProfile.Group, studentProfileRetrieved.Group)
	}

	return nil
}

func (s *suite) createACourse(ID string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithTokenForGrpcCall(s, ctx)

	req := &yasuo_pb.UpsertCoursesRequest{
		Courses: []*yasuo_pb.UpsertCoursesRequest_Course{},
	}
	course := newUpsertCourseReq(ID, "course name for id: "+ID)
	req.Courses = append(req.Courses, course)
	_, err := yasuo_pb.NewCourseServiceClient(s.yasuoConn).UpsertCourses(contextWithToken(s, ctx), req)
	if err != nil {
		return err
	}

	s.RequestStack.Push(req)

	return nil
}

func newUpsertCourseReq(ID, name string) *yasuo_pb.UpsertCoursesRequest_Course {
	r := &yasuo_pb.UpsertCoursesRequest_Course{
		Id:           ID,
		Name:         name,
		Country:      bob_pb.COUNTRY_MASTER,
		Subject:      bob_pb.SUBJECT_ENGLISH,
		Grade:        "Grade 12",
		DisplayOrder: 1,
		ChapterIds:   nil,
		SchoolId:     1,
		BookIds:      nil,
		Icon:         "link-icon",
	}
	return r
}

func (s *suite) allParentSeesStudentsStatsOnLearnerApp() error {
	// Check for existed parent, existed parent should
	// see 2 student stats
	err := s.existedParentLoginsLearnerAppSuccessfullyWithHisExistedCredentials()
	if err != nil {
		return err
	}
	err = s.parentSeesStudentsStatsOnLearnerApp(2)
	if err != nil {
		return err
	}

	// Check for new parent, new parent should
	// see only one student stats
	err = s.newParentLoginsLearnerAppSuccessfullyWithCredentialsWhichSchoolAdminGives()
	if err != nil {
		return err
	}
	err = s.parentSeesStudentsStatsOnLearnerApp(1)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) schoolAdminCreatesANewStudentWithNewParentExistedParentAndVisibleCourse() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithTokenForGrpcCall(s, ctx)

	// Make create student request
	req := reqWithOnlyStudentInfo()
	// Add new parent profile to request
	newParentProfiles := s.newParentProfiles()
	req.ParentProfiles =
		append(req.ParentProfiles, newParentProfiles...)
	// Add existed parent profile to request
	prevResp, err := s.ResponseStack.Peek()
	if err != nil {
		return err
	}
	prevRespParentProfiles := prevResp.(*yasuo_pbv1.CreateStudentResponse).ParentProfiles
	req.ParentProfiles =
		append(req.ParentProfiles, &yasuo_pbv1.CreateStudentRequest_ParentProfile{
			Id: prevRespParentProfiles[0].Parent.UserProfile.UserId,
		})
	// Add visible course to request
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	newStudentPackageProfiles, err := s.newStudentPackageProfiles(today.Add(-24*time.Hour), today.Add(24*time.Hour))
	if err != nil {
		return err
	}
	req.StudentPackageProfiles = newStudentPackageProfiles

	// Create new student with existed parent info using
	// yasuo v1 CreateStudent api
	resp, err :=
		yasuo_pbv1.
			NewUserModifierServiceClient(s.yasuoConn).
			CreateStudent(ctx, req)
	if err != nil {
		return err
	}
	s.ResponseStack.Push(resp)
	s.RequestStack.Push(req)

	return nil
}

func (s *suite) studentSeesTheCourseOnLearnerApp() error {
	return s.studentTheCourseOnLearnerAppWhen("sees", "start date <= current date <= end date")
}

func (s *suite) exchangeTokenAndUpdateUserGroupCredential(token, userID, userGroup string) error {
	schoolID := s.getSchoolId()
	token, err := helper.ExchangeToken(token, userID, userGroup, applicantID, schoolID, shamirConn)
	if err != nil {
		return err
	}
	s.UserGroupCredentials[userGroup] = &userCredential{
		UserID:    userID,
		AuthToken: token,
		UserGroup: userGroup,
	}
	return nil
}
