package communication

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	"github.com/manabie-com/backend/features/eibanam/communication/helper"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/cucumber/godog"
)

type AnotherSchoolAdminEditScheduledNotificationSuite struct {
	util *helper.CommunicationHelper
}

func NewAnotherSchoolAdminEditScheduledNotificationSuite(util *helper.CommunicationHelper) *AnotherSchoolAdminEditScheduledNotificationSuite {
	return &AnotherSchoolAdminEditScheduledNotificationSuite{
		util: util,
	}
}

func (s *AnotherSchoolAdminEditScheduledNotificationSuite) InitScenario(ctx *godog.ScenarioContext) {
	stepsMapping := map[string]interface{}{
		`^"([^"]*)" logins CMS$`:                                 s.loginCms,
		`^"([^"]*)" has created student S1 with parent P1 info$`: s.hasCreatedAStudentWithParentInfo,
		`^"([^"]*)" has created a scheduled notification$`:       s.hasCreateNotification,

		`^"([^"]*)" has opened editor full-screen dialog of scheduled notification$`: s.hasOpenedEditorFullscreenDialogOfScheduledNotification,
		`^"([^"]*)" edits "([^"]*)" of scheduled notification$`:                      s.editsOfScheduledNotification,
		`^"([^"]*)" clicks "([^"]*)" button$`:                                        s.clicksButton,
		`^"([^"]*)" sees updated scheduled notification on CMS$`:                     s.seesUpdatedScheduledNotificationOnCMS,
		`^"([^"]*)" sees name of composer updated to "([^"]*)"$`:                     s.seesNameOfComposerUpdatedTo,
	}
	for pattern, function := range stepsMapping {
		ctx.Step(pattern, function)
	}
}

func (s *AnotherSchoolAdminEditScheduledNotificationSuite) loginCms(ctx context.Context, accountName string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	switch accountName {
	case "school admin 1":
		sysAdmin, school, err := s.util.CreateSchoolAdminAndLoginToCMS(ctx, helper.AccountTypeSchoolAdmin)
		if err != nil {
			return ctx, err
		}
		state.SystemAdmin = sysAdmin
		state.School = school
	case "school admin 2":
		// create admin 2
		schoolAdmin2, err := s.util.CreateSchoolAdmin(state.SystemAdmin, int64(state.School.ID))
		if err != nil {
			return ctx, err
		}

		// login to cms and exchange the token for using later
		if err = s.util.SchoolAdminLoginToCms(ctx, schoolAdmin2); err != nil {
			return ctx, fmt.Errorf("SchoolAdminLoginToCms.Error %v", err)
		}

		// map data to  suit state
		state.School.Admins = append(state.School.Admins, schoolAdmin2)
	}
	return util.StateToContext(ctx, state), nil
}

func (s *AnotherSchoolAdminEditScheduledNotificationSuite) hasCreatedAStudentWithParentInfo(ctx context.Context, adminName string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	if state.School == nil || len(state.School.Admins) == 0 {
		return ctx, errors.New("missing create school and admin step")
	}
	switch adminName {
	case "school admin 1":
		// create student
		newStudent, err := s.util.CreateStudent(state.School.Admins[0], 4, []string{state.School.DefaultLocation}, true, 1)
		if err != nil {
			return ctx, fmt.Errorf("CreateStudent.Error %v", err)
		}

		courses, err := s.util.CreateCourses(state.School.Admins[0], state.School.ID, newStudent.Grade.ID, 1)
		if err != nil {
			return ctx, fmt.Errorf("CreateCourses.Error %v", err)
		}

		if err = s.util.AddCourseToStudent(state.School.Admins[0], newStudent, courses); err != nil {
			return ctx, fmt.Errorf("AddCourseToStudent.Error %v", err)
		}

		state.Students = append(state.Students, newStudent)
	default:
		return ctx, fmt.Errorf("%v don't have to create new student", adminName)

	}

	return util.StateToContext(ctx, state), nil
}

