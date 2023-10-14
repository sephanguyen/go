package communication

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/j4/infras"
	serviceutil "github.com/manabie-com/backend/j4/serviceutil"
	commuserviceutil "github.com/manabie-com/backend/j4/serviceutil/communication"
	legacytpb "github.com/manabie-com/backend/pkg/genproto/tom"
	j4 "github.com/manabie-com/j4/pkg/runner"

	grpcpool "github.com/processout/grpc-go-pool"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type SpamMsg struct {
	tokenGenerator *serviceutil.TokenGenerator
	logger         *zap.SugaredLogger
	j4cfg          *infras.ManabieJ4Config
	conns          *infras.Connections
}
type ChatSvc interface {
	PingSubscribeV2(ctx context.Context, in *legacytpb.PingSubscribeV2Request, opts ...grpc.CallOption) (*legacytpb.PingSubscribeV2Response, error)
	SubscribeV2(ctx context.Context, in *legacytpb.SubscribeV2Request, opts ...grpc.CallOption) (legacytpb.ChatService_SubscribeV2Client, error)
	SendMessage(ctx context.Context, in *legacytpb.SendMessageRequest, opts ...grpc.CallOption) (*legacytpb.SendMessageResponse, error)
}

func ScenarioIntializer(ctx context.Context, c *infras.ManabieJ4Config, dep *infras.Dep) ([]*j4.Scenario, error) {
	ctx = interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: c.SchoolID,
			UserID:       c.AdminID,
		},
	})

	tokenGenerator := serviceutil.NewTokenGenerator(c, dep.Connections)

	s := &SpamMsg{
		tokenGenerator: tokenGenerator,
		logger:         logger.NewZapLogger("debug", false).Sugar(),
		conns:          dep.Connections,
		j4cfg:          c,
	}
	spamMsg, err := s.setupMsgSpam(ctx, c)
	if err != nil {
		return nil, err
	}

	// filterNoti, err := setupNoti(ctx, c, dep)
	// if err != nil {
	// 	return nil, err
	// }
	return []*j4.Scenario{
		spamMsg,
		// filterNoti,
	}, nil
}

func (s *SpamMsg) resetToken(ctx context.Context, stu commuserviceutil.StudentConvo) (string, error) {
	return s.tokenGenerator.GetTokenFromShamir(ctx, stu.UserID, s.j4cfg.SchoolID)
}

func (s *SpamMsg) setupMsgSpam(ctx context.Context, c *infras.ManabieJ4Config) (*j4.Scenario, error) {
	studentGenerator := commuserviceutil.InitStudentConvoPool(ctx, c, s.conns)
	testFuncClojure := func(childCtx context.Context) j4.TestFunc {
		stu := studentGenerator.GetOne(ctx)
		if stu == nil {
			return nil
		}
		tickChan := make(chan struct{}, 1)
		s.makeUserSendReceiveMsg(childCtx, *stu, tickChan)
		return func(_ context.Context) error {
			tickChan <- struct{}{}
			return nil
		}
	}

	// generating student is expensive, so we do it in a background goroutine by internal
	runConfig, err := c.GetScenarioConfig("spam_msg")
	if err != nil {
		return nil, err
	}
	runCfg := infras.MustOptionFromConfig(&runConfig)
	runCfg.TestClosure = testFuncClojure

	// TODO:
	// - suffix each scenario with an id
	// - Allow user to control scenario config using github param
	return j4.NewScenario("spam_msg", *runCfg)
}

// get or context cancel
func (s *SpamMsg) getGrpcConnFromPool(ctx context.Context) *grpcpool.ClientConn {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		ret, err := s.conns.PoolToGateWay.Get(ctx)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		return ret
	}
}

func (s *SpamMsg) makeUserSendReceiveMsg(ctx context.Context, stu commuserviceutil.StudentConvo, tick chan struct{}) {
	userID := stu.UserID
	convID := stu.ConvID

	dedicatedConn := s.getGrpcConnFromPool(ctx)
	if dedicatedConn == nil {
		panic("todo")
	}
	defer dedicatedConn.Close()
	chatSvc := legacytpb.NewChatServiceClient(dedicatedConn.ClientConn)

	reconnect := func(ctx2 context.Context) (string, string, legacytpb.ChatService_SubscribeV2Client, error) {
		ctx2 = metadata.AppendToOutgoingContext(ctx2, "x-chat-userhash", userID)
		var tok string
	tryResetToken:
		for {
			select {
			case <-ctx2.Done():
				// let the caller cancel
				return "", "", nil, fmt.Errorf("ctx canceled")
			default:
				newTok, err := s.resetToken(ctx2, stu)
				if err != nil {
					s.logger.Warnf("failed to generate token %s for user %s, %s", err, userID, stu.UserID)
					time.Sleep(2 * time.Second)
					continue
				}
				tok = newTok
				break tryResetToken
			}
		}

		ctx2 = contextWithToken(ctx2, tok)
		streamV2, err := chatSvc.SubscribeV2(ctx2, &legacytpb.SubscribeV2Request{})
		if err != nil {
			return "", "", nil, err
		}

		sessionID := ""
		for try := 0; try < 10; try++ {
			resp, err := streamV2.Recv()
			if err != nil {
				return "", "", nil, fmt.Errorf("streamV2.Recv %w", err)
			}

			if resp.Event.GetEventPing() == nil {
				continue
			}
			sessionID = resp.Event.GetEventPing().SessionId
			break
		}

		return sessionID, tok, streamV2, nil
	}

	go func() {
		count := 0
		for {
			ctx2, cancel := context.WithCancel(ctx)
			select {
			case <-ctx.Done():
				cancel()
				return
			default:
			}
			count++
			sessionID, token, stream, err := reconnect(ctx2)
			fmt.Printf("reconnecting====%d\n", count)
			if err != nil {
				s.logger.Warnf("failed to reconnect client connection %s", err)
				time.Sleep(2 * time.Second)
				continue
			}
			errg := errgroup.Group{}

			ctx3 := contextWithToken(ctx2, token)
			errg.Go(func() error {
				for {
					select {
					case <-ctx2.Done():
						return nil
					default:
						_, err := chatSvc.PingSubscribeV2(ctx3, &legacytpb.PingSubscribeV2Request{SessionId: sessionID})
						if err != nil {
							s.logger.Warnf("failed to ping using subscribe v2 api: %s, reconnecting\n", err)
							cancel()
							return err
						}
						time.Sleep(3 * time.Second)
					}
				}
			})
			errg.Go(func() error {
				for {
					select {
					case <-ctx.Done():
						return nil
					default:
					}
					_, err := stream.Recv()
					if err != nil {
						return err
					}
				}
			})
			errg.Go(func() error {
				for {
					select {
					case <-ctx.Done():
						return nil
					case <-tick:
					}
					_, err := chatSvc.SendMessage(ctx3, &legacytpb.SendMessageRequest{
						ConversationId: convID,
						Message:        "DDOS-ing Tom",
						Type:           legacytpb.MESSAGE_TYPE_TEXT,
					})
					if err != nil {
						s.logger.Warnf("SendMessage %s", err)
						cancel()
						return err
					}
				}
			})
			errg.Wait()
			err = stream.CloseSend()
			if err != nil {
				s.logger.Warnf("failed to close connection %s", err)
			}
		}
	}()
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(ctx context.Context, token string) context.Context {
	ctx = contextWithValidVersion(ctx)
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}
