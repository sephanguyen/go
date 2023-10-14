package stresstest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type UserSignInPayload struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

type UserSignInResponse struct {
	Kind         string `json:"kind"`
	LocalId      string `json:"localId"`
	Email        string `json:"email"`
	DisplayName  string `json:"displayName"`
	IdToken      string `json:"idToken"`
	Registered   bool   `json:"registered"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
}

// UserSignInWithPassword will call Google Cloud Identity Platform
// to sign in with user and password, and we will receive some info and tokens
func (s *StressTest) UserSignInWithPassword(ctx context.Context, acc *AccountInfo, returnSecureToken bool) (*UserSignInResponse, error) {
	payload := &UserSignInPayload{
		Email:             acc.Email,
		Password:          acc.Password,
		ReturnSecureToken: returnSecureToken,
	}
	jsonReq, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/accounts:signInWithPassword?key=%s", s.cfg.IdentityToolkitAPI, s.cfg.FirebaseAPIKey)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonReq))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	request = request.WithContext(ctx)
	resp, err := s.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("the HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected status 200 but got %d", resp.StatusCode)
	}

	var response UserSignInResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// ExchangeUserToken will fetch manabie token to auth calling to backend api
func (s *StressTest) ExchangeUserToken(ctx context.Context, idToken string) (string, error) {
	res, err := bpb.NewUserModifierServiceClient(s.connection.BobConn).
		ExchangeToken(ctx, &bpb.ExchangeTokenRequest{
			Token: idToken,
		})
	if err != nil {
		return "", fmt.Errorf("UserModifierService.ExchangeToken: %w", err)
	}

	return res.Token, nil
}

func SimplifiedDial(host string, insecureParam bool) (*grpc.ClientConn, error) {
	dialWithTransportSecurityOption := grpc.
		WithTransportCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12}))
	if insecureParam {
		dialWithTransportSecurityOption = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	conn, err := grpc.Dial(host, dialWithTransportSecurityOption, grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func LoadListAccounts(path string) (admins []*AccountInfo, teachers []*AccountInfo, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	type JsonStruct struct {
		Admins   []*AccountInfo `json:"admins"`
		Teachers []*AccountInfo `json:"teachers"`
	}
	var acc JsonStruct
	err = json.Unmarshal(data, &acc)
	if err != nil {
		return nil, nil, fmt.Errorf("could not unmarshal file %s to get list accounts: %w", path, err)
	}

	return acc.Admins, acc.Teachers, nil
}

func (s *StressTest) LoginAllAccounts(ctx context.Context) error {
	// get token for testing accounts
	errChan := make(chan error)
	defer close(errChan)

	var wg sync.WaitGroup
	waitCh := make(chan struct{})
	go func() {
		for i := range s.adminAccounts {
			wg.Add(1)
			go func(a *AccountInfo) {
				defer wg.Done()
				var err error
				a.SignInInfo, err = s.UserSignInWithPassword(ctx, a, true)
				if err != nil {
					errChan <- fmt.Errorf("UserSignInWithPassword: %w", err)
					return
				}
				token, err := s.ExchangeUserToken(ctx, a.SignInInfo.IdToken)
				if err != nil {
					errChan <- fmt.Errorf("ExchangeUserToken: %w", err)
					return
				}
				a.Token = token
			}(s.adminAccounts[i])
		}
		wg.Wait()
		close(waitCh)
	}()

	select {
	case err := <-errChan:
		return err
	case <-waitCh:
		break
	}

	return nil
}

func (s *StressTest) QueryHasura(ctx context.Context, body []byte, jwt string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/graphql", s.cfg.BobHasuraAdminURL)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("accept", "application/json")
	request.Header.Set("authorization", fmt.Sprintf("Bearer %s", jwt))

	request = request.WithContext(ctx)
	resp, err := s.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("the HTTP request failed: %w", err)
	}

	return resp, nil
}