func (s *AnotherSchoolAdminEditScheduledNotificationSuite) hasCreateNotification(ctx context.Context, adminName string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	if state.School == nil || len(state.School.Admins) == 0 {
		return ctx, errors.New("no school admin found")
	}
	switch adminName {
	case "school admin 1":
		notify := &entity.Notification{
			SchoolID:      state.School.ID,
			Title:         "scheduled notification",
			Content:       "hello world",
			HTMLContent:   "<b> hello world </b>",
			ReceiverGroup: []cpb.UserGroup{cpb.UserGroup_USER_GROUP_PARENT, cpb.UserGroup_USER_GROUP_STUDENT},
			Status:        cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED,
			Data:          nil,
			ScheduledAt:   time.Now().Add(1 * time.Hour),
			FilterByGrade: entity.GradeFilter{
				Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			},
			FilterByCourse: entity.CourseFilter{
				Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			},
			IndividualReceivers: []string{state.Students[0].ID},
		}

		if err := s.util.CreateNotification(state.School.Admins[0], notify); err != nil {
			return ctx, err
		}
		state.Notify = notify

	default:
		return ctx, fmt.Errorf("%v has not to create notification", adminName)
	}
	return util.StateToContext(ctx, state), nil
}

func (s *AnotherSchoolAdminEditScheduledNotificationSuite) hasOpenedEditorFullscreenDialogOfScheduledNotification(ctx context.Context, adminName string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	if state.School == nil || len(state.School.Admins) < 2 || state.School.Admins[0].Token == "" {
		return ctx, errors.New("admin 2 still not login")
	}
	if state.Notify == nil {
		return ctx, errors.New("no scheduled notification created")
	}
	return ctx, nil
}

func (s *AnotherSchoolAdminEditScheduledNotificationSuite) editsOfScheduledNotification(ctx context.Context, adminName, field string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	switch adminName {
	case "school admin 2":
		switch field {
		case "Title":
			state.Notify.Title += " updated"
		case "Content":
			state.Notify.Content += " updated"
		case "Date", "Time":
			state.Notify.ScheduledAt = state.Notify.ScheduledAt.Add(24 * time.Hour)
		case "Course":
			state.Notify.FilterByCourse.Courses = []string{state.Students[0].Courses[0].ID}
		case "Grade":
			state.Notify.FilterByGrade.Grades = []int32{state.Students[0].Grade.ID}
		case "Recipient email":
			state.Notify.IndividualReceivers = []string{state.Students[0].ID}
		case "Type filter":
			state.Notify.ReceiverGroup = []cpb.UserGroup{cpb.UserGroup_USER_GROUP_STUDENT}
		case "All fields":
			state.Notify.Title += " updated"
			state.Notify.Content += " updated"
			state.Notify.ScheduledAt = state.Notify.ScheduledAt.Add(24 * time.Hour)
			state.Notify.FilterByCourse.Courses = []string{state.Students[0].Courses[0].ID}
			state.Notify.FilterByGrade.Grades = []int32{state.Students[0].Grade.ID}
			state.Notify.IndividualReceivers = []string{state.Students[0].ID}
			state.Notify.ReceiverGroup = []cpb.UserGroup{cpb.UserGroup_USER_GROUP_STUDENT}
		default:
			return ctx, errors.New("unsupported field")
		}
	default:
		return ctx, nil
	}
	return util.StateToContext(ctx, state), nil
}

func (s *AnotherSchoolAdminEditScheduledNotificationSuite) clicksButton(ctx context.Context, adminName, button string) (context.Context, error) {
	state := util.StateFromContext(ctx)

	var admin *entity.Admin

	switch adminName {
	case "school admin 1":
		admin = state.School.Admins[0]
	case "school admin 2":
		admin = state.School.Admins[1]
	default:
		return ctx, fmt.Errorf("invalid admin name %v", adminName)
	}

	if err := s.util.UpdateNotification(admin, state.Notify); err != nil {
		return ctx, err
	}

	return util.StateToContext(ctx, state), nil
}

