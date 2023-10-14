package firebase

import (
	"context"
	"fmt"

	"firebase.google.com/go/v4/messaging"
	"github.com/gogo/protobuf/types"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
)

const (
	fcmBatchLimit = 500
	ClickAction   = "FLUTTER_NOTIFICATION_CLICK"
)

type FCMClientV4 interface {
	Send(context.Context, *messaging.Message) (string, error)
	SendMulticast(context.Context, *messaging.MulticastMessage) (*messaging.BatchResponse, error)
}

type NotificationPusher interface {
	SendTokens(ctx context.Context, msg *messaging.MulticastMessage, tokens []string) (successCount, failureCount int, err *SendTokensError)
	SendToken(ctx context.Context, msg *messaging.Message, token string) error
	RetrievePushedMessages(ctx context.Context, deviceToken string, limit int, since *types.Timestamp) ([]*messaging.MulticastMessage, error)
}

type notificationPusherImpl struct {
	Client FCMClientV4
}

type SendTokensError struct {
	DirectError        error
	BatchCombinedError error
}

func NewNotificationPusher(client FCMClientV4) NotificationPusher {
	return &notificationPusherImpl{Client: client}
}

func (fcm *notificationPusherImpl) SendTokens(ctx context.Context, msg *messaging.MulticastMessage, tokens []string) (successCount, failureCount int, errRet *SendTokensError) {
	if len(tokens) == 0 {
		return
	}

	numOfTokens := len(tokens)
	eg, egCtx := errgroup.WithContext(ctx)
	batchRespChan := make(chan *messaging.BatchResponse)

	for len(tokens) > 0 {
		length := min(fcmBatchLimit, len(tokens))
		batchTokens := make([]string, length)
		copy(batchTokens, tokens[:length])

		tokens = tokens[length:]
		msg.Tokens = batchTokens

		eg.Go(func() error {
			resp, err := fcm.Client.SendMulticast(ctx, msg)
			if err != nil {
				return err
			}
			select {
			case batchRespChan <- resp:
				return nil
			case <-egCtx.Done():
				return egCtx.Err()
			}
		})
	}
	go func() {
		_ = eg.Wait()
		close(batchRespChan)
	}()

	errRet = &SendTokensError{}
	for resp := range batchRespChan {
		successCount += resp.SuccessCount
		failureCount += resp.FailureCount

		for _, respDetail := range resp.Responses {
			if respDetail.Error != nil {
				errRet.BatchCombinedError = multierr.Combine(errRet.BatchCombinedError, respDetail.Error)
			}
		}
	}

	if werr := eg.Wait(); werr != nil {
		errRet.DirectError = multierr.Combine(errRet.DirectError, fmt.Errorf("error when call fcm.Client.SendMulticast(): %w", werr))
		failureCount = numOfTokens - successCount
	}

	// For case all success -> error returned should be nil
	if errRet.DirectError == nil && errRet.BatchCombinedError == nil {
		errRet = nil
	}

	return
}

func (fcm *notificationPusherImpl) SendToken(ctx context.Context, msg *messaging.Message, token string) error {
	msg.Token = token
	_, err := fcm.Client.Send(ctx, msg)
	if err != nil && !messaging.IsRegistrationTokenNotRegistered(err) {
		return err
	}
	return nil
}

// nolint:revive
func (fcm *notificationPusherImpl) RetrievePushedMessages(ctx context.Context, deviceToken string, limit int, since *types.Timestamp) ([]*messaging.MulticastMessage, error) {
	// Not implemented yet.
	//
	// TBD: we can query the DB or firebase API to retrieve this information.
	return nil, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
