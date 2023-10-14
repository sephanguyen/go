package communication

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/bxcodec/faker/v3/support/slice"
	"github.com/cucumber/godog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetNotificationsByFilterSuite struct {
	*common.NotificationSuite
	tagIDs                         []string
	notifications                  []*cpb.Notification
	notificationsRes               *npb.GetNotificationsByFilterResponse
	notificationsTags              map[string][]string
	typeStatusFilter               cpb.NotificationStatus
	mapNotiCoutingByStatusWithTags map[string]int
	createdQN                      *cpb.Questionnaire
}

func (c *SuiteConstructor) InitGetNotificationsByFilter(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &GetNotificationsByFilterSuite{
		NotificationSuite:              dep.notiCommonSuite,
		notificationsTags:              make(map[string][]string, 0),
		mapNotiCoutingByStatusWithTags: make(map[string]int, 0),
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students$`:                                       s.CreatesNumberOfStudents,
		`^school admin creates "([^"]*)" courses with "([^"]*)" classes for each course$`: s.CreatesNumberOfCoursesWithClass,
		`^school admin add packages data of those courses for each student$`:              s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^current staff upsert notification to "([^"]*)" and "([^"]*)" course and "([^"]*)" grade and "([^"]*)" location and "([^"]*)" class and "([^"]*)" school and "([^"]*)" individuals and "([^"]*)" scheduled time with "([^"]*)" and important is "([^"]*)"$`: s.CurrentStaffUpsertNotificationWithFilter,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a current organization$`:                                                                                                                                                        s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfCurrentOrg,
		`^returns "([^"]*)" status code$`:      s.CheckReturnStatusCode,
		`^school admin has created some tags$`: s.schoolAdminHasCreatedTags,
		`^current staff upsert some notifications with "([^"]*)" draft and "([^"]*)" schedule after one day to some students$`: s.aSchoolAdminUpsertSomeNotifications,
		`^sends "([^"]*)" of drafted notifications$`:               s.sendsDraftedNotifications,
		`^discards "([^"]*)" of drafted notifications$`:            s.discardsDraftedNotifications,
		`^see "([^"]*)" notifications in CMS with corrected data$`: s.seeNotificationsInCMSWithCorrectedData,
		`^see "([^"]*)" drafted and "([^"]*)" sent and "([^"]*)" scheduled and "([^"]*)" total notifications count in CMS notification tab$`: s.seeNotificationsCountForAllAndStatusIsCorrected,
		`^attach some tags for "([^"]*)" notifications$`: s.attachSomeTagsForNotifications,
		`^current staff get notifications by filter with status is "([^"]*)" and tags is "([^"]*)" and send time from is "([^"]*)" and send time to is "([^"]*)" and title is "([^"]*)" and target_group filter is "([^"]*)" and limit is "([^"]*)" and offset is "([^"]*)" and fully questionnaire submitted is "([^"]*)" and composer is "([^"]*)"$`: s.aSchoolAdminGetNotificationsByFilter,
		`^current staff upsert "([^"]*)" draft notifications with "([^"]*)" have title is "([^"]*)" and the rest have random title$`: s.aSchoolAdminUpsertSomeNotificationsWithTitle,
		`^see previous offset is "([^"]*)" and next offset is "([^"]*)"$`:                                                            s.checkOffsetPaging,
		`^remove all tags on all notifications$`:                                                                                     s.removeAllTagsOnAllNotifications,
		`^a questionnaire with resubmit allowed "([^"]*)", questions "([^"]*)" respectively$`:                                        s.aQuestionnaireWithResubmitAllowedQuestionsRespectively,
		`^current staff send a notification with attached questionnaire to "([^"]*)"$`:                                               s.currentStaffSendANotificationWithAttachedQuestionnaire,
		`^parent answer questionnaire for "([^"]*)"$`:                                                                                s.parentAnswerQuestionnaire,
		`^school admin creates "([^"]*)" students with the same parent$`:                                                             s.CreatesNumberOfStudentsWithSameParentsInfo,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *GetNotificationsByFilterSuite) aSchoolAdminUpsertSomeNotifications(ctx context.Context, numDraft int, numSchedule int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	for i := 0; i < numDraft; i++ {
		student := commonState.Students[common.RandRangeIn(0, len(commonState.Students))]

		opts := &common.NotificationWithOpts{
			UserGroups:         "student",
			CourseFilter:       "random",
			GradeFilter:        "random",
			LocationFilter:     "random",
			ClassFilter:        "random",
			IndividualFilter:   "none",
			ScheduledStatus:    "none",
			Status:             "NOTIFICATION_STATUS_DRAFT",
			IsImportant:        false,
			GenericReceiverIds: []string{student.ID},
		}
		ctx, noti, err := s.GetNotificationWithOptions(ctx, opts)
		if err != nil {
			return ctx, fmt.Errorf("failed GetNotificationWithOptions Draft: %v", err)
		}
		s.notifications = append(s.notifications, noti)

		req := &npb.UpsertNotificationRequest{
			Notification: noti,
		}

		res, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), req)
		if err != nil {
			return ctx, fmt.Errorf("UpsertNotification %s", err)
		}

		noti.NotificationId = res.NotificationId
	}

	for i := 0; i < numSchedule; i++ {
		student := commonState.Students[common.RandRangeIn(0, len(commonState.Students))]

		opts := &common.NotificationWithOpts{
			UserGroups:         "student",
			CourseFilter:       "random",
			GradeFilter:        "random",
			LocationFilter:     "random",
			ClassFilter:        "random",
			IndividualFilter:   "none",
			ScheduledStatus:    "none",
			Status:             cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String(),
			IsImportant:        false,
			GenericReceiverIds: []string{student.ID},
		}
		ctx, noti, err := s.GetNotificationWithOptions(ctx, opts)
		if err != nil {
			return ctx, fmt.Errorf("failed GetNotificationWithOptions Schedule: %v", err)
		}
		next := time.Now().Add(time.Duration(24 * time.Hour))
		noti.ScheduledAt = timestamppb.New(next)

		s.notifications = append(s.notifications, noti)

		req := &npb.UpsertNotificationRequest{
			Notification: noti,
		}

		res, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), req)
		if err != nil {
			return ctx, fmt.Errorf("UpsertNotification %s", err)
		}

		noti.NotificationId = res.NotificationId
	}

	return ctx, nil
}

