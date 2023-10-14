package tom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/tom/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"

	"github.com/jackc/pgtype"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

func (s *suite) aValidUserToken(ctx context.Context) (context.Context, error) {
	var err error
	if !s.CommonSuite.ContextHasToken(ctx) {
		ctx, err = s.CommonSuite.ASignedInWithSchool(ctx, "school admin", int32ResourcePathFromCtx(ctx))
		if err != nil {
			return ctx, err
		}
	}
	_, token, err := s.CommonSuite.CreateTeacher(ctx)
	if err != nil {
		return ctx, err
	}
	s.AuthToken = token
	return ctx, nil
}
func (s *suite) tomShouldThisConnectionMoreThanSeconds(ctx context.Context, action string, durationSec int) (context.Context, error) {
	t, _ := jwt.ParseString(s.AuthToken)
	streamClient, ok := s.StreamClients[t.Subject()]
	if !ok {
		return ctx, errors.New("user did not subscribe before")
	}

	// add more 2 seconds
	ctx2, cancel := context.WithTimeout(ctx, time.Duration(durationSec+2)*time.Second)
	defer cancel()

	errChan := make(chan error)

	go func() {
		for {
			_, err := streamClient.Recv()
			if err != nil {
				errChan <- err
				close(errChan)
				return
			}
		}
	}()

	var aliveDuration time.Duration
	select {
	case <-ctx2.Done():
	case <-errChan:
	}

	aliveDuration = time.Since(s.RequestAt)

	switch action {
	case "keep":
		if aliveDuration.Seconds() < float64(durationSec+1) {
			return ctx, fmt.Errorf("connection disconnect after %s, expected: %ds", aliveDuration, durationSec+1)
		}

		_ = streamClient.CloseSend()
	case "disconnect":
		if aliveDuration.Seconds()-float64(durationSec) > 1 {
			return ctx, fmt.Errorf("connection still alive after %s, expected disconnected after: %ds", aliveDuration, durationSec)
		}
	}

	return ctx, nil
}
func (s *suite) userSendPingEventToStreamEverySeconds(ctx context.Context, durationSec int) (context.Context, error) {
	t, _ := jwt.ParseString(s.AuthToken)
	streamClient, ok := s.StreamClients[t.Subject()]
	if !ok {
		return ctx, errors.New("user did not subscribe before")
	}

	go func() {
		for {
			err := streamClient.Send(&pb.StreamingEventRequest{
				Event: &pb.StreamingEventRequest_EventPing{
					EventPing: &pb.EventPing{},
				},
			})
			if err != nil {
				return
			}

			time.Sleep(time.Duration(durationSec) * time.Second)
		}
	}()

	return ctx, nil
}
func (s *suite) userSubscribeToEndpointStreamingEvent(ctx context.Context) (context.Context, error) {
	streamClient, err := pb.NewChatServiceClient(s.Conn).StreamingEvent(
		metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", s.AuthToken),
	)
	if err != nil {
		return ctx, err
	}
	s.RequestAt = time.Now()

	t, _ := jwt.ParseString(s.AuthToken)
	s.StreamClients[t.Subject()] = streamClient

	return ctx, nil
}
func (s *suite) userSendPingSubscribeV2ToStreamViaPingEndpointEverySeconds(ctx context.Context, durationSec int) (context.Context, error) {
	time.Sleep(time.Second)
	c := pb.NewChatServiceClient(s.Conn)

	t, _ := jwt.ParseString(s.AuthToken)
	r := &repositories.OnlineUserRepo{}
	_ = try.Do(func(attempt int) (retry bool, err error) {
		users, err := r.OnlineUserDBRepo.Find(ctx, s.DB, database.TextArray([]string{t.Subject()}), pgtype.Timestamptz{Time: time.Now().Add(-5 * time.Second), Status: 2})
		if err != nil {
			return false, err
		}

		if len(users) == 0 {
			time.Sleep(1 * time.Second)
			return attempt < 5, fmt.Errorf("not found user online")
		}

		return false, nil
	})

	streamV2, ok := s.SubV2Clients[t.Subject()]
	if !ok {
		return ctx, errors.New("user did not subscribe before")
	}

	sessionID := ""

	for try := 0; try < 4; try++ {
		resp, err := streamV2.Recv()
		if err != nil {
			return ctx, err
		}

		if resp.Event.GetEventPing() == nil {
			continue
		}
		sessionID = resp.Event.GetEventPing().SessionId
		break
	}

	token := s.AuthToken
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, err := c.PingSubscribeV2(
				s.signedTokenCtx(ctx, token), &pb.PingSubscribeV2Request{
					SessionId: sessionID,
				})
			if err != nil {
				s.ZapLogger.Error(fmt.Sprintf("err: %v", err))
			}

			time.Sleep(time.Duration(durationSec) * time.Second)
		}
	}()

	return ctx, nil
}
func (s *suite) userSubscribeToEndpointSubscribeV2(ctx context.Context) (context.Context, error) {
	ctx, cancel := context.WithCancel(ctx)
	err := try.Do(func(attempt int) (bool, error) {
		time.Sleep(1 * time.Second)

		sub, err := pb.NewChatServiceClient(s.Conn).SubscribeV2(
			contextWithToken(ctx, s.AuthToken),
			&pb.SubscribeV2Request{},
		)
		if err != nil {
			return attempt < 5, err
		}

		s.RequestAt = time.Now()

		t, _ := jwt.ParseString(s.AuthToken)
		s.SubV2Clients[t.Subject()] = cancellableStream{sub, cancel}

		return false, nil
	})

	if err != nil {
		cancel()
		return ctx, err
	}

	return ctx, nil
}
func (s *suite) tomShouldThisConnectionOfSubscribeV2MoreThanSeconds(ctx context.Context, action string, durationSec int) (context.Context, error) {
	t, _ := jwt.ParseString(s.AuthToken)
	streamClient, ok := s.SubV2Clients[t.Subject()]
	if !ok {
		return ctx, errors.New("user did not subscribe before")
	}

	// add more 2 seconds
	ctx2, cancel := context.WithTimeout(ctx, time.Duration(durationSec+20)*time.Second)
	defer cancel()

	errChan := make(chan error)

	go func() {
		for {
			_, err := streamClient.Recv()
			if err != nil {
				errChan <- err
				close(errChan)
				return
			}
		}
	}()

	var aliveDuration time.Duration
	select {
	case <-ctx2.Done():
	case <-errChan:
	}

	aliveDuration = time.Since(s.RequestAt)

	switch action {
	case "keep":
		if aliveDuration.Seconds() < float64(durationSec+1) {
			return ctx, fmt.Errorf("connection disconnect after %s, expected: %ds", aliveDuration, durationSec+1)
		}

		_ = streamClient.CloseSend()
	case "disconnect":
		if aliveDuration.Seconds()-float64(durationSec) > 1 {
			return ctx, fmt.Errorf("connection still alive after %s, expected disconnected after: %ds", aliveDuration, durationSec)
		}
	}

	return ctx, nil
}
