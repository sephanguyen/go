package common

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	bdd_entities "github.com/manabie-com/backend/features/communication/common/entities"
	"github.com/manabie-com/backend/features/communication/common/helpers"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgtype"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *NotificationSuite) CurrentStaffUpsertNotificationWithFilter(ctx context.Context, userGroups, courseFilter, gradeFilter, locationFilter, classFilter, schoolFilter, individualFilter, scheduledStatus, status string, isImportantStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	isImportant := false
	if isImportantStr == "true" {
		isImportant = true
	}

	opts := &NotificationWithOpts{
		UserGroups:       userGroups,
		CourseFilter:     courseFilter,
		GradeFilter:      gradeFilter,
		LocationFilter:   locationFilter,
		ClassFilter:      classFilter,
		SchoolFilter:     schoolFilter,
		IndividualFilter: individualFilter,
		ScheduledStatus:  scheduledStatus,
		Status:           status,
		IsImportant:      isImportant,
	}
	var err error
	ctx, stepState.Notification, err = s.GetNotificationWithOptions(ctx, opts)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("GetNotificationWithOptions: %v", err)
	}

	stepState.Request = &npb.UpsertNotificationRequest{
		Notification: stepState.Notification,
	}

	if stepState.Questionnaire != nil {
		stepState.Request.(*npb.UpsertNotificationRequest).Questionnaire = stepState.Questionnaire
	}

	stepState.Response, stepState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, stepState.CurrentStaff.Token), stepState.Request.(*npb.UpsertNotificationRequest))

	if stepState.ResponseErr == nil {
		resp := stepState.Response.(*npb.UpsertNotificationResponse)
		stepState.Notification.NotificationId = resp.NotificationId
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *NotificationSuite) NotificationMgmtMustStoreTheNotification(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	infoNotification := &entities.InfoNotification{}
	fields := database.GetFieldNames(infoNotification)
	queryGetNotication := fmt.Sprintf(`SELECT %s FROM %s WHERE notification_id = $1 AND deleted_at IS NULL;`, strings.Join(fields, ","), infoNotification.TableName())

	err := database.Select(ctx, s.BobDBConn, queryGetNotication, stepState.Notification.NotificationId).ScanOne(infoNotification)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can't get notification: %v", err)
	}

	req, ok := stepState.Request.(*npb.UpsertNotificationRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect npb.UpsertNotificationRequest but got %v", stepState.Request)
	}

	ctx, err = s.CheckInfoNotificationResponse(ctx, req.Notification, infoNotification)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("CheckInfoNotificationResponse: %v", err)
	}

	infoNotificationMsg := &entities.InfoNotificationMsg{}
	fields = database.GetFieldNames(infoNotificationMsg)
	queryGetInfoNoticationMsg := fmt.Sprintf(`SELECT %s FROM %s WHERE notification_msg_id = $1 AND deleted_at IS NULL;`, strings.Join(fields, ","), infoNotificationMsg.TableName())

	err = database.Select(ctx, s.BobDBConn, queryGetInfoNoticationMsg, infoNotification.NotificationMsgID).ScanOne(infoNotificationMsg)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can't get notification message: %v", err)
	}

	msgEnt, err := mappers.PbToInfoNotificationMsgEnt(req.Notification.Message)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("mappers.PbToInfoNotificationMsgEnt: %v", err)
	}

	err = s.CheckInfoNotificationMsgResponse(msgEnt, infoNotificationMsg)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

