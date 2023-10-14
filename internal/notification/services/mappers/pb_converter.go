// nolint
package mappers

import (
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	npbv2 "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v2"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
)

// nolint
func BobToNotiV1_RetrieveNotificationDetailRequest(data *bpb.RetrieveNotificationDetailRequest) *npb.RetrieveNotificationDetailRequest {
	if data == nil {
		return &npb.RetrieveNotificationDetailRequest{}
	}

	return &npb.RetrieveNotificationDetailRequest{
		NotificationId: data.NotificationId,
		TargetId:       data.TargetId,
	}
}

func NotiV1ToBob_RetrieveNotificationDetailResponse(data *npb.RetrieveNotificationDetailResponse) *bpb.RetrieveNotificationDetailResponse {
	if data == nil {
		return &bpb.RetrieveNotificationDetailResponse{}
	}

	return &bpb.RetrieveNotificationDetailResponse{
		Item:              data.Item,
		UserNotification:  data.UserNotification,
		UserQuestionnaire: data.UserQuestionnaire,
	}
}

func BobToNotiV1_RetrieveNotificationsRequest(data *bpb.RetrieveNotificationsRequest) *npb.RetrieveNotificationsRequest {
	if data == nil {
		return &npb.RetrieveNotificationsRequest{}
	}

	return &npb.RetrieveNotificationsRequest{
		Paging:        data.Paging,
		ImportantOnly: data.ImportantOnly,
	}
}

func NotiV1ToBob_RetrieveNotificationsResponse(data *npb.RetrieveNotificationsResponse) *bpb.RetrieveNotificationsResponse {
	if data == nil {
		return &bpb.RetrieveNotificationsResponse{}
	}

	items := []*bpb.RetrieveNotificationsResponse_NotificationInfo{}
	for _, i := range data.Items {
		items = append(items, &bpb.RetrieveNotificationsResponse_NotificationInfo{
			Title:            i.Title,
			Description:      i.Description,
			UserNotification: i.UserNotification,
			SentAt:           i.SentAt,
			IsImportant:      i.IsImportant,
			QuestionnaireId:  i.QuestionnaireId,
			TargetId:         i.TargetId,
		})
	}
	return &bpb.RetrieveNotificationsResponse{
		NextPage: data.NextPage,
		Items:    items,
	}
}

func BobToNotiV1_GetAnswersByFilterRequest(data *bpb.GetAnswersByFilterRequest) *npb.GetAnswersByFilterRequest {
	if data == nil {
		return &npb.GetAnswersByFilterRequest{}
	}
	return &npb.GetAnswersByFilterRequest{
		QuestionnaireId: data.QuestionnaireId,
		Keyword:         data.Keyword,
		Paging:          data.Paging,
	}
}

func NotiV1ToBob_GetAnswersByFilterResponse(data *npb.GetAnswersByFilterResponse) *bpb.GetAnswersByFilterResponse {
	if data == nil {
		return &bpb.GetAnswersByFilterResponse{}
	}

	userAnswers := []*bpb.GetAnswersByFilterResponse_UserAnswer{}
	for _, i := range data.UserAnswers {
		userAnswers = append(userAnswers, &bpb.GetAnswersByFilterResponse_UserAnswer{
			ResponderName:      i.ResponderName,
			UserId:             i.UserId,
			TargetId:           i.TargetId,
			TargetName:         i.TargetName,
			IsParent:           i.IsParent,
			SubmittedAt:        i.SubmittedAt,
			Answers:            i.Answers,
			UserNotificationId: i.UserNotificationId,
			IsIndividual:       i.IsIndividual,
		})
	}
	return &bpb.GetAnswersByFilterResponse{
		UserAnswers:  userAnswers,
		TotalItems:   data.TotalItems,
		NextPage:     data.NextPage,
		PreviousPage: data.PreviousPage,
		Questions:    data.Questions,
	}
}

func BobToNotiV1_CountUserNotificationRequest(data *bpb.CountUserNotificationRequest) *npb.CountUserNotificationRequest {
	if data == nil {
		return &npb.CountUserNotificationRequest{}
	}

	return &npb.CountUserNotificationRequest{
		Status: data.Status,
	}
}

func NotiV1ToBob_CountUserNotificationResponse(data *npb.CountUserNotificationResponse) *bpb.CountUserNotificationResponse {
	if data == nil {
		return &bpb.CountUserNotificationResponse{}
	}

	return &bpb.CountUserNotificationResponse{
		NumByStatus: data.NumByStatus,
		Total:       data.Total,
	}
}

func BobToNotiV1_SetUserNotificationStatusRequest(data *bpb.SetUserNotificationStatusRequest) *npb.SetUserNotificationStatusRequest {
	if data == nil {
		return &npb.SetUserNotificationStatusRequest{}
	}

	return &npb.SetUserNotificationStatusRequest{
		NotificationIds: data.NotificationIds,
		Status:          data.Status,
	}
}

func NotiV1ToBob_SetUserNotificationStatusResponse(data *npb.SetUserNotificationStatusResponse) *bpb.SetUserNotificationStatusResponse {
	return &bpb.SetUserNotificationStatusResponse{}
}

