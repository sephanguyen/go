package learning

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/jackc/pgtype"
	"github.com/lestrrat-go/jwx/jwt"
	"google.golang.org/grpc/metadata"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/tom/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
)

type genAuthTokenOption func(values url.Values)

func generateValidAuthenticationToken(sub string) (string, error) {
	return generateAuthenticationToken(sub, "templates/phone.template")
}

func generateAuthenticationToken(sub, template string, opts ...genAuthTokenOption) (string, error) {
	v := url.Values{}
	v.Set("template", template)
	v.Set("UserID", sub)
	for _, opt := range opts {
		opt(v)
	}
	resp, err := http.Get("http://" + firebaseAddr + "/token?" + v.Encode())
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken:cannot generate new user token, err: %v", err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()
	return string(b), nil
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(ctx context.Context, token string) context.Context {
	ctx = contextWithValidVersion(ctx)
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}

func generateNumber(n int) string {
	codeLetters := "0123456789"
	b := make([]byte, n)
	len := len(codeLetters)
	for i := range b {
		b[i] = codeLetters[rand.Intn(len)]
	}
	return "0" + string(b)
}

func generateText(n int, isEmail bool) string {
	codeLetters := "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	len := len(codeLetters)
	for i := range b {
		b[i] = codeLetters[rand.Intn(len)]
	}
	if isEmail {
		return string(b) + "gmail.com"
	}
	return string(b)
}

func (s *suite) userSendPingEventToStreamEverySeconds(ctx context.Context, token string, durationSec int) (context.Context, error) {
	c := pb.NewChatServiceClient(s.tomConn)
	t, _ := jwt.ParseString(token)
	r := &repositories.OnlineUserRepo{}
	err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(time.Millisecond * 500)
		users, err := r.OnlineUserDBRepo.Find(context.Background(), s.tomDB, database.TextArray([]string{t.Subject()}),
			pgtype.Timestamptz{Time: time.Now().Add(-5 * time.Second), Status: 2})
		if err != nil {
			return false, err
		}
		if len(users) == 0 {
			return true, fmt.Errorf("not found user online")
		}
		return false, nil
	})
	if err != nil {
		return ctx, fmt.Errorf("OnlineUserDBRepo.Find: %s", err)
	}

	streamV2, ok := s.SubV2Clients[t.Subject()]
	if !ok {
		return ctx, errors.New("user did not subscribe before")
	}

	sessionID := ""
	resp, err := streamV2.Recv()
	if err != nil {
		return ctx, err

	}

	if resp.Event.GetEventPing() == nil {
		return ctx, fmt.Errorf("stream must receive pingEvent first")
	}

	sessionID = resp.Event.GetEventPing().SessionId

	go func() {
		for {
			_, err := c.PingSubscribeV2(
				metadata.AppendToOutgoingContext(metadata.AppendToOutgoingContext(contextWithValidVersion(context.Background()), "token", token)),
				&pb.PingSubscribeV2Request{
					SessionId: sessionID,
				})
			if err != nil {
				panic(err)
			}

			time.Sleep(time.Duration(durationSec) * time.Second)
		}
	}()

	return ctx, nil
}
