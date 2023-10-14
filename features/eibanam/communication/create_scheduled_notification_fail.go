package communication

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	"github.com/manabie-com/backend/features/eibanam/communication/helper"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/cucumber/godog"
)

type CreateScheduledNotificationFailedSuite struct {
	util *helper.CommunicationHelper
}

func NewCreateScheduledNotificationFailedSuite(util *helper.CommunicationHelper) *CreateScheduledNotificationFailedSuite {
	return &CreateScheduledNotificationFailedSuite{
		util: util,
	}
}

func (s *CreateScheduledNotificationFailedSuite) InitScenario(ctx *godog.ScenarioContext) {
	stepsMapping := map[string]interface{}{
		`^school admin logins CMS$`: s.schoolAdminLoginCms,
		`^school admin has created a student with grade, course and parent info$`:                   s.schoolAdminHasCreatedAStudentWithGradeCourseAndParentInfo,
		`^school admin has opened compose new notification full-screen dialog$`:                     s.schoolAdminHasOpenedComposeNewNotificationFullscreenDialog,
		`^school admin is at "([^"]*)" page on CMS$`:                                                s.schoolAdminIsAtPageOnCMS,
		`^school admin leaves "([^"]*)" blank and click "([^"]*)" button$`:                          s.schoolAdminLeavesBlankAndClickButton,
		`^school admin sees new notification full-screen dialog closed$`:                            s.schoolAdminSeesNewNotificationFullscreenDialogClosed,
		`^school admin sees new notification full-screen dialog still opened$`:                      s.schoolAdminSeesNewNotificationFullscreenDialogStillOpened,
		`^school admin sees scheduled notification is not created with "([^"]*)" error validation$`: s.schoolAdminSeesScheduledNotificationIsNotCreatedWithErrorValidation,
		`^school admin selects "([^"]*)"$`:                                                          s.schoolAdminSelects,
	}
	for pattern, function := range stepsMapping {
		ctx.Step(pattern, function)
	}
}

func (s *CreateScheduledNotificationFailedSuite) schoolAdminLoginCms(ctx context.Context) (context.Context, error) {
	sysAdmin, school, err := s.util.CreateSchoolAdminAndLoginToCMS(ctx, schoolAdmin)
	if err != nil {
		return ctx, err
	}
	state := util.StateFromContext(ctx)
	state.SystemAdmin = sysAdmin
	state.School = school
	return util.StateToContext(ctx, state), nil
}

func (s *CreateScheduledNotificationFailedSuite) schoolAdminHasCreatedAStudentWithGradeCourseAndParentInfo(ctx context.Context) (context.Context, error) {

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

	// create course
	courses1, err := s.util.CreateCourses(state.School.Admins[0], state.School.ID, newStudent1.Grade.ID, 1)
	if err != nil {
		return ctx, err
	}
	if err = s.util.AddCourseToStudent(state.School.Admins[0], newStudent1, courses1); err != nil {
		return ctx, err
	}

	return util.StateToContext(ctx, state), nil
}

func (s *CreateScheduledNotificationFailedSuite) schoolAdminHasOpenedComposeNewNotificationFullscreenDialog(ctx context.Context) (context.Context, error) {
	state := util.StateFromContext(ctx)
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
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
		},
		FilterByCourse: entity.CourseFilter{
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
		},
		IndividualReceivers: []string{state.Students[0].ID},
	}
	state.Notify = notify
	return util.StateToContext(ctx, state), nil
}

func (s *CreateScheduledNotificationFailedSuite) schoolAdminIsAtPageOnCMS(ctx context.Context, page string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	if state.School == nil || len(state.School.Admins) == 0 || state.School.Admins[0].Token == "" {
		return ctx, errors.New("admin is not at notification page")
	}
	return util.StateToContext(ctx, state), nil
}

func (s *CreateScheduledNotificationFailedSuite) schoolAdminLeavesBlankAndClickButton(ctx context.Context, field, button string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	notify := state.Notify
	switch field {
	case "Title":
		notify.Title = ""
	case "Content":
		notify.Content = ""
		notify.HTMLContent = ""
	case "Time":
		notify.ScheduledAt = time.Time{}
	case "Recipient Group":
		notify.IndividualReceivers = []string{}
		notify.FilterByGrade = entity.GradeFilter{
			Grades: []int32{},
			Type:   cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
		}
		notify.FilterByCourse = entity.CourseFilter{
			Courses: []string{},
			Type:    cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
		}
	default:
		return ctx, errors.New("field is not defined")
	}
	return util.StateToContext(ctx, state), nil
}

func (s *CreateScheduledNotificationFailedSuite) schoolAdminSeesNewNotificationFullscreenDialogClosed(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (s *CreateScheduledNotificationFailedSuite) schoolAdminSeesNewNotificationFullscreenDialogStillOpened(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (s *CreateScheduledNotificationFailedSuite) schoolAdminSeesScheduledNotificationIsNotCreatedWithErrorValidation(ctx context.Context, numOfErr string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	notify := state.Notify

	err := s.util.CreateNotification(state.School.Admins[0], notify)
	if err == nil {
		notifyString, _ := json.Marshal(notify)
		fmt.Println(string(notifyString))
		return ctx, errors.New(fmt.Sprintf("expected %v error but get no one", numOfErr))
	}
	return util.StateToContext(ctx, state), nil
}

func (s *CreateScheduledNotificationFailedSuite) schoolAdminSelects(ctx context.Context, notifyType string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	notify := state.Notify
	switch notifyType {
	case "Schedule":
		notify.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED
	default:
		return ctx, errors.New("type is not defined")
	}
	return util.StateToContext(ctx, state), nil
}