func (s *GetNotificationsByFilterSuite) aSchoolAdminUpsertSomeNotificationsWithTitle(ctx context.Context, numNoti int, numHaveTitle int, title string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	numNotiTitle := 0
	for i := 0; i < numNoti; i++ {
		student := commonState.Students[common.RandRangeIn(0, len(commonState.Students))]

		noti := aSampleComposedNotification([]string{student.ID}, commonState.Organization.ID, "student", false)

		if numNotiTitle < numHaveTitle {
			timeNow := time.Now().Nanosecond()

			randTitleType := timeNow % 4
			switch randTitleType {
			case 0:
				noti.Message.Title = title
			case 1:
				noti.Message.Title = idutil.ULIDNow() + title
			case 2:
				noti.Message.Title = title + idutil.ULIDNow()
			default:
				noti.Message.Title = idutil.ULIDNow() + title + idutil.ULIDNow()
			}

			numNotiTitle++
		} else {
			noti.Message.Title = idutil.ULIDNow()
		}

		s.notifications = append(s.notifications, noti)

		req := &npb.UpsertNotificationRequest{
			Notification: noti,
		}

		res, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), req)
		if err != nil {
			return ctx, fmt.Errorf("UpsertNotification %s", err)
		}

		noti.NotificationId = res.NotificationId
	}

	return ctx, nil
}

