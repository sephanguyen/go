package communication

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	legacybpb "github.com/manabie-com/backend/pkg/genproto/bob"
	legacytpb "github.com/manabie-com/backend/pkg/genproto/tom"
	legacyypb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
)

func (s *suite) createChatGroupSteps() map[string]interface{} {
	return map[string]interface{}{
		`^"([^"]*)" logins CMS$`:         s.loginsCMS,
		`^"([^"]*)" logins Teacher App$`: s.loginsTeacherApp,
		`^"([^"]*)" logins Learner App$`: s.loginsLearnerApp,
		`^"([^"]*)" logins Learner App successfully with credentials which school admin gives$`:     s.loginsLearnerAppSuccessfullyWithCredentialsWhichSchoolAdminGives,
		`^school admin has created student with student info only$`:                                 s.schoolAdminHasCreatedStudentWithStudentInfoOnly,
		`^student sees student chat group on Learner App$`:                                          s.studentSeesStudentChatGroupOnLearnerApp,
		`^school admin has created student with parent info$`:                                       s.schoolAdminHasCreatedStudentWithParentInfo,
		`^teacher sees both student chat group & parent chat group in Unjoined tab on Teacher App$`: s.teacherSeesBothStudentChatGroupParentChatGroupInUnjoinedTabOnTeacherApp,
		`^all student\'s parents see same parent chat group on Learner App$`:                        s.allStudentsParentsSeeSameParentChatGroupOnLearnerApp,
		`^"([^"]*)" login Learner App successfully with credentials which school admin gives$`:      s.multiplePeopleloginLearnerAppSuccessfullyWithCredentialsWhichSchoolAdminGives,
		`^school admin has created student with many parents\' info$`:                               s.schoolAdminHasCreatedStudentWithManyParentsInfo,
		`^school admin created new student with same parent info$`:                                  s.schoolAdminCreatedNewStudentWithSameParentInfo,
		`^parent sees (\d+) chat groups on Learner App$`:                                            s.parentSeesNChatGroupsOnLearnerApp,
		`^"([^"]*)" sees new chat group on Learner App$`:                                            s.personSeesNewChatGroupOnLearnerApp,
		`^teacher sees (\d+) new chat groups in Unjoined tab on Teacher App$`:                       s.teacherSeesNewChatGroupsInUnjoinedTabOnTeacherApp,
	}
}

func (s *suite) insertAdminSchoolAdmin(schoolID int64) error {
	// new admin
	// ctx := auth.InjectFakeJwtToken(context.Background(), fmt.Sprint(schoolID))
	s.profile.schoolAdmin.email = idutil.ULIDNow() + "-admin@gamil.com"
	// adminID := idutil.ULIDNow()
	// err := s.insertNewAdmin(ctx, adminID, s.profile.admin.email)
	// if err != nil {
	// 	return fmt.Errorf("insertNewAdmin.Err %v", err)
	// }
	// firebaseToken, err := generateFakeAuthenticationToken(adminID, constant.UserGroupAdmin)
	// if err != nil {
	// 	return fmt.Errorf("generateFakeAuthenticationToken.Err %v", err)
	// }

	// authToken, err := helper.ExchangeToken(firebaseToken, adminID, constant.UserGroupAdmin, s.helper.ApplicantID, schoolID, s.shamirConn)
	// if err != nil {
	// 	return fmt.Errorf("failed to generate exchange token: %w", err)
	// }
	ctx, err := s.commuHelper.ASignedInWithSchool(contextWithResourcePath(context.Background(), strconv.Itoa(int(schoolID))), "school admin", int32(schoolID))
	if err != nil {
		return err
	}
	commonStepState := common.StepStateFromContext(ctx)
	authToken := commonStepState.AuthToken
	s.profile.schoolAdmin.id = commonStepState.CurrentUserID
	// tok := s.commuHelper.AuthToken
	s.updateToken(schoolAdmin, authToken)

	// new school admin
	// err = s.insertNewSchoolAdmin()
	// if err != nil {
	// 	return fmt.Errorf("insertNewSchoolAdmin.Err %v", err)
	// }
	return nil
}

