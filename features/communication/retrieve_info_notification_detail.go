package communication

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/manabie-com/backend/features/communication/common"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"google.golang.org/grpc/status"
)

type RetrieveInfoNotificationDetailSuite struct {
	*common.NotificationSuite
	NotificationList         []*cpb.Notification
	NotificationInfoListResp []*npb.RetrieveNotificationsResponse_NotificationInfo
	ReadNotiCount            int
	studentToken             string
	studentID                string
}

func (c *SuiteConstructor) InitRetrieveInfoNotificationDetail(dep *DependencyV2, ctx *godog.ScenarioContext) {
	s := &RetrieveInfoNotificationDetailSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^school admin sends some notificationss to a student$`: s.schoolAdminSendsSomeNotificationssToAStudent,
		`^student retrieve notification detail$`:                s.studentRetrieveNotificationDetail,
		`^returns correct notification detail$`:                 s.returnsCorrectNotificationDetail,
		`^student retrieves list of notifications$`:             s.studentRetrievesListOfNotifications,
		`^returns correct list of notifications$`:               s.returnsCorrectListOfNotifications,
		`^student reads some notifications$`:                    s.studentReadsSomeNotifications,
		`^returns correct number of read notification$`:         s.returnsCorrectNumberOfReadNotification,
		`^student counts number of read notification$`:          s.studentCountsNumberOfReadNotification,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students$`: s.CreatesNumberOfStudents,
		`^student logins to Learner App$`:           s.studentLoginsToLearnerApp,
	}

	c.InitScenarioStepMapping(ctx, stepsMapping)
}

func (s *RetrieveInfoNotificationDetailSuite) studentLoginsToLearnerApp(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	student := commonState.Students[0]
	studentToken, err := s.GenerateExchangeTokenCtx(ctx, student.ID, "student")
	if err != nil {
		return ctx, fmt.Errorf("failed login learner app: %v", err)
	}
	s.studentToken = studentToken
	s.studentID = student.ID
	return ctx, nil
}

func (s *RetrieveInfoNotificationDetailSuite) schoolAdminSendsSomeNotificationssToAStudent(ctx context.Context) (context.Context, error) {
	randNum := common.RandRangeIn(1, 5)
	opts := &common.NotificationWithOpts{
		UserGroups:       "student",
		CourseFilter:     "random",
		GradeFilter:      "random",
		LocationFilter:   "none",
		ClassFilter:      "none",
		IndividualFilter: "none",
		ScheduledStatus:  "none",
		Status:           "NOTIFICATION_STATUS_DRAFT",
		IsImportant:      false,
		ReceiverIds:      []string{s.studentID},
	}
	for i := 0; i < randNum; i++ {
		_, notification, err := s.GetNotificationWithOptions(ctx, opts)
		if err != nil {
			return ctx, fmt.Errorf("failed GetNotificationWithOptions %v", err)
		}

		res, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(
			ctx,
			&npb.UpsertNotificationRequest{
				Notification: notification,
			},
		)
		if err != nil {
			return ctx, fmt.Errorf("failed UpsertNotification %v", err)
		}
		notification.NotificationId = res.NotificationId

		_, err = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).SendNotification(
			ctx,
			&npb.SendNotificationRequest{
				NotificationId: res.NotificationId,
			},
		)
		if err != nil {
			return ctx, fmt.Errorf("failed SendNotification %v", err)
		}
		s.NotificationList = append(s.NotificationList, notification)
	}

	return ctx, nil
}

func (s *RetrieveInfoNotificationDetailSuite) returnsCorrectNotificationDetail(ctx context.Context) (context.Context, error) {
	resp := s.Response.(*npb.RetrieveNotificationDetailResponse)

	noti := resp.Item

	expectNoti := s.NotificationList[0]

	// check notification
	ctx, err := s.checkNotification(ctx, noti, expectNoti)
	if err != nil {
		return ctx, fmt.Errorf("checkNotification %v", err)
	}

	// check notification message
	ctx, err = s.checkNotificationMessage(ctx, noti, expectNoti)
	if err != nil {
		return ctx, fmt.Errorf("checkNotificationMessage %v", err)
	}

	return ctx, nil
}

func (s *RetrieveInfoNotificationDetailSuite) checkNotification(ctx context.Context, cur *cpb.Notification, expect *cpb.Notification) (context.Context, error) {
	if cur.NotificationId != expect.NotificationId {
		return ctx, fmt.Errorf("expect notification id %v but got %v", expect.NotificationId, cur.NotificationId)
	}

	return ctx, nil
}

func (s *RetrieveInfoNotificationDetailSuite) checkNotificationMessage(ctx context.Context, cur *cpb.Notification, expect *cpb.Notification) (context.Context, error) {
	if cur.Message.Title != expect.Message.Title {
		return ctx, fmt.Errorf("expect notification tile %v but got %v", cur.Message.Title, expect.Message.Title)
	}

	content := expect.Message.GetContent()
	if cur.Message.Content.Raw != content.Raw {
		return ctx, fmt.Errorf("expect notification content raw %v but got %v", content.Raw, cur.Message.Content.Raw)
	}

	if len(cur.Message.MediaIds) != len(expect.Message.MediaIds) {
		return ctx, fmt.Errorf("expect notification number of medias %v but got %v", len(expect.Message.MediaIds), len(cur.Message.MediaIds))
	}
	return ctx, nil
}