// Use for cases that are not NATS notifications
// nolint
func (s *NotificationSuite) GetNotificationWithOptions(ctx context.Context, opts *NotificationWithOpts) (context.Context, *cpb.Notification, error) {
	stepState := StepStateFromContext(ctx)

	notification := makeSampleNotification(stepState.CurrentStaff.ID, stepState.CurrentStaff.ID, stepState.Organization.ID)

	notification.TargetGroup.UserGroupFilter = &cpb.NotificationTargetGroup_UserGroupFilter{}
	if opts.UserGroups != "none" {
		notification.TargetGroup.UserGroupFilter.UserGroups = make([]cpb.UserGroup, 0)
		usrGroups := strings.Split(opts.UserGroups, ",")
		for _, gr := range usrGroups {
			gr = strings.TrimSpace(gr)
			switch gr {
			case "student":
				notification.TargetGroup.UserGroupFilter.UserGroups = append(notification.TargetGroup.UserGroupFilter.UserGroups, cpb.UserGroup_USER_GROUP_STUDENT)
			case "parent":
				notification.TargetGroup.UserGroupFilter.UserGroups = append(notification.TargetGroup.UserGroupFilter.UserGroups, cpb.UserGroup_USER_GROUP_PARENT)
			}
		}
	}

	switch opts.CourseFilter {
	case "all":
		notification.TargetGroup.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL,
		}
	case "random":
		courseIDs := make([]string, 0)
		courses := make([]*cpb.NotificationTargetGroup_CourseFilter_Course, 0)
		for i, course := range stepState.Courses {
			if i == 0 || rand.Intn(2) == 1 {
				courseIDs = append(courseIDs, course.ID)
				courses = append(courses, &cpb.NotificationTargetGroup_CourseFilter_Course{
					CourseId:   course.ID,
					CourseName: course.Name,
				})
			}
		}
		notification.TargetGroup.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{
			Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			CourseIds: courseIDs,
			Courses:   courses, // simulate FE's request to always send this data
		}
	default:
		notification.TargetGroup.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
		}
	}

	switch opts.GradeFilter {
	case "all":
		notification.TargetGroup.GradeFilter = &cpb.NotificationTargetGroup_GradeFilter{
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL,
		}
	case "random":
		gradeIDs := make([]string, 0)
		for i, grade := range stepState.GradeAssigneds {
			if i == 0 || rand.Intn(2) == 1 {
				gradeIDs = append(gradeIDs, grade.ID)
			}
		}
		gradeIDs = golibs.GetUniqueElementStringArray(gradeIDs)
		if len(gradeIDs) == 0 {
			notification.TargetGroup.GradeFilter = &cpb.NotificationTargetGroup_GradeFilter{
				Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
			}
		} else {
			notification.TargetGroup.GradeFilter = &cpb.NotificationTargetGroup_GradeFilter{
				Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
				GradeIds: gradeIDs,
			}
		}
	default:
		notification.TargetGroup.GradeFilter = &cpb.NotificationTargetGroup_GradeFilter{
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
		}
	}

	//nolint
	switch opts.LocationFilter {
	case "random":
		locationIDs := make([]string, 0)
		locations := make([]*cpb.NotificationTargetGroup_LocationFilter_Location, 0)
		for i, location := range stepState.Organization.DescendantLocations {
			if i == 0 || rand.Intn(2) == 1 {
				locationIDs = append(locationIDs, location.ID)
				locations = append(locations, &cpb.NotificationTargetGroup_LocationFilter_Location{
					LocationId:   location.ID,
					LocationName: location.Name,
				})
			}
		}
		notification.TargetGroup.LocationFilter = &cpb.NotificationTargetGroup_LocationFilter{
			Type:        cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			LocationIds: locationIDs,
			Locations:   locations, // simulate FE's request to always send this data
		}
	case "none":
		notification.TargetGroup.LocationFilter = &cpb.NotificationTargetGroup_LocationFilter{
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
		}
	case "all":
		notification.TargetGroup.LocationFilter = &cpb.NotificationTargetGroup_LocationFilter{
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL,
		}
	default:
		locationIDs := make([]string, 0)
		locations := make([]*cpb.NotificationTargetGroup_LocationFilter_Location, 0)
		idxsLocsStr := strings.Split(opts.LocationFilter, ",")
		for _, idxLocStr := range idxsLocsStr {
			if idxLocStr == "default" {
				locationIDs = append(locationIDs, stepState.Organization.DefaultLocation.ID)
				locations = append(locations, &cpb.NotificationTargetGroup_LocationFilter_Location{
					LocationId:   stepState.Organization.DefaultLocation.ID,
					LocationName: stepState.Organization.DefaultLocation.Name,
				})
				continue
			}

			idxLoc, err := strconv.Atoi(idxLocStr)
			if err != nil {
				return ctx, nil, fmt.Errorf("can't convert descendant location index: %v", err)
			}
			if idxLoc <= 0 || idxLoc > helpers.NumberOfNewCenterLocationCreated {
				return ctx, nil, fmt.Errorf("index descendant location out of range")
			}
			locationIDs = append(locationIDs, stepState.Organization.DescendantLocations[idxLoc-1].ID)
			locations = append(locations, &cpb.NotificationTargetGroup_LocationFilter_Location{
				LocationId:   stepState.Organization.DescendantLocations[idxLoc-1].ID,
				LocationName: stepState.Organization.DescendantLocations[idxLoc-1].Name,
			})
		}
		notification.TargetGroup.LocationFilter = &cpb.NotificationTargetGroup_LocationFilter{
			Type:        cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			LocationIds: locationIDs,
			Locations:   locations, // simulate FE's request to always send this data
		}
	}

	//nolint
	switch opts.ClassFilter {
	case "random":
		classList := make([]*bdd_entities.Class, 0)
		classIDs := make([]string, 0)
		classes := make([]*cpb.NotificationTargetGroup_ClassFilter_Class, 0)
		for _, course := range stepState.Courses {
			if len(course.Classes) > 0 {
				for _, class := range course.Classes {
					classList = append(classList, class)
				}
			}
		}
		for i, class := range classList {
			if i == 0 || rand.Intn(2) == 1 {
				classIDs = append(classIDs, class.ID)
				classes = append(classes, &cpb.NotificationTargetGroup_ClassFilter_Class{
					ClassId:   class.ID,
					ClassName: class.Name,
				})
			}
		}
		notification.TargetGroup.ClassFilter = &cpb.NotificationTargetGroup_ClassFilter{
			Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			ClassIds: classIDs,
			Classes:  classes, // simulate FE's request to always send this data
		}
	case "all":
		notification.TargetGroup.ClassFilter = &cpb.NotificationTargetGroup_ClassFilter{
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL,
		}
	default:
		notification.TargetGroup.ClassFilter = &cpb.NotificationTargetGroup_ClassFilter{
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
		}
	}

	switch opts.SchoolFilter {
	case "random":
		schoolIDs := make([]string, 0)
		schools := make([]*cpb.NotificationTargetGroup_SchoolFilter_School, 0)
		for i, school := range stepState.CurrentSchools {
			if i == 0 || rand.Intn(2) == 1 {
				schoolIDs = append(schoolIDs, school.ID)
				schools = append(schools, &cpb.NotificationTargetGroup_SchoolFilter_School{
					SchoolId:   school.ID,
					SchoolName: school.Name,
				})
			}
		}
		if len(schools) == 0 {
			notification.TargetGroup.SchoolFilter = &cpb.NotificationTargetGroup_SchoolFilter{
				Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
			}
		} else {
			notification.TargetGroup.SchoolFilter = &cpb.NotificationTargetGroup_SchoolFilter{
				Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
				SchoolIds: schoolIDs,
				Schools:   schools, // simulate FE's request to always send this data
			}
		}
	case "all":
		notification.TargetGroup.SchoolFilter = &cpb.NotificationTargetGroup_SchoolFilter{
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL,
		}
	default:
		notification.TargetGroup.SchoolFilter = &cpb.NotificationTargetGroup_SchoolFilter{
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
		}
	}

	switch opts.IndividualFilter {
	case "receiver":
		numRand := RandRangeIn(1, 3)
		studentIDs, err := s.GenerateIndividualIDs(ctx, false, numRand)
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		notification.ReceiverIds = studentIDs
	default: // case generic
		if opts.IndividualFilter != "none" {
			if opts.IndividualFilter == "random" || opts.IndividualFilter == "all" {
				numRand := RandRangeIn(1, 3)
				genericUserIDs, err := s.GenerateIndividualIDs(ctx, true, numRand)
				if err != nil {
					return StepStateToContext(ctx, stepState), nil, err
				}
				notification.GenericReceiverIds = genericUserIDs
			} else {
				numStudentNew, err := strconv.Atoi(opts.IndividualFilter)
				if err != nil {
					return ctx, nil, err
				}
				genericUserIDs, err := s.GenerateIndividualIDs(ctx, true, numStudentNew)
				if err != nil {
					return StepStateToContext(ctx, stepState), nil, err
				}
				notification.GenericReceiverIds = genericUserIDs
			}
		}
	}

	validScheduledTimeMin := regexp.MustCompile(`(^[0-9]+)\s[min]+$`) // ex: "1 min", "2 min", etc...
	validScheduledTimeSec := regexp.MustCompile(`(^[0-9]+)\s[sec]+$`) // ex: "1 sec", "2 sec", etc...
	if opts.ScheduledStatus == "random" {
		next := time.Now().Add(time.Duration(rand.Intn(100)+1) * time.Hour)
		notification.ScheduledAt = timestamppb.New(next)
	} else if validScheduledTimeMin.MatchString(opts.ScheduledStatus) {
		matches := validScheduledTimeMin.FindStringSubmatch(opts.ScheduledStatus)
		if len(matches) == 0 {
			return StepStateToContext(ctx, stepState), nil, fmt.Errorf("cannot extract ScheduledStatus")
		}
		minuteValue, _ := strconv.Atoi(matches[1])
		next := time.Now().Add(time.Duration(time.Duration(minuteValue) * time.Minute))
		notification.ScheduledAt = timestamppb.New(next)
	} else if validScheduledTimeSec.MatchString(opts.ScheduledStatus) {
		matches := validScheduledTimeSec.FindStringSubmatch(opts.ScheduledStatus)
		if len(matches) == 0 {
			return StepStateToContext(ctx, stepState), nil, fmt.Errorf("cannot extract ScheduledStatus")
		}
		secValue, _ := strconv.Atoi(matches[1])
		next := time.Now().Add(time.Duration(time.Duration(secValue) * time.Second))
		notification.ScheduledAt = timestamppb.New(next)
	}

	notification.Status = cpb.NotificationStatus(cpb.NotificationStatus_value[opts.Status])

	notification.IsImportant = opts.IsImportant

	if opts.ExcludedGenericReceiverStr == "excluded" {
		res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveGroupAudience(
			ctx,
			&npb.RetrieveGroupAudienceRequest{
				Keyword:     "",
				TargetGroup: notification.GetTargetGroup(),
			},
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, fmt.Errorf("failed RetrieveGroupAudience: %v", err)
		}
		// get first audience to exclude
		notification.ExcludedGenericReceiverIds = []string{res.Audiences[0].UserId}
	}

	if len(opts.ReceiverIds) > 0 {
		notification.ReceiverIds = opts.ReceiverIds
	}

	if len(opts.GenericReceiverIds) > 0 {
		notification.GenericReceiverIds = opts.GenericReceiverIds
	}

	if len(opts.ExcludedGenericReceiverIds) > 0 {
		notification.ExcludedGenericReceiverIds = opts.ExcludedGenericReceiverIds
	}

	if len(opts.MediaIds) > 0 {
		notification.Message.MediaIds = opts.MediaIds
	}

	return StepStateToContext(ctx, stepState), notification, nil
}

