package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	statusCoutings = [...]string{
		cpb.NotificationStatus_NOTIFICATION_STATUS_NONE.String(),
		cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String(),
		cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String(),
		cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String(),
	}
)

func (svc *NotificationReaderService) GetNotificationsByFilter(ctx context.Context, req *npb.GetNotificationsByFilterRequest) (*npb.GetNotificationsByFilterResponse, error) {
	if req.Paging == nil {
		req.Paging = &cpb.Paging{
			Limit:  100,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0},
		}
	}

	if req.Paging.Limit == 0 {
		req.Paging.Limit = 100
	}

	notificationsFilter, countNotificationsForStatusFilter, err := svc.makeNotificationsFilter(ctx, req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "an error occurred when makeNotificationsFilter: "+err.Error())
	}

	var notifications entities.InfoNotifications
	var totalNotificationsForStatus map[string]uint32
	var wg sync.WaitGroup
	var errGetListAndGetCounting error
	errChan := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		notifications, err = svc.InfoNotificationRepo.Find(ctx, svc.DB, notificationsFilter)
		if err != nil {
			errChan <- err
		}
	}()
	go func() {
		defer wg.Done()
		totalNotificationsForStatus, err = svc.InfoNotificationRepo.CountTotalNotificationForStatus(ctx, svc.DB, countNotificationsForStatusFilter)
		if err != nil {
			errChan <- err
		}
	}()
	go func() {
		wg.Wait()
		close(errChan)
	}()
	for errInChan := range errChan {
		if errInChan != nil {
			errGetListAndGetCounting = multierr.Append(errGetListAndGetCounting, errInChan)
		}
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "an error occurred when finding/counting notifications: "+errGetListAndGetCounting.Error())
	}

	notiIDs := make([]string, 0, len(notifications))
	for _, notification := range notifications {
		notiIDs = append(notiIDs, notification.NotificationID.String)
	}

	notiMsgMap, err := svc.InfoNotificationMsgRepo.GetByNotificationIDs(ctx, svc.DB, database.TextArray(notiIDs))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("an error occurred when get notification message: %v", err))
	}

	notificationTags, err := svc.InfoNotificationTagRepo.GetByNotificationIDs(ctx, svc.DB, database.TextArray(notiIDs))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("an error occurred when get notification tags: %v", err))
	}

	notificationsPb, err := mappers.NotificationsFilteredToPb(notifications, notiMsgMap, notificationTags)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("an error occurred when convert NotificationsFilteredToPb: %v", err))
	}

	totalNotificationsForStatusPb := []*npb.GetNotificationsByFilterResponse_TotalNotificationForStatus{}
	var totalItem uint32

	for _, statusCouting := range statusCoutings {
		totalNotificationForStatus := &npb.GetNotificationsByFilterResponse_TotalNotificationForStatus{
			Status:     cpb.NotificationStatus(cpb.NotificationStatus_value[statusCouting]),
			TotalItems: 0,
		}
		totalCount, ok := totalNotificationsForStatus[statusCouting]
		if ok {
			totalNotificationForStatus.TotalItems = totalCount
		}
		totalNotificationsForStatusPb = append(totalNotificationsForStatusPb, totalNotificationForStatus)

		// Note: total items here is shown the number of notifications in the tab user is selected on the notification page
		if statusCouting == req.Status.String() {
			totalItem = totalNotificationForStatus.TotalItems
		}
	}

	offsetPre := req.Paging.GetOffsetInteger() - int64(req.Paging.Limit)
	if offsetPre < 0 {
		offsetPre = 0
	}

	return &npb.GetNotificationsByFilterResponse{
		Notifications:       notificationsPb,
		TotalItemsForStatus: totalNotificationsForStatusPb,
		TotalItems:          totalItem,
		NextPage: &cpb.Paging{
			Limit:  req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: req.Paging.GetOffsetInteger() + int64(len(notifications))},
		},
		PreviousPage: &cpb.Paging{
			Limit:  req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: offsetPre},
		},
	}, nil
}

