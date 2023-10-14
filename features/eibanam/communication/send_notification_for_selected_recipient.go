package communication

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bobPb "github.com/manabie-com/backend/pkg/genproto/bob"
	yasuoPb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"go.uber.org/multierr"
)

func (s *suite) hasAddedCourseForStudent(ctx context.Context, _, courseName, studentName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	student := stepState.studentInfos[studentName]
	course := stepState.courseInfos[courseName]
	student.courseIDs = append(student.courseIDs, course.id)
	return s.schoolAdminAddCoursesForStudent(ctx, student)
}

func (s *suite) newCourseInfo(schoolIDStr string) *yasuoPb.UpsertCoursesRequest_Course {
	courseID := idutil.ULIDNow()
	schoolID, err := strconv.ParseInt(schoolIDStr, 10, 64)
	if err != nil {
		return nil
	}
	id := rand.Int31()
	country := bobPb.COUNTRY_VN
	grade, _ := i18n.ConvertIntGradeToString(country, 7)

	c := &yasuoPb.UpsertCoursesRequest_Course{
		Id:       courseID,
		Name:     fmt.Sprintf("course-%d", id),
		Country:  country,
		Subject:  bobPb.SUBJECT_BIOLOGY,
		SchoolId: int32(schoolID),
		Grade:    grade,
	}
	return c
}

func (s *suite) hasCreatedCoursesAnd(ctx context.Context, role string, numCourse int, courseC1, courseC2 string) (context.Context, error) {
	courseNames := []string{courseC1, courseC2}
	return s.schoolAdminCreateCourses(ctx, courseNames)
}