// create teacher using school admin account, but override teacher password by admin account
func (s *suite) schoolAdminCreateNewTeacher(ctx context.Context) (context.Context, error) {
	intschool, err := strconv.ParseInt(s.schoolID, 10, 64)
	if err != nil {
		return ctx, err
	}
	req := &legacyypb.CreateUserRequest{
		UserGroup: legacyypb.USER_GROUP_TEACHER,
		SchoolId:  intschool,
	}
	id := idutil.ULIDNow()
	num := rand.Int()
	email := fmt.Sprintf("e2euser-%s+%d@gmail.com", id, num)
	user := &legacyypb.CreateUserProfile{
		Name:        fmt.Sprintf("e2euser-%s", id),
		PhoneNumber: fmt.Sprintf("+848%d", num),
		Email:       email,
		Country:     legacybpb.COUNTRY_VN,
		Grade:       1,
	}
	req.Users = []*legacyypb.CreateUserProfile{user}
	svc := legacyypb.NewUserServiceClient(s.yasuoConn)
	ctx2 := contextWithToken(ctx, s.getToken(schoolAdmin))

	res, err := svc.CreateUser(ctx2, req)
	if err != nil {
		return ctx, err
	}

	teacherID := res.GetUsers()[0].GetId()
	s.profile.defaultTeacher = profile{
		email: email,
		id:    teacherID,
	}
	// newPassword := idutil.ULIDNow()
	// req2 := &upb.ReissueUserPasswordRequest{
	// 	UserId:      teacherID,
	// 	NewPassword: newPassword,
	// }
	// s.setPassword(teacher, newPassword)

	// ctx2, cancel2 := contextWithTokenAndTimeOut(context.Background(), s.getToken("school admin"))
	// defer cancel2()
	authToken, err := s.commuHelper.GenerateExchangeTokenCtx(ctx, teacherID, cpb.UserGroup_USER_GROUP_TEACHER.String())
	// res2, err := upb.NewUserModifierServiceClient(s.usermgmtConn).ReissueUserPassword(ctx2, req2)
	// if err != nil {
	// 	return err
	// }
	// if !res2.GetSuccessful() {
	// 	return fmt.Errorf("failed to changed password of teacher")
	// }
	// authToken, err := s.loginFirebaseAccount(context.Background(), s.profile.defaultTeacher.email, newPassword)
	if err != nil {
		return ctx, err
	}
	// ctx3, cancel3 := context.WithTimeout(context.Background(), timeout)
	// defer cancel3()
	// tokenRes, err := bpb.NewUserModifierServiceClient(s.bobConn).ExchangeToken(
	// 	ctx3, &bpb.ExchangeTokenRequest{Token: authToken})
	// if err != nil {
	// 	return err
	// }

	s.updateToken(teacher, authToken)

	return ctx, nil
}

func (s *suite) loginsLearnerAppSuccessfullyWithCredentialsWhichSchoolAdminGives(ctx context.Context, person string) (context.Context, error) {
	st := StepStateFromContext(ctx)
	alloweds := []string{
		student, parent, "newly created student", newParent, existingParent, "parent P1",
	}
	valid := false
	for _, allowedPerson := range alloweds {
		if allowedPerson == person {
			valid = true
			break
		}
	}
	if !valid {
		return ctx, fmt.Errorf("do not support person %s to login learner app", person)
	}

	id := st.getID(person)
	err := try.Do(func(attempt int) (bool, error) {
		token, err := s.commuHelper.GenerateExchangeTokenCtx(ctx, id, s.getUserGroup(person))
		if err != nil {
			return false, err
		}
		st.updateToken(person, token)
		return false, nil
	})
	if err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, st), nil
}

