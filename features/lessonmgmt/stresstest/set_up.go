package stresstest

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/lessonmgmt"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
)

type AccountInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string
	ID       string

	SignInInfo *UserSignInResponse
}

type Suite struct {
	lessonSuite *lessonmgmt.Suite
	st          *StressTest
}

// random an account
func (s *Suite) ASignedInAsSchoolAdmin(ctx context.Context) error {
	acc := s.st.GetRandomAdminAccount()

	token, err := s.st.ExchangeUserToken(ctx, acc.SignInInfo.IdToken)
	if err != nil {
		return fmt.Errorf("ExchangeUserToken: %w", err)
	}
	acc.Token = token
	s.lessonSuite.CommonSuite.AuthToken = token
	return nil
}

func (s *Suite) ASignedInWithAccInfo(ctx context.Context, acc *AccountInfo) error {
	token, err := s.st.ExchangeUserToken(ctx, acc.SignInInfo.IdToken)
	if err != nil {
		return fmt.Errorf("ExchangeUserToken: %w", err)
	}
	acc.Token = token
	s.lessonSuite.CommonSuite.AuthToken = token
	return nil
}

type StressTest struct {
	cfg    *common.Config
	client *http.Client

	tenantID        string
	schoolID        int32
	adminAccounts   []*AccountInfo
	teacherAccounts []*AccountInfo
	studentAccounts []*AccountInfo
	connection      *common.Connections
}

func NewStressTest(cfg *common.Config, schoolID int32, tenantID, accountsPath string) (*StressTest, error) {
	rsc := bootstrap.NewResources().WithLoggerC(&cfg.Common)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	st := &StressTest{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cfg:      cfg,
		tenantID: tenantID,
		schoolID: schoolID,
	}

	// connect to backend
	bobConn, err := SimplifiedDial(rsc.GetAddress("bob"), false)
	if err != nil {
		return nil, fmt.Errorf("SimplifiedDial: %w", err)
	}
	st.connection = &common.Connections{
		BobConn:   bobConn,
		YasuoConn: bobConn,
	}

	// load accounts
	admins, teachers, students, err := LoadListAccounts(accountsPath)
	if err != nil {
		return nil, fmt.Errorf("LoadListAccounts: %w", err)
	}
	st.adminAccounts = admins
	st.teacherAccounts = teachers
	st.studentAccounts = students

	// login all account to get token
	err = st.LoginAllAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("LoginAllAccounts: %w", err)
	}
	err = st.GetAllUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllUserID: %w", err)
	}

	return st, nil
}

func (s *StressTest) NewSuite() *Suite {
	lessonSuite := &lessonmgmt.Suite{
		Connections: s.connection,
		Cfg:         s.cfg,
		CommonSuite: &common.Suite{},
	}
	lessonSuite.CommonSuite.Connections = s.connection
	lessonSuite.CommonSuite.StepState = &common.StepState{
		CurrentSchoolID: s.schoolID,
	}
	lessonSuite.CommonSuite.SubV2Clients = make(map[string]common.CancellableStream)

	suite := &Suite{
		lessonSuite: lessonSuite,
		st:          s,
	}

	return suite
}

func (s *StressTest) GetRandomAdminAccount() *AccountInfo {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	n := r.Int31n(int32(len(s.adminAccounts)))
	return s.adminAccounts[n]
}
