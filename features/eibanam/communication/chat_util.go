package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"google.golang.org/grpc/metadata"
)

type credential struct {
	password string
	token    string
}
type profile struct {
	email string
	name  string
	id    string
	credential
}

var (
	timeout = 10 * time.Second
)

// input type "1 of []" need to be cached, so later call can find back previous input
func (s *suite) loadFromCacheIfIsOneOfSyntax(key string, value string) string {
	if cached, ok := s.memo[key]; ok {
		return cached
	}
	// if match "1 of []" syntax
	choices := parseOneOf(value)
	if choices != nil {
		s.memo[key] = selectOneOf(value)
	} else {
		s.memo[key] = value
	}
	return s.memo[key]
}

var (
	newParent        = "new parent"
	existingParent   = "an existed parent"
	teacher          = "teacher"
	student          = "student"
	parent           = "parent"
	schoolAdmin      = "school admin"
	jprepSchoolAdmin = "jprep school admin"
)

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithChatHashKey(ctx context.Context, userID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-chat-userhash", userID)
}

func contextWithTokenAndTimeOut(ctx context.Context, token string) (context.Context, context.CancelFunc) {
	newCtx, cancel := context.WithTimeout(ctx, timeout)
	return contextWithToken(newCtx, token), cancel
}

func contextWithToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}

func (s *suite) loginsLearnerApp(ctx context.Context, person string) (context.Context, error) {
	return s.loginsLearnerAppSuccessfullyWithCredentialsWhichSchoolAdminGives(ctx, person)
}

func generateFakeAuthenticationToken(sub string, userGroup string) (string, error) {
	template := "templates/" + userGroup + ".template"
	resp, err := http.Get("http://" + firebaseAddr + "/token?template=" + template + "&UserID=" + sub)
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

func (s *suite) loginFirebaseAccount(ctx context.Context, email, password string) (string, error) {
	childCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", firebaseKey)

	loginInfo := struct {
		Email             string `json:"email"`
		Password          string `json:"password"`
		ReturnSecureToken bool   `json:"returnSecureToken"`
	}{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}
	body, err := json.Marshal(&loginInfo)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(childCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to login")
	}
	tokenContainer := struct {
		Token string `json:"idToken"`
	}{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&tokenContainer)
	if err != nil {
		return "", err
	}
	return tokenContainer.Token, nil
}
