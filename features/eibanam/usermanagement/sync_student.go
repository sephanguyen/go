package usermanagement

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/enigma/dto"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bob_pbv1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	common_pbv1 "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	eureka_pbv1 "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	fatima_pbv1 "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const StudyPlanCreate = `ID,Book ID,Book name,Chapter ID,Chapter name,Topic ID,Topic name,Assignment/ LO,Content ID,Name,Available from,Available until,Start time,Due time
,Book-ID-1,Book 1,B1-Chapter-1,Book 1-Chapter 1,Topic-ID-1,,Learning Objective,Book 1 - Topic 1 - LO 1,LO name 1 - 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-2,Book 1-Chapter 2,Topic-ID-2,,Assignment,assignment-1,Assignment name 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-2,Book 1-Chapter 2,Topic-ID-2,,Assignment,assignment-2,Assignment name 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-2,Book 1-Chapter 2,Topic-ID-2,,Assignment,assignment-3,Assignment name 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-3,Book 1-Chapter 3,Topic-ID-3,,Learning Objective,Book 1 - Topic 1 - LO 3,LO name 1 - 3,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-4,Book 1-Chapter 4,Topic-ID-4,,Learning Objective,Book 1 - Topic 1 - LO 4,LO name 1 - 4,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-5,Book 1-Chapter 5,Topic-ID-5,,Learning Objective,Book 1 - Topic 1 - LO 5,LO name 1 - 5,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-6,Book 1-Chapter 6,Topic-ID-6,,Learning Objective,Book 1 - Topic 1 - LO 6,LO name 1 - 6,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-2,Book 2,B2-Chapter-1,Book 1-Chapter 1,Topic-ID-7,,Learning Objective,Book 2 - Topic 1 - LO 1,LO name 1 - 7,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-2,Book 2,B2-Chapter-2,Book 2-Chapter 2,Topic-ID-8,,Learning Objective,Book 2 - Topic 1 - LO 2,LO name 2 - 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-2,Book 2,B2-Chapter-3,Book 2-Chapter 3,Topic-ID-9,,Learning Objective,Book 2 - Topic 1 - LO 3,LO name 2 - 2,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-2,Book 2,B2-Chapter-4,Book 2-Chapter 4,Topic-ID-08,,Learning Objective,Book 2 - Topic 1 - LO 4,LO name 2 - 3,2052-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-3,Book 3,B3-Chapter-1,Book 3-Chapter 1,Topic-ID-11,,Learning Objective,Book 3 - Topic 1 - LO 1,LO name 2 - 4,2052-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-3,Book 3,B3-Chapter-2,Book 3-Chapter 1,Topic-ID-12,,Learning Objective,Book 3 - Topic 1 - LO 2,LO name 3 - 1,2052-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-3,Book 3,B3-Chapter-3,Book 3-Chapter 1,Topic-ID-13,,Learning Objective,Book 3 - Topic 1 - LO 3,LO name 3 - 2,2052-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-3,Book 3,B3-Chapter-4,Book 3-Chapter 1,Topic-ID-13,,Learning Objective,Book 3 - Topic 2 - LO 4,LO name 3 - 3,2052-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-3,Book 3,B3-Chapter-5,Book 3-Chapter 1,Topic-ID-13,,Learning Objective,Book 3 - Topic 3 - LO 5,LO name 3 - 4,2052-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-3,Book 3,B3-Chapter-6,Book 3-Chapter 1,Topic-ID-13,,Learning Objective,Book 3 - Topic 1 - LO 6,LO name 3 - 5,2052-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-3,Book 3,B3-Chapter-7,Book 3-Chapter 1,Topic-ID-13,,Learning Objective,Book 3 - Topic 1 - LO 7,LO name 3 - 6,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-3,Book 3,B3-Chapter-8,Book 3-Chapter 1,Topic-ID-13,,Learning Objective,Book 3 - Topic 1 - LO 8,LO name 3 - 7,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-4,Book 4,B4-Chapter-1,Book 4-Chapter 1,Topic-ID-13,,Learning Objective,Book 4 - Topic 1 - LO 9,LO name 3 - 9,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-4,Book 4,B4-Chapter-2,Book 4-Chapter 1,Topic-ID-13,,Learning Objective,Book 4 - Topic 5 - LO 08,LO name 4 -1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00`

