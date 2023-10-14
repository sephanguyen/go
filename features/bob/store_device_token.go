package bob

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) userTryToStoreDeviceToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.createUserDeviceTokenCreatedSubscription(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createUserDeviceTokenCreatedSubscription: %v", err)
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewUserServiceClient(s.Conn).UpdateUserDeviceToken(s.signedCtx(ctx), stepState.Request.(*pb.UpdateUserDeviceTokenRequest))
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when call pb.NewUserServiceClient.UpdateUserDeviceToken: %w", stepState.ResponseErr)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) UserTryToStoreDeviceToken(ctx context.Context) (context.Context, error) {
	return s.userTryToStoreDeviceToken(ctx)
}
func (s *suite) aValidDeviceToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	userID := t.Subject()
	stepState.Request = &pb.UpdateUserDeviceTokenRequest{
		UserId:            userID,
		DeviceToken:       "some device token",
		AllowNotification: true,
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) AValidDeviceToken(ctx context.Context) (context.Context, error) {
	return s.aValidDeviceToken(ctx)
}
func (s *suite) bobMustStoreTheUsersDeviceToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	query := "SELECT COUNT(*) from users where  user_id=$1 and device_token =$2 and allow_notification= $3"
	var userID, deviceToken pgtype.Text
	var allowNotification pgtype.Bool
	req := stepState.Request.(*pb.UpdateUserDeviceTokenRequest)
	userID.Set(req.UserId)
	deviceToken.Set(req.DeviceToken)
	allowNotification.Set(req.AllowNotification)
	row := s.DB.QueryRow(ctx, query, &userID, &deviceToken, &allowNotification)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	if count < 1 {
		return StepStateToContext(ctx, stepState), errors.New("bob doesn't store user device token")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) BobMustStoreTheUsersDeviceToken(ctx context.Context) (context.Context, error) {
	return s.bobMustStoreTheUsersDeviceToken(ctx)
}

func (s *suite) createUserDeviceTokenCreatedSubscription(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{nats.StartTime(time.Now()), nats.AckWait(2 * time.Second), nats.AckExplicit()},
	}

	stepState.FoundChanForJetStream = make(chan interface{}, 1)

	handleUserDeviceTokenUpdatedEvent := func(ctx context.Context, data []byte) (bool, error) {
		msg := &pb.EvtUserInfo{}
		err := msg.Unmarshal(data)
		if err != nil {
			return true, err
		}

		switch req := stepState.Request.(type) {
		case *pb.UpdateUserProfileRequest:
			if req.Profile.Name == msg.Name {
				stepState.FoundChanForJetStream <- stepState.Request
				return false, nil
			}
		case *pb.UpdateProfileRequest:
			if req.Name == msg.Name {
				stepState.FoundChanForJetStream <- stepState.Request
				return false, nil
			}
		case *pb.UpdateUserDeviceTokenRequest:
			if req.UserId == msg.UserId && req.DeviceToken == msg.DeviceToken {
				stepState.FoundChanForJetStream <- stepState.Request
				return false, nil
			}
		}
		return false, nil
	}

	sub, err := s.JSM.Subscribe(constants.SubjectUserDeviceTokenUpdated, opts, handleUserDeviceTokenUpdatedEvent)
	if err != nil {
		return fmt.Errorf("S.JSM.Subscribe: %v", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return nil
}

func (s *suite) BobMustPublishEventToUserDeviceTokenChannel(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(5 * time.Second)

	timer := time.NewTimer(time.Minute)
	defer timer.Stop()

	select {
	case <-stepState.FoundChanForJetStream:
		return StepStateToContext(ctx, stepState), nil
	case <-timer.C:
		return StepStateToContext(ctx, stepState), errors.New("time out")
	}
}
func (s *suite) BobMustPublishEventToUser_device_tokenChannel(ctx context.Context) (context.Context, error) {
	return s.BobMustPublishEventToUserDeviceTokenChannel(ctx)
}
func (s *suite) aDeviceTokenWithDevice_tokenEmpty(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.UpdateUserDeviceTokenRequest{
		UserId:            "",
		DeviceToken:       "",
		AllowNotification: true,
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) bobMustStoreTheUsersDeviceTokenToUser_device_tokensTable(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	query := "SELECT COUNT(*) from user_device_tokens where user_id=$1 and device_token=$2 and allow_notification=$3"
	var userID, deviceToken pgtype.Text
	var allowNotification pgtype.Bool
	req := stepState.Request.(*pb.UpdateUserDeviceTokenRequest)
	userID.Set(req.UserId)
	deviceToken.Set(req.DeviceToken)
	allowNotification.Set(req.AllowNotification)
	row := s.DB.QueryRow(ctx, query, &userID, &deviceToken, &allowNotification)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count < 1 {
		return StepStateToContext(ctx, stepState), errors.New("bob doesn't store user device token to table user_device_tokens")
	}
	return StepStateToContext(ctx, stepState), nil
}