func (s *GetNotificationsByFilterSuite) sendsDraftedNotifications(ctx context.Context, numSend int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	numSent := 0
	for _, noti := range s.notifications {
		if noti.Status == cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT {
			_, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).SendNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), &npb.SendNotificationRequest{
				NotificationId: noti.NotificationId,
			})

			if err != nil {
				return ctx, fmt.Errorf("SendNotification %s", err)
			}

			noti.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SENT
			numSent++
		}

		if numSend == numSent {
			return ctx, nil
		}
	}

	return ctx, nil
}

func (s *GetNotificationsByFilterSuite) discardsDraftedNotifications(ctx context.Context, numDiscard int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	numDiscarded := 0
	for _, noti := range s.notifications {
		if noti.Status == cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT {
			_, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).DiscardNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), &npb.DiscardNotificationRequest{
				NotificationId: noti.NotificationId,
			})

			if err != nil {
				return ctx, fmt.Errorf("DiscardNotification %s", err)
			}

			noti.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_DISCARD
			numDiscarded++
		}

		if numDiscard == numDiscarded {
			return ctx, nil
		}
	}

	return ctx, nil
}

func (s *GetNotificationsByFilterSuite) aSchoolAdminGetNotificationsByFilter(ctx context.Context, statusFilter string, tagFilter string, sendTimeFromFilter string, sendTimeToFilter string, title string, targetGroup string, limit string, offset string, isFullyQnSubmitted string, composer string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	reqFilter := &npb.GetNotificationsByFilterRequest{
		Keyword: "",
		TagIds:  []string{},
		Status:  cpb.NotificationStatus_NOTIFICATION_STATUS_NONE,
		Paging: &cpb.Paging{
			Limit: math.MaxUint32,
		},
	}

	commonState.Request = reqFilter

	if title != "none" {
		reqFilter.Keyword = title
	}

	if composer == "current" {
		reqFilter.ComposerIds = []string{commonState.CurrentStaff.ID}
	}

	switch statusFilter {
	case "all":
		reqFilter.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_NONE
	case "scheduled":
		reqFilter.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED
	case "draft":
		reqFilter.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT
	case "sent":
		reqFilter.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SENT
	}
	s.typeStatusFilter = reqFilter.Status

	switch tagFilter {
	case "none":
		reqFilter.TagIds = []string{}
	case "all tags added":
		reqFilter.TagIds = s.tagIDs
	}

	switch sendTimeFromFilter {
	case "1 min before":
		reqFilter.SentFrom = timestamppb.New(time.Now().Add(-1 * time.Minute))
	case "1 min after":
		reqFilter.SentFrom = timestamppb.New(time.Now().Add(1 * time.Minute))
	}

	switch sendTimeToFilter {
	case "1 min before":
		reqFilter.SentTo = timestamppb.New(time.Now().Add(-1 * time.Minute))
	case "1 min after":
		reqFilter.SentTo = timestamppb.New(time.Now().Add(1 * time.Minute))
	}

	if targetGroup == "none" {
		reqFilter.TargetGroup = nil
	} else {
		targetGroupFilters := strings.Split(targetGroup, ",")
		reqFilter.TargetGroup = &cpb.NotificationTargetGroup{
			LocationFilter: &cpb.NotificationTargetGroup_LocationFilter{
				LocationIds: []string{},
			},
			CourseFilter: &cpb.NotificationTargetGroup_CourseFilter{
				CourseIds: []string{},
			},
			ClassFilter: &cpb.NotificationTargetGroup_ClassFilter{
				ClassIds: []string{},
			},
			SchoolFilter: &cpb.NotificationTargetGroup_SchoolFilter{
				SchoolIds: []string{},
			},
		}
		for _, targetGroupFilter := range targetGroupFilters {
			switch targetGroupFilter {
			case "all location":
				reqFilter.TargetGroup.LocationFilter.Type = cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL
			case "all course":
				reqFilter.TargetGroup.CourseFilter.Type = cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL
			case "all class":
				reqFilter.TargetGroup.ClassFilter.Type = cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL
			case "list location":
				reqFilter.TargetGroup.LocationFilter.LocationIds = append(reqFilter.TargetGroup.LocationFilter.LocationIds, commonState.Notification.TargetGroup.LocationFilter.LocationIds...)
			case "list course":
				reqFilter.TargetGroup.CourseFilter.CourseIds = append(reqFilter.TargetGroup.CourseFilter.CourseIds, commonState.Notification.TargetGroup.CourseFilter.CourseIds...)
			case "list class":
				reqFilter.TargetGroup.ClassFilter.ClassIds = append(reqFilter.TargetGroup.ClassFilter.ClassIds, commonState.Notification.TargetGroup.ClassFilter.ClassIds...)
			}
		}
	}

	if limit != "none" {
		limitReq, err := strconv.Atoi(limit)
		if err != nil {
			return ctx, fmt.Errorf("limit convert (strconv.Atoi): %s", err)
		}
		reqFilter.Paging.Limit = uint32(limitReq)
	}

	if offset != "none" {
		offsetReq, err := strconv.Atoi(offset)
		if err != nil {
			return ctx, fmt.Errorf("offset convert (strconv.Atoi): %s", err)
		}
		reqFilter.Paging.Offset = &cpb.Paging_OffsetInteger{OffsetInteger: int64(offsetReq)}
	}

	if isFullyQnSubmitted == "true" {
		reqFilter.IsQuestionnaireFullySubmitted = true
	} else {
		reqFilter.IsQuestionnaireFullySubmitted = false
	}

	commonState.Response, commonState.ResponseErr = npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).GetNotificationsByFilter(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), reqFilter)

	if commonState.ResponseErr == nil {
		s.notificationsRes = commonState.Response.(*npb.GetNotificationsByFilterResponse)
	}

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *GetNotificationsByFilterSuite) seeNotificationsInCMSWithCorrectedData(ctx context.Context, numNoti int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	if numNoti != len(s.notificationsRes.Notifications) {
		return ctx, fmt.Errorf("invalid total notifications in list, expected: %v, got %v, notification_id: %s", numNoti, len(s.notificationsRes.Notifications), commonState.Notification.NotificationId)
	}

	if numNoti == 0 {
		return ctx, nil
	}

	notificationMap := make(map[string]*cpb.Notification, 0)

	for _, noti := range s.notifications {
		notificationMap[noti.NotificationId] = noti
	}

	for _, notiRes := range s.notificationsRes.Notifications {
		notiReq := notificationMap[notiRes.NotificationId]

		if notiReq == nil {
			notiReq = commonState.Notification
		}
		if notiReq.Message.Title != notiRes.Title {
			return ctx, fmt.Errorf("invalid Title, expected: %v, got %v", notiReq.Message.Title, notiRes.Title)
		}

		if len(notiRes.TagIds) > 0 {
			if ok, diff := protoEqualWithoutOrder(notiRes.TagIds, s.notificationsTags[notiRes.NotificationId], nil); !ok {
				return ctx, fmt.Errorf("notification tags is invalid: %s", diff)
			}
		}

		if s.typeStatusFilter != cpb.NotificationStatus_NOTIFICATION_STATUS_NONE && s.typeStatusFilter != notiRes.Status {
			return ctx, fmt.Errorf("notification status is invalid, expected: %s, got %s", s.typeStatusFilter.String(), notiRes.Status.String())
		}

		if notiRes.TargetGroup == nil {
			return ctx, fmt.Errorf("expected TargetGroup not nil")
		}

		expectTargetGroupEnt := mappers.PbToNotificationTargetEnt(notiReq.TargetGroup)
		actualTargetGroupEnt := mappers.PbToNotificationTargetEnt(notiRes.TargetGroup)
		err := s.CheckTargetGroupFilter(expectTargetGroupEnt, actualTargetGroupEnt)
		if err != nil {
			return ctx, fmt.Errorf("failed CheckTargetGroupFilter: %v", err)
		}
	}

	return ctx, nil
}