func (s *suite) schoolAdminCreatesStudyPlanForCourse() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithTokenForGrpcCall(s, ctx)
	s.validAssignmentInDB(ctx)

	reqs, err := s.RequestStack.PeekMulti(2)
	if err != nil {
		return err
	}
	courseSyncReq := reqs[0].(*dto.MasterRegistrationRequest)
	_, err = eureka_pbv1.NewStudyPlanWriteServiceClient(s.eurekaConn).
		ImportStudyPlan(ctx, &eureka_pbv1.ImportStudyPlanRequest{
			CourseId: toJprepCourseID(courseSyncReq.Payload.Courses[0].CourseID),
			Name:     "study-plan-name.csv",
			Mode:     eureka_pbv1.ImportMode_IMPORT_MODE_CREATE,
			Type:     eureka_pbv1.StudyPlanType_STUDY_PLAN_TYPE_COURSE,
			SchoolId: constants.ManabieSchool,
			Payload:  []byte(StudyPlanCreate),
		})
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) validAssignmentInDB(ctx context.Context) error {
	assignments := []*eureka_pbv1.Assignment{
		s.generateAssignment("assignment-1", false, false, true),
		s.generateAssignment("assignment-2", false, false, true),
		s.generateAssignment("assignment-3", false, false, true),
	}

	req := &eureka_pbv1.UpsertAssignmentsRequest{
		Assignments: assignments,
	}
	var err error
	_, err = eureka_pbv1.NewAssignmentModifierServiceClient(s.eurekaConn).UpsertAssignments(contextWithToken(s, ctx), req)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) generateAssignment(assignmentID string, allowLate, allowResubmit, gradingMethod bool) *eureka_pbv1.Assignment {
	id := assignmentID
	if id == "" {
		id = idutil.ULIDNow()
	}
	return &eureka_pbv1.Assignment{
		AssignmentId: id,
		Name:         fmt.Sprintf("assignment-%s", idutil.ULIDNow()),
		Content: &eureka_pbv1.AssignmentContent{
			TopicId: idutil.ULIDNow(),
			LoId:    []string{"lo-id-1", "lo-id-2"},
		},
		CheckList: &eureka_pbv1.CheckList{
			Items: []*eureka_pbv1.CheckListItem{
				{
					Content:   "Complete all learning objectives",
					IsChecked: true,
				},
				{
					Content:   "Submitted required videos",
					IsChecked: false,
				},
			},
		},
		Instruction:    "teacher's instruction",
		MaxGrade:       100,
		Attachments:    []string{"media-id-1", "media-id-2"},
		AssignmentType: eureka_pbv1.AssignmentType_ASSIGNMENT_TYPE_LEARNING_OBJECTIVE,
		Setting: &eureka_pbv1.AssignmentSetting{
			AllowLateSubmission: allowLate,
			AllowResubmission:   allowResubmit,
		},
		RequiredGrade: gradingMethod,
		DisplayOrder:  0,
	}
}

func (s *suite) schoolAdminSeesThisStudentOnStudentstudyPlanPage() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithToken(s, ctx)

	// Pre-setup for hasura query using admin secret
	if err := trackTableForHasuraQuery(eurekaHasuraAdminUrl+"/eureka/v1/query", "course_students"); err != nil {
		return errors.Wrap(err, "trackTableForHasuraQuery()")
	}
	if err := createSelectPermissionForHasuraQuery(eurekaHasuraAdminUrl+"/eureka/v1/query", "course_students"); err != nil {
		return errors.Wrap(err, "createSelectPermissionForHasuraQuery()")
	}

	query :=
		`
		query ($course_id: String!) {
			course_students(where: {course_id: {_eq: $course_id}}) {
					student_id
					course_id
				}
		}
		`

	if err := addQueryToAllowListForHasuraQuery(eurekaHasuraAdminUrl+"/eureka/v1/query", query); err != nil {
		return errors.Wrap(err, "addQueryToAllowListForHasuraQuery()")
	}

	// Query newly student id by course from hasura
	var profileQuery struct {
		CourseStudents []struct {
			StudentID string `graphql:"student_id"`
			CourseID  string `graphql:"course_id"`
		} `graphql:"course_students(where: {course_id: {_eq: $course_id}})"`
	}

	reqs, err := s.RequestStack.PeekMulti(2)
	if err != nil {
		return err
	}
	courseID := reqs[0].(*dto.MasterRegistrationRequest).Payload.Courses[0].CourseID

	variables := map[string]interface{}{
		"course_id": graphql.String(toJprepCourseID(courseID)),
	}
	err = queryHasura(ctx, &profileQuery, variables, eurekaHasuraAdminUrl+"/eureka/v1/graphql")
	if err != nil {
		return errors.Wrap(err, "queryHasura")
	}

	studentID := reqs[1].(*dto.UserRegistrationRequest).Payload.Students[0].StudentID

	for _, q := range profileQuery.CourseStudents {
		if q.StudentID == studentID {
			return nil
		}
	}

	return fmt.Errorf("school admin does not see student id=%v in student studyplan page", studentID)
}

