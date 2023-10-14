package stress

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/tom/configurations"
	legacytpb "github.com/manabie-com/backend/pkg/genproto/tom"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// TO gen mock only
type GrpcClient interface {
	ExchangeToken(ctx context.Context, in *bpb.ExchangeTokenRequest, opts ...grpc.CallOption) (*bpb.ExchangeTokenResponse, error)
	CreateStudent(ctx context.Context, in *upb.CreateStudentRequest, opts ...grpc.CallOption) (*upb.CreateStudentResponse, error)
	PingSubscribeV2(ctx context.Context, in *legacytpb.PingSubscribeV2Request, opts ...grpc.CallOption) (*legacytpb.PingSubscribeV2Response, error)
	SubscribeV2(ctx context.Context, in *legacytpb.SubscribeV2Request, opts ...grpc.CallOption) (legacytpb.ChatService_SubscribeV2Client, error)
	SendMessage(ctx context.Context, in *legacytpb.SendMessageRequest, opts ...grpc.CallOption) (*legacytpb.SendMessageResponse, error)
}
type ClientStream interface {
	legacytpb.ChatService_SubscribeV2Client
}

type StagingStress struct {
	logger *zap.SugaredLogger
	// deprecated
	runtimeConfig *StagingStressConfig
	// conn            *grpc.ClientConn
	userModifierSvc interface {
		CreateStudent(ctx context.Context, in *upb.CreateStudentRequest, opts ...grpc.CallOption) (*upb.CreateStudentResponse, error)
	}

	bobSvc interface {
		ExchangeToken(ctx context.Context, in *bpb.ExchangeTokenRequest, opts ...grpc.CallOption) (*bpb.ExchangeTokenResponse, error)
	}
	chatSvc interface {
		PingSubscribeV2(ctx context.Context, in *legacytpb.PingSubscribeV2Request, opts ...grpc.CallOption) (*legacytpb.PingSubscribeV2Response, error)
		SubscribeV2(ctx context.Context, in *legacytpb.SubscribeV2Request, opts ...grpc.CallOption) (legacytpb.ChatService_SubscribeV2Client, error)
		SendMessage(ctx context.Context, in *legacytpb.SendMessageRequest, opts ...grpc.CallOption) (*legacytpb.SendMessageResponse, error)
	}
}
type Student struct {
	UserID   string
	Email    string
	Password string
	Token    string
}

type StagingStressConfig struct {
	FirebaseAPIKey             string
	TotalStudent               int
	ConPerUser                 int
	SchoolID                   int
	AdminEmail                 string
	AdminPassword              string
	Duration                   string
	Addr                       string
	FirebaseIdentityToolkitURL string
}

func BindCobra(f *pflag.FlagSet, conf *StagingStressConfig) {
	f.StringVar(&conf.AdminEmail, "adminEmail", "", "email of stress test school admin to create test user")
	f.StringVar(&conf.AdminPassword, "adminPassword", "", "password of stress test school admin")
	f.StringVar(&conf.FirebaseAPIKey, "apiKey", "", "staging api key")
	f.IntVar(&conf.SchoolID, "schoolID", 0, "id of the school used to stress test")
	f.IntVar(&conf.ConPerUser, "conPerUser", 0, "connection per user")
	f.IntVar(&conf.TotalStudent, "totalStudent", 0, "number of student created")
	f.StringVar(&conf.Duration, "duration", "1h", "test for how long?")
	f.StringVar(&conf.Addr, "addr", "api.staging.manabie.io:443", "addr of env")
	conf.FirebaseIdentityToolkitURL = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword"
}

func RunStagingStressTest(ctx context.Context, c *configurations.Config, runTimeConf *StagingStressConfig) {
	timeDur, err := time.ParseDuration(runTimeConf.Duration)
	if err != nil {
		panic(err)
	}
	ctx2, cancel := context.WithTimeout(ctx, timeDur)
	defer cancel()
	s := &StagingStress{
		runtimeConfig: runTimeConf,
		logger:        logger.NewZapLogger(c.Common.Log.ApplicationLevel, false).Sugar(),
	}
	opts := []grpc.DialOption{grpc.WithBlock()}
	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})))

	conn, err := grpc.DialContext(ctx, s.runtimeConfig.Addr, opts...)
	if err != nil {
		s.logger.Fatalf("connection grpc %s", err)
	}
	s.bobSvc = bpb.NewUserModifierServiceClient(conn)
	s.chatSvc = legacytpb.NewChatServiceClient(conn)
	s.userModifierSvc = upb.NewUserModifierServiceClient(conn)
	s.StartPingStressTest(ctx2, c)
}

func (s *StagingStress) StartPingStressTest(ctx context.Context, c *configurations.Config) {
	tok, err := s.GetManabieToken(ctx, s.runtimeConfig.AdminEmail, s.runtimeConfig.AdminPassword)
	if err != nil {
		s.logger.Fatalf("failed to get token for admin: %s", err)
	}
	var createdStudent int
	var continuousErr int
	for createdStudent < s.runtimeConfig.TotalStudent {
		if continuousErr >= 10 {
			s.logger.Fatalf("too much continuous error")
		}
		stuprof, err := s.CreateStudent(ctx, tok)
		if err != nil {
			time.Sleep(2 * time.Second)
			s.logger.Warnf("error creating stress test user: %s", err)
			continuousErr++
			continue
		}

		resetTokFunc := func() (string, error) {
			studenTok, err := s.GetManabieToken(ctx, stuprof.Student.UserProfile.Email, stuprof.StudentPassword)
			if err != nil {
				return "", err
			}
			return studenTok, err
		}
		go s.makeUserPing(ctx, resetTokFunc, stuprof.GetStudent().UserProfile.UserId, s.runtimeConfig.ConPerUser)
		continuousErr = 0
		createdStudent++
	}

	s.logger.Infof("created %d students, pinging until ctx cancelled", s.runtimeConfig.TotalStudent)
	<-ctx.Done()
}