func (s *suite) newstudentInfo() *ypb.CreateStudentRequest {
	intSchool, _ := strconv.ParseInt(s.schoolID, 10, 64)
	randomID := idutil.ULIDNow()
	password := fmt.Sprintf("password-%v", randomID)
	email := fmt.Sprintf("%v@example.com", randomID)
	// name := fmt.Sprintf("user-%v", randomID)
	name := randomNameFromSamples()
	req := &ypb.CreateStudentRequest{
		SchoolId: int32(intSchool),
		StudentProfile: &ypb.CreateStudentRequest_StudentProfile{
			Email:            email,
			Password:         password,
			Name:             name,
			CountryCode:      cpb.Country_COUNTRY_VN,
			PhoneNumber:      fmt.Sprintf("phone-number-%v", randomID),
			Grade:            5,
			EnrollmentStatus: cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
		},
	}
	return req
}

// TODO: use state from context
func (s *suite) schoolAdminHasCreatedStudentWithStudentInfoOnly(ctx context.Context) (context.Context, error) {
	st := StepStateFromContext(ctx)
	// req := s.newstudentInfo()
	authCtx := s.commuHelper.CtxWithAuthToken(ctx, s.getToken(schoolAdmin))
	stu, err := s.commuHelper.Suite.CreateStudent(authCtx, nil, nil)
	if err != nil {
		return ctx, err
	}
	s.profile.defaultStudent = profile{
		email: stu.UserProfile.Email,
		name:  stu.UserProfile.Name,
		id:    stu.UserProfile.UserId,
	}

	st.profile = s.profile
	return StepStateToContext(ctx, st), nil
}

func (s *suite) newParentInfo() *ypb.CreateStudentRequest_ParentProfile {
	randomStr := idutil.ULIDNow()
	return &ypb.CreateStudentRequest_ParentProfile{
		Email:        fmt.Sprintf("parent-%s@gmail.com", randomStr),
		Password:     randomStr,
		Name:         fmt.Sprintf("parent-%s", randomStr),
		PhoneNumber:  fmt.Sprintf("+84%d", rand.Int()),
		CountryCode:  cpb.Country_COUNTRY_VN,
		Relationship: ypb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
	}
}

func (s *suite) createStudentWithParent(ctx context.Context) (studentprof, parentprof profile, studentcred, parentcred credential, err error) {
	authedCtx := s.commuHelper.CtxWithAuthToken(ctx, s.getToken(schoolAdmin))
	stu, par, err2 := s.commuHelper.CreateStudentWithParent(authedCtx, nil, nil)
	if err2 != nil {
		err = err2
		return
	}

	studentprof = profile{
		id:    stu.UserProfile.UserId,
		email: stu.UserProfile.Email,
		name:  stu.UserProfile.Name,
	}

	parentprof = profile{
		id:    par.UserProfile.UserId,
		email: par.UserProfile.Email,
		name:  par.UserProfile.Name,
	}
	return
}

func (s *suite) schoolAdminHasCreatedStudentWithParentInfo(ctx context.Context) (context.Context, error) {
	st := StepStateFromContext(ctx)
	student, parent, studentcred, parentcred, err := s.createStudentWithParent(ctx)
	if err != nil {
		return ctx, err
	}
	st.profile.defaultParent = parent
	st.profile.defaultStudent = student
	st.credentials[student.email] = studentcred
	st.credentials[parent.email] = parentcred
	return StepStateToContext(ctx, st), nil
}

func eventually(fn func() error) error {
	return try.Do(func(attempt int) (bool, error) {
		err := fn()
		if err != nil {
			time.Sleep(2 * time.Second)
			return attempt < 10, err
		}
		return false, nil
	})
}