func (s *NotificationSuite) CheckInfoNotificationResponse(ctx context.Context, expectNotification *cpb.Notification, infoNotification *entities.InfoNotification) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if string(infoNotification.Data.Bytes) != expectNotification.Data {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification data %v but got %v", expectNotification.Data, string(infoNotification.Data.Bytes))
	}

	if infoNotification.EditorID.String != expectNotification.EditorId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification editor id %v but got %v", expectNotification.EditorId, infoNotification.EditorID.String)
	}

	if infoNotification.CreatedUserID.String != expectNotification.CreatedUserId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification created userz id %v but got %v", expectNotification.CreatedUserId, infoNotification.CreatedUserID.String)
	}

	if infoNotification.Type.String != expectNotification.Type.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification type %v but got %v", expectNotification.Type, infoNotification.Type.String)
	}

	actualTargetGroup, err := infoNotification.GetTargetGroup()
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("infoNotification.GetTargetGroup: %v", err)
	}

	expectedTargetGroupEnt := mappers.PbToNotificationTargetEnt(expectNotification.TargetGroup)
	err = s.CheckTargetGroupFilter(expectedTargetGroupEnt, actualTargetGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("checkTargetGroupFilter: %v", err)
	}

	err = s.CheckNotificationFilterData(ctx, actualTargetGroup)
	if err != nil {
		return ctx, fmt.Errorf("CheckNotificationFilterData: %v", err)
	}

	if infoNotification.Event.String != expectNotification.Event.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification event %v but got %v", expectNotification.Event, infoNotification.Event.String)
	}

	if infoNotification.Status.String != expectNotification.Status.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification status %v but got %v", expectNotification.Status, infoNotification.Status.String)
	}

	if expectNotification.ScheduledAt.AsTime().Round(time.Second).Equal(infoNotification.ScheduledAt.Time) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification scheduled at %v but got %v", expectNotification.ScheduledAt.AsTime(), infoNotification.ScheduledAt.Time)
	}

	schoolAdminRepo := &repositories.SchoolAdminRepo{}
	schoolAdmin, err := schoolAdminRepo.Get(ctx, s.BobDBConn, database.Text(expectNotification.EditorId))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("schoolAdminRepo.Get: %v", err)
	}

	if infoNotification.Owner.Int != schoolAdmin.SchoolID.Int {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification school id %v but got %v", schoolAdmin.SchoolID, infoNotification.Owner.Int)
	}

	if expectNotification.IsImportant != infoNotification.IsImportant.Bool {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification is_important at %v but got %v", expectNotification.IsImportant, infoNotification.IsImportant.Bool)
	}

	if len(expectNotification.GenericReceiverIds) != len(infoNotification.GenericReceiverIDs.Elements) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification has %v generic_receiver_ids, got %v", len(expectNotification.GenericReceiverIds), len(infoNotification.GenericReceiverIDs.Elements))
	}

	ifnGenericReceiverIDs := []string{}
	err = infoNotification.GenericReceiverIDs.AssignTo(&ifnGenericReceiverIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed assign generic_receiver_ids: %v", err)
	}
	if !sliceutils.UnorderedEqual(expectNotification.GenericReceiverIds, ifnGenericReceiverIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected generic_receiver_ids array %v, got %v", expectNotification.GenericReceiverIds, ifnGenericReceiverIDs)
	}

	if len(expectNotification.ReceiverIds) != len(infoNotification.ReceiverIDs.Elements) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification has %v receiver_ids, got %v", len(expectNotification.ReceiverIds), len(infoNotification.ReceiverIDs.Elements))
	}

	ifnReceiverIDs := []string{}
	err = infoNotification.ReceiverIDs.AssignTo(&ifnReceiverIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed assign receiver_ids: %v", err)
	}
	if !sliceutils.UnorderedEqual(expectNotification.ReceiverIds, ifnReceiverIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected receiver_ids array %v, got %v", expectNotification.ReceiverIds, ifnReceiverIDs)
	}

	ifnExcludedGenericReceiverIDs := []string{}
	err = infoNotification.ExcludedGenericReceiverIDs.AssignTo(&ifnExcludedGenericReceiverIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed assign excluded_generic_receiver_ids: %v", err)
	}
	if len(expectNotification.ExcludedGenericReceiverIds) != len(infoNotification.ExcludedGenericReceiverIDs.Elements) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification has %v excluded_generic_receiver_ids, got %v", len(expectNotification.ExcludedGenericReceiverIds), len(infoNotification.ExcludedGenericReceiverIDs.Elements))
	}
	if !sliceutils.UnorderedEqual(expectNotification.ExcludedGenericReceiverIds, ifnExcludedGenericReceiverIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected excluded_generic_receiver_ids array %v, got %v", expectNotification.ExcludedGenericReceiverIds, ifnExcludedGenericReceiverIDs)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *NotificationSuite) CheckTargetGroupFilter(expectEnt *entities.InfoNotificationTarget, cur *entities.InfoNotificationTarget) error {
	if expectEnt.CourseFilter.Type != cur.CourseFilter.Type {
		return fmt.Errorf("expect course filter select type %v but got %v", expectEnt.CourseFilter.Type, cur.CourseFilter.Type)
	}
	if len(expectEnt.CourseFilter.CourseIDs) != len(cur.CourseFilter.CourseIDs) {
		return fmt.Errorf("expect number of target course_ids %v but got %v", len(expectEnt.CourseFilter.CourseIDs), len(cur.CourseFilter.CourseIDs))
	}
	if len(expectEnt.CourseFilter.Courses) != len(cur.CourseFilter.Courses) {
		return fmt.Errorf("expect number of target courses %v but got %v", len(expectEnt.CourseFilter.Courses), len(cur.CourseFilter.Courses))
	}
	if err := CheckTargetGroupFilterNameValues(&expectEnt.CourseFilter, &cur.CourseFilter); err != nil {
		return fmt.Errorf("CheckTargetGroupFilterNameValues Course: %v", err)
	}

	if expectEnt.GradeFilter.Type != cur.GradeFilter.Type {
		return fmt.Errorf("expect grade filter select type %v but got %v", expectEnt.GradeFilter.Type, cur.GradeFilter.Type)
	}
	if len(expectEnt.GradeFilter.GradeIDs) != len(cur.GradeFilter.GradeIDs) {
		return fmt.Errorf("expect number of target grade IDs %v but got %v", len(expectEnt.GradeFilter.GradeIDs), len(cur.GradeFilter.GradeIDs))
	}
	if !stringutil.SliceElementsMatch(expectEnt.GradeFilter.GradeIDs, cur.GradeFilter.GradeIDs) {
		return fmt.Errorf("expect elements of target grade IDs %+v, got %+v", expectEnt.GradeFilter.GradeIDs, cur.GradeFilter.GradeIDs)
	}

	if expectEnt.LocationFilter.Type != cur.LocationFilter.Type {
		return fmt.Errorf("expect location filter select type %v but got %v", expectEnt.LocationFilter.Type, cur.LocationFilter.Type)
	}
	if expectEnt.LocationFilter.Type == consts.TargetGroupSelectTypeList.String() && len(expectEnt.LocationFilter.LocationIDs) != len(cur.LocationFilter.LocationIDs) {
		return fmt.Errorf("expect number of target location_ids %v but got %v", len(expectEnt.LocationFilter.LocationIDs), len(cur.LocationFilter.LocationIDs))
	}
	if len(expectEnt.LocationFilter.Locations) != len(cur.LocationFilter.Locations) {
		return fmt.Errorf("expect number of target locations %v but got %v", len(expectEnt.LocationFilter.Locations), len(cur.LocationFilter.Locations))
	}
	if err := CheckTargetGroupFilterNameValues(&expectEnt.LocationFilter, &cur.LocationFilter); err != nil {
		return fmt.Errorf("CheckTargetGroupFilterNameValues Location: %v", err)
	}

	if expectEnt.ClassFilter.Type != cur.ClassFilter.Type {
		return fmt.Errorf("expect class filter select type %v but got %v", expectEnt.ClassFilter.Type, cur.ClassFilter.Type)
	}
	if len(expectEnt.ClassFilter.ClassIDs) != len(cur.ClassFilter.ClassIDs) {
		return fmt.Errorf("expect number of target class_ids %v but got %v", len(expectEnt.ClassFilter.ClassIDs), len(cur.ClassFilter.ClassIDs))
	}
	if len(expectEnt.ClassFilter.Classes) != len(cur.ClassFilter.Classes) {
		return fmt.Errorf("expect number of target classes %v but got %v", len(expectEnt.ClassFilter.Classes), len(cur.ClassFilter.Classes))
	}
	if err := CheckTargetGroupFilterNameValues(&expectEnt.ClassFilter, &cur.ClassFilter); err != nil {
		return fmt.Errorf("CheckTargetGroupFilterNameValues Class: %v", err)
	}

	if expectEnt.SchoolFilter.Type != cur.SchoolFilter.Type {
		return fmt.Errorf("expect school filter select type %v but got %v", expectEnt.SchoolFilter.Type, cur.SchoolFilter.Type)
	}
	if len(expectEnt.SchoolFilter.SchoolIDs) != len(cur.SchoolFilter.SchoolIDs) {
		return fmt.Errorf("expect number of target school_ids %v but got %v", len(expectEnt.SchoolFilter.SchoolIDs), len(cur.SchoolFilter.SchoolIDs))
	}
	if len(expectEnt.SchoolFilter.Schools) != len(cur.SchoolFilter.Schools) {
		return fmt.Errorf("expect number of target schools %v but got %v", len(expectEnt.SchoolFilter.Schools), len(cur.SchoolFilter.Schools))
	}
	if err := CheckTargetGroupFilterNameValues(&expectEnt.SchoolFilter, &cur.SchoolFilter); err != nil {
		return fmt.Errorf("CheckTargetGroupFilterNameValues School: %v", err)
	}

	if len(expectEnt.UserGroupFilter.UserGroups) != len(cur.UserGroupFilter.UserGroups) {
		return fmt.Errorf("expect number of target user group %v but got %v", len(expectEnt.UserGroupFilter.UserGroups), len(cur.UserGroupFilter.UserGroups))
	}

	return nil
}