func (s *GetNotificationsByFilterSuite) seeNotificationsCountForAllAndStatusIsCorrected(ctx context.Context, numDraftStr string, numSendStr string, numScheduleStr string, numTotal int) (context.Context, error) {
	mapNotiStatusAndTotal := make(map[string]int, 0)

	for _, total := range s.notificationsRes.TotalItemsForStatus {
		mapNotiStatusAndTotal[total.Status.String()] = int(total.TotalItems)
	}

	numDraft := 0
	var err error
	if numDraftStr == "auto detect" {
		numDraft = s.mapNotiCoutingByStatusWithTags[cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String()]
	} else {
		numDraft, err = strconv.Atoi(numDraftStr)
		if err != nil {
			return ctx, err
		}
	}

	numSend := 0
	if numSendStr == "auto detect" {
		numSend = s.mapNotiCoutingByStatusWithTags[cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String()]
	} else {
		numSend, err = strconv.Atoi(numSendStr)
		if err != nil {
			return ctx, err
		}
	}

	numSchedule := 0
	if numScheduleStr == "auto detect" {
		numSchedule = s.mapNotiCoutingByStatusWithTags[cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String()]
	} else {
		numSchedule, err = strconv.Atoi(numScheduleStr)
		if err != nil {
			return ctx, err
		}
	}

	actualTotalDraft := mapNotiStatusAndTotal[cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String()]
	if numDraft != actualTotalDraft {
		return ctx, fmt.Errorf("invalid total drafted count, expected: %v, got %v", numDraft, actualTotalDraft)
	}

	actualTotalSent := mapNotiStatusAndTotal[cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String()]
	if numSend != actualTotalSent {
		return ctx, fmt.Errorf("invalid total sent count, expected: %v, got %v", numSend, actualTotalSent)
	}

	actualTotalSchedule := mapNotiStatusAndTotal[cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String()]
	if numSchedule != actualTotalSchedule {
		return ctx, fmt.Errorf("invalid total scheduled count, expected: %v, got %v", numSchedule, actualTotalSchedule)
	}

	actualTotal := mapNotiStatusAndTotal[cpb.NotificationStatus_NOTIFICATION_STATUS_NONE.String()]
	if numTotal != actualTotal {
		return ctx, fmt.Errorf("invalid total count, expected: %v, got %v", numTotal, actualTotal)
	}

	return ctx, nil
}

