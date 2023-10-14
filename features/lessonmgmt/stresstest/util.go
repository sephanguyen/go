package stresstest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type UserSignInPayload struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
	TenantID          string `json:"tenantId"`
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
		TenantID:          s.tenantID,
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
		return nil, fmt.Errorf("expected status 200 but got %d when login email '%s', psw: '%s'", resp.StatusCode, acc.Email, acc.Password)
	}

	var response UserSignInResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *StressTest) UserVerifyWithPassword(ctx context.Context, acc *AccountInfo, returnSecureToken bool) (*UserSignInResponse, error) {
	payload := &UserSignInPayload{
		Email:             acc.Email,
		Password:          acc.Password,
		ReturnSecureToken: returnSecureToken,
	}
	jsonReq, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPassword?key=%s", s.cfg.FirebaseAPIKey)
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
		return nil, fmt.Errorf("expected status 200 but got %d when login email '%s', psw: '%s'", resp.StatusCode, acc.Email, acc.Password)
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

type AccJsonStruct struct {
	Admins   []*AccountInfo `json:"admins"`
	Teachers []*AccountInfo `json:"teachers"`
	Students []*AccountInfo `json:"students"`
}

func LoadListAccounts(path string) (admins []*AccountInfo, teachers []*AccountInfo, students []*AccountInfo, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, err
	}

	var acc AccJsonStruct
	err = json.Unmarshal(data, &acc)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not unmarshal file %s to get list accounts: %w", path, err)
	}

	return acc.Admins, acc.Teachers, acc.Students, nil
}

func WriteAccounts(path string, admins []*AccountInfo, teachers []*AccountInfo, students []*AccountInfo) error {
	data := &AccJsonStruct{
		Admins:   admins,
		Teachers: teachers,
		Students: students,
	}
	content, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, content, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (s *StressTest) LoginAllAccounts(ctx context.Context) error {
	signInFunc := func(a *AccountInfo) error {
		var err error
		a.SignInInfo, err = s.UserSignInWithPassword(ctx, a, true)
		if err != nil {
			return fmt.Errorf("UserSignInWithPassword: %w", err)
		}
		token, err := s.ExchangeUserToken(ctx, a.SignInInfo.IdToken)
		if err != nil {
			// errChan <- fmt.Errorf("ExchangeUserToken: %w", err)
			return fmt.Errorf("ExchangeUserToken: %w", err)
		}
		// get token for testing accounts
		a.Token = token
		return nil
	}

	studentSignInFunc := func(a *AccountInfo) error {
		var err error
		a.SignInInfo, err = s.UserVerifyWithPassword(ctx, a, true)
		if err != nil {
			return fmt.Errorf("UserSignInWithPassword: %w", err)
		}
		token, err := s.ExchangeUserToken(ctx, a.SignInInfo.IdToken)
		if err != nil {
			return fmt.Errorf("ExchangeUserToken: %w", err)
		}
		// get token for testing accounts
		a.Token = token
		return nil
	}

	g := new(errgroup.Group)
	for i := range s.adminAccounts {
		acc := s.adminAccounts[i]
		g.Go(func() error {
			return signInFunc(acc)
		})
	}
	for i := range s.teacherAccounts {
		acc := s.teacherAccounts[i]
		g.Go(func() error {
			return signInFunc(acc)
		})
	}
	for i := range s.studentAccounts {
		acc := s.studentAccounts[i]
		g.Go(func() error {
			return studentSignInFunc(acc)
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *StressTest) GetAllUserID(ctx context.Context) error {
	getTeacherIDFunc := func(a *AccountInfo) error {
		resp, err := pb.NewUserServiceClient(s.connection.BobConn).
			GetTeacherProfiles(helper.GRPCContext(ctx, "token", a.Token), &pb.GetTeacherProfilesRequest{})
		if err != nil {
			return fmt.Errorf("GetTeacherProfiles: %w", err)
		}
		a.ID = resp.Profiles[0].Id
		return nil
	}
	getStudentIDFunc := func(a *AccountInfo) error {
		resp, err := bpb.NewStudentReaderServiceClient(s.connection.BobConn).
			RetrieveStudentProfile(helper.GRPCContext(ctx, "token", a.Token), &bpb.RetrieveStudentProfileRequest{})
		if err != nil {
			return fmt.Errorf("RetrieveStudentProfile: %w", err)
		}
		a.ID = resp.Items[0].Profile.Id
		return nil
	}

	g := new(errgroup.Group)
	for i := range s.teacherAccounts {
		go getTeacherIDFunc(s.teacherAccounts[i])
		acc := s.teacherAccounts[i]
		g.Go(func() error {
			return getTeacherIDFunc(acc)
		})
	}
	for i := range s.studentAccounts {
		go getStudentIDFunc(s.studentAccounts[i])
		acc := s.studentAccounts[i]
		g.Go(func() error {
			return getStudentIDFunc(acc)
		})
	}

	if err := g.Wait(); err != nil {
		return err
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

func HaveNoResponseError(ctx context.Context) error {
	stepState := common.StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return stepState.ResponseErr
	}

	return nil
}

func ContextForSuite(s *Suite) context.Context {
	return common.StepStateToContext(context.Background(), s.lessonSuite.CommonSuite.StepState)
}
