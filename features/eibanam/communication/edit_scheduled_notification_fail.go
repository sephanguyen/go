package communication

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	"github.com/manabie-com/backend/features/eibanam/communication/helper"
	"github.com/manabie-com/backend/features/eibanam/communication/util"

	"github.com/cucumber/godog"
)

type UpdateScheduledNotificationFailedSuite struct {
	util *helper.CommunicationHelper
}

func NewUpdateScheduledNotificationFailedSuite(util *helper.CommunicationHelper) *UpdateScheduledNotificationFailedSuite {
	return &UpdateScheduledNotificationFailedSuite{
		util: util,
	}
}

func (s *UpdateScheduledNotificationFailedSuite) InitScenario(ctx *godog.ScenarioContext) {
	stepsMapping := map[string]interface{}{
		`^"([^"]*)" logins CMS$`:         s.loginsCMS,
		`^"([^"]*)" logins Learner App$`: s.loginsLearnerApp,
		`^school admin clear value of "([^"]*)" field and "([^"]*)" notification$`:                 s.schoolAdminClearValueOfFieldAndNotification,
		`^school admin has created a scheduled notification$`:                                      s.schoolAdminHasCreatedAScheduledNotification,
		`^school admin has created a student with grade, course and parent info$`:                  s.schoolAdminHasCreatedAStudentWithGradeCourseAndParentInfo,
		`^school admin has opened a scheduled notification dialog$`:                                s.schoolAdminHasOpenedAScheduledNotificationDialog,
		`^school admin is at "([^"]*)" page on CMS$`:                                               s.schoolAdminIsAtPageOnCMS,
		`^school admin sees new notification full-screen dialog closed$`:                           s.schoolAdminSeesNewNotificationFullscreenDialogClosed,
		`^school admin sees "([^"]*)" of required errors validation message in form of "([^"]*)"$`: s.schoolAdminSeesOfRequiredErrorsValidationMessageInFormOf,
		`^school admin sees scheduled notification is not updated$`:                                s.schoolAdminSeesScheduledNotificationIsNotUpdated,
	}
	for pattern, function := range stepsMapping {
		ctx.Step(pattern, function)
	}
}

func (s *UpdateScheduledNotificationFailedSuite) loginsCMS(ctx context.Context, roleName string) (context.Context, error) {
	sysAdmin, school, err := s.util.CreateSchoolAdminAndLoginToCMS(ctx, roleName)
	if err != nil {
		return ctx, err
	}
	state := util.StateFromContext(ctx)
	state.SystemAdmin = sysAdmin
	state.School = school
	return util.StateToContext(ctx, state), nil
}

func (s *UpdateScheduledNotificationFailedSuite) loginsLearnerApp(ctx context.Context, student string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	token, err := s.util.LoginLeanerApp(state.Students[0].Email, state.Students[0].Password)
	if err != nil {
		return ctx, err
	}
	state.Students[0].Token = token

	return util.StateToContext(ctx, state), nil
}

func (s *UpdateScheduledNotificationFailedSuite) schoolAdminClearValueOfFieldAndNotification(ctx context.Context, field, button string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	switch field {
	case "Title":
		state.Notify.Title = ""
	case "Content":
		state.Notify.Content = ""
		state.Notify.HTMLContent = ""
	case "Time":
		state.Notify.ScheduledAt = time.Time{}
	case "Recipient Group":
		state.Notify.IndividualReceivers = []string{}
		state.Notify.FilterByGrade.Grades = []int32{}
		state.Notify.FilterByCourse.Courses = []string{}
	default:
		return ctx, errors.New("unsupported field")
	}
	return util.StateToContext(ctx, state), nil
}

func (s *UpdateScheduledNotificationFailedSuite) schoolAdminHasCreatedAScheduledNotification(ctx context.Context) (context.Context, error) {
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
		ScheduledAt:   time.Now().Add(5 * time.Minute),
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
	return util.StateToContext(ctx, state), nil
}

func (s *UpdateScheduledNotificationFailedSuite) schoolAdminHasCreatedAStudentWithGradeCourseAndParentInfo(ctx context.Context) (context.Context, error) {
	state := util.StateFromContext(ctx)
	if state.School == nil || len(state.School.Admins) == 0 {
		return ctx, errors.New("missing create school and admin step")
	}

	// create student
	newStudent1, err := s.util.CreateStudent(state.School.Admins[0], 4, []string{state.School.DefaultLocation}, true, 1)
	if err != nil {
		return ctx, err
	}
	state.Students = []*entity.Student{newStudent1}

	courses1, err := s.util.CreateCourses(state.School.Admins[0], state.School.ID, newStudent1.Grade.ID, 1)
	if err != nil {
		return ctx, err
	}
	if err = s.util.AddCourseToStudent(state.School.Admins[0], newStudent1, courses1); err != nil {
		return ctx, err
	}

	return util.StateToContext(ctx, state), nil
}

func (s *UpdateScheduledNotificationFailedSuite) schoolAdminHasOpenedAScheduledNotificationDialog(ctx context.Context) (context.Context, error) {
	state := util.StateFromContext(ctx)
	if state.School == nil || len(state.School.Admins) == 0 || state.School.Admins[0].Token == "" {
		return ctx, errors.New("admin still not login")
	}
	if state.Notify == nil {
		return ctx, errors.New("no scheduled notification created")
	}
	return ctx, nil
}

func (s *UpdateScheduledNotificationFailedSuite) schoolAdminIsAtPageOnCMS(ctx context.Context, page string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	switch page {
	case "Notification":
		if state.School == nil || len(state.School.Admins) == 0 || state.School.Admins[0].Token == "" {
			return ctx, errors.New("school admin is not logged in")
		}
	}
	return util.StateToContext(ctx, state), nil
}

func (s *UpdateScheduledNotificationFailedSuite) schoolAdminSeesNewNotificationFullscreenDialogClosed(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (s *UpdateScheduledNotificationFailedSuite) schoolAdminSeesOfRequiredErrorsValidationMessageInFormOf(ctx context.Context, numOfErr, field string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	if err := s.util.UpdateNotification(state.School.Admins[0], state.Notify); err == nil {
		data, _ := json.Marshal(state.Notify)
		return ctx, fmt.Errorf("update empty field %v with data %v want %v err but no one found", field, string(data), numOfErr)
	}
	return util.StateToContext(ctx, state), nil
}

func (s *UpdateScheduledNotificationFailedSuite) schoolAdminSeesScheduledNotificationIsNotUpdated(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