func (s *suite) parentSeesNChatGroupsOnLearnerApp(chatNum int) error {
	err := eventually(func() error {
		res, err := s.personGetChatOnLearnerApp(parent)
		if err != nil {
			return err
		}
		convs := res.GetConversations()
		if len(convs) != chatNum {
			return fmt.Errorf("expect parent to have %d conversations, %d returned", chatNum, len(convs))
		}
		studentsCheck := map[string]bool{}
		studentsCheck[s.profile.defaultStudent.id] = true
		studentsCheck[s.profile.newlyCreatedStudent.id] = true
		for _, item := range convs {
			if !studentsCheck[item.GetStudentId()] {
				return fmt.Errorf("chat for student %s is not expected", item.GetStudentId())
			}
		}
		return nil
	})
	return err
}

// func (s *suite) parentSeesParentChatGroupOnLearnerApp() error {
// 	return s.parentSeesNChatGroupsOnLearnerApp(1)
// }

func (s *suite) studentSeesStudentChatGroupOnLearnerApp() error {
	return s.personSeesNewChatGroupOnLearnerApp(student)
}
func (s *suite) teacherSeesNewChatGroupsInUnjoinedTabOnTeacherApp(numchat int) error {
	err := try.Do(func(attempt int) (bool, error) {
		req := &tpb.ListConversationsInSchoolRequest{
			Paging: &cpb.Paging{
				Limit: 100,
			},
			JoinStatus: tpb.ConversationJoinStatus_CONVERSATION_JOIN_STATUS_NOT_JOINED,
		}
		ctx, cancel := contextWithTokenAndTimeOut(context.Background(), s.getToken(teacher))
		defer cancel()
		res, err := tpb.NewChatReaderServiceClient(s.tomConn).ListConversationsInSchool(ctx, req)
		if err != nil {
			return false, err
		}
		if len(res.GetItems()) != numchat {
			time.Sleep(2 * time.Second)
			return attempt < 5, fmt.Errorf("expect %d conversations, %d returned", numchat, len(res.GetItems()))
		}
		return false, nil
	})
	return err
}

// func (s *suite) teacherSeesListOfChatGroupsOnTeacherApp() error {
// 	return nil
// }

// func (s *suite) schoolAdminCompletesEdit() error { return nil }

// func (s *suite) schoolAdminSelectsToEditThatStudent() error { return nil }

// func (s *suite) schoolAdminHasCreatedStudentWithoutParentInfo(ctx context.Context) (context.Context, error) {
// 	return s.schoolAdminHasCreatedStudentWithStudentInfoOnly(ctx)
// }

// func (s *suite) teacherGoesToMessageScreen() error {
// 	req := &tpb.ListConversationsInSchoolRequest{
// 		Paging: &cpb.Paging{
// 			Limit: 100,
// 		},
// 		JoinStatus: tpb.ConversationJoinStatus_CONVERSATION_JOIN_STATUS_NOT_JOINED,
// 	}
// 	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), s.getToken(teacher))
// 	defer cancel()
// 	res, err := tpb.NewChatReaderServiceClient(s.tomConn).ListConversationsInSchool(ctx, req)
// 	if err != nil {
// 		return err
// 	}
// 	s.studentChatState.teacherChats = res
// 	return nil
// }

// use default student and default parent to check
func (s *suite) teacherSeesBothStudentChatGroupParentChatGroupInUnjoinedTabOnTeacherApp() error {
	req := &tpb.ListConversationsInSchoolRequest{
		Paging: &cpb.Paging{
			Limit: 100,
		},
		JoinStatus: tpb.ConversationJoinStatus_CONVERSATION_JOIN_STATUS_NOT_JOINED,
	}
	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), s.getToken(teacher))
	defer cancel()
	res, err := tpb.NewChatReaderServiceClient(s.tomConn).ListConversationsInSchool(ctx, req)
	if err != nil {
		return err
	}
	if len(res.GetItems()) != 2 {
		return fmt.Errorf("expect 2 conversations, %d returned", len(res.GetItems()))
	}

	studentName := s.profile.defaultStudent.name
	var getParent, getStudent bool
	for _, chat := range res.GetItems() {
		if chat.GetConversationName() != studentName {
			return fmt.Errorf("expect returned chat to have student's name, %s returned", chat.GetConversationName())
		}
		if chat.GetConversationType() == tpb.ConversationType_CONVERSATION_STUDENT {
			getStudent = true
		}
		if chat.GetConversationType() == tpb.ConversationType_CONVERSATION_PARENT {
			getParent = true
		}
	}
	if !getParent || !getStudent {
		return fmt.Errorf("either parent chat is not returned: %v or student chat is not returned: %v", getParent, getStudent)
	}
	return nil
}

