package mappers

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/manabie-com/backend/internal/notification/services/utils"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	npbv2 "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v2"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func Test_LegacyToNewProto(t *testing.T) {
	type testCase struct {
		msg       proto.Message
		converter interface{}
	}

	tcases := []testCase{
		{
			&bpb.RetrieveNotificationsRequest{},
			BobToNotiV1_RetrieveNotificationsRequest,
		},
		{
			&npb.RetrieveNotificationDetailResponse{},
			NotiV1ToBob_RetrieveNotificationDetailResponse,
		},
		{
			&bpb.RetrieveNotificationsRequest{},
			BobToNotiV1_RetrieveNotificationsRequest,
		},
		{
			&npb.RetrieveNotificationsResponse{},
			NotiV1ToBob_RetrieveNotificationsResponse,
		},
		{
			&bpb.GetAnswersByFilterRequest{},
			BobToNotiV1_GetAnswersByFilterRequest,
		},
		{
			&npb.GetAnswersByFilterResponse{},
			NotiV1ToBob_GetAnswersByFilterResponse,
		},
		{
			&bpb.CountUserNotificationRequest{},
			BobToNotiV1_CountUserNotificationRequest,
		},
		{
			&npb.CountUserNotificationResponse{},
			NotiV1ToBob_CountUserNotificationResponse,
		},
		{
			&bpb.SetUserNotificationStatusRequest{},
			BobToNotiV1_SetUserNotificationStatusRequest,
		},
		{
			&npb.SetUserNotificationStatusResponse{},
			NotiV1ToBob_SetUserNotificationStatusResponse,
		},
		{
			&ypb.SubmitQuestionnaireRequest{},
			YasuoToNotiV1_SubmitQuestionnaireRequest,
		},
		{
			&npb.SubmitQuestionnaireResponse{},
			NotiV1ToYasuo_SubmitQuestionnaireResponse,
		},
		{
			&ypb.SendScheduledNotificationRequest{},
			YasuoToNotiV1_SendScheduledNotificationRequest,
		},
		{
			&npb.SendScheduledNotificationResponse{},
			NotiV1ToYasuo_SendScheduledNotificationResponse,
		},
		{
			&ypb.SendNotificationRequest{},
			YasuoToNotiV1_SendNotificationRequest,
		},
		{
			&npb.SendNotificationResponse{},
			NotiV1ToYasuo_SendNotificationResponse,
		},
		{
			&ypb.NotifyUnreadUserRequest{},
			YasuoToNotiV1_NotifiUnreadUserRequest,
		},
		{
			&npb.NotifyUnreadUserResponse{},
			NotiV1ToYasuo_NotifiUnreadUserResponse,
		},
		{
			&ypb.DiscardNotificationRequest{},
			YasuoToNotiV1_DiscardNotificationRequest,
		},
		{
			&npb.DiscardNotificationResponse{},
			NotiV1ToYasuo_DiscardNotificationResponse,
		},
		{
			&ypb.UpsertNotificationRequest{},
			YasuoToNotiV1_UpsertNotificationRequest,
		},
		{
			&npb.UpsertNotificationResponse{},
			NotiV1ToYasuo_UpsertNotificationResponse,
		},
		{
			&bpb.RetrieveNotificationDetailRequest{},
			BobToNotiV1_RetrieveNotificationDetailRequest,
		},
	}
	rand := utils.NewProtoRand()
	for _, tcase := range tcases {
		randomMsg, err := rand.Gen(tcase.msg)
		assert.NoError(t, err)
		gotTo := reflect.ValueOf(tcase.converter).Call([]reflect.Value{reflect.ValueOf(randomMsg)})[0].Interface()
		protoGot := gotTo.(proto.Message)
		jsonExpect, err := json.Marshal(randomMsg)
		assert.NoError(t, err)
		jsonGot, err := json.Marshal(protoGot)
		assert.NoError(t, err)
		assert.JSONEq(t, string(jsonExpect), string(jsonGot))
	}
}

func Test_BobToNotiV1_UpdateUserDeviceTokenRequest(t *testing.T) {
	t.Parallel()
	t.Run("happy case", func(t *testing.T) {
		req := &pb.UpdateUserDeviceTokenRequest{}
		expcRes := &npb.UpdateUserDeviceTokenRequest{}
		res := BobToNotiV1_UpdateUserDeviceTokenRequest(req)
		jsonEpect, err := json.Marshal(expcRes)
		assert.NoError(t, err)
		jsonGot, err := json.Marshal(res)
		assert.NoError(t, err)
		assert.JSONEq(t, string(jsonEpect), string(jsonGot))
	})
}

func Test_NotiV1ToBob_UpdateUserDeviceTokenResponse(t *testing.T) {
	t.Parallel()
	t.Run("happy case", func(t *testing.T) {
		req := &npb.UpdateUserDeviceTokenResponse{}
		expcRes := &pb.UpdateUserDeviceTokenResponse{}
		res := NotiV1ToBob_UpdateUserDeviceTokenResponse(req)
		jsonEpect, err := json.Marshal(expcRes)
		assert.NoError(t, err)
		jsonGot, err := json.Marshal(res)
		assert.NoError(t, err)
		assert.JSONEq(t, string(jsonEpect), string(jsonGot))
	})
}

func Test_NotiV1ToV2_RetrieveNotificationDetailResponse(t *testing.T) {
	t.Parallel()
	t.Run("happy case", func(t *testing.T) {
		input := &npb.RetrieveNotificationDetailResponse{}
		expectRes := &npbv2.RetrieveNotificationDetailResponse{}
		res := NotiV1ToV2_RetrieveNotificationDetailResponse(input)
		jsonExpect, err := json.Marshal(expectRes)
		assert.NoError(t, err)
		jsonGot, err := json.Marshal(res)
		assert.NoError(t, err)
		assert.JSONEq(t, string(jsonExpect), string(jsonGot))
	})
}
