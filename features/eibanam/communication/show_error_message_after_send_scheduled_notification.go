package communication

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	"github.com/manabie-com/backend/features/eibanam/communication/helper"
	"github.com/manabie-com/backend/features/eibanam/communication/util"

	"github.com/cucumber/godog"
)

type EditScheduledNotificationAfterSentSuite struct {
	util *helper.CommunicationHelper
}

func NewEditScheduledNotificationAfterSentSuite(util *helper.CommunicationHelper) *EditScheduledNotificationAfterSentSuite {
	return &EditScheduledNotificationAfterSentSuite{
		util: util,
	}
}

func (s *EditScheduledNotificationAfterSentSuite) InitScenario(ctx *godog.ScenarioContext) {
	stepsMapping := map[string]interface{}{
		`^"([^"]*)" logins CMS$`:                                                                    s.loginsCMS,
		`^"([^"]*)" logins Learner App$`:                                                            s.loginsLearnerApp,
		`^scheduled notification sent successfully on CMS$`:                                         s.scheduledNotificationSentSuccessfullyOnCMS,
		`^school admin clicks "([^"]*)" in notification dialog$`:                                    s.schoolAdminClicksInNotificationDialog,
		`^school admin has created a scheduled notification which will be sent (\d+) minute later$`: s.schoolAdminHasCreatedAScheduledNotificationWhichWillBeSentMinuteLater,
		`^school admin has created a student with grade, course and parent info$`:                   s.schoolAdminHasCreatedAStudentWithGradeCourseAndParentInfo,
		`^school admin has edited scheduled notification$`:                                          s.schoolAdminHasEditedScheduledNotification,
		`^school admin has opened a scheduled notification dialog$`:                                 s.schoolAdminHasOpenedAScheduledNotificationDialog,
		`^school admin is at "([^"]*)" page on CMS$`:                                                s.schoolAdminIsAtPageOnCMS,
		`^school admin sees message "([^"]*)"$`:                                                     s.schoolAdminSeesMessage,
		`^school admin selects "([^"]*)"$`:                                                          s.schoolAdminSelects,
	}
	for pattern, function := range stepsMapping {
		ctx.Step(pattern, function)
	}
}
func (s *EditScheduledNotificationAfterSentSuite) loginsCMS(ctx context.Context, accountType string) (context.Context, error) {
	sysAdmin, school, err := s.util.CreateSchoolAdminAndLoginToCMS(ctx, accountType)
	if err != nil {
		return ctx, err
	}
	state := util.StateFromContext(ctx)
	state.SystemAdmin = sysAdmin
	state.School = school
	return util.StateToContext(ctx, state), nil
}

func (s *EditScheduledNotificationAfterSentSuite) loginsLearnerApp(ctx context.Context, student string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	token, err := s.util.LoginLeanerApp(state.Students[0].Email, state.Students[0].Password)
	if err != nil {
		return ctx, err
	}
	state.Students[0].Token = token

	return util.StateToContext(ctx, state), nil
}

func (s *EditScheduledNotificationAfterSentSuite) scheduledNotificationSentSuccessfullyOnCMS(ctx context.Context) (context.Context, error) {
	state := util.StateFromContext(ctx)

	if state.Notify.ScheduledAt.After(time.Now()) {
		waitTime := state.Notify.ScheduledAt.Unix() - time.Now().Unix()
		time.Sleep(time.Duration(waitTime+5) * time.Second)
	}

	cmsNotify, err := s.util.GetCmsNotification(state.School.Admins[0], state.Notify.ID)
	if err != nil {
		return ctx, err
	}

	if cmsNotify.InfoNotifications[0].Status != "NOTIFICATION_STATUS_SENT" {
		return ctx, errors.New("scheduled notification still not sent")
	}

	return util.StateToContext(ctx, state), nil
}

func (s *EditScheduledNotificationAfterSentSuite) schoolAdminClicksInNotificationDialog(ctx context.Context, button string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	admin := state.School.Admins[0]
	switch button {
	case "Save schedule", "Close schedule":
		state.NotifyErr.UpsertError = s.util.UpdateNotification(admin, state.Notify)
	case "Discard and confirm":
		state.NotifyErr.DiscardError = s.util.DiscardNotification(admin, state.Notify)
	case "Send":
		state.NotifyErr.SentError = s.util.SendNotification(state.School.Admins[0], state.Notify)
	}
	return util.StateToContext(ctx, state), nil
}

func (s *EditScheduledNotificationAfterSentSuite) schoolAdminHasCreatedAScheduledNotificationWhichWillBeSentMinuteLater(ctx context.Context, arg1 int) (context.Context, error) {
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
		ScheduledAt:   time.Now().Add(2 * time.Minute),
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

func (s *EditScheduledNotificationAfterSentSuite) schoolAdminHasCreatedAStudentWithGradeCourseAndParentInfo(ctx context.Context) (context.Context, error) {
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

func (s *EditScheduledNotificationAfterSentSuite) schoolAdminHasEditedScheduledNotification(ctx context.Context) (context.Context, error) {
	state := util.StateFromContext(ctx)
	state.Notify.Title += " updated"
	return util.StateToContext(ctx, state), nil
}

func (s *EditScheduledNotificationAfterSentSuite) schoolAdminHasOpenedAScheduledNotificationDialog(ctx context.Context) (context.Context, error) {
	state := util.StateFromContext(ctx)
	if state.School == nil || len(state.School.Admins) == 0 || state.School.Admins[0].Token == "" {
		return ctx, errors.New("admin still not login")
	}
	if state.Notify == nil {
		return ctx, errors.New("no scheduled notification created")
	}
	return ctx, nil
}

func (s *EditScheduledNotificationAfterSentSuite) schoolAdminIsAtPageOnCMS(ctx context.Context, page string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	switch page {
	case "Notification":
		if state.School == nil || len(state.School.Admins) == 0 || state.School.Admins[0].Token == "" {
			return ctx, errors.New("school admin is not logged in")
		}
	}
	return util.StateToContext(ctx, state), nil
}

func (s *EditScheduledNotificationAfterSentSuite) schoolAdminSeesMessage(ctx context.Context, message string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	serverMessage := ""
	switch message {
	case "The notification has been sent, you can no longer edit this notification":
		if state.NotifyErr.UpsertError == nil {
			return ctx, errors.New("no error message about upsert found")
		}
		serverMessage = state.NotifyErr.UpsertError.Error()
	case "The notification has been sent, you can no longer discard this notification":
		if state.NotifyErr.DiscardError == nil {
			return ctx, errors.New("no error message about discard found")
		}
		serverMessage = state.NotifyErr.DiscardError.Error()
	case "The notification has been sent":
		if state.NotifyErr.SentError == nil {
			return ctx, errors.New("no error message about sent found")
		}
		serverMessage = state.NotifyErr.SentError.Error()
	}

	if !strings.Contains(strings.ToLower(serverMessage), strings.ToLower(message)) {
		return ctx, fmt.Errorf("want err [%v] but get [%v] ", message, serverMessage)
	}

	return util.StateToContext(ctx, state), nil
}

func (s *EditScheduledNotificationAfterSentSuite) schoolAdminSelects(ctx context.Context, notifyType string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	switch notifyType {
	case "Now":
		state.Notify.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT
	}
	return util.StateToContext(ctx, state), nil
}