func (s *NotificationSuite) CheckInfoNotificationMsgResponse(expect *entities.InfoNotificationMsg, current *entities.InfoNotificationMsg) error {
	if expect.Title.String != current.Title.String {
		return fmt.Errorf("expect notification message title %v but got %v", expect.Title, current.Title.String)
	}

	expectContent, _ := expect.GetContent()
	currentContent, err := current.GetContent()
	if err != nil {
		return err
	}

	if expectContent.Raw != currentContent.Raw {
		return fmt.Errorf("expect notification message content raw %v but got %v", expect.Content, current.Content)
	}

	url, _ := generateUploadURL(s.Storage.Endpoint, s.Storage.Bucket, expectContent.RenderedURL)
	if url != currentContent.RenderedURL {
		return fmt.Errorf("expect notification message content rendered url %v but got %v", url, currentContent.RenderedURL)
	}
	if len(expect.MediaIDs.Elements) != len(current.MediaIDs.Elements) {
		return fmt.Errorf("expect notification message medias %v but got %v", expect.MediaIDs, current.MediaIDs)
	}

	for i := range expect.MediaIDs.Elements {
		if expect.MediaIDs.Elements[i].String != current.MediaIDs.Elements[i].String {
			return fmt.Errorf("expect notification message medias %v but got %v", expect.MediaIDs, current.MediaIDs)
		}
	}

	return nil
}