func (s *GetNotificationsByFilterSuite) schoolAdminHasCreatedTags(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	numRandTag := common.RandRangeIn(10, 15)
	for i := 0; i < numRandTag; i++ {
		res, err := npb.NewTagMgmtModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertTag(
			s.ContextWithToken(ctx, commonState.CurrentStaff.Token),
			&npb.UpsertTagRequest{
				Name: "Notification_Tag_" + idutil.ULIDNow(),
			},
		)
		if err != nil {
			return ctx, fmt.Errorf("UpsertTag: %w", err)
		}
		tagID := res.TagId
		s.tagIDs = append(s.tagIDs, tagID)
	}
	return ctx, nil
}

func (s *GetNotificationsByFilterSuite) attachSomeTagsForNotifications(ctx context.Context, numNotiTag int) (context.Context, error) {
	//Todo: replace by api enpoint
	countNumNotiTagAdded := 0
	for _, noti := range s.notifications {
		if noti.Status == cpb.NotificationStatus_NOTIFICATION_STATUS_DISCARD {
			continue
		}

		numTagAdd := common.RandRangeIn(1, 4)

		for i := 0; i < numTagAdd; i++ {
			tagID := s.tagIDs[common.RandRangeIn(0, len(s.tagIDs))]
			// prevent duplicate tag for a notification
			for slice.Contains(s.notificationsTags[noti.NotificationId], tagID) {
				tagID = s.tagIDs[common.RandRangeIn(0, len(s.tagIDs))]
			}

			commonState := common.StepStateFromContext(ctx)
			ctxRp := contextWithResourcePath(ctx, strconv.Itoa(int(commonState.Organization.ID)))
			stmt := `
				INSERT INTO public.info_notifications_tags
				(notification_tag_id, notification_id, tag_id, updated_at, created_at, deleted_at)
				VALUES ($1, $2, $3, now(), now(), NULL)
			`
			s.tagIDs = append(s.tagIDs, tagID)
			if _, err := s.BobDBConn.Exec(ctxRp, stmt, idutil.ULIDNow(), noti.NotificationId, tagID); err != nil {
				return ctx, err
			}

			s.notificationsTags[noti.NotificationId] = append(s.notificationsTags[noti.NotificationId], tagID)
		}

		if _, ok := s.mapNotiCoutingByStatusWithTags[noti.Status.String()]; !ok {
			s.mapNotiCoutingByStatusWithTags[noti.Status.String()] = 1
		} else {
			s.mapNotiCoutingByStatusWithTags[noti.Status.String()]++
		}

		countNumNotiTagAdded++
		if countNumNotiTagAdded == numNotiTag {
			return ctx, nil
		}
	}

	return ctx, nil
}

