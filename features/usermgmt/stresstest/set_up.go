package stresstest

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
)

type AccountInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string

	SignInInfo *UserSignInResponse
}

type Suite struct {
	userSuite *usermgmt.Suite
	st        *StressTest
}

// random an account
func (s *Suite) ASignedInAsAccounts(ctx context.Context) error {
	acc := s.st.GetRandomAccount()

	token, err := s.st.ExchangeUserToken(ctx, acc.SignInInfo.IdToken)
	if err != nil {
		return fmt.Errorf("ExchangeUserToken: %w", err)
	}
	acc.Token = token
	s.userSuite.CommonSuite.AuthToken = token
	return nil
}

type StressTest struct {
	cfg    *common.Config
	client *http.Client

	schoolID      int32
	adminAccounts []*AccountInfo
	connection    *common.Connections
}

func NewStressTest(cfg *common.Config, schoolID int32, accountsPath string) (*StressTest, error) {
	rsc := bootstrap.NewResources().WithLoggerC(&cfg.Common)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	st := &StressTest{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cfg: cfg,
	}

	// connect to backend
	bobConn, err := SimplifiedDial(rsc.GetAddress("bob"), false)
	if err != nil {
		return nil, fmt.Errorf("SimplifiedDial: %w", err)
	}
	st.connection = &common.Connections{
		BobConn: bobConn,
	}

	// load accounts
	admins, _, err := LoadListAccounts(accountsPath)
	if err != nil {
		return nil, fmt.Errorf("LoadListAccounts: %w", err)
	}
	st.adminAccounts = admins
	// login all account to get token
	err = st.LoginAllAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("LoginAllAccounts: %w", err)
	}
	st.schoolID = schoolID

	return st, nil
}

func (s *StressTest) NewSuite() *Suite {
	userSuite := &usermgmt.Suite{
		Connections: s.connection,
		Cfg:         s.cfg,
		CommonSuite: &common.Suite{},
	}
	userSuite.CommonSuite.Connections = s.connection
	userSuite.CommonSuite.StepState = &common.StepState{
		CurrentSchoolID: s.schoolID,
	}

	suite := &Suite{
		userSuite: userSuite,
		st:        s,
	}

	return suite
}

func (s *StressTest) GetRandomAccount() *AccountInfo {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	n := r.Int31n(int32(len(s.adminAccounts)))
	return s.adminAccounts[n]
}
