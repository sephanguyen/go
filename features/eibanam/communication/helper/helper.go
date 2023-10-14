package helper

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	"github.com/manabie-com/backend/internal/golibs/nats"

	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
)

type CommunicationHelper struct {
	*common.Suite
	bobDBConn        *pgxpool.Pool
	bobGRPCConn      *grpc.ClientConn
	yasuoGRPCConn    *grpc.ClientConn
	firebaseAddress  string
	firebaseKey      string
	exampleNames     []string
	hasuraAdminUrl   string
	jsm              nats.JetStreamManagement
	tomGRPCConn      *grpc.ClientConn
	userMgmtGrpcConn *grpc.ClientConn
	shamirGRPCConn   *grpc.ClientConn
	applicantID      string
}

func NewCommunicationHelper(bobDBConn *pgxpool.Pool,
	bobGRPCConn *grpc.ClientConn,
	tomGrpcConn *grpc.ClientConn,
	yasuoConn *grpc.ClientConn,
	userManagementConn *grpc.ClientConn,
	firebaseAddress string,
	firebaseKey string,
	hasuraAdminUrl string,
	jsm nats.JetStreamManagement,
	shamirGRPCConn *grpc.ClientConn,
	applicantID string,
	connections *common.Connections,
) *CommunicationHelper {
	exampleName, _ := util.LoadExampleName()
	commonSuite := &common.Suite{}
	commonSuite.Connections = connections
	commonSuite.StepState = &common.StepState{}
	commonSuite.StepState.FirebaseAddress = firebaseAddress
	commonSuite.StepState.ApplicantID = applicantID
	return &CommunicationHelper{
		Suite:            commonSuite,
		bobDBConn:        bobDBConn,
		bobGRPCConn:      bobGRPCConn,
		yasuoGRPCConn:    yasuoConn,
		firebaseAddress:  firebaseAddress,
		firebaseKey:      firebaseKey,
		exampleNames:     exampleName,
		hasuraAdminUrl:   hasuraAdminUrl,
		jsm:              jsm,
		tomGRPCConn:      tomGrpcConn,
		userMgmtGrpcConn: userManagementConn,
		shamirGRPCConn:   shamirGRPCConn,
		applicantID:      applicantID,
	}
}

type userState struct {
	usersByID    map[string]*entity.User
	defaultAdmin *entity.Admin
	school       *entity.School
}

func (h *CommunicationHelper) NewStateful() *StatefulHelper {
	return &StatefulHelper{
		userState: &userState{
			usersByID: map[string]*entity.User{},
		},
		CommunicationHelper: h,
	}
}

type StatefulHelper struct {
	*CommunicationHelper
	userState *userState
}

func (h *CommunicationHelper) CreateSchoolAdminAndLoginToCMS(ctx context.Context, accountType string) (*entity.Admin, *entity.School, error) {
	// create school
	school, err := h.CreateNewSchool(accountType)
	if err != nil {
		return nil, nil, fmt.Errorf("CreateSysAdmin.Error %v", err)
	}

	// create system admin
	sysAdmin, err := h.CreateSysAdmin(int64(school.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("CreateNewSchool.Error %v", err)
	}

	// using system admin to create school admin
	schoolAdmin, err := h.CreateSchoolAdmin(sysAdmin, int64(school.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("CreateSchoolAdmin.Error %v", err)
	}

	// creat school admin password
	// if err = h.GenerateSchoolAdminPassword(sysAdmin, schoolAdmin); err != nil {
	// 	return nil, nil, fmt.Errorf("GenerateSchoolAdminPassword.Error %v", err)
	// }

	// login to cms and exchange the token for using later
	if err = h.SchoolAdminLoginToCms(ctx, schoolAdmin); err != nil {
		return nil, nil, fmt.Errorf("SchoolAdminLoginToCms.Error %v", err)
	}

	// map data to  suit state
	school.Admins = []*entity.Admin{schoolAdmin}

	return sysAdmin, school, nil
}

func (h *CommunicationHelper) PickName() string {
	l := len(h.exampleNames)
	if l < 1 {
		return ""
	}
	return h.exampleNames[rand.Intn(l)]
}

func (h *CommunicationHelper) BobDB() *pgxpool.Pool {
	return h.bobDBConn
}