func (s *suite) staffCreatesSchoolAdminAccountForPartnerManually() error {
	return nil
}

func (s *suite) studentLoginsLearnerAppSuccessfullyWithStudentPartnerAccount() error {
	// since we cannot simulate login using rest api to get token from jprep
	// which uses sso authorization code flow
	// (https://stackoverflow.com/questions/52311757/keycloak-get-authorization-code-in-json)
	// we will use fake token from firebase emulator instead
	req, err := s.RequestStack.Peek()
	if err != nil {
		return err
	}
	studentID := req.(*dto.UserRegistrationRequest).Payload.Students[0].StudentID
	err = s.saveCredential(studentID, constant.UserGroupStudent, s.getSchoolId())
	if err != nil {
		return err
	}
	return nil
}

func (s *suite) studentSeesCourseWhichStudentJoinsOnLearnerApp() error {
	// Setup context
	time.Sleep(3 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupStudent
	ctx = contextWithTokenForGrpcCall(s, ctx)
	reqs, err := s.RequestStack.PeekMulti(2)
	if err != nil {
		return err
	}
	courseRequested := reqs[0].(*dto.MasterRegistrationRequest).Payload.Courses[0]

	// Check if student can see course on learner app by calling
	// RetrieveAccessibility api to get course IDs and
	// ListCourses api to get course info and compare with course in request
	retrieveAccessibilityResp, err := fatima_pbv1.NewAccessibilityReadServiceClient(s.fatimaConn).
		RetrieveAccessibility(ctx, &fatima_pbv1.RetrieveAccessibilityRequest{})
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
			ListCourses(ctx, &bob_pbv1.ListCoursesRequest{
				Paging: &common_pbv1.Paging{Limit: 100},
				Filter: &common_pbv1.CommonFilter{
					Ids: courseIDs,
				},
			})
		if err != nil {
			return err
		}
		for _, course := range retrieveCoursesResp.Items {
			if toJprepCourseID(courseRequested.CourseID) == course.Info.Id {
				courseRetrieved = course
				break
			}
		}
	}

	if courseRetrieved == nil {
		return fmt.Errorf(`student cannot see course`)
	}
	if courseRequested.CourseName != courseRetrieved.Info.Name {
		return fmt.Errorf(`expected course "name": %v but actual is %v`, courseRequested.CourseName, courseRetrieved.Info.Name)
	}

	return nil
}

func (s *suite) systemHasSyncedCourseAndClassesFromPartner(numClasses int) error {
	// now := time.Now().Format("2006/01/02")
	courseID := rand.Intn(1000)
	classes := []dto.Class{}
	for i := 1; i <= numClasses; i++ {
		classes = append(classes, dto.Class{
			ActionKind:     dto.ActionKindUpserted,
			ClassName:      "class name " + idutil.ULIDNow(),
			ClassID:        rand.Intn(99999999 + i*1000000),
			CourseID:       courseID,
			StartDate:      time.Now().Add(-48 * time.Hour).Format("2006/01/02"),
			EndDate:        time.Now().Add(48 * time.Hour).Format("2006/01/02"),
			AcademicYearID: 2021,
		})
	}
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Classes: classes,
			Courses: []dto.Course{
				{
					ActionKind:         dto.ActionKindUpserted,
					CourseID:           courseID,
					CourseName:         "course-name-with-actionKind-upsert",
					CourseStudentDivID: dto.CourseIDKid,
				},
			},
		},
	}
	err := s.attachValidJPREPSignature(request)
	if err != nil {
		return err
	}
	err = s.performMasterRegistrationRequest(request)
	if err != nil {
		return err
	}
	s.RequestStack.Push(request)

	return nil
}