type NotificationWithOpts struct {
	UserGroups                 string
	CourseFilter               string
	GradeFilter                string
	LocationFilter             string
	ClassFilter                string
	SchoolFilter               string
	IndividualFilter           string
	ScheduledStatus            string
	Status                     string
	IsImportant                bool
	ReceiverIds                []string
	GenericReceiverIds         []string
	ExcludedGenericReceiverIds []string
	MediaIds                   []string
	ExcludedGenericReceiverStr string
}

func (s *NotificationSuite) CurrentStaffUpsertNotificationWithOpts(ctx context.Context, opts *NotificationWithOpts) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	ctx, stepState.Notification, err = s.GetNotificationWithOptions(ctx, opts)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("GetNotificationWithOptions: %v", err)
	}

	stepState.Request = &npb.UpsertNotificationRequest{
		Notification: stepState.Notification,
	}

	if stepState.Questionnaire != nil {
		stepState.Request.(*npb.UpsertNotificationRequest).Questionnaire = stepState.Questionnaire
	}

	stepState.Response, stepState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, stepState.CurrentStaff.Token), stepState.Request.(*npb.UpsertNotificationRequest))

	if stepState.ResponseErr == nil {
		resp := stepState.Response.(*npb.UpsertNotificationResponse)
		stepState.Notification.NotificationId = resp.NotificationId
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *NotificationSuite) StaffUpdateNotificationWithLocationFilterChange(ctx context.Context, role string, locationFilter string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	notification := stepState.Notification
	infoNotification := &entities.InfoNotification{}
	fields := database.GetFieldNames(infoNotification)
	queryGetNotication := fmt.Sprintf(`SELECT %s FROM %s WHERE notification_id = $1;`, strings.Join(fields, ","), infoNotification.TableName())
	notification.EditorId = stepState.CurrentStaff.ID

	err := database.Select(ctx, s.BobDBConn, queryGetNotication, stepState.Notification.NotificationId).ScanOne(infoNotification)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	notification.Message.NotificationMsgId = infoNotification.NotificationMsgID.String

	locationIDs, selectType, err := s.GetLocationIDsFromString(ctx, locationFilter)
	if err != nil {
		return ctx, err
	}
	notification.TargetGroup.LocationFilter = &cpb.NotificationTargetGroup_LocationFilter{
		Type:        selectType,
		LocationIds: locationIDs,
	}

	stepState.Request = &npb.UpsertNotificationRequest{
		Notification: notification,
	}

	token := ""
	switch role {
	case "current staff":
		token = stepState.CurrentStaff.Token
	default:
		token = stepState.Organization.Staffs[0].Token
	}

	stepState.Response, stepState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, token), stepState.Request.(*npb.UpsertNotificationRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *NotificationSuite) CheckNotificationFilterData(ctx context.Context, targetGroup *entities.InfoNotificationTarget) error {
	stepState := StepStateFromContext(ctx)

	queryGetLocationIDs := `
			SELECT location_id 
			FROM notification_location_filter nlf 
			WHERE notification_id = $1 AND deleted_at IS NULL;
		`
	rows, err := s.BobDBConn.Query(ctx, queryGetLocationIDs, stepState.Notification.NotificationId)
	if err != nil {
		return fmt.Errorf("cannot get location filter: %v", err)
	}
	defer rows.Close()

	locationIDs := make([]string, 0)
	for rows.Next() {
		var locationID pgtype.Text
		err = rows.Scan(&locationID)
		if err != nil {
			return fmt.Errorf("cannot assign location when query: %v", err)
		}

		locationIDs = append(locationIDs, locationID.String)
		if !slices.Contains(targetGroup.LocationFilter.LocationIDs, locationID.String) {
			return fmt.Errorf("unexpect location %s", locationID.String)
		}
	}

	switch targetGroup.LocationFilter.Type {
	case consts.TargetGroupSelectTypeList.String():
		if len(locationIDs) != len(targetGroup.LocationFilter.LocationIDs) {
			return fmt.Errorf("expect number of target locations %v but got %v", len(targetGroup.LocationFilter.LocationIDs), len(locationIDs))
		}
	default:
		if len(locationIDs) != 0 {
			return fmt.Errorf("expect number of target locations is 0 for type select all or none but got %v", len(locationIDs))
		}
	}

	queryGetCourseIDs := `
			SELECT course_id 
			FROM notification_course_filter ncf 
			WHERE notification_id = $1 AND deleted_at IS NULL;
		`
	rows, err = s.BobDBConn.Query(ctx, queryGetCourseIDs, stepState.Notification.NotificationId)
	if err != nil {
		return fmt.Errorf("cannot get course filter: %v", err)
	}
	defer rows.Close()

	courseIDs := make([]string, 0)
	for rows.Next() {
		var courseID pgtype.Text
		err = rows.Scan(&courseID)
		if err != nil {
			return fmt.Errorf("cannot assign course when query: %v", err)
		}

		courseIDs = append(courseIDs, courseID.String)
		if !slices.Contains(targetGroup.CourseFilter.CourseIDs, courseID.String) {
			return fmt.Errorf("unexpect course %s", courseID.String)
		}
	}

	switch targetGroup.CourseFilter.Type {
	case consts.TargetGroupSelectTypeList.String():
		if len(courseIDs) != len(targetGroup.CourseFilter.CourseIDs) {
			return fmt.Errorf("expect number of target course %v but got %v", len(targetGroup.CourseFilter.CourseIDs), len(courseIDs))
		}
	default:
		if len(courseIDs) != 0 {
			return fmt.Errorf("expect number of target course is 0 for type select all or none but got %v", len(courseIDs))
		}
	}

	queryGetClassIDs := `
			SELECT class_id 
			FROM notification_class_filter ncf 
			WHERE notification_id = $1 AND deleted_at IS NULL;
		`
	rows, err = s.BobDBConn.Query(ctx, queryGetClassIDs, stepState.Notification.NotificationId)
	if err != nil {
		return fmt.Errorf("cannot get class filter: %v", err)
	}
	defer rows.Close()

	classIDs := make([]string, 0)
	for rows.Next() {
		var classID pgtype.Text
		err = rows.Scan(&classID)
		if err != nil {
			return fmt.Errorf("cannot assign class when query: %v", err)
		}

		classIDs = append(classIDs, classID.String)
		if !slices.Contains(targetGroup.ClassFilter.ClassIDs, classID.String) {
			return fmt.Errorf("unexpect class %s", classID.String)
		}
	}

	switch targetGroup.ClassFilter.Type {
	case consts.TargetGroupSelectTypeList.String():
		if len(classIDs) != len(targetGroup.ClassFilter.ClassIDs) {
			return fmt.Errorf("expect number of target class %v but got %v", len(targetGroup.ClassFilter.ClassIDs), len(classIDs))
		}
	default:
		if len(classIDs) != 0 {
			return fmt.Errorf("expect number of target class is 0 for type select all or none but got %v", len(classIDs))
		}
	}

	return nil
}
