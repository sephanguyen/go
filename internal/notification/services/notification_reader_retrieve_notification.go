package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (svc *NotificationReaderService) RetrieveNotificationDetail(ctx context.Context, req *npb.RetrieveNotificationDetailRequest) (*npb.RetrieveNotificationDetailResponse, error) {
	filter := mappers.NotificationDetailToUserNotificationFilter(ctx, req)

	es, err := svc.UserInfoNotificationRepo.Find(ctx, svc.DB, filter)
	if err != nil {
		return nil, fmt.Errorf("UserInfoNotificationRepo.Find: %v", err)
	}

	if len(es) == 0 {
		return &npb.RetrieveNotificationDetailResponse{}, nil
	}

	userNoti := es[0]
	noti, err := svc.findSentNotification(ctx, userNoti.NotificationID.String)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("RetrieveNotificationDetail.FindNotification: %v", err))
	}
	notiMsg, err := svc.findNotificationMsg(ctx, noti.NotificationMsgID.String)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("RetrieveNotificationDetail.FindNotificationMsg: %v", err))
	}
	resp := &npb.RetrieveNotificationDetailResponse{
		Item:             mappers.NotificationToPb(noti, notiMsg),
		UserNotification: mappers.ToUserNotificationPb(userNoti),
	}
	if noti.QuestionnaireID.Status != pgtype.Null && noti.QuestionnaireID.String != "" {
		isSubmitted := userNoti.QuestionnaireStatus.String == cpb.UserNotificationQuestionnaireStatus_USER_NOTIFICATION_QUESTIONNAIRE_STATUS_ANSWERED.String()
		userQn, err := svc.findAttachedQuestionnaireDetail(
			ctx, noti.QuestionnaireID.String, userNoti.UserNotificationID.String, isSubmitted)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("findAttachedQuestionnaireDetail: %v", err))
		}
		resp.UserQuestionnaire = userQn
	}

	return resp, nil
}

func (svc *NotificationReaderService) RetrieveNotifications(ctx context.Context, req *npb.RetrieveNotificationsRequest) (*npb.RetrieveNotificationsResponse, error) {
	if req.Paging == nil {
		req.Paging = &cpb.Paging{
			Limit:  100,
			Offset: nil,
		}
	}

	if req.Paging.Limit == 0 {
		req.Paging.Limit = 100
	}
	userID := interceptors.UserIDFromContext(ctx)
	userNoti, err := svc.findUserNotification(ctx, userID, req.Paging, req.ImportantOnly)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("FindUserNotification %v", err))
	}
	if len(userNoti) == 0 {
		return &npb.RetrieveNotificationsResponse{}, nil
	}

	notiIDs := make([]string, 0, len(userNoti))
	for _, un := range userNoti {
		notiIDs = append(notiIDs, un.NotificationID.String)
	}

	notiMsgMap, err := svc.InfoNotificationMsgRepo.GetByNotificationIDs(ctx, svc.DB, database.TextArray(notiIDs))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("InfoNotificationMsgRepo.GetByNotificationIDs: %v", err))
	}

	filter := repositories.NewFindNotificationFilter()
	_ = filter.NotiIDs.Set(notiIDs)

	notifications, err := svc.InfoNotificationRepo.Find(ctx, svc.DB, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("InfoNotificationRepo.GetNotificationByIDs: %v", err))
	}

	notiMap := make(map[string]*entities.InfoNotification)
	for _, el := range notifications {
		notiMap[el.NotificationID.String] = el
	}

	itemsPb := make([]*npb.RetrieveNotificationsResponse_NotificationInfo, 0, len(userNoti))
	for _, un := range userNoti {
		noti, has := notiMap[un.NotificationID.String]
		if !has {
			return nil, fmt.Errorf("expect find notification has id: %v", un.NotificationID.String)
		}
		notiMsg, ok := notiMsgMap[un.NotificationID.String]
		if !ok {
			return nil, fmt.Errorf("expect find message of notification id: %v", un.NotificationID.String)
		}
		title := notiMsg.Title
		content, err := notiMsg.GetContent()
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("notiMsg.GetContent cannot get content %v", err))
		}

		notificationResponse := &npb.RetrieveNotificationsResponse_NotificationInfo{
			Title:            title.String,
			Description:      content.GetText(),
			TargetId:         un.StudentID.String,
			IsImportant:      noti.IsImportant.Bool,
			QuestionnaireId:  noti.QuestionnaireID.String,
			SentAt:           timestamppb.New(noti.SentAt.Time),
			UserNotification: mappers.ToUserNotificationPb(un),
		}

		if notificationResponse.TargetId == "" {
			notificationResponse.TargetId = un.ParentID.String
		}

		notificationResponse.UserNotification.Type = cpb.NotificationType(cpb.NotificationType_value[noti.Type.String])
		notificationResponse.UserNotification.Data = string(noti.Data.Bytes)
		itemsPb = append(itemsPb, notificationResponse)
	}
	lastItem := userNoti[len(userNoti)-1]
	resp := &npb.RetrieveNotificationsResponse{
		Items: itemsPb,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetCombined{
				OffsetCombined: &cpb.Paging_Combined{
					OffsetTime:   timestamppb.New(lastItem.UpdatedAt.Time),
					OffsetString: lastItem.NotificationID.String,
				},
			},
		},
	}
	return resp, nil
}