func (s *suite) systemHasSyncedTeacherFromPartner() error {
	request := &dto.UserRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Students []dto.Student `json:"m_student"`
			Staffs   []dto.Staff   `json:"m_staff"`
		}{
			Staffs: []dto.Staff{
				{
					ActionKind: dto.ActionKindUpserted,
					StaffID:    idutil.ULIDNow(),
					Name:       "teacher name " + idutil.ULIDNow(),
				},
			},
		},
	}

	err := s.attachValidJPREPSignature(request)
	if err != nil {
		return err
	}
	err = s.performUserRegistrationRequest(request)
	if err != nil {
		return err
	}

	s.RequestStack.Push(request)
	return nil
}

func (s *suite) systemSyncsStudentAccountWhichAssociateWithAllAvailableCourseclass() error {
	req, err := s.RequestStack.Peek()
	if err != nil {
		return err
	}
	masterRegReq := req.(*dto.MasterRegistrationRequest)
	regularCourses := []struct {
		ClassID   int    `json:"m_course_id"`
		Startdate string `json:"startdate"`
		Enddate   string `json:"enddate"`
	}{}
	for _, class := range masterRegReq.Payload.Classes {
		regularCourses = append(regularCourses, struct {
			ClassID   int    `json:"m_course_id"`
			Startdate string `json:"startdate"`
			Enddate   string `json:"enddate"`
		}{
			ClassID:   class.ClassID,
			Startdate: time.Now().Add(-48 * time.Hour).Format("2006/01/02"),
			Enddate:   time.Now().Add(48 * time.Hour).Format("2006/01/02"),
		})
	}
	request := &dto.UserRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Students []dto.Student `json:"m_student"`
			Staffs   []dto.Staff   `json:"m_staff"`
		}{
			Students: []dto.Student{
				{
					ActionKind: dto.ActionKindUpserted,
					StudentID:  idutil.ULIDNow(),
					LastName:   "Last name " + idutil.ULIDNow(),
					GivenName:  "Given name " + idutil.ULIDNow(),
					StudentDivs: []struct {
						MStudentDivID int `json:"m_student_div_id"`
					}{
						{MStudentDivID: 1},
					},
					Regularcourses: regularCourses,
				},
			},
		},
	}
	err = s.attachValidJPREPSignature(request)
	if err != nil {
		return err
	}
	err = s.performUserRegistrationRequest(request)
	if err != nil {
		return err
	}

	s.RequestStack.Push(request)
	return nil
}

func (s *suite) teacherSeesThisStudentInfoOnTeacherApp() error {
	// Setup context
	reqs, err := s.RequestStack.PeekMulti(3)
	if err != nil {
		return err
	}
	teacherID := reqs[0].(*dto.UserRegistrationRequest).Payload.Staffs[0].StaffID
	err = s.saveCredential(teacherID, constant.UserGroupTeacher, s.getSchoolId())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupTeacher
	ctx = contextWithTokenForGrpcCall(s, ctx)

	classes := reqs[1].(*dto.MasterRegistrationRequest).Payload.Classes
	classIDs := []int{}
	for _, class := range classes {
		classIDs = append(classIDs, class.ClassID)
	}
	studentID := reqs[2].(*dto.UserRegistrationRequest).Payload.Students[0].StudentID

	limit := uint32(10)
	req := &bob_pbv1.RetrieveClassMembersRequest{
		Paging:    &common_pbv1.Paging{Limit: limit},
		ClassIds:  convertIntArrayToStringArray(classIDs),
		UserGroup: common_pbv1.UserGroup_USER_GROUP_STUDENT,
	}

	var res *bob_pbv1.RetrieveClassMembersResponse

	res, err = bob_pbv1.NewClassReaderServiceClient(s.bobConn).RetrieveClassMembers(ctx, req)
	if err != nil {
		return fmt.Errorf("ClassReaderService.RetrieveClassMembers: %w", err)
	}
	members := res.GetMembers()
	for _, member := range members {
		if member.UserId == studentID {
			return nil
		}
	}
	return fmt.Errorf("teacher does not see student id=%v in teacher app", studentID)
}