func (s *AnotherSchoolAdminEditScheduledNotificationSuite) seesUpdatedScheduledNotificationOnCMS(ctx context.Context, adminName string) (context.Context, error) {
	state := util.StateFromContext(ctx)

	var admin *entity.Admin
	switch adminName {
	case "school admin 1":
		admin = state.School.Admins[0]
	case "school admin 2":
		admin = state.School.Admins[1]
	default:
		return ctx, fmt.Errorf("invalid admin name %v", adminName)
	}

	cmsNotify, err := s.util.GetCmsNotification(admin, state.Notify.ID)
	if err != nil {
		return ctx, err
	}

	cmsNotifyMsg, err := s.util.GetCmsNotificationMsg(admin, cmsNotify.InfoNotifications[0].NotificationMsgID)
	if err != nil {
		return ctx, err
	}

	if !util.DeepEqualInt32(state.Notify.FilterByGrade.Grades, cmsNotify.InfoNotifications[0].TargetGroups.GradeFilter.Grades) {
		return ctx, fmt.Errorf("grades ids was not updated [notificationID %v]", state.Notify.ID)
	}

	if !util.DeepEqualString(state.Notify.FilterByCourse.Courses, cmsNotify.InfoNotifications[0].TargetGroups.CourseFilter.CourseIDs) {
		return ctx, fmt.Errorf("course ids was not updated [notificationID %v]", state.Notify.ID)
	}

	if !util.DeepEqualString(state.Notify.IndividualReceivers, cmsNotify.InfoNotifications[0].ReceiverIDs) {
		return ctx, fmt.Errorf("reciverId ids was not updated [notificationID %v]", state.Notify.ID)
	}

	if state.Notify.Title != cmsNotifyMsg.InfoNotificationMsgs[0].Title {
		return ctx, fmt.Errorf("title ids was not updated [notificationID %v]", state.Notify.ID)
	}

	if state.Notify.Content != cmsNotifyMsg.InfoNotificationMsgs[0].Content.Raw {
		return ctx, fmt.Errorf("content ids was not updated [notificationID %v]", state.Notify.ID)
	}

	if state.Notify.ScheduledAt.Format("2006-01-02T15:04:05Z07:00") != cmsNotify.InfoNotifications[0].ScheduledAt.Format("2006-01-02T15:04:05Z07:00") {
		return ctx, fmt.Errorf("time ids was not updated [notificationID %v]", state.Notify.ID)
	}

	if len(state.Notify.ReceiverGroup) != len(cmsNotify.InfoNotifications[0].TargetGroups.UserGroupFilter.UserGroup) {
		return ctx, fmt.Errorf("user group was not updated [notificationID %v]", state.Notify.ID)
	}

	return util.StateToContext(ctx, state), nil
}

func (s *AnotherSchoolAdminEditScheduledNotificationSuite) seesNameOfComposerUpdatedTo(ctx context.Context, adminName, composerName string) (context.Context, error) {
	state := util.StateFromContext(ctx)

	var admin *entity.Admin
	switch adminName {
	case "school admin 1":
		admin = state.School.Admins[0]
	case "school admin 2":
		admin = state.School.Admins[1]
	default:
		return ctx, fmt.Errorf("invalid admin name %v", adminName)
	}

	cmsNotify, err := s.util.GetCmsNotification(admin, state.Notify.ID)
	if err != nil {
		return ctx, err
	}

	switch composerName {
	case "school admin 1":
		if cmsNotify.InfoNotifications[0].EditorID != state.School.Admins[0].ID {
			return ctx, fmt.Errorf("scheduled notification composer is not %v", composerName)
		}
	case "school admin 2":
		if cmsNotify.InfoNotifications[0].EditorID != state.School.Admins[1].ID {
			return ctx, fmt.Errorf("scheduled notification composer is not %v", composerName)
		}
	}
	return util.StateToContext(ctx, state), nil
}