func (s *RetrieveInfoNotificationDetailSuite) studentRetrieveNotificationDetail(ctx context.Context) (context.Context, error) {
	// resp, err := s.GetDetailNotification(s.student.Token, s.notifications[0].NotificationId)
	notiID := s.NotificationList[0].NotificationId
	commonState := common.StepStateFromContext(ctx)
	resp, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotificationDetail(
		common.ContextWithToken(ctx, s.studentToken),
		&npb.RetrieveNotificationDetailRequest{
			NotificationId: notiID,
			TargetId:       commonState.Students[0].ID,
		},
	)

	s.Response = resp

	if err != nil {
		return ctx, err
	}

	if resp.Item == nil {
		return ctx, fmt.Errorf("expect get notification with id %s but not found", notiID)
	}

	if resp.Item.NotificationId != notiID {
		return ctx, fmt.Errorf("expect notification id %s but got %v", notiID, resp.Item.NotificationId)
	}

	return ctx, nil
}

func (s *RetrieveInfoNotificationDetailSuite) returnsCorrectListOfNotifications(ctx context.Context) (context.Context, error) {
	items := s.NotificationInfoListResp
	expectItems := s.NotificationList
	if len(items) != len(expectItems) {
		return ctx, fmt.Errorf("expect number of item %v but got %v", len(expectItems), len(items))
	}

	// check items is sorted by updated_at DESC
	// items[i-1] has been updated after items[i]
	for i := 1; i < len(items); i++ {
		prev := items[i-1]
		cur := items[i]

		prevUpdatedAt := prev.UserNotification.UpdatedAt.AsTime()
		curUpdatedAt := cur.UserNotification.UpdatedAt.AsTime()
		if prevUpdatedAt.Before(curUpdatedAt) {
			return ctx, fmt.Errorf("expect items is sorted decreasingly by updated_at, but got %v and %v", prevUpdatedAt, curUpdatedAt)
		}

		if prevUpdatedAt.After(curUpdatedAt) {
			continue
		}

		prevNotificationID := prev.UserNotification.NotificationId
		curNotificationID := cur.UserNotification.NotificationId
		if prevNotificationID < curNotificationID {
			return ctx, fmt.Errorf("expect items is sorted decreasingly by notification id, but got %v and %v", prevNotificationID, curNotificationID)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].UserNotification.NotificationId < items[j].UserNotification.NotificationId
	})
	sort.Slice(expectItems, func(i, j int) bool {
		return expectItems[i].NotificationId < expectItems[j].NotificationId
	})

	for i, noti := range expectItems {
		if s.studentID != items[i].UserNotification.UserId {
			return ctx, fmt.Errorf("expect user id %v but got %v", s.studentID, items[i].UserNotification.UserId)
		}

		if expectItems[i].NotificationId != items[i].UserNotification.NotificationId {
			return ctx, fmt.Errorf("expect notification id %v but got %v", expectItems[i].NotificationId, items[i].UserNotification.NotificationId)
		}

		notiMsg := noti.Message
		if notiMsg.Title != items[i].Title {
			return ctx, fmt.Errorf("expect title %v but got %v", notiMsg.Title, items[i].Title)
		}
	}
	return ctx, nil
}

func (s *RetrieveInfoNotificationDetailSuite) studentRetrievesListOfNotifications(ctx context.Context) (context.Context, error) {
	nextPage := &cpb.Paging{
		Limit:  2,
		Offset: nil,
	}

	ctx, cancel := common.ContextWithTokenAndTimeOut(ctx, s.studentToken)
	defer cancel()

	for {
		req := &npb.RetrieveNotificationsRequest{Paging: nextPage}

		resp, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotifications(ctx, req)
		if err != nil {
			return ctx, err
		}

		if len(resp.Items) == 0 {
			break
		}

		s.NotificationInfoListResp = append(s.NotificationInfoListResp, resp.Items...)
		nextPage = resp.NextPage
	}

	return ctx, nil
}

func (s *RetrieveInfoNotificationDetailSuite) studentReadsSomeNotifications(ctx context.Context) (context.Context, error) {
	ctx = common.ContextWithToken(ctx, s.studentToken)

	for _, noti := range s.NotificationList {
		if rand.Intn(2) > 0 {
			req := &bpb.SetUserNotificationStatusRequest{
				NotificationIds: []string{noti.NotificationId},
				Status:          cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ,
			}
			s.Response, s.ResponseErr = bpb.NewNotificationModifierServiceClient(s.BobGRPCConn).SetUserNotificationStatus(ctx, req)

			ctx, err := s.returnsStatusCode(ctx, "OK")
			if err != nil {
				return ctx, err
			}

			s.ReadNotiCount++
		}
	}

	return ctx, nil
}

func (s *RetrieveInfoNotificationDetailSuite) returnsCorrectNumberOfReadNotification(ctx context.Context) (context.Context, error) {
	resp := s.Response.(*npb.CountUserNotificationResponse)

	if s.ReadNotiCount != int(resp.NumByStatus) {
		return ctx, fmt.Errorf("expect number of read noti is %v but got %v", s.ReadNotiCount, resp.NumByStatus)
	}

	if len(s.NotificationList) != int(resp.Total) {
		return ctx, fmt.Errorf("expect total number of noti is %v but got %v", len(s.NotificationList), resp.Total)
	}

	return ctx, nil
}

func (s *RetrieveInfoNotificationDetailSuite) returnsStatusCode(ctx context.Context, arg1 string) (context.Context, error) {
	stt, ok := status.FromError(s.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", s.ResponseErr.Error())
	}
	if stt.Code().String() != arg1 {
		return ctx, fmt.Errorf("expecting %s, got %s status code, message: %s", arg1, stt.Code().String(), stt.Message())
	}
	return ctx, nil
}

func (s *RetrieveInfoNotificationDetailSuite) studentCountsNumberOfReadNotification(ctx context.Context) (context.Context, error) {
	req := &npb.CountUserNotificationRequest{
		Status: cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ,
	}

	resp, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).CountUserNotification(ctx, req)
	if err != nil {
		return ctx, err
	}

	s.Response = resp
	return ctx, nil
}