func (s *suite) attachValidJPREPSignature(request interface{}) error {
	data, err := json.Marshal(request)
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

func (s *suite) generateSignature(key, message string) (string, error) {
	sig := hmac.New(sha256.New, []byte(key))
	if _, err := sig.Write([]byte(message)); err != nil {
		return "", err
	}
	return hex.EncodeToString(sig.Sum(nil)), nil
}

func (s *suite) performUserRegistrationRequest(request interface{}) error {
	url := fmt.Sprintf("%s/jprep/user-registration", s.enigmaSrvURL)
	err := s.makeJPREPHTTPRequest(http.MethodPut, url, request)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) performMasterRegistrationRequest(request interface{}) error {
	url := fmt.Sprintf("%s/jprep/master-registration", s.enigmaSrvURL)
	err := s.makeJPREPHTTPRequest(http.MethodPut, url, request)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) makeJPREPHTTPRequest(method, url string, request interface{}) error {
	bodyRequest, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyRequest))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("JPREP-Signature", s.JPREPSignature)
	if err != nil {
		return err
	}

	client := http.Client{Timeout: time.Duration(30) * time.Second}
	s.RequestAt = &timestamppb.Timestamp{Seconds: time.Now().Unix()}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	if bodyBytes == nil {
		return fmt.Errorf("body is nil")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"expected status code %d, got %d in response %v",
			http.StatusOK, resp.StatusCode, resp,
		)
	}

	return nil
}

func toJprepCourseID(v int) string {
	return toJprepID("COURSE", v)
}

func toJprepID(typeID string, v int) string {
	return fmt.Sprintf("JPREP_%s_%09d", typeID, v)
}

func convertIntArrayToStringArray(ss []int) []string {
	result := make([]string, 0, len(ss))
	for _, element := range ss {
		val := strconv.Itoa(element)

		result = append(result, (val))
	}
	return result
}

func (s *suite) schoolAdminSeesTheEditedNameOnStudentstudyPlanPage() error {
	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupSchoolAdmin
	ctx = contextWithToken(s, ctx)

	// Pre-setup for hasura query using admin secret
	if err := trackTableForHasuraQuery(bobHasuraAdminUrl+"/v1/query", "users"); err != nil {
		return errors.Wrap(err, "trackTableForHasuraQuery()")
	}
	if err := createSelectPermissionForHasuraQuery(bobHasuraAdminUrl+"/v1/query", "users"); err != nil {
		return errors.Wrap(err, "createSelectPermissionForHasuraQuery()")
	}

	query :=
		`
			query ($user_id: String!) {
				users(where: {user_id: {_eq: $user_id}}) {
						user_id
						name
					}
			}
			`

	if err := addQueryToAllowListForHasuraQuery(bobHasuraAdminUrl+"/v1/query", query); err != nil {
		return errors.Wrap(err, "addQueryToAllowListForHasuraQuery()")
	}

	// Query user name by user id from hasura
	var profileQuery struct {
		Users []struct {
			UserID string `graphql:"user_id"`
			Name   string `graphql:"name"`
		} `graphql:"users(where: {user_id: {_eq: $user_id}})"`
	}

	req, err := s.RequestStack.Peek()
	if err != nil {
		return err
	}
	editedStudent := req.(*dto.UserRegistrationRequest).Payload.Students[0]

	variables := map[string]interface{}{
		"user_id": graphql.String(editedStudent.StudentID),
	}
	err = queryHasura(ctx, &profileQuery, variables, bobHasuraAdminUrl+"/v1/graphql")
	if err != nil {
		return errors.Wrap(err, "queryHasura")
	}

	for _, q := range profileQuery.Users {
		if q.Name == editedStudent.LastName {
			return nil
		}
	}

	return fmt.Errorf("school admin does not see edited student name=%v in student studyplan page", editedStudent.LastName)
}

func (s *suite) systemHasSyncedStudentAccountWhichAssociateWithAllAvailableClass() error {
	return s.systemSyncsStudentAccountWhichAssociateWithAllAvailableCourseclass()
}

