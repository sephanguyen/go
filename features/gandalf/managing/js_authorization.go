package managing

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

func (s *suite) aUserWithUserInfo(ctx context.Context, userName, passWord string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	var opts []nats.Option
	opts = append(opts, nats.UserInfo(userName, passWord),
		nats.MaxReconnects(5),
		nats.ReconnectWait(5*time.Second))

	conn, err := nats.Connect(stepState.GandalfStateJetStreamAddress, opts...)
	if err != nil {
		return ctx, fmt.Errorf("failed to connect to NATS server: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		return ctx, fmt.Errorf("failed to get JetStream context: %w", err)
	}
	stepState.ZeusStepState.MapJSContext[userName] = js
	stepState.ZeusStepState.ListNatJSConnection = append(stepState.ZeusStepState.ListNatJSConnection, conn)
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) publishesAMessageWithSubject(ctx context.Context, userName, subjectName string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	jsm := stepState.ZeusStepState.MapJSContext[userName]
	_, err := jsm.Publish(subjectName, []byte("message demo"))
	stepState.ZeusStepState.MapPublishStatus[userName] = err
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) publishesMessageStatus(ctx context.Context, userName, status string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	defer s.closeAllJetStreamConnection(ctx)

	if status == "successfully" {
		return ctx, stepState.ZeusStepState.MapPublishStatus[userName]
	}
	err := stepState.ZeusStepState.MapPublishStatus[userName]
	if err == nil {
		return ctx, errors.New("expect publish fail, but the fact is successfully")
	}
	return ctx, nil
}

func (s *suite) subscribesAMessageWithSubject(ctx context.Context, userName, subjectName string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	jsm := stepState.ZeusStepState.MapJSContext[userName]
	sub, err := jsm.Subscribe(subjectName, func(msg *nats.Msg) {
		msg.Ack()
	}, nats.ManualAck())

	defer sub.Unsubscribe()

	stepState.ZeusStepState.MapSubscribeStatus[userName] = err
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) subscribesThisMessageSuccessfully(ctx context.Context, userName, status string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	defer s.closeAllJetStreamConnection(ctx)

	if status == "successfully" {
		return ctx, stepState.ZeusStepState.MapSubscribeStatus[userName]
	}
	err := stepState.ZeusStepState.MapSubscribeStatus[userName]
	if err == nil {
		return ctx, errors.New("expect subscribe failed, but the fact is successfully")
	}
	return ctx, nil
}

func (s *suite) closeAllJetStreamConnection(ctx context.Context) {
	stepState := GandalfStepStateFromContext(ctx)
	for _, v := range stepState.ZeusStepState.ListNatJSConnection {
		v.Close()
	}
}