func (s *GetNotificationsByFilterSuite) removeAllTagsOnAllNotifications(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	for _, noti := range s.notifications {
		ctxRp := contextWithResourcePath(ctx, strconv.Itoa(int(commonState.Organization.ID)))
		if tagIDs, ok := s.notificationsTags[noti.NotificationId]; ok {
			for _, tagID := range tagIDs {
				stmt := `
				UPDATE info_notifications_tags 
				SET deleted_at = now()
				WHERE tag_id = $1 AND notification_id = $2
			`
				if _, err := s.BobDBConn.Exec(ctxRp, stmt, tagID, noti.NotificationId); err != nil {
					return ctx, err
				}
			}
		}
	}
	return ctx, nil
}

func (s *GetNotificationsByFilterSuite) checkOffsetPaging(ctx context.Context, previousOffset int, nextOffset int) (context.Context, error) {
	if int(s.notificationsRes.NextPage.GetOffsetInteger()) != nextOffset {
		return ctx, fmt.Errorf("expected next offset is %d, got %d", nextOffset, int(s.notificationsRes.NextPage.GetOffsetInteger()))
	}

	if int(s.notificationsRes.PreviousPage.GetOffsetInteger()) != previousOffset {
		return ctx, fmt.Errorf("expected previous offset is %d, got %d", previousOffset, int(s.notificationsRes.PreviousPage.GetOffsetInteger()))
	}
	return ctx, nil
}

// QUESTIONNAIRE

func (s *GetNotificationsByFilterSuite) aQuestionnaireWithResubmitAllowedQuestionsRespectively(ctx context.Context, resubmit string, questionStr string) (context.Context, error) {
	questions := parseQuestionFromString(questionStr)
	qn := &cpb.Questionnaire{
		ResubmitAllowed: common.StrToBool(resubmit),
		Questions:       questions,
		ExpirationDate:  timestamppb.New(time.Now().Add(24 * time.Hour)),
	}
	s.createdQN = qn
	return ctx, nil
}