func (s *suite) systemSyncsExistedStudentWithNewName() error {
	req, err := s.RequestStack.Peek()
	if err != nil {
		return err
	}
	student := req.(*dto.UserRegistrationRequest).Payload.Students[0]

	request := &dto.UserRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Students []dto.Student `json:"m_student"`
			Staffs   []dto.Staff   `json:"m_staff"`
		}{
			Students: []dto.Student{
				{
					ActionKind: dto.ActionKindUpserted,
					StudentID:  student.StudentID,
					LastName:   student.GivenName + " edited",
					GivenName:  student.LastName + " edited",
					StudentDivs: []struct {
						MStudentDivID int `json:"m_student_div_id"`
					}{
						{MStudentDivID: 1},
					},
				},
			},
		},
	}
	err = s.attachValidJPREPSignature(request)
	if err != nil {
		return err
	}
	err = s.performUserRegistrationRequest(request)
	if err != nil {
		return err
	}

	s.RequestStack.Push(request)
	return nil
}

func (s *suite) teacherSeesTheEditedNameOnTeacherApp() error {
	time.Sleep(2 * time.Second)
	// Setup context
	reqs, err := s.RequestStack.PeekMulti(4)
	if err != nil {
		return err
	}
	teacherID := reqs[0].(*dto.UserRegistrationRequest).Payload.Staffs[0].StaffID
	err = s.saveCredential(teacherID, constant.UserGroupTeacher, s.getSchoolId())
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupTeacher
	ctx = contextWithTokenForGrpcCall(s, ctx)

	classes := reqs[1].(*dto.MasterRegistrationRequest).Payload.Classes
	classIDs := []int{}
	for _, class := range classes {
		classIDs = append(classIDs, class.ClassID)
	}
	student := reqs[3].(*dto.UserRegistrationRequest).Payload.Students[0]

	limit := uint32(10)
	req := &bob_pbv1.RetrieveClassMembersRequest{
		Paging:    &common_pbv1.Paging{Limit: limit},
		ClassIds:  convertIntArrayToStringArray(classIDs),
		UserGroup: common_pbv1.UserGroup_USER_GROUP_STUDENT,
	}

	var res *bob_pbv1.RetrieveClassMembersResponse

	res, err = bob_pbv1.NewClassReaderServiceClient(s.bobConn).RetrieveClassMembers(ctx, req)
	if err != nil {
		return fmt.Errorf("ClassReaderService.RetrieveClassMembers: %w", err)
	}
	members := res.GetMembers()
	memberIDs := []string{}
	for _, member := range members {
		memberIDs = append(memberIDs, member.UserId)
	}

	getProfileReq := &bob_pb.GetStudentProfileRequest{
		StudentIds: memberIDs,
	}

	resp, err := bob_pb.NewStudentClient(s.bobConn).GetStudentProfile(ctx, getProfileReq)
	if err != nil {
		return err
	}
	for _, data := range resp.Datas {
		if data.Profile.Name == student.GivenName+" "+student.LastName {
			return nil
		}
	}

	return fmt.Errorf("teacher does not see edited student name=%v in teacher app", student.LastName)
}

func (s *suite) studentSeesTheEditedNameOnLearnerApp() error {
	time.Sleep(1 * time.Second)
	// Setup context
	req, err := s.RequestStack.Peek()
	if err != nil {
		return err
	}
	student := req.(*dto.UserRegistrationRequest).Payload.Students[0]
	s.saveCredential(student.StudentID, constant.UserGroupStudent, s.getSchoolId())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.UserGroupInContext = constant.UserGroupStudent
	ctx = contextWithTokenForGrpcCall(s, ctx)

	getProfileReq := &bob_pb.GetStudentProfileRequest{
		StudentIds: []string{student.StudentID},
	}

	resp, err := bob_pb.NewStudentClient(s.bobConn).GetStudentProfile(ctx, getProfileReq)
	if err != nil {
		return err
	}
	for _, data := range resp.Datas {
		if data.Profile.Name == student.GivenName+" "+student.LastName {
			return nil
		}
	}

	return fmt.Errorf("student does not see edited student name=%v in learner app", student.LastName)
}