func (svc *NotificationReaderService) makeNotificationsFilter(ctx context.Context, req *npb.GetNotificationsByFilterRequest) (*repositories.FindNotificationFilter, *repositories.FindNotificationFilter, error) {
	notificationsFilter := repositories.NewFindNotificationFilter()
	countNotificationsForStatusFilter := repositories.NewFindNotificationFilter()

	// Notification type must be NOTIFICATION_TYPE_COMPOSED
	_ = notificationsFilter.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())
	_ = countNotificationsForStatusFilter.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())

	if req.Status == cpb.NotificationStatus_NOTIFICATION_STATUS_NONE {
		_ = notificationsFilter.Status.Set([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String()})
	} else {
		_ = notificationsFilter.Status.Set([]string{req.Status.String()})
	}

	err := multierr.Combine(
		notificationsFilter.Offset.Set(req.Paging.GetOffsetInteger()),
		notificationsFilter.Limit.Set(req.Paging.Limit),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot assign filter (limit, offset): %v", err)
	}

	notificationsFilter.FromSent = database.TimestamptzFromPb(req.SentFrom)
	notificationsFilter.ToSent = database.TimestamptzFromPb(req.SentTo)
	countNotificationsForStatusFilter.FromSent = database.TimestamptzFromPb(req.SentFrom)
	countNotificationsForStatusFilter.ToSent = database.TimestamptzFromPb(req.SentTo)

	if len(req.ComposerIds) > 0 {
		notificationsFilter.EditorIDs = database.TextArray(req.ComposerIds)
		countNotificationsForStatusFilter.EditorIDs = database.TextArray(req.ComposerIds)
	}

	if req.Keyword != "" {
		notificationMsgID, err := svc.InfoNotificationMsgRepo.GetIDsByTitle(ctx, svc.DB, database.Text(req.Keyword))
		if err != nil {
			return nil, nil, fmt.Errorf("an error occurred when get notification message by title: %v", err)
		}
		_ = notificationsFilter.NotificationMsgIDs.Set(notificationMsgID)
		countNotificationsForStatusFilter.NotificationMsgIDs = notificationsFilter.NotificationMsgIDs
	}

	// Filter by notification_id (tag, location, course, class filter)
	// Noted:
	//   - If notificationIDsFilter == NULL, it will not be affected to the next filter
	//   - If notificationIDsFilter != NULL, the next filter will have empty results
	var notificationIDsFilter []string = nil
	if len(req.TagIds) > 0 {
		notificationIDs, err := svc.InfoNotificationTagRepo.GetNotificationIDsByTagIDs(ctx, svc.DB, database.TextArray(req.TagIds))
		if err != nil {
			return nil, nil, fmt.Errorf("an error occurred when get notification tag by title: %v", err)
		}

		notificationIDsFilter = notificationIDs
		if notificationIDsFilter == nil {
			notificationIDsFilter = make([]string, 0)
		}
	}

	targetGroup := req.TargetGroup
	if targetGroup != nil {
		if targetGroup.LocationFilter != nil {
			locationIDs := targetGroup.LocationFilter.LocationIds
			if targetGroup.LocationFilter.Type.String() == consts.TargetGroupSelectTypeAll.String() {
				_ = countNotificationsForStatusFilter.IsLocationSelectionAll.Set(true)
				_ = notificationsFilter.IsLocationSelectionAll.Set(true)
			} else if len(locationIDs) > 0 {
				notificationIDs, err := svc.NotificationLocationFilterRepo.GetNotificationIDsByLocationIDs(ctx, svc.DB, database.TextArray(notificationIDsFilter), database.TextArray(targetGroup.LocationFilter.LocationIds))
				if err != nil {
					return nil, nil, fmt.Errorf("an error occurred when get notification location filter: %v", err)
				}

				notificationIDsFilter = notificationIDs
				if notificationIDsFilter == nil {
					notificationIDsFilter = make([]string, 0)
				}
			}
		}

		if targetGroup.CourseFilter != nil {
			courseIDs := targetGroup.CourseFilter.CourseIds
			if targetGroup.CourseFilter.Type.String() == consts.TargetGroupSelectTypeAll.String() {
				_ = countNotificationsForStatusFilter.IsCourseSelectionAll.Set(true)
				_ = notificationsFilter.IsCourseSelectionAll.Set(true)
			} else if len(courseIDs) > 0 {
				notificationIDs, err := svc.NotificationCourseFilterRepo.GetNotificationIDsByCourseIDs(ctx, svc.DB, database.TextArray(notificationIDsFilter), database.TextArray(targetGroup.CourseFilter.CourseIds))
				if err != nil {
					return nil, nil, fmt.Errorf("an error occurred when get notification course filter: %v", err)
				}

				notificationIDsFilter = notificationIDs
				if notificationIDsFilter == nil {
					notificationIDsFilter = make([]string, 0)
				}
			}
		}

		if targetGroup.ClassFilter != nil {
			classIDs := targetGroup.ClassFilter.ClassIds
			if targetGroup.ClassFilter.Type.String() == consts.TargetGroupSelectTypeAll.String() {
				_ = countNotificationsForStatusFilter.IsClassSelectionAll.Set(true)
				_ = notificationsFilter.IsClassSelectionAll.Set(true)
			} else if len(classIDs) > 0 {
				notificationIDs, err := svc.NotificationClassFilterRepo.GetNotificationIDsByClassIDs(ctx, svc.DB, database.TextArray(notificationIDsFilter), database.TextArray(targetGroup.ClassFilter.ClassIds))
				if err != nil {
					return nil, nil, fmt.Errorf("an error occurred when get notification class filter: %v", err)
				}

				notificationIDsFilter = notificationIDs
				if notificationIDsFilter == nil {
					notificationIDsFilter = make([]string, 0)
				}
			}
		}
	}

	if req.IsQuestionnaireFullySubmitted {
		submittedStatus := database.Text(cpb.UserNotificationQuestionnaireStatus_USER_NOTIFICATION_QUESTIONNAIRE_STATUS_ANSWERED.String())
		notificationIDs, err := svc.UserInfoNotificationRepo.GetNotificationIDWithFullyQnStatus(ctx, svc.DB, database.TextArray(notificationIDsFilter), submittedStatus)
		if err != nil {
			return nil, nil, fmt.Errorf("an error occurred when get notification with fully questionnaire submitted: %v", err)
		}

		notificationIDsFilter = notificationIDs
		if notificationIDsFilter == nil {
			notificationIDsFilter = make([]string, 0)
		}
	}

	if len(notificationIDsFilter) > 0 {
		notificationIDsFilter = golibs.GetUniqueElementStringArray(notificationIDsFilter)
	}

	_ = notificationsFilter.NotiIDs.Set(notificationIDsFilter)
	countNotificationsForStatusFilter.NotiIDs = notificationsFilter.NotiIDs

	return notificationsFilter, countNotificationsForStatusFilter, nil
}
