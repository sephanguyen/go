package mappers

import (
	"fmt"
	"strconv"

	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	natspb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func PbToInfoNotificationEnt(notification *cpb.Notification) (*entities.InfoNotification, error) {
	scheduledAt := database.TimestamptzFromPb(notification.ScheduledAt)

	e := &entities.InfoNotification{}
	database.AllNullEntity(e)
	err := multierr.Combine(
		e.NotificationID.Set(notification.NotificationId),
		e.NotificationMsgID.Set(notification.Message.NotificationMsgId),
		e.Type.Set(notification.Type.String()),
		e.EditorID.Set(notification.EditorId),
		e.CreatedUserID.Set(notification.CreatedUserId),
		e.ReceiverIDs.Set(notification.ReceiverIds),
		e.Event.Set(notification.Event.String()),
		e.Status.Set(notification.Status.String()),
		e.Owner.Set(notification.SchoolId),
		e.ScheduledAt.Set(scheduledAt),
		e.QuestionnaireID.Set(nil),
		e.IsImportant.Set(notification.IsImportant),
		e.GenericReceiverIDs.Set(notification.GenericReceiverIds),
		e.ExcludedGenericReceiverIDs.Set(notification.ExcludedGenericReceiverIds),
	)
	if err != nil {
		return nil, err
	}

	if notification.Data != "" {
		err := e.Data.Set(notification.Data)
		if err != nil {
			return nil, err
		}
	}

	targetEnt := PbToNotificationTargetEnt(notification.TargetGroup)

	err = e.TargetGroups.Set(targetEnt)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func PbToInfoNotificationMsgEnt(notificationMsg *cpb.NotificationMessage) (*entities.InfoNotificationMsg, error) {
	e := &entities.InfoNotificationMsg{}
	database.AllNullEntity(e)
	err := multierr.Combine(
		e.NotificationMsgID.Set(notificationMsg.NotificationMsgId),
		e.Title.Set(notificationMsg.Title),
		e.Content.Set(&entities.RichText{
			Raw:         notificationMsg.Content.Raw,
			RenderedURL: notificationMsg.Content.Rendered,
		}),
		e.MediaIDs.Set(notificationMsg.MediaIds),
	)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func ToStudentNotificationEnts(students []*entities.Audience, notificationID string, studentGradeMap map[string]int, studentCoursesMap map[string][]string) ([]*entities.UserInfoNotification, error) {
	ents := make([]*entities.UserInfoNotification, 0, len(students))
	checkList := map[string]struct{}{}
	for _, st := range students {
		_, exist := checkList[st.StudentID.String]
		if exist {
			continue
		}
		checkList[st.StudentID.String] = struct{}{}
		un, err := toUserNotificationEnt(st.StudentID.String, notificationID, false, cpb.UserGroup_USER_GROUP_STUDENT.String(), "", st.StudentID.String, studentGradeMap, studentCoursesMap)
		if err != nil {
			return nil, fmt.Errorf("can not get user info notification for student: %v", err)
		}

		ents = append(ents, un)
	}
	return ents, nil
}

func ToParentNotificationEnts(studentParent []*bobEntities.StudentParent, notificationID string) ([]*entities.UserInfoNotification, error) {
	ents := make([]*entities.UserInfoNotification, 0, len(studentParent))

	for _, sp := range studentParent {
		parentID := sp.ParentID.String
		studentID := sp.StudentID.String
		un, err := toUserNotificationEnt(parentID, notificationID, false, cpb.UserGroup_USER_GROUP_PARENT.String(), parentID, studentID, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("can not get user info notification for student: %v", err)
		}

		ents = append(ents, un)
	}
	return ents, nil
}

func toUserNotificationEnt(userID, notificationID string, isIndividual bool, userGroup string, parentID string, studentID string, userGradeMap map[string]int, userCoursesMap map[string][]string) (*entities.UserInfoNotification, error) {
	un := &entities.UserInfoNotification{}
	database.AllNullEntity(un)
	err := multierr.Combine(
		un.NotificationID.Set(notificationID),
		un.UserID.Set(userID),
		un.Status.Set(cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String()),
		un.QuestionnaireStatus.Set(cpb.UserNotificationQuestionnaireStatus_USER_NOTIFICATION_QUESTIONNAIRE_STATUS_UNANSWERED),
		un.IsIndividual.Set(isIndividual),
	)
	if err != nil {
		return nil, fmt.Errorf("set user infor notification %v", err)
	}

	if userCoursesMap != nil {
		if courses, ok := userCoursesMap[userID]; ok {
			_ = un.Courses.Set(courses)
		}
	}
	if userGradeMap != nil {
		if grade, ok := userGradeMap[userID]; ok {
			_ = un.CurrentGrade.Set(grade)
		}
	}

	if userGroup != "" {
		_ = un.UserGroup.Set(userGroup)
	}

	if parentID != "" {
		_ = un.ParentID.Set(parentID)
	}

	if studentID != "" {
		_ = un.StudentID.Set(studentID)
	}

	return un, nil
}

func PbToNotificationTargetEnt(src *cpb.NotificationTargetGroup) *entities.InfoNotificationTarget {
	var i entities.InfoNotificationTarget
	switch src.CourseFilter.Type {
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE:
		i.CourseFilter = entities.InfoNotificationTarget_CourseFilter{
			Type:      src.CourseFilter.Type.String(),
			CourseIDs: make([]string, 0),
		}
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL:
		i.CourseFilter = entities.InfoNotificationTarget_CourseFilter{
			Type:      src.CourseFilter.Type.String(),
			CourseIDs: nil,
		}
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST:
		i.CourseFilter = entities.InfoNotificationTarget_CourseFilter{
			Type:      src.CourseFilter.Type.String(),
			CourseIDs: src.CourseFilter.CourseIds,
			Courses:   PbToTargetGroupFilterCoursesEnt(src.CourseFilter.Courses),
		}
	}

	switch src.GradeFilter.Type {
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE:
		i.GradeFilter = entities.InfoNotificationTarget_GradeFilter{
			Type:     src.GradeFilter.Type.String(),
			GradeIDs: make([]string, 0),
		}
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL:
		i.GradeFilter = entities.InfoNotificationTarget_GradeFilter{
			Type:     src.GradeFilter.Type.String(),
			GradeIDs: nil,
		}
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST:
		i.GradeFilter = entities.InfoNotificationTarget_GradeFilter{
			Type:     src.GradeFilter.Type.String(),
			GradeIDs: src.GradeFilter.GradeIds,
		}
	}

	switch src.LocationFilter.Type {
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE:
		i.LocationFilter = entities.InfoNotificationTarget_LocationFilter{
			Type:        src.LocationFilter.Type.String(),
			LocationIDs: make([]string, 0),
		}
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL:
		i.LocationFilter = entities.InfoNotificationTarget_LocationFilter{
			Type:        src.LocationFilter.Type.String(),
			LocationIDs: nil,
		}
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST:
		i.LocationFilter = entities.InfoNotificationTarget_LocationFilter{
			Type:        src.LocationFilter.Type.String(),
			LocationIDs: src.LocationFilter.LocationIds,
			Locations:   PbToTargetGroupFilterLocationsEnt(src.LocationFilter.Locations),
		}
	}

	switch src.ClassFilter.Type {
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE:
		i.ClassFilter = entities.InfoNotificationTarget_ClassFilter{
			Type:     src.ClassFilter.Type.String(),
			ClassIDs: make([]string, 0),
		}
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL:
		i.ClassFilter = entities.InfoNotificationTarget_ClassFilter{
			Type:     src.ClassFilter.Type.String(),
			ClassIDs: nil,
		}
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST:
		i.ClassFilter = entities.InfoNotificationTarget_ClassFilter{
			Type:     src.ClassFilter.Type.String(),
			ClassIDs: src.ClassFilter.ClassIds,
			Classes:  PbToTargetGroupFilterClassesEnt(src.ClassFilter.Classes),
		}
	}

	switch src.SchoolFilter.Type {
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE:
		i.SchoolFilter = entities.InfoNotificationTarget_SchoolFilter{
			Type:      src.SchoolFilter.Type.String(),
			SchoolIDs: make([]string, 0),
		}
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL:
		i.SchoolFilter = entities.InfoNotificationTarget_SchoolFilter{
			Type:      src.SchoolFilter.Type.String(),
			SchoolIDs: nil,
		}
	case cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST:
		i.SchoolFilter = entities.InfoNotificationTarget_SchoolFilter{
			Type:      src.SchoolFilter.Type.String(),
			SchoolIDs: src.SchoolFilter.SchoolIds,
			Schools:   PbToTargetGroupFilterSchoolsEnt(src.SchoolFilter.Schools),
		}
	}

	for _, gr := range src.UserGroupFilter.UserGroups {
		i.UserGroupFilter.UserGroups = append(i.UserGroupFilter.UserGroups, gr.String())
	}
	return &i
}

func PbToTargetGroupFilterCoursesEnt(courses []*cpb.NotificationTargetGroup_CourseFilter_Course) []entities.InfoNotificationTarget_CourseFilter_Course {
	entCourses := make([]entities.InfoNotificationTarget_CourseFilter_Course, 0, len(courses))
	for _, course := range courses {
		entCourses = append(entCourses, entities.InfoNotificationTarget_CourseFilter_Course{
			CourseID:   course.CourseId,
			CourseName: course.CourseName,
		})
	}
	return entCourses
}

func PbToTargetGroupFilterLocationsEnt(locations []*cpb.NotificationTargetGroup_LocationFilter_Location) []entities.InfoNotificationTarget_LocationFilter_Location {
	entLocations := make([]entities.InfoNotificationTarget_LocationFilter_Location, 0, len(locations))
	for _, location := range locations {
		entLocations = append(entLocations, entities.InfoNotificationTarget_LocationFilter_Location{
			LocationID:   location.LocationId,
			LocationName: location.LocationName,
		})
	}
	return entLocations
}

func PbToTargetGroupFilterClassesEnt(classes []*cpb.NotificationTargetGroup_ClassFilter_Class) []entities.InfoNotificationTarget_ClassFilter_Class {
	entClasses := make([]entities.InfoNotificationTarget_ClassFilter_Class, 0, len(classes))
	for _, class := range classes {
		entClasses = append(entClasses, entities.InfoNotificationTarget_ClassFilter_Class{
			ClassID:   class.ClassId,
			ClassName: class.ClassName,
		})
	}
	return entClasses
}

func PbToTargetGroupFilterSchoolsEnt(schools []*cpb.NotificationTargetGroup_SchoolFilter_School) []entities.InfoNotificationTarget_SchoolFilter_School {
	entSchools := make([]entities.InfoNotificationTarget_SchoolFilter_School, 0, len(schools))
	for _, school := range schools {
		entSchools = append(entSchools, entities.InfoNotificationTarget_SchoolFilter_School{
			SchoolID:   school.SchoolId,
			SchoolName: school.SchoolName,
		})
	}
	return entSchools
}

func PbToQuestionnaireUserAnswerEnts(userID string, targetID string, answerReq *npb.SubmitQuestionnaireRequest) (entities.QuestionnaireUserAnswers, error) {
	questionnaireUserAnswers := make(entities.QuestionnaireUserAnswers, 0)

	for _, answer := range answerReq.Answers {
		if answer.Answer != "" {
			e := &entities.QuestionnaireUserAnswer{}
			database.AllNullEntity(e)

			err := multierr.Combine(
				e.UserNotificationID.Set(answerReq.UserInfoNotificationId),
				e.QuestionnaireQuestionID.Set(answer.QuestionnaireQuestionId),
				e.Answer.Set(answer.Answer),
				e.UserID.Set(userID),
				e.TargetID.Set(targetID),
			)

			if err != nil {
				return nil, err
			}

			questionnaireUserAnswers = append(questionnaireUserAnswers, e)
		}
	}

	return questionnaireUserAnswers, nil
}

func PbToQuestionnaireEnt(qn *cpb.Questionnaire) (*entities.Questionnaire, error) {
	e := &entities.Questionnaire{}
	database.AllNullEntity(e)
	err := multierr.Combine(
		e.QuestionnaireID.Set(qn.QuestionnaireId),
		e.ResubmitAllowed.Set(qn.ResubmitAllowed),
		e.ExpirationDate.Set(database.TimestamptzFromPb(qn.ExpirationDate)),
		e.QuestionnaireTemplateID.Set(nil),
	)
	if err != nil {
		return nil, err
	}

	if qn.QuestionnaireTemplateId != "" {
		_ = e.QuestionnaireTemplateID.Set(qn.QuestionnaireTemplateId)
	}

	return e, nil
}

func PbToQuestionnaireQuestionEnts(qn *cpb.Questionnaire) (entities.QuestionnaireQuestions, error) {
	questionsReq := qn.Questions
	questions := make(entities.QuestionnaireQuestions, 0)

	for _, question := range questionsReq {
		e := &entities.QuestionnaireQuestion{}
		database.AllNullEntity(e)
		err := multierr.Combine(
			e.QuestionnaireQuestionID.Set(question.QuestionnaireQuestionId),
			e.QuestionnaireID.Set(qn.QuestionnaireId),
			e.Choices.Set(question.Choices),
			e.OrderIndex.Set(question.OrderIndex),
			e.Type.Set(question.Type.String()),
			e.IsRequired.Set(question.Required),
			e.Title.Set(question.Title),
		)
		if err != nil {
			return nil, err
		}

		questions = append(questions, e)
	}

	return questions, nil
}

func AudienceToUserNotificationEnt(notificationID string, audience *entities.Audience) (*entities.UserInfoNotification, error) {
	un := &entities.UserInfoNotification{}
	database.AllNullEntity(un)
	err := multierr.Combine(
		un.NotificationID.Set(notificationID),
		un.UserID.Set(audience.UserID),
		un.Status.Set(cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String()),
		un.QuestionnaireStatus.Set(cpb.UserNotificationQuestionnaireStatus_USER_NOTIFICATION_QUESTIONNAIRE_STATUS_UNANSWERED),
		un.IsIndividual.Set(audience.IsIndividual.Bool),
		un.StudentID.Set(audience.StudentID),
		un.ParentID.Set(audience.ParentID),
	)
	if err != nil {
		return nil, fmt.Errorf("set user infor notification %v", err)
	}

	if audience.CurrentGrade.Status == pgtype.Present {
		err = un.CurrentGrade.Set(audience.CurrentGrade)
		if err != nil {
			return nil, fmt.Errorf("failed set CurrentGrade: %v", err)
		}
	}

	if audience.UserGroup.Status == pgtype.Present {
		err = un.UserGroup.Set(audience.UserGroup)
		if err != nil {
			return nil, fmt.Errorf("failed set UserGroup: %v", err)
		}
	}

	if audience.GradeID.Status == pgtype.Present {
		err = un.GradeID.Set(audience.GradeID)
		if err != nil {
			return nil, fmt.Errorf("failed set GradeID: %v", err)
		}
	}

	return un, nil
}

func EventStudentPackageV2PbToNotificationStudentCourseEnt(studentPackage *natspb.EventStudentPackageV2) (*entities.NotificationStudentCourse, error) {
	notiStudentCourse := &entities.NotificationStudentCourse{}
	database.AllNullEntity(notiStudentCourse)

	id := idutil.ULIDNow()
	err := multierr.Combine(
		notiStudentCourse.StudentCourseID.Set(id),
		notiStudentCourse.StudentID.Set(studentPackage.StudentPackage.StudentId),
		notiStudentCourse.CourseID.Set(studentPackage.StudentPackage.Package.CourseId),
		notiStudentCourse.LocationID.Set(studentPackage.StudentPackage.Package.LocationId),
		notiStudentCourse.StartAt.Set(studentPackage.StudentPackage.Package.StartDate.AsTime()),
		notiStudentCourse.EndAt.Set(studentPackage.StudentPackage.Package.EndDate.AsTime()),
	)

	if err != nil {
		return nil, err
	}

	return notiStudentCourse, nil
}

func EventStudentPackageJPRPEPbToNotificationStudentCourseEnts(studentPackage *natspb.EventSyncStudentPackage_StudentPackage) ([]*entities.NotificationStudentCourse, error) {
	notiStudentCourses := []*entities.NotificationStudentCourse{}
	studentID := studentPackage.StudentId
	for _, pkg := range studentPackage.Packages {
		for _, course := range pkg.CourseIds {
			notiStudentCourse := &entities.NotificationStudentCourse{}
			database.AllNullEntity(notiStudentCourse)

			id := idutil.ULIDNow()
			err := multierr.Combine(
				notiStudentCourse.StudentCourseID.Set(id),
				notiStudentCourse.StudentID.Set(studentID),
				notiStudentCourse.CourseID.Set(course),
				notiStudentCourse.StartAt.Set(pkg.StartDate.AsTime().UTC()),
				notiStudentCourse.EndAt.Set(pkg.EndDate.AsTime().UTC()),
				notiStudentCourse.LocationID.Set(constants.JPREPOrgLocation),
			)

			if err != nil {
				return nil, err
			}

			notiStudentCourses = append(notiStudentCourses, notiStudentCourse)
		}
	}

	return notiStudentCourses, nil
}

func EventStudentPackageV2PbToNotificationClassMemberEnt(studentPackage *natspb.EventStudentPackageV2) (*entities.NotificationClassMember, error) {
	notiClassMember := &entities.NotificationClassMember{}
	database.AllNullEntity(notiClassMember)
	err := multierr.Combine(
		notiClassMember.StudentID.Set(studentPackage.StudentPackage.StudentId),
		notiClassMember.ClassID.Set(studentPackage.StudentPackage.Package.ClassId),
		notiClassMember.StartAt.Set(studentPackage.StudentPackage.Package.StartDate.AsTime()),
		notiClassMember.EndAt.Set(studentPackage.StudentPackage.Package.EndDate.AsTime()),
		notiClassMember.LocationID.Set(studentPackage.StudentPackage.Package.LocationId),
		notiClassMember.CourseID.Set(studentPackage.StudentPackage.Package.CourseId),
	)

	if err != nil {
		return nil, err
	}

	return notiClassMember, nil
}

func EventLeaveClassRoomToNotificationClassMemberEnts(e *pb.EvtClassRoom_LeaveClass, courseID, locationID string, deletedAt interface{}) ([]*entities.NotificationClassMember, error) {
	notiClassMembers := []*entities.NotificationClassMember{}
	for _, userID := range e.GetUserIds() {
		notiClassMember := &entities.NotificationClassMember{}
		database.AllNullEntity(notiClassMember)
		strClassID := strconv.Itoa(int(e.GetClassId()))
		err := multierr.Combine(
			notiClassMember.StudentID.Set(userID),
			notiClassMember.ClassID.Set(strClassID),
			notiClassMember.CourseID.Set(courseID),
			notiClassMember.LocationID.Set(locationID),
			notiClassMember.DeletedAt.Set(deletedAt),
		)
		if err != nil {
			return nil, err
		}

		notiClassMembers = append(notiClassMembers, notiClassMember)
	}

	return notiClassMembers, nil
}

func EventJoinClassRoomToNotificationClassMemberEnt(e *pb.EvtClassRoom_JoinClass, courseID, locationID string) (*entities.NotificationClassMember, error) {
	notiClassMember := &entities.NotificationClassMember{}
	database.AllNullEntity(notiClassMember)
	strClassID := strconv.Itoa(int(e.GetClassId()))
	err := multierr.Combine(
		notiClassMember.StudentID.Set(e.GetUserId()),
		notiClassMember.ClassID.Set(strClassID),
		notiClassMember.CourseID.Set(courseID),
		notiClassMember.LocationID.Set(locationID),
		notiClassMember.StartAt.Set(nil),
		notiClassMember.EndAt.Set(nil),
		notiClassMember.DeletedAt.Set(nil),
	)
	if err != nil {
		return nil, err
	}
	return notiClassMember, nil
}

func PbToQuestionnaireTemplateEnt(qt *npb.QuestionnaireTemplate) (*entities.QuestionnaireTemplate, error) {
	e := &entities.QuestionnaireTemplate{}
	database.AllNullEntity(e)
	err := multierr.Combine(
		e.QuestionnaireTemplateID.Set(qt.QuestionnaireTemplateId),
		e.Name.Set(qt.Name),
		e.ResubmitAllowed.Set(qt.ResubmitAllowed),
		e.ExpirationDate.Set(database.TimestamptzFromPb(qt.ExpirationDate)),
		e.Type.Set(npb.QuestionnaireTemplateType_QUESTION_TEMPLATE_TYPE_DEFAULT.String()),
	)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func PbToQuestionnaireTemplateQuestionEnts(qt *npb.QuestionnaireTemplate) (entities.QuestionnaireTemplateQuestions, error) {
	questionsReq := qt.Questions
	questions := make(entities.QuestionnaireTemplateQuestions, 0)

	for _, question := range questionsReq {
		e := &entities.QuestionnaireTemplateQuestion{}
		database.AllNullEntity(e)
		err := multierr.Combine(
			e.QuestionnaireTemplateQuestionID.Set(question.QuestionnaireTemplateQuestionId),
			e.QuestionnaireTemplateID.Set(qt.QuestionnaireTemplateId),
			e.Choices.Set(question.Choices),
			e.OrderIndex.Set(question.OrderIndex),
			e.Type.Set(question.Type.String()),
			e.IsRequired.Set(question.Required),
			e.Title.Set(question.Title),
		)
		if err != nil {
			return nil, err
		}

		questions = append(questions, e)
	}

	return questions, nil
}