func (s *suite) schoolAdminCreateCourses(ctx context.Context, courseNames []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	reqs := make([]*yasuoPb.UpsertCoursesRequest, 0)

	courses := make([]*yasuoPb.UpsertCoursesRequest_Course, 0)
	for _, courseName := range courseNames {
		c := s.newCourseInfo(stepState.schoolID)
		c.Name = courseName
		stepState.courseInfos[c.Name] = courseInfo{name: c.Name, id: c.Id}
		stepState.courseIDs = append(stepState.courseIDs, c.Id)
		courses = append(courses, c)
	}
	req := &yasuoPb.UpsertCoursesRequest{
		Courses: courses,
	}
	reqs = append(reqs, req)
	token := s.getToken(schoolAdmin)
	for _, req := range reqs {
		_, err := yasuoPb.NewCourseServiceClient(s.yasuoConn).UpsertCourses(contextWithToken(ctx, token), req)
		if err != nil {
			return ctx, err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasCreatedStudentWithGradeAndParentInfo(ctx context.Context, _, studentS2NameArg, parentP3NameArg string) (context.Context, error) {
	parentNames := []string{parentP3NameArg}
	return s.schoolAdminCreateStudentWithMultipleParent(ctx, studentS2NameArg, parentNames)
}

func (s *suite) hasCreatedStudentWithGradeAndParentParentInfo(ctx context.Context, _, studentS1NameArg, parentP1NameArg, parentP2NameArg string) (context.Context, error) {
	parentNames := []string{parentP1NameArg, parentP2NameArg}
	return s.schoolAdminCreateStudentWithMultipleParent(ctx, studentS1NameArg, parentNames)
}

func (s *suite) schoolAdminCreateStudentWithMultipleParent(ctx context.Context, studentName string, parentNames []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := s.newstudentInfo()
	req.StudentProfile.Name = studentName
	studentProf := profile{
		email: req.StudentProfile.Email,
		name:  req.StudentProfile.Name,
		credential: credential{
			password: req.StudentProfile.Password,
		},
	}
	parentProfs := make([]*profile, 0)
	for _, name := range parentNames {
		p := s.newParentInfo()
		p.Name = name

		prof := &profile{
			email: p.Email,
			name:  p.Name,
			credential: credential{
				password: p.Password,
			},
		}
		req.ParentProfiles = append(req.ParentProfiles, p)
		parentProfs = append(parentProfs, prof)
	}
	res, err := ypb.NewUserModifierServiceClient(s.yasuoConn).CreateStudent(contextWithToken(ctx, s.getToken(schoolAdmin)), req)
	if err != nil {
		return ctx, err
	}
	studentProf.id = res.GetStudentProfile().GetStudent().GetUserProfile().GetUserId()

	parentInfos := make([]parentInfo, 0)
	for i, p := range parentProfs {
		p.id = res.GetParentProfiles()[i].GetParent().GetUserProfile().GetUserId()
		parentInfos = append(parentInfos, parentInfo{
			id:    p.id,
			email: p.email,
			name:  p.name,
		})
	}
	stepState.studentInfos[studentName] = studentInfo{
		id:      studentProf.id,
		name:    studentProf.name,
		email:   studentProf.email,
		parents: parentInfos,
		grade:   req.StudentProfile.Grade,
	}

	stepState.users[studentName] = &studentProf
	for _, p := range parentProfs {
		stepState.users[p.name] = p
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) isAtNotificationPage(ctx context.Context, arg1 string) (context.Context, error) {
	return ctx, nil
}

func (s *suite) loginLearnerApp(ctx context.Context, userNamesArg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	userNames := strings.Split(userNamesArg, ",")
	for i := range userNames {
		userNames[i] = strings.TrimSpace(userNames[i])
	}

	for _, userName := range userNames {
		user := stepState.users[userName]
		group := s.getUserGroup(userName)
		err := try.Do(func(attempt int) (bool, error) {
			token, err := s.commuHelper.GenerateExchangeTokenCtx(ctx, user.id, group)
			if err != nil {
				return false, err
			}
			if err != nil {
				time.Sleep(1 * time.Second)
				return attempt < 5, err
			}

			user.token = token
			return false, nil
		})
		if err != nil {
			return ctx, err
		}
		stepState.users[userName] = user
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) schoolAdminHasCreatedNotification(ctx context.Context) (context.Context, error) {
	return s.schoolAdminHasSavedADraftNotificationWithRequiredFields(ctx)
}

func (s *suite) schoolAdminSendsANotificationToTheListIn(ctx context.Context, recipientTypeArg, courseTypeArg, gradeTypeArg, individualTypeArg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	recipientType := ""
	if choices := parseOneOf(recipientTypeArg); choices != nil {
		recipientType = selectOneOf(recipientTypeArg)
	}
	individualType := ""
	if choices := parseOneOf(recipientTypeArg); choices != nil {
		individualType = selectOneOf(individualTypeArg)
	}
	courseType := ""
	if choices := parseOneOf(courseTypeArg); choices != nil {
		courseType = selectOneOf(courseTypeArg)
	} else {
		courseType = courseTypeArg
	}
	gradeType := ""
	if choices := parseOneOf(gradeTypeArg); choices != nil {
		gradeType = selectOneOf(gradeTypeArg)
	} else {
		gradeType = gradeTypeArg
	}

	noti := stepState.notification
	noti = s.notificationWithRecipientType(recipientType, noti)
	if noti == nil {
		return ctx, fmt.Errorf("cannot parse recipientType %s", recipientType)
	}

	noti = s.notificationWithCourseType(ctx, courseType, noti)
	if noti == nil {
		return ctx, fmt.Errorf("cannot parse courseType %s", courseType)
	}

	noti = s.notificationWithGradeType(ctx, gradeType, noti)
	if noti == nil {
		return ctx, fmt.Errorf("cannot parse gradeType %s", gradeType)
	}

	// we can have many individuals email which separated by &
	individuals := strings.Split(individualType, "&")
	noti.ReceiverIds = make([]string, 0)
	for i := range individuals {
		individuals[i] = strings.TrimSpace(individuals[i])
		noti = s.notificationWithIndividualType(ctx, individuals[i], noti)
		if noti == nil {
			return ctx, fmt.Errorf("cannot parse individualType %s", individuals[i])
		}
	}

	// store infor recipient type and individual for later step
	stepState.notiRecipientType = recipientType
	for i := range individuals {
		individuals[i] = strings.TrimSpace(individuals[i])
		if individuals[i] == "empty" {
			continue
		}
		valid, userName := userNameIs(individuals[i])
		if !valid {
			return ctx, fmt.Errorf("cannot parse userName from argument %s", individuals[i])
		}
		user, ok := stepState.users[userName]
		if !ok {
			return ctx, fmt.Errorf("cannot find user %v", userName)
		}
		stepState.notiIndividuals = append(stepState.notiIndividuals, user)
	}
	var err error
	ctx, err = s.upsertNotification(ctx, noti)
	if err != nil {
		return ctx, fmt.Errorf("update notification error: %v", err)
	}
	return s.sendNotification(ctx, noti)
}

func (s *suite) whoRelatesToReceiveTheNotification(ctx context.Context, _, _ string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	recipientType := stepState.notiRecipientType
	recipient := make([]cpb.UserGroup, 0)
	switch recipientType {
	case "student and parent":
		recipient = append(recipient, cpb.UserGroup_USER_GROUP_STUDENT)
		recipient = append(recipient, cpb.UserGroup_USER_GROUP_PARENT)
	case "student":
		recipient = append(recipient, cpb.UserGroup_USER_GROUP_STUDENT)
	case "parent":
		recipient = append(recipient, cpb.UserGroup_USER_GROUP_PARENT)
	default:
		return ctx, nil
	}

	for _, user := range stepState.notiIndividuals {
		student := stepState.studentInfos[user.name]
		for _, gr := range recipient {
			var err error
			switch gr {
			case cpb.UserGroup_USER_GROUP_STUDENT:
				token := s.users[user.name].token
				ctx, err = s.RetrieveNotificationDetail(ctx, stepState.notification.NotificationId, token, student.id, cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW)
				if err != nil {
					return ctx, err
				}
			case cpb.UserGroup_USER_GROUP_PARENT:
				parents := student.parents
				for _, p := range parents {
					token := s.users[p.name].token
					ctx, err = s.RetrieveNotificationDetail(ctx, stepState.notification.NotificationId, token, p.id, cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW)
					if err != nil {
						return ctx, err
					}
				}
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkRecipientWithFilterCourseAndGrade(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseFilter := stepState.notification.TargetGroup.CourseFilter
	gradeFilter := stepState.notification.TargetGroup.GradeFilter

	studentIDs, err := s.studentIDsInCourse(ctx, courseFilter)
	if err != nil {
		return ctx, err
	}
	if len(studentIDs) > 0 {
		studentIDs, err = s.studentIDsWithGrade(ctx, studentIDs, gradeFilter)
		if err != nil {
			return ctx, err
		}
	}
	studentParents, err := s.findParentOfStudent(ctx, studentIDs)
	if err != nil {
		return ctx, err
	}

	studentInfos := make([]*studentInfo, 0)
	studentInfoMap := make(map[string]*studentInfo)
	for _, sp := range studentParents {
		pInfo := parentInfo{
			id: sp.ParentID.String,
		}
		if sti, ok := studentInfoMap[sp.StudentID.String]; ok {
			sti.parents = append(sti.parents, pInfo)
		} else {
			sInfo := &studentInfo{
				id:      sp.StudentID.String,
				parents: []parentInfo{pInfo},
			}
			studentInfoMap[sp.StudentID.String] = sInfo
		}
	}

	for _, sInfo := range studentInfoMap {
		studentInfos = append(studentInfos, sInfo)
	}

	ctx, err = s.checkRecipient(ctx, studentInfos)
	if err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentIDsInCourse(ctx context.Context, courseFilter *cpb.NotificationTargetGroup_CourseFilter) ([]string, error) {
	stepState := StepStateFromContext(ctx)
	switch courseFilter.Type {
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE:
		return nil, nil
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL:
		token := s.getToken(schoolAdmin)

		studentIDs := make([]string, 0)
		stream, err := epb.NewCourseReaderServiceClient(s.eurekaConn).ListStudentIDsByCourseV2(contextWithToken(ctx, token), &epb.ListStudentIDsByCourseV2Request{
			CourseIds: nil,
		})
		if err != nil {
			return nil, err
		}

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			if resp.StudentCourses != nil {
				studentIDs = append(studentIDs, resp.StudentCourses.StudentId)
			}
		}
		return studentIDs, nil
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST:
		studentIDs := make([]string, 0)
		courseIDs := courseFilter.CourseIds
		// search student id in course manually
		for _, student := range stepState.studentInfos {
			found := false
			for _, courseID := range student.courseIDs {
				if golibs.InArrayString(courseID, courseIDs) {
					found = true
					break
				}
			}
			if found {
				studentIDs = append(studentIDs, student.id)
			}
		}

		return studentIDs, nil
	default:
		return nil, nil
	}
}

func (s *suite) studentIDsWithGrade(ctx context.Context, studentIDs []string, gradeFilter *cpb.NotificationTargetGroup_GradeFilter) ([]string, error) {
	stepState := StepStateFromContext(ctx)
	repo := repositories.StudentRepo{}

	filter := repositories.FindStudentFilter{}
	err := multierr.Combine(
		filter.StudentIDs.Set(studentIDs),
		filter.SchoolID.Set(stepState.schoolID),
	)
	if err != nil {
		return nil, fmt.Errorf("set filter student by grade: %v", err)
	}

	students, err := repo.FindStudents(ctx, s.bobDB, filter)
	if err != nil {
		return nil, fmt.Errorf("StudentRepo.FindStudents: %v", err)
	}
	studentIDsWithGrade := make([]string, 0)
	for _, st := range students {
		studentIDsWithGrade = append(studentIDsWithGrade, st.ID.String)
	}

	return studentIDsWithGrade, nil
}

func (s *suite) findParentOfStudent(ctx context.Context, studentIDs []string) ([]*entities.StudentParent, error) {
	repo := repositories.StudentParentRepo{}
	studentParents, err := repo.GetStudentParents(ctx, s.bobDB, database.TextArray(studentIDs))
	if err != nil {
		return nil, err
	}
	return studentParents, nil
}

func (s *suite) checkRecipient(ctx context.Context, studentInfos []*studentInfo) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	recipientType := stepState.notiRecipientType
	recipient := make([]cpb.UserGroup, 0)
	switch recipientType {
	case "student and parent":
		recipient = append(recipient, cpb.UserGroup_USER_GROUP_STUDENT)
		recipient = append(recipient, cpb.UserGroup_USER_GROUP_PARENT)
	case "student":
		recipient = append(recipient, cpb.UserGroup_USER_GROUP_STUDENT)
	case "parent":
		recipient = append(recipient, cpb.UserGroup_USER_GROUP_PARENT)
	default:
		return ctx, nil
	}

	for _, student := range studentInfos {
		for _, gr := range recipient {
			switch gr {
			case cpb.UserGroup_USER_GROUP_STUDENT:
				token, err := s.commuHelper.GenerateExchangeTokenCtx(ctx, student.id, constant.UserGroupStudent)
				if err != nil {
					return ctx, err
				}
				ctx, err = s.RetrieveNotificationDetail(ctx, stepState.notification.NotificationId, token, student.id, cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW)
				if err != nil {
					return ctx, err
				}
			case cpb.UserGroup_USER_GROUP_PARENT:
				parents := student.parents
				for _, p := range parents {
					token, err := s.commuHelper.GenerateExchangeTokenCtx(ctx, p.id, constant.UserGroupParent)
					if err != nil {
						return ctx, err
					}
					ctx, err = s.RetrieveNotificationDetail(ctx, stepState.notification.NotificationId, token, p.id, cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW)
					if err != nil {
						return ctx, err
					}
				}
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func InArrayInt32(n int32, arr []int32) bool {
	for _, i := range arr {
		if n == i {
			return true
		}
	}
	return false
}

func (s *suite) matchingRecipientsReceiveTheNotification(ctx context.Context) (context.Context, error) {
	var err error
	ctx, err = s.checkRecipientWithFilterCourseAndGrade(ctx)
	if err != nil {
		return ctx, fmt.Errorf("check recipient with filter course and grade %r", err)
	}
	ctx, err = s.whoRelatesToReceiveTheNotification(ctx, "", "")
	if err != nil {
		return ctx, fmt.Errorf("who relates to receive noti %r", err)
	}

	return ctx, err
}

func userNameIs(str string) (bool, string) {
	matches := userWithEmailRegex.FindStringSubmatch(str)
	if matches == nil {
		return false, ""
	}
	return true, strings.TrimSpace(matches[1])
}

func userNameOfGradeIs(str string) (bool, string) {
	matches := userWithGradeRegex.FindStringSubmatch(str)
	if matches == nil {
		return false, ""
	}
	return true, strings.TrimSpace(matches[1])
}
