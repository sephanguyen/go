package communication

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	"github.com/manabie-com/backend/features/eibanam/communication/helper"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/cucumber/godog"
	"github.com/pkg/errors"
)

type EditScheduledNotificationSuite struct {
	util *helper.CommunicationHelper
}

func NewEditScheduledNotificationSuite(util *helper.CommunicationHelper) *EditScheduledNotificationSuite {
	return &EditScheduledNotificationSuite{
		util: util,
	}
}

func (s *EditScheduledNotificationSuite) InitScenario(ctx *godog.ScenarioContext) {
	stepsMapping := map[string]interface{}{
		`^"([^"]*)" logins CMS$`: s.createSchoolAdminAndLoginToCms,
		`^school admin has created a student with grade, course and parent info$`: s.schoolAdminHasCreatedAStudentWithGradeCourseAndParentInfo,
		`^"([^"]*)" logins Learner App$`:                                          s.loginsLearnerApp,
		`^school admin has created a scheduled notification$`:                     s.CreateNotification,
		`^school admin has opened a scheduled notification dialog$`:               s.schoolAdminHasOpenedAScheduledNotificationDialog,
		`^school admin edits "([^"]*)" of scheduled notification$`:                s.schoolAdminEditsOfScheduledNotification,
		`^school admin clicks "([^"]*)" button$`:                                  s.schoolAdminClicksButton,
		`^school admin sees updated scheduled notification on CMS$`:               s.schoolAdminSeesUpdatedScheduledNotificationOnCMS,
	}
	for pattern, function := range stepsMapping {
		ctx.Step(pattern, function)
	}
}

func (s *EditScheduledNotificationSuite) createSchoolAdminAndLoginToCms(ctx context.Context, accountType string) (context.Context, error) {
	sysAdmin, school, err := s.util.CreateSchoolAdminAndLoginToCMS(ctx, accountType)
	if err != nil {
		return ctx, err
	}
	state := util.StateFromContext(ctx)
	state.SystemAdmin = sysAdmin
	state.School = school
	return util.StateToContext(ctx, state), nil
}

func (s *EditScheduledNotificationSuite) schoolAdminHasCreatedAStudentWithGradeCourseAndParentInfo(ctx context.Context) (context.Context, error) {
	state := util.StateFromContext(ctx)
	if state.School == nil || len(state.School.Admins) == 0 {
		return ctx, errors.New("missing create school and admin step")
	}

	// create student
	newStudent1, err := s.util.CreateStudent(state.School.Admins[0], 4, []string{state.School.DefaultLocation}, true, 1)
	if err != nil {
		return ctx, err
	}

	// create student
	newStudent2, err := s.util.CreateStudent(state.School.Admins[0], 5, []string{state.School.DefaultLocation}, true, 1)
	if err != nil {
		return ctx, err
	}
	state.Students = []*entity.Student{newStudent1, newStudent2}

	courses1, err := s.util.CreateCourses(state.School.Admins[0], state.School.ID, newStudent1.Grade.ID, 1)
	if err != nil {
		return ctx, err
	}

	if err = s.util.AddCourseToStudent(state.School.Admins[0], newStudent1, courses1); err != nil {
		return ctx, err
	}

	courses2, err := s.util.CreateCourses(state.School.Admins[0], state.School.ID, newStudent2.Grade.ID, 1)
	if err != nil {
		return ctx, err
	}

	if err = s.util.AddCourseToStudent(state.School.Admins[0], newStudent2, courses2); err != nil {
		return ctx, err
	}

	return util.StateToContext(ctx, state), nil
}

func (s *EditScheduledNotificationSuite) loginsLearnerApp(ctx context.Context, role string) (context.Context, error) {
	state := util.StateFromContext(ctx)

	token, err := s.util.LoginLeanerApp(state.Students[0].Email, state.Students[0].Password)
	if err != nil {
		return ctx, err
	}
	state.Students[0].Token = token

	token, err = s.util.LoginLeanerApp(state.Students[1].Email, state.Students[1].Password)
	if err != nil {
		return ctx, err
	}
	state.Students[1].Token = token

	return util.StateToContext(ctx, state), nil
}

func (s *EditScheduledNotificationSuite) CreateNotification(ctx context.Context) (context.Context, error) {
	state := util.StateFromContext(ctx)
	if state.School == nil || len(state.School.Admins) == 0 {
		return ctx, errors.New("no school admin found")
	}

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
			Type:   cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			Grades: []int32{state.Students[0].Grade.ID},
		},
		FilterByCourse: entity.CourseFilter{
			Type:    cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			Courses: []string{state.Students[0].Courses[0].ID},
		},
		IndividualReceivers: []string{state.Students[0].ID},
	}

	if err := s.util.CreateNotification(state.School.Admins[0], notify); err != nil {
		return ctx, err
	}
	state.Notify = notify
	return util.StateToContext(ctx, state), nil
}

func (s *EditScheduledNotificationSuite) schoolAdminHasOpenedAScheduledNotificationDialog(ctx context.Context) (context.Context, error) {
	state := util.StateFromContext(ctx)
	if state.Notify == nil || state.Notify.ID == "" {
		return ctx, errors.New("no notification created")
	}
	return ctx, nil
}

func (s *EditScheduledNotificationSuite) schoolAdminEditsOfScheduledNotification(ctx context.Context, field string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	switch field {
	case "Title":
		state.Notify.Title += " updated"
	case "Content":
		state.Notify.Content += " updated"
	case "Date", "Time":
		state.Notify.ScheduledAt = state.Notify.ScheduledAt.Add(24 * time.Hour)
	case "Course":
		state.Notify.FilterByCourse.Courses = []string{state.Students[1].Courses[0].ID}
	case "Grade":
		state.Notify.FilterByGrade.Grades = []int32{state.Students[1].Grade.ID}
	case "Individual Recipient":
		state.Notify.IndividualReceivers = []string{state.Students[1].ID}
	case "All fields":
		state.Notify.Title += " updated"
		state.Notify.Content += " updated"
		state.Notify.ScheduledAt = state.Notify.ScheduledAt.Add(24 * time.Hour)
		state.Notify.FilterByCourse.Courses = []string{state.Students[1].Courses[0].ID}
		state.Notify.FilterByGrade.Grades = []int32{state.Students[1].Grade.ID}
		state.Notify.IndividualReceivers = []string{state.Students[1].ID}
		state.Notify.ReceiverGroup = []cpb.UserGroup{cpb.UserGroup_USER_GROUP_STUDENT}
	default:
		return ctx, errors.New("unsupported field")
	}
	return util.StateToContext(ctx, state), nil
}

func (s *EditScheduledNotificationSuite) schoolAdminClicksButton(ctx context.Context, buttonName string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	if err := s.util.UpdateNotification(state.School.Admins[0], state.Notify); err != nil {
		return ctx, err
	}
	return util.StateToContext(ctx, state), nil
}

func (s *EditScheduledNotificationSuite) schoolAdminSeesUpdatedScheduledNotificationOnCMS(ctx context.Context) (context.Context, error) {
	state := util.StateFromContext(ctx)
	cmsNotify, err := s.util.GetCmsNotification(state.School.Admins[0], state.Notify.ID)
	if err != nil {
		return ctx, err
	}

	cmsNotifyMsg, err := s.util.GetCmsNotificationMsg(state.School.Admins[0], cmsNotify.InfoNotifications[0].NotificationMsgID)
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
