package common

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/mock"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
)

func (s *NotificationSuite) UpdateDeviceTokenForLeanerUser(ctx context.Context, typeDeviceToken string) (context.Context, error) {
	commonState := StepStateFromContext(ctx)
	for _, student := range commonState.Students {
		studentToken, err := s.GenerateExchangeTokenCtx(ctx, student.ID, constant.UserGroupStudent)
		if err != nil {
			return ctx, fmt.Errorf("s.GenerateExchangeTokenCtx: %v", err)
		}

		deviceToken := mock.MockNotificationPusherValidDeviceToken + "-" + idutil.ULIDNow()
		if typeDeviceToken == "invalid" {
			deviceToken = mock.MockNotificationPusherInvalidDeviceToken + "-" + idutil.ULIDNow()
		} else if typeDeviceToken == "unexpected" {
			deviceToken = mock.MockNotificationPusherDeviceTokenWithUnexpectedError + "-" + idutil.ULIDNow()
		}

		err = s.UpdateDeviceToken(studentToken, deviceToken, student.ID)
		if err != nil {
			return ctx, fmt.Errorf("s.UpdateDeviceToken: %v", err)
		}
		student.DeviceToken = deviceToken

		for _, parent := range student.Parents {
			parentToken, err := s.GenerateExchangeTokenCtx(ctx, parent.ID, constant.UserGroupParent)
			if err != nil {
				return ctx, fmt.Errorf("s.GenerateExchangeTokenCtx: %v", err)
			}

			deviceToken := mock.MockNotificationPusherValidDeviceToken + "-" + idutil.ULIDNow()
			if typeDeviceToken == "invalid" {
				deviceToken = mock.MockNotificationPusherInvalidDeviceToken + "-" + idutil.ULIDNow()
			} else if typeDeviceToken == "unexpected" {
				deviceToken = mock.MockNotificationPusherDeviceTokenWithUnexpectedError + "-" + idutil.ULIDNow()
			}
			err = s.UpdateDeviceToken(parentToken, deviceToken, parent.ID)
			if err != nil {
				return ctx, fmt.Errorf("s.UpdateDeviceToken: %v", err)
			}
			parent.DeviceToken = deviceToken
		}
	}

	return ctx, nil
}

// This method is to simulate scenario with % failures in pushing FCM
func (s *NotificationSuite) UpdateDeviceTokenForLeanerUserWithPercentageFailure(ctx context.Context, typeDeviceToken, failRate string) (context.Context, error) {
	commonState := StepStateFromContext(ctx)
	percentage := StrToInt(failRate)

	totalNumberUsers := len(commonState.Students)
	for _, student := range commonState.Students {
		totalNumberUsers += len(student.Parents)
	}

	failureRate := float64(totalNumberUsers) * float64(percentage) / 100

	countInvalidToken := int(failureRate)
	for _, student := range commonState.Students {
		studentToken, err := s.GenerateExchangeTokenCtx(ctx, student.ID, constant.UserGroupStudent)
		if err != nil {
			return ctx, fmt.Errorf("s.GenerateExchangeTokenCtx: %v", err)
		}

		deviceToken := mock.MockNotificationPusherValidDeviceToken + "-" + idutil.ULIDNow()
		if typeDeviceToken == "invalid" && countInvalidToken > 0 {
			deviceToken = mock.MockNotificationPusherInvalidDeviceToken + "-" + idutil.ULIDNow()
			countInvalidToken--
		} else if typeDeviceToken == "unexpected" {
			deviceToken = mock.MockNotificationPusherDeviceTokenWithUnexpectedError + "-" + idutil.ULIDNow()
		}

		err = s.UpdateDeviceToken(studentToken, deviceToken, student.ID)
		if err != nil {
			return ctx, fmt.Errorf("s.UpdateDeviceToken: %v", err)
		}
		student.DeviceToken = deviceToken

		for _, parent := range student.Parents {
			parentToken, err := s.GenerateExchangeTokenCtx(ctx, parent.ID, constant.UserGroupParent)
			if err != nil {
				return ctx, fmt.Errorf("s.GenerateExchangeTokenCtx: %v", err)
			}

			deviceToken := mock.MockNotificationPusherValidDeviceToken + "-" + idutil.ULIDNow()
			if typeDeviceToken == "invalid" && countInvalidToken > 0 {
				deviceToken = mock.MockNotificationPusherInvalidDeviceToken + "-" + idutil.ULIDNow()
				countInvalidToken--
			} else if typeDeviceToken == "unexpected" {
				deviceToken = mock.MockNotificationPusherDeviceTokenWithUnexpectedError + "-" + idutil.ULIDNow()
			}
			err = s.UpdateDeviceToken(parentToken, deviceToken, parent.ID)
			if err != nil {
				return ctx, fmt.Errorf("s.UpdateDeviceToken: %v", err)
			}
			parent.DeviceToken = deviceToken
		}
	}

	return StepStateToContext(ctx, commonState), nil
}
