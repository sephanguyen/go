package communication

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CreateAndUpdateScheduledNotificationSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitCreateAndUpdateScheduledNotification(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &CreateAndUpdateScheduledNotificationSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^school admin creates "([^"]*)" courses$`:                                                                                  s.CreatesNumberOfCourses,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a current organization$`:                       s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfCurrentOrg,
		`^current staff upsert notification to "([^"]*)" and "([^"]*)" course and "([^"]*)" grade and "([^"]*)" location and "([^"]*)" class and "([^"]*)" school and "([^"]*)" individuals and "([^"]*)" scheduled time with "([^"]*)" and important is "([^"]*)"$`: s.CurrentStaffUpsertNotificationWithFilter,
		`^returns "([^"]*)" status code$`:                                                               s.CheckReturnStatusCode,
		`^notificationmgmt services must store the notification with correctly info$`:                   s.NotificationMgmtMustStoreTheNotification,
		`^current staff send notification$`:                                                             s.CurrentStaffSendNotification,
		`^notificationmgmt services must send notification to user$`:                                    s.NotificationMgmtMustSendNotificationToUser,
		`^current staff discards notification$`:                                                         s.CurrentStaffDiscardsNotification,
		`^notification is discarded$`:                                                                   s.NotificationIsDiscarded,
		`^returns error message "([^"]*)"$`:                                                             s.CheckReturnsErrorMessage,
		`^current staff upsert notification with valid filter for scheduled notification$`:              s.upsertNotificationWithValidFilterForScheduledNotification,
		`^current staff update notification with change "([^"]*)"$`:                                     s.currentStaffUpdateNotificationWithChange,
		`^update correctly corresponding field$`:                                                        s.updateCorrectlyCorrespondingField,
		`^current staff upsert notification with invalid field scheduled_at which before current time$`: s.upsertNotificationWithInvalidFieldScheduledAtWhichBeforeCurrentTime,
		`^current staff upsert notification with missing field scheduled_at$`:                           s.upsertNotificationWithMissingFieldScheduledAt,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *CreateAndUpdateScheduledNotificationSuite) updateCorrectlyCorrespondingField(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	resp, ok := commonState.Response.(*npb.UpsertNotificationResponse)
	if !ok {
		return common.StepStateToContext(ctx, commonState), fmt.Errorf("expect npb.UpsertNotificationResponse but got %v", commonState.Response)
	}

	if commonState.Notification.NotificationId != resp.NotificationId {
		return common.StepStateToContext(ctx, commonState), fmt.Errorf("expect upsert with the same notifition id %v but got %v", commonState.Notification.NotificationId, resp.NotificationId)
	}

	return s.NotificationMgmtMustStoreTheNotification(ctx)
}