func (s *GetNotificationsByFilterSuite) currentStaffSendANotificationWithAttachedQuestionnaire(ctx context.Context, receiver string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	opts := &common.NotificationWithOpts{
		CourseFilter:     "all",
		GradeFilter:      "all",
		LocationFilter:   "none",
		ClassFilter:      "none",
		IndividualFilter: "none",
		ScheduledStatus:  "none",
		Status:           "NOTIFICATION_STATUS_DRAFT",
		IsImportant:      false,
	}
	switch receiver {
	case "none":
		opts.CourseFilter = "none"
		opts.GradeFilter = "none"
		opts.UserGroups = "student, parent"
	case "student":
		opts.UserGroups = "student"
	case "parent":
		opts.UserGroups = "parent"
	default:
		opts.UserGroups = "student, parent"
	}

	genericUserIDs := []string{}
	opts.GenericReceiverIds = genericUserIDs

	var err error
	ctx, commonState.Notification, err = s.GetNotificationWithOptions(ctx, opts)
	if err != nil {
		return ctx, fmt.Errorf("failed GetNotificationWithOptions: %v", err)
	}

	req := &npb.UpsertNotificationRequest{
		Notification:  commonState.Notification,
		Questionnaire: s.createdQN,
	}

	res, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), req)
	if err != nil {
		return ctx, fmt.Errorf("UpsertNotification %s", err)
	}

	commonState.Notification.NotificationId = res.NotificationId
	_, err = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).SendNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), &npb.SendNotificationRequest{
		NotificationId: res.NotificationId,
	})
	if err != nil {
		return ctx, fmt.Errorf("SendNotification %s", err)
	}

	s.notifications = append(s.notifications, commonState.Notification)

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *GetNotificationsByFilterSuite) parentAnswerQuestionnaire(ctx context.Context, answerFor string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	parentID := commonState.Students[0].Parents[0].ID
	if parentID == "" {
		return ctx, fmt.Errorf("parentID is empty")
	}
	token, err := s.GenerateExchangeTokenCtx(ctx, parentID, constant.UserGroupParent)
	ctxWithToken := s.ContextWithToken(ctx, token)
	if err != nil {
		return ctx, err
	}

	res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotifications(ctxWithToken, &npb.RetrieveNotificationsRequest{
		Paging: &cpb.Paging{Limit: 100},
	})
	if err != nil {
		return ctx, err
	}

	studentIDMap := make(map[string]string)
	for index, student := range commonState.Students {
		studentIDMap[strconv.Itoa(index+1)] = student.ID
	}

	stuNotiIDMap := map[string]string{}
	for _, item := range res.Items {
		stuNotiIDMap[item.TargetId] = item.UserNotification.UserNotificationId
		s.createdQN.QuestionnaireId = item.QuestionnaireId
	}

	studentIdxes := strings.Split(answerFor, ",")
	userNotiIDs := make([]string, 0, len(studentIdxes))
	for _, answeredStu := range studentIdxes {
		stuID := studentIDMap[answeredStu]
		userNotiIDs = append(userNotiIDs, stuNotiIDMap[stuID])
	}

	err = s.userAnswerQuesionnaire(ctx, parentID, commonState.Students[0].ID, userNotiIDs)

	return ctx, err
}

func (s *GetNotificationsByFilterSuite) userAnswerQuesionnaire(ctx context.Context, userID, targetID string, userNotiIDs []string) error {
	commonState := common.StepStateFromContext(ctx)
	token, err := s.GenerateExchangeTokenCtx(ctx, userID, constant.UserGroupStudent)
	ctxWithToken := s.ContextWithToken(ctx, token)

	if err != nil {
		return err
	}

	notiInfo, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotificationDetail(ctxWithToken, &npb.RetrieveNotificationDetailRequest{
		NotificationId: commonState.Notification.NotificationId,
		TargetId:       targetID,
	})
	if err != nil {
		return err
	}

	questions := notiInfo.UserQuestionnaire.Questionnaire.Questions

	answers := makeAnswersListForOnlyRequiredQuestion(questions)

	for _, userNoti := range userNotiIDs {
		submitReq := &npb.SubmitQuestionnaireRequest{
			QuestionnaireId:        s.createdQN.QuestionnaireId,
			Answers:                answers,
			UserInfoNotificationId: userNoti,
		}
		_, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).SubmitQuestionnaire(ctxWithToken, submitReq)
		if err != nil {
			return err
		}
	}
	return nil
}