// Ping until ctx is cancelled
func (s *StagingStress) makeUserPing(ctx context.Context, resetToken func() (string, error), id string, connPerUser int) {
	reconnect := func() (string, string, legacytpb.ChatService_SubscribeV2Client, error) {
		// chatSvc := legacytpb.NewChatServiceClient(s.conn)
		ctx = metadata.AppendToOutgoingContext(ctx, "x-chat-userhash", id)
		var tok string
	tryResetToken:
		for {
			select {
			case <-ctx.Done():
				// let the caller cancel
				return "", "", nil, fmt.Errorf("ctx canceled")
			default:
				newTok, err := resetToken()
				if err != nil {
					s.logger.Warnf("faled to generate token %s", err)
					time.Sleep(2 * time.Second)
				}
				tok = newTok
				break tryResetToken
			}
		}

		ctx = contextWithToken(ctx, tok)
		streamV2, err := s.chatSvc.SubscribeV2(ctx, &legacytpb.SubscribeV2Request{})
		if err != nil {
			return "", "", nil, err
		}

		sessionID := ""
		resp, err := streamV2.Recv()
		if err != nil {
			return "", "", nil, err
		}

		if resp.Event.GetEventPing() == nil {
			return "", "", nil, fmt.Errorf("stream must receive pingEvent first")
		}
		sessionID = resp.Event.GetEventPing().SessionId
		return sessionID, tok, streamV2, nil
	}

	// chatSvc := legacytpb.NewChatServiceClient(s.conn)
	for i := 0; i < connPerUser; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					break
				}
				sessionID, token, stream, err := reconnect()
				if err != nil {
					s.logger.Warnf("faled to reconnect client connection %s", err)
					time.Sleep(2 * time.Second)
					continue
				}
			pingloop:
				for {
					select {
					case <-ctx.Done():
						err = stream.CloseSend()
						if err != nil {
							s.logger.Warnf("faled to close connection %s", err)
						}
						return
					default:
						ctx, cancel := contextWithTokenAndTimeOut(context.Background(), token)
						defer cancel()
						_, err := s.chatSvc.PingSubscribeV2(ctx, &legacytpb.PingSubscribeV2Request{SessionId: sessionID})
						if err != nil {
							s.logger.Warnf("failed to ping using subscribe v2 api: %s, reconnecting\n", err)
							err = stream.CloseSend()
							if err != nil {
								s.logger.Warnf("faled to close connection %s", err)
							}
							break pingloop
						}
						time.Sleep(1 * time.Second)
					}
				}
			}
		}()
	}
}

func contextWithTokenAndTimeOut(ctx context.Context, token string) (context.Context, context.CancelFunc) {
	newCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	return contextWithToken(newCtx, token), cancel
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(ctx context.Context, token string) context.Context {
	ctx = contextWithValidVersion(ctx)
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}

func (s *StagingStress) CreateStudent(ctx context.Context, token string) (*upb.CreateStudentResponse_StudentProfile, error) {
	randomID := idutil.ULIDNow()
	password := fmt.Sprintf("password-%v", randomID)
	email := fmt.Sprintf("commu-stresstest-%v@example.com", randomID)
	name := fmt.Sprintf("commu-stresstest-%v", randomID)
	req := &upb.CreateStudentRequest{
		SchoolId: int32(s.runtimeConfig.SchoolID),
		StudentProfile: &upb.CreateStudentRequest_StudentProfile{
			Email:            email,
			Password:         password,
			Name:             name,
			CountryCode:      cpb.Country_COUNTRY_VN,
			PhoneNumber:      fmt.Sprintf("phone-number-%v", randomID),
			Grade:            5,
			EnrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			LocationIds:      nil,
		},
	}

	res, err := s.userModifierSvc.CreateStudent(contextWithToken(ctx, token), req)
	if err != nil {
		return nil, err
	}
	return res.StudentProfile, nil
}

func (s *StagingStress) GetManabieToken(ctx context.Context, email, password string) (string, error) {
	firebaseTok, _, err := s.ExchangeFirebaseToken(email, password, s.runtimeConfig.FirebaseAPIKey)
	if err != nil {
		return "", err
	}

	rsp, err := s.bobSvc.ExchangeToken(contextWithValidVersion(ctx), &bpb.ExchangeTokenRequest{
		Token: firebaseTok,
	})
	if err != nil {
		return "", err
	}
	return rsp.Token, nil
}

func (s *StagingStress) ExchangeFirebaseToken(email, password string, apiKey string) (string, string, error) {
	childCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s?key=%s", s.runtimeConfig.FirebaseIdentityToolkitURL, apiKey)

	loginInfo := struct {
		Email             string `json:"email"`
		Password          string `json:"password"`
		ReturnSecureToken bool   `json:"returnSecureToken"`
	}{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}

	body, _ := json.Marshal(&loginInfo)

	req, err := http.NewRequestWithContext(childCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", errors.New("failed to login")
	}

	tokenContainer := struct {
		Token string `json:"idToken"`
		ID    string `json:"localId"`
	}{}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&tokenContainer)

	if err != nil {
		return "", "", err
	}

	return tokenContainer.Token, tokenContainer.ID, nil
}
