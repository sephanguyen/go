package communication

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/utils/strings/slices"
)

type CreateAndUpdateNotificationSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitCreateAndUpdateNotification(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &CreateAndUpdateNotificationSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^school admin creates "([^"]*)" courses with "([^"]*)" classes for each course$`:                                           s.CreatesNumberOfCoursesWithClass,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a current organization$`:                       s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfCurrentOrg,
		`^current staff upsert notification to "([^"]*)" and "([^"]*)" course and "([^"]*)" grade and "([^"]*)" location and "([^"]*)" class and "([^"]*)" school and "([^"]*)" individuals and "([^"]*)" scheduled time with "([^"]*)" and important is "([^"]*)"$`: s.CurrentStaffUpsertNotificationWithFilter,
		`^returns "([^"]*)" status code$`:                                             s.CheckReturnStatusCode,
		`^notificationmgmt services must store the notification with correctly info$`: s.NotificationMgmtMustStoreTheNotification,
		`^current staff upsert notification with valid filter$`:                       s.upsertNotificationWithValidFilter,
		`^current staff update notification with change "([^"]*)"$`:                   s.currentStaffUpdateNotificationWithChange,
		`^update correctly corresponding field$`:                                      s.updateCorrectlyCorrespondingField,
		`^individual name is saved successfully$`:                                     s.individualNameIsSavedSuccessfully,
		`^current staff upsert notification again with new individual targets$`:       s.currentStaffUpsertNotificationAgainWithNewIndividualTargets,
		`^current staff update notification filter with change selection all$`:        s.currentStaffUpdateNotificationFilterWithAllSelection,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *CreateAndUpdateNotificationSuite) upsertNotificationWithValidFilter(ctx context.Context) (context.Context, error) {
	// TODO: Update "none" for location/class filter to "random"
	ctx, err := s.CurrentStaffUpsertNotificationWithFilter(ctx, "student, parent", "random", "random", "none", "none", "none", "random", "random", "NOTIFICATION_STATUS_DRAFT", "false")
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

// nolint:gosec
func (s *CreateAndUpdateNotificationSuite) currentStaffUpdateNotificationWithChange(ctx context.Context, field string) (context.Context, error) {
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
			notification.TargetGroup.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{
				Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
			}
		} else {
			notification.TargetGroup.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{
				Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
				CourseIds: courseIDs,
			}
		}
	case "grade_filter":
		gradeIDs := make([]string, 0)
		for i, grade := range commonState.GradeAssigneds {
			if i == 0 || rand.Intn(2) == 1 {
				// make sure we have at least 1 target grade
				gradeIDs = append(gradeIDs, grade.ID)
			}
		}

		notification.TargetGroup.GradeFilter = &cpb.NotificationTargetGroup_GradeFilter{
			Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			GradeIds: gradeIDs,
		}
	case "class_filter":
		classIDs := make([]string, 0)
		for _, class := range commonState.Classes {
			if rand.Intn(2) == 1 {
				// make sure we have at least 1 target course
				classIDs = append(classIDs, class.ID)
			}
		}
		if len(classIDs) == 0 {
			notification.TargetGroup.ClassFilter = &cpb.NotificationTargetGroup_ClassFilter{
				Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
			}
		} else {
			notification.TargetGroup.ClassFilter = &cpb.NotificationTargetGroup_ClassFilter{
				Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
				ClassIds: classIDs,
			}
		}
	case "location_filter":
		locationIDs := make([]string, 0)
		for _, location := range commonState.Organization.DescendantLocations {
			if rand.Intn(2) == 1 {
				// make sure we have at least 1 target course
				locationIDs = append(locationIDs, location.ID)
			}
		}
		if len(locationIDs) == 0 {
			notification.TargetGroup.LocationFilter = &cpb.NotificationTargetGroup_LocationFilter{
				Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL,
			}
		} else {
			notification.TargetGroup.LocationFilter = &cpb.NotificationTargetGroup_LocationFilter{
				Type:        cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
				LocationIds: locationIDs,
			}
		}
	case "school_filter":
		schoolIDs := make([]string, 0)
		for _, school := range commonState.Schools {
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
	case "excluded_generic_receiver_ids":
		res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveGroupAudience(
			ctx,
			&npb.RetrieveGroupAudienceRequest{
				Keyword:     "",
				TargetGroup: commonState.Notification.TargetGroup,
			},
		)
		if err != nil {
			return common.StepStateToContext(ctx, commonState), fmt.Errorf("failed RetrieveGroupAudience: %v", err)
		}
		if len(res.Audiences) == 0 || res.TotalItems == 0 {
			return common.StepStateToContext(ctx, commonState), fmt.Errorf("expected RetrieveGroupAudience to not empty result")
		}
		// get second element
		notification.ExcludedGenericReceiverIds = []string{res.Audiences[1].UserId}
	}

	commonState.Request = &npb.UpsertNotificationRequest{
		Notification: notification,
	}
	commonState.Response, commonState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), commonState.Request.(*npb.UpsertNotificationRequest))

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *CreateAndUpdateNotificationSuite) updateCorrectlyCorrespondingField(ctx context.Context) (context.Context, error) {
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

func (s *CreateAndUpdateNotificationSuite) individualNameIsSavedSuccessfully(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	notificationUserRepo := repositories.UserRepo{}
	receivers, _, err := notificationUserRepo.FindUser(ctx, s.BobDBConn, &repositories.FindUserFilter{UserIDs: database.TextArray(commonState.Notification.GenericReceiverIds)})
	if err != nil {
		return nil, fmt.Errorf("svc.NotificationUserRepo.Get: %v", err)
	}

	expectedReceiverNames := make([]string, 0)
	for _, receiver := range receivers {
		expectedReceiverNames = append(expectedReceiverNames, receiver.Name.String)
	}

	queryGetReceiverNames := `SELECT receiver_names FROM info_notifications WHERE notification_id = $1 AND deleted_at IS NULL;`

	var actualReceiverNames pgtype.TextArray
	err = database.Select(ctx, s.BobDBConn, queryGetReceiverNames, commonState.Notification.NotificationId).ScanFields(&actualReceiverNames)
	if err != nil {
		return ctx, err
	}

	if !stringutil.SliceElementsMatch(expectedReceiverNames, database.FromTextArray(actualReceiverNames)) {
		return ctx, fmt.Errorf("expect receiver names %+v, got %+v", expectedReceiverNames, database.FromTextArray(actualReceiverNames))
	}

	return ctx, nil
}

func (s *CreateAndUpdateNotificationSuite) currentStaffUpsertNotificationAgainWithNewIndividualTargets(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	newStudentIDs := []string{}
	for _, student := range commonState.Students {
		if !slices.Contains(commonState.Notification.ReceiverIds, student.ID) {
			newStudentIDs = append(newStudentIDs, student.ID)
		}
	}
	commonState.Notification.ReceiverIds = newStudentIDs
	commonState.Notification.GenericReceiverIds, _ = s.GenerateIndividualIDs(ctx, true, 2)
	commonState.Response, commonState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).
		UpsertNotification(ctx, &npb.UpsertNotificationRequest{
			Notification: commonState.Notification,
		})
	if commonState.ResponseErr == nil {
		resp := commonState.Response.(*npb.UpsertNotificationResponse)
		commonState.Notification.NotificationId = resp.NotificationId
	}
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *CreateAndUpdateNotificationSuite) currentStaffUpdateNotificationFilterWithAllSelection(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	notification := commonState.Notification
	infoNotification := &entities.InfoNotification{}
	fields := database.GetFieldNames(infoNotification)
	queryGetNotication := fmt.Sprintf(`SELECT %s FROM %s WHERE notification_id = $1;`, strings.Join(fields, ","), infoNotification.TableName())

	err := database.Select(ctx, s.BobDBConn, queryGetNotication, commonState.Notification.NotificationId).ScanOne(infoNotification)
	if err != nil {
		return common.StepStateToContext(ctx, commonState), err
	}

	notification.TargetGroup.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{
		Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL,
	}
	notification.TargetGroup.LocationFilter = &cpb.NotificationTargetGroup_LocationFilter{
		Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL,
	}
	notification.TargetGroup.ClassFilter = &cpb.NotificationTargetGroup_ClassFilter{
		Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL,
	}

	commonState.Request = &npb.UpsertNotificationRequest{
		Notification: notification,
	}
	commonState.Response, commonState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), commonState.Request.(*npb.UpsertNotificationRequest))

	return common.StepStateToContext(ctx, commonState), nil
}