func YasuoToNotiV1_UpsertNotificationRequest(data *ypb.UpsertNotificationRequest) *npb.UpsertNotificationRequest {
	if data == nil {
		return &npb.UpsertNotificationRequest{}
	}

	return &npb.UpsertNotificationRequest{
		Notification:  data.Notification,
		Questionnaire: data.Questionnaire,
	}
}

func NotiV1ToYasuo_UpsertNotificationResponse(data *npb.UpsertNotificationResponse) *ypb.UpsertNotificationResponse {
	if data == nil {
		return &ypb.UpsertNotificationResponse{}
	}

	return &ypb.UpsertNotificationResponse{
		NotificationId: data.NotificationId,
	}
}

func YasuoToNotiV1_SubmitQuestionnaireRequest(data *ypb.SubmitQuestionnaireRequest) *npb.SubmitQuestionnaireRequest {
	if data == nil {
		return &npb.SubmitQuestionnaireRequest{}
	}

	return &npb.SubmitQuestionnaireRequest{
		UserInfoNotificationId: data.UserInfoNotificationId,
		QuestionnaireId:        data.QuestionnaireId,
		Answers:                data.Answers,
	}
}

func NotiV1ToYasuo_SubmitQuestionnaireResponse(data *npb.SubmitQuestionnaireResponse) *ypb.SubmitQuestionnaireResponse {
	return &ypb.SubmitQuestionnaireResponse{}
}

func YasuoToNotiV1_SendScheduledNotificationRequest(data *ypb.SendScheduledNotificationRequest) *npb.SendScheduledNotificationRequest {
	if data == nil {
		return &npb.SendScheduledNotificationRequest{}
	}

	return &npb.SendScheduledNotificationRequest{
		OrganizationId:         data.OrganizationId,
		From:                   data.From,
		To:                     data.To,
		TenantIds:              data.TenantIds,
		IsRunningForAllTenants: data.IsRunningForAllTenants,
	}
}

func NotiV1ToYasuo_SendScheduledNotificationResponse(data *npb.SendScheduledNotificationResponse) *ypb.SendScheduledNotificationResponse {
	return &ypb.SendScheduledNotificationResponse{}
}

func YasuoToNotiV1_SendNotificationRequest(data *ypb.SendNotificationRequest) *npb.SendNotificationRequest {
	if data == nil {
		return &npb.SendNotificationRequest{}
	}

	return &npb.SendNotificationRequest{
		NotificationId: data.NotificationId,
	}
}

func NotiV1ToYasuo_SendNotificationResponse(data *npb.SendNotificationResponse) *ypb.SendNotificationResponse {
	return &ypb.SendNotificationResponse{}
}

func YasuoToNotiV1_NotifiUnreadUserRequest(data *ypb.NotifyUnreadUserRequest) *npb.NotifyUnreadUserRequest {
	if data == nil {
		return &npb.NotifyUnreadUserRequest{}
	}

	return &npb.NotifyUnreadUserRequest{
		NotificationId: data.NotificationId,
	}
}

func NotiV1ToYasuo_NotifiUnreadUserResponse(data *npb.NotifyUnreadUserResponse) *ypb.NotifyUnreadUserResponse {
	return &ypb.NotifyUnreadUserResponse{}
}

func YasuoToNotiV1_DiscardNotificationRequest(data *ypb.DiscardNotificationRequest) *npb.DiscardNotificationRequest {
	if data == nil {
		return &npb.DiscardNotificationRequest{}
	}

	return &npb.DiscardNotificationRequest{
		NotificationId: data.NotificationId,
	}
}

func NotiV1ToYasuo_DiscardNotificationResponse(data *npb.DiscardNotificationResponse) *ypb.DiscardNotificationResponse {
	return &ypb.DiscardNotificationResponse{}
}

func BobToNotiV1_UpdateUserDeviceTokenRequest(data *pb.UpdateUserDeviceTokenRequest) *npb.UpdateUserDeviceTokenRequest {
	return &npb.UpdateUserDeviceTokenRequest{
		UserId:            data.UserId,
		DeviceToken:       data.DeviceToken,
		AllowNotification: data.AllowNotification,
	}
}

func NotiV1ToBob_UpdateUserDeviceTokenResponse(data *npb.UpdateUserDeviceTokenResponse) *pb.UpdateUserDeviceTokenResponse {
	if data == nil {
		return &pb.UpdateUserDeviceTokenResponse{}
	}
	return &pb.UpdateUserDeviceTokenResponse{
		Successful: data.Successful,
	}
}

func NotiV1ToV2_RetrieveNotificationDetailResponse(data *npb.RetrieveNotificationDetailResponse) *npbv2.RetrieveNotificationDetailResponse {
	if data == nil {
		return &npbv2.RetrieveNotificationDetailResponse{}
	}
	return &npbv2.RetrieveNotificationDetailResponse{
		Item:              data.Item,
		UserNotification:  data.UserNotification,
		UserQuestionnaire: data.UserQuestionnaire,
	}
}