func (s *suite) allStudentsParentsSeeSameParentChatGroupOnLearnerApp() error {
	tokens := s.getParentTokens()
	uniqueID := ""

	for _, tok := range tokens {
		thisParentToken := tok
		err := eventually(func() error {
			req := &legacytpb.ConversationListRequest{
				Limit: 100,
			}
			ctx, cancel := contextWithTokenAndTimeOut(context.Background(), thisParentToken)
			defer cancel()
			res, err := legacytpb.NewChatServiceClient(s.tomConn).ConversationList(ctx, req)
			if err != nil {
				return err
			}
			convs := res.GetConversations()
			if len(convs) != 1 {
				return fmt.Errorf("expect parent to only have one conversation, %d returned", len(convs))
			}

			id := convs[0].GetConversationId()
			if uniqueID == "" {
				uniqueID = id
				return nil
			}
			if id != uniqueID {
				return fmt.Errorf("parents have different chat id: %s and %s eventhough they have same children", id, uniqueID)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *suite) personSeesNewChatGroupOnLearnerApp(person string) error {
	req := &legacytpb.ConversationListRequest{
		Limit: 100,
	}
	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), s.getToken(person))
	defer cancel()
	res, err := legacytpb.NewChatServiceClient(s.tomConn).ConversationList(ctx, req)
	if err != nil {
		return err
	}

	convs := res.GetConversations()
	if len(convs) != 1 {
		return fmt.Errorf("expect student to only have one conversation, %d returned", len(convs))
	}
	prof := s.getProfile(person)
	if prof.id != convs[0].StudentId {
		return fmt.Errorf("returned conversation does not have student's id equal to %s's id", person)
	}
	return nil
}

// use default parent to assign for newly created student
func (s *suite) schoolAdminCreatedNewStudentWithSameParentInfo(ctx context.Context) (context.Context, error) {
	authed := s.commuHelper.CtxWithAuthToken(ctx, s.getToken(schoolAdmin))
	stu, err := s.commuHelper.Suite.CreateStudent(authed, nil, nil)
	if err != nil {
		return ctx, err
	}

	par := s.profile.defaultParent
	err = s.commuHelper.UpdateStudentParent(authed, stu.UserProfile.UserId, par.id, par.email)
	if err != nil {
		return ctx, err
	}

	s.profile.newlyCreatedStudent = profile{
		email: stu.UserProfile.Email,
		name:  stu.UserProfile.Name,
		id:    stu.UserProfile.UserId,
	}

	return ctx, nil
}

func (s *suite) schoolAdminHasCreatedStudentWithManyParentsInfo(ctx context.Context) (context.Context, error) {
	authedCtx := s.commuHelper.CtxWithAuthToken(ctx, s.getToken(schoolAdmin))
	stu, par1, err := s.commuHelper.CreateStudentWithParent(authedCtx, nil, nil)
	if err != nil {
		return ctx, nil
	}
	par2, err := s.commuHelper.CreateParentForStudent(authedCtx, stu.UserProfile.UserId)
	if err != nil {
		return ctx, nil
	}
	// req := s.newstudentInfo()
	// parent1 := s.newParentInfo()
	// parent2 := s.newParentInfo()
	profiles := []profile{}
	profiles = append(profiles, profile{
		email: par1.UserProfile.Email,
		name:  par1.UserProfile.Name,
		id:    par1.UserProfile.UserId,
	})
	profiles = append(profiles, profile{
		id:    par2.UserProfile.UserId,
		email: par2.UserProfile.Email,
		name:  par2.UserProfile.Name,
	})

	s.profile.defaultStudent = profile{
		id:    stu.UserProfile.UserId,
		email: stu.UserProfile.Email,
		name:  stu.UserProfile.Name,
	}

	s.profile.multipleParents = profiles

	return ctx, nil
}

func (s *suite) multiplePeopleloginLearnerAppSuccessfullyWithCredentialsWhichSchoolAdminGives(ctx context.Context, people string) (context.Context, error) {
	if people != "parents" {
		return ctx, fmt.Errorf("only accept multiple parent")
	}
	// mails := s.getParentMails()
	for idx := range s.profile.multipleParents {
		id := s.profile.multipleParents[idx].id
		m := s.profile.multipleParents[idx].email
		// cred := s.credentials[m]
		err := try.Do(func(attempt int) (bool, error) {
			token, err := s.commuHelper.GenerateExchangeTokenCtx(ctx, id, cpb.UserGroup_USER_GROUP_PARENT.String())
			// token, err := s.loginFirebaseAccount(context.Background(), m, cred.password)
			if err != nil {
				time.Sleep(1 * time.Second)
				return attempt < 5, err
			}

			// ctx, cancel := context.WithTimeout(context.Background(), timeout)
			// defer cancel()
			// res, err := bpb.NewUserModifierServiceClient(s.bobConn).ExchangeToken(
			// 	ctx, &bpb.ExchangeTokenRequest{Token: token})
			// if err != nil {
			// 	return false, err
			// }
			s.credentials[m] = credential{
				// password: cred.password,
				token: token,
			}
			return false, nil
		})
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

// use both state from context and suite
// TODO: only use state from context
func (s *suite) loginsTeacherApp(ctx context.Context, person string) (context.Context, error) {
	if person != teacher {
		return ctx, fmt.Errorf("expecting only teacher to login teacher app")
	}
	ctx, err := s.schoolAdminCreateNewTeacher(ctx)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, s.stepState), nil
}

// use fake token
// this function is shared between old and new e2e script, new one uses state from StepStateFromContext
// and old one uses suite state.
// TODO: migrate old state from suite to state in context for all steps
func (s *suite) loginsCMS(ctx context.Context, accountType string) (context.Context, error) {
	var schoolID int
	switch accountType {
	case schoolAdmin:
		newschool, _, _, err := s.commuHelper.Suite.NewOrgWithOrgLocation(ctx)
		// newschool, err := s.insertNewSchool()
		if err != nil {
			return ctx, fmt.Errorf("insertNewSchool.Err %v", err)
		}

		schoolID = int(newschool)
	case jprepSchoolAdmin:
		err := s.helper.InsertJprepSchool()
		if err != nil {
			return ctx, fmt.Errorf("InsertJprepSchool.Err %v", err)
		}
		schoolID = constants.JPREPSchool
	default:
		panic(fmt.Sprintf("unsupported %s logging in CMS", accountType))
	}

	s.schoolID = strconv.Itoa(schoolID)
	ctx = contextWithResourcePath(ctx, s.schoolID)
	err := s.insertAdminSchoolAdmin(int64(schoolID))
	if err != nil {
		return ctx, fmt.Errorf("insertAdminSchoolAdmin.Err %v", err)
	}
	step := StepStateFromContext(ctx)
	step.SchoolID = s.schoolID
	return StepStateToContext(ctx, step), nil
}

func (s *suite) personGetChatOnLearnerApp(person string) (*legacytpb.ConversationListResponse, error) {
	req := &legacytpb.ConversationListRequest{
		Limit: 100,
	}
	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), s.getToken(person))
	defer cancel()
	res, err := legacytpb.NewChatServiceClient(s.tomConn).ConversationList(ctx, req)
	return res, err
}