func (s *CreateAndUpdateScheduledNotificationSuite) currentStaffUpdateNotificationWithChange(ctx context.Context, field string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	notification := commonState.Notification
	infoNotification := &entities.InfoNotification{}
	fields := database.GetFieldNames(infoNotification)
	queryGetNotication := fmt.Sprintf(`SELECT %s FROM %s WHERE notification_id = $1;`, strings.Join(fields, ","), infoNotification.TableName())

	err := database.Select(ctx, s.BobDBConn, queryGetNotication, commonState.Notification.NotificationId).ScanOne(infoNotification)
	if err != nil {
		return common.StepStateToContext(ctx, commonState), err
	}
	notification.Message.NotificationMsgId = infoNotification.NotificationMsgID.String
	switch field {
	case "content":
		notification.Message.Content.Raw = "modified raw: " + commonState.CurrentUserID
		notification.Message.Content.Rendered = "modified rendered: " + commonState.CurrentUserID
	case "title":
		notification.Message.Title = "update title of notification"
	case "user_groups":
		notification.TargetGroup.UserGroupFilter.UserGroups = make([]cpb.UserGroup, 0)
		if rand.Intn(2) > 0 {
			notification.TargetGroup.UserGroupFilter.UserGroups = append(notification.TargetGroup.UserGroupFilter.UserGroups, cpb.UserGroup_USER_GROUP_PARENT)
		}
		if len(notification.TargetGroup.UserGroupFilter.UserGroups) == 0 || rand.Intn(2) > 0 {
			notification.TargetGroup.UserGroupFilter.UserGroups = append(notification.TargetGroup.UserGroupFilter.UserGroups, cpb.UserGroup_USER_GROUP_STUDENT)
		}

	case "course_filter":
		courseIDs := make([]string, 0)
		for _, course := range commonState.Courses {
			if rand.Intn(2) == 1 {
				// make sure we have at least 1 target course
				courseIDs = append(courseIDs, course.ID)
			}
		}
		if len(courseIDs) == 0 {
			notification.TargetGroup.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{}
		} else {
			notification.TargetGroup.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{
				Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
				CourseIds: courseIDs,
			}
		}
	// nolint
	case "grade_filter":
		gradeIDs := make([]string, 0)
		for i, grade := range commonState.GradeAssigneds {
			if i == 0 && rand.Intn(2) == 1 {
				// make sure we have at least 1 target course
				gradeIDs = append(gradeIDs, grade.ID)
			}
		}

		notification.TargetGroup.GradeFilter = &cpb.NotificationTargetGroup_GradeFilter{
			Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			GradeIds: gradeIDs,
		}
	case "school_filter":
		schoolIDs := make([]string, 0)
		for _, school := range commonState.Schools {
			// nolint
			if rand.Intn(2) == 1 {
				schoolIDs = append(schoolIDs, school.ID)
			}
		}
		if len(schoolIDs) == 0 {
			notification.TargetGroup.SchoolFilter = &cpb.NotificationTargetGroup_SchoolFilter{
				Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL,
			}
		} else {
			notification.TargetGroup.SchoolFilter = &cpb.NotificationTargetGroup_SchoolFilter{
				Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
				SchoolIds: schoolIDs,
			}
		}
	case "individuals_filter":
		numStudentNew := common.RandRangeIn(2, 10)
		ctx, err := s.CreatesNumberOfStudentsWithParentsInfo(ctx, fmt.Sprint(numStudentNew), "1")
		if err != nil {
			return common.StepStateToContext(ctx, commonState), fmt.Errorf("CreatesNumberOfStudentsWithParentsInfo: %v", err)
		}
		commonState = common.StepStateFromContext(ctx)
		studentIDs := []string{}

		// Only add new students for individual list
		for idx, student := range commonState.Students {
			if idx >= len(commonState.Students)-numStudentNew {
				studentIDs = append(studentIDs, student.ID)
			}
		}
		notification.GenericReceiverIds = studentIDs

	case "status":
		notification.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED

	case "scheduled_time":
		notification.ScheduledAt = timestamppb.New(time.Now().Add(time.Duration(rand.Intn(5)+1) * time.Hour))
	case "is_important":
		notification.IsImportant = true
	}

	commonState.Request = &npb.UpsertNotificationRequest{
		Notification: notification,
	}
	commonState.Response, commonState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), commonState.Request.(*npb.UpsertNotificationRequest))

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *CreateAndUpdateScheduledNotificationSuite) upsertNotificationWithValidFilterForScheduledNotification(ctx context.Context) (context.Context, error) {
	// TODO: Update "none" for location/class filter to "random"
	ctx, err := s.CurrentStaffUpsertNotificationWithFilter(ctx, "student, parent", "random", "random", "default", "random", "random", "random", "random", "NOTIFICATION_STATUS_SCHEDULED", "false")
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *CreateAndUpdateScheduledNotificationSuite) upsertNotificationWithMissingFieldScheduledAt(ctx context.Context) (context.Context, error) {
	// TODO: Update "none" for location/class filter to "random"
	ctx, err := s.CurrentStaffUpsertNotificationWithFilter(ctx, "student, parent", "random", "random", "default", "random", "random", "random", "", "NOTIFICATION_STATUS_SCHEDULED", "false")
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *CreateAndUpdateScheduledNotificationSuite) upsertNotificationWithInvalidFieldScheduledAtWhichBeforeCurrentTime(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	opts := &common.NotificationWithOpts{
		UserGroups:       "student",
		CourseFilter:     "random",
		GradeFilter:      "random",
		LocationFilter:   "none",
		ClassFilter:      "none",
		IndividualFilter: "random",
		ScheduledStatus:  "random",
		Status:           "NOTIFICATION_STATUS_SCHEDULED",
		IsImportant:      false,
	}
	var err error
	ctx, commonState.Notification, err = s.GetNotificationWithOptions(ctx, opts)
	if err != nil {
		return common.StepStateToContext(ctx, commonState), err
	}

	// before current time
	commonState.Notification.ScheduledAt = timestamppb.New(time.Now().Add(-1 * time.Minute))

	commonState.Request = &npb.UpsertNotificationRequest{
		Notification: commonState.Notification,
	}
	commonState.Response, commonState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), commonState.Request.(*npb.UpsertNotificationRequest))
	if commonState.ResponseErr == nil {
		resp := commonState.Response.(*npb.UpsertNotificationResponse)
		commonState.Notification.NotificationId = resp.NotificationId
	}

	return common.StepStateToContext(ctx, commonState), nil
}
