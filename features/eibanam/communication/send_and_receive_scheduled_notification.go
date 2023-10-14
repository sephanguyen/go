package communication

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	"github.com/manabie-com/backend/features/eibanam/communication/helper"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

type SendAndReceiveScheduledNotificationSuite struct {
	util *helper.CommunicationHelper
}

func NewSendAndReceiveScheduledNotificationSuite(util *helper.CommunicationHelper) *SendAndReceiveScheduledNotificationSuite {
	return &SendAndReceiveScheduledNotificationSuite{
		util: util,
	}
}

func (s *SendAndReceiveScheduledNotificationSuite) InitScenario(ctx *godog.ScenarioContext) {
	stepsMapping := map[string]interface{}{
		`^"([^"]*)" logins CMS$`: s.createSchoolAdminAndLoginToCms,
		`^school admin has created a student with grade, course and parent info$`: s.schoolAdminHasCreatedAStudentWithGradeCourseAndParentInfo,
		`^"([^"]*)" logins Learner App$`:                                          s.loginsLearnerApp,
		`^"([^"]*)" of "([^"]*)" logins Learner App$`:                             s.ofLoginsLearnerApp,
		`^school admin has created scheduled notification$`:                       s.CreateNotification,
		`^school admin is at "([^"]*)" page on CMS$`:                              s.schoolAdminIsAtPageOnCMS,

		`^school admin waits for scheduled notification to be sent on time$`: s.schoolAdminWaitsForScheduledNotificationToBeSentOnTime,
		`^scheduled notification is sent successfully on CMS$`:               s.scheduledNotificationIsSentSuccessfullyOnCMS,
		`^"([^"]*)" receives the scheduled notification in their device$`:    s.receivesTheScheduledNotificationInTheirDevice,

		`^school admin has edited sending time of scheduled notification$`: s.schoolAdminHasEditedSendingTimeOfScheduledNotification,
	}
	for pattern, function := range stepsMapping {
		ctx.Step(pattern, function)
	}
}

func (s *SendAndReceiveScheduledNotificationSuite) createSchoolAdminAndLoginToCms(ctx context.Context, accountType string) (context.Context, error) {
	sysAdmin, school, err := s.util.CreateSchoolAdminAndLoginToCMS(ctx, accountType)
	if err != nil {
		return ctx, err
	}
	state := util.StateFromContext(ctx)
	state.SystemAdmin = sysAdmin
	state.School = school
	return util.StateToContext(ctx, state), nil
}

func (s *SendAndReceiveScheduledNotificationSuite) schoolAdminHasCreatedAStudentWithGradeCourseAndParentInfo(ctx context.Context) (context.Context, error) {
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

func (s *SendAndReceiveScheduledNotificationSuite) loginsLearnerApp(ctx context.Context, role string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	token, err := s.util.LoginLeanerApp(state.Students[0].Email, state.Students[0].Password)
	if err != nil {
		return ctx, err
	}
	state.Students[0].Token = token

	return util.StateToContext(ctx, state), nil
}

func (s *SendAndReceiveScheduledNotificationSuite) ofLoginsLearnerApp(ctx context.Context, parent, student string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	token, err := s.util.LoginLeanerApp(state.Students[0].Parents[0].Email, state.Students[0].Parents[0].Password)
	if err != nil {
		return ctx, err
	}
	state.Students[0].Parents[0].Token = token
	return util.StateToContext(ctx, state), nil
}

func (s *SendAndReceiveScheduledNotificationSuite) CreateNotification(ctx context.Context) (context.Context, error) {
	state := util.StateFromContext(ctx)
	if state.School == nil || len(state.School.Admins) == 0 {
		return ctx, errors.New("no school admin found")
	}

	notify := &entity.Notification{
		SchoolID:            state.School.ID,
		Title:               "scheduled notification",
		Content:             "hello world",
		HTMLContent:         "<b> hello world </b>",
		ReceiverGroup:       []cpb.UserGroup{cpb.UserGroup_USER_GROUP_PARENT, cpb.UserGroup_USER_GROUP_STUDENT},
		Status:              cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED,
		Data:                nil,
		ScheduledAt:         time.Now().Add(1 * time.Minute),
		FilterByGrade:       entity.GradeFilter{},
		FilterByCourse:      entity.CourseFilter{},
		IndividualReceivers: []string{state.Students[0].ID},
	}

	if err := s.util.CreateNotification(state.School.Admins[0], notify); err != nil {
		return ctx, err
	}
	state.Notify = notify
	return util.StateToContext(ctx, state), nil
}

func (s *SendAndReceiveScheduledNotificationSuite) schoolAdminIsAtPageOnCMS(ctx context.Context, page string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	switch page {
	case "Notification":
		if state.School == nil || len(state.School.Admins) == 0 || state.School.Admins[0].Token == "" {
			return ctx, errors.New("school admin is not logged in")
		}
	}
	return util.StateToContext(ctx, state), nil
}

func (s *SendAndReceiveScheduledNotificationSuite) schoolAdminWaitsForScheduledNotificationToBeSentOnTime(ctx context.Context) (context.Context, error) {
	state := util.StateFromContext(ctx)

	if state.Notify.ScheduledAt.After(time.Now()) {
		waitTime := state.Notify.ScheduledAt.Unix() - time.Now().Unix()
		time.Sleep(time.Duration(waitTime+5) * time.Second)
	}

	return util.StateToContext(ctx, state), nil
}

func (s *SendAndReceiveScheduledNotificationSuite) scheduledNotificationIsSentSuccessfullyOnCMS(ctx context.Context) (context.Context, error) {
	state := util.StateFromContext(ctx)
	cmsNotify, err := s.util.GetCmsNotification(state.School.Admins[0], state.Notify.ID)
	if err != nil {
		return ctx, err
	}

	if cmsNotify.InfoNotifications[0].Status != "NOTIFICATION_STATUS_SENT" {
		return ctx, errors.New("scheduled notification still not sent")
	}
	return util.StateToContext(ctx, state), nil
}

func (s *SendAndReceiveScheduledNotificationSuite) receivesTheScheduledNotificationInTheirDevice(ctx context.Context, role string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	token := ""
	switch role {
	case "student":
		token = state.Students[0].Token
	case "parent P1":
		token = state.Students[0].Parents[0].Token
	}
	notifications, err := s.util.GetUserNotification(token, false)
	if err != nil {
		return ctx, err
	}
	for _, notify := range notifications {
		if notify.UserNotification.NotificationId == state.Notify.ID {
			return util.StateToContext(ctx, state), nil
		}
	}

	return ctx, fmt.Errorf("notification was not send to %v", role)
}

func (s *SendAndReceiveScheduledNotificationSuite) schoolAdminHasEditedSendingTimeOfScheduledNotification(ctx context.Context) (context.Context, error) {
	state := util.StateFromContext(ctx)

	state.Notify.ScheduledAt = state.Notify.ScheduledAt.Add(1 * time.Minute)
	err := s.util.UpdateNotification(state.School.Admins[0], state.Notify)
	if err != nil {
		return ctx, err
	}

	return util.StateToContext(ctx, state), nil
}
