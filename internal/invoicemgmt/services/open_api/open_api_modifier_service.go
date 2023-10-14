package openapisvc

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type MasterConfigurationService interface {
	GetConfigurations(context.Context, *mpb.GetConfigurationsRequest, ...grpc.CallOption) (*mpb.GetConfigurationsResponse, error)
}

type OpenAPIModifierServiceRepositories struct {
	StudentPaymentDetailRepo          *repositories.StudentPaymentDetailRepo
	BillingAddressRepo                *repositories.BillingAddressRepo
	PrefectureRepo                    *repositories.PrefectureRepo
	UserRepo                          *repositories.UserRepo
	BankRepo                          *repositories.BankRepo
	BankBranchRepo                    *repositories.BankBranchRepo
	BankAccountRepo                   *repositories.BankAccountRepo
	StudentPaymentDetailActionLogRepo *repositories.StudentPaymentDetailActionLogRepo
}

type OpenAPIModifierService struct {
	logger                   zap.SugaredLogger
	JSM                      nats.JetStreamManagement
	DB                       database.Ext
	StudentPaymentDetailRepo interface {
		FindByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.StudentPaymentDetail, error)
		Upsert(ctx context.Context, db database.QueryExecer, studentPaymentDetail ...*entities.StudentPaymentDetail) error
	}
	BillingAddressRepo interface {
		FindByUserID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.BillingAddress, error)
		Upsert(ctx context.Context, db database.QueryExecer, billingAddress ...*entities.BillingAddress) error
	}
	PrefectureRepo interface {
		FindByPrefectureID(ctx context.Context, db database.QueryExecer, prefectureID string) (*entities.Prefecture, error)
	}
	UserRepo interface {
		FindByUserExternalID(ctx context.Context, db database.QueryExecer, externalUserID string) (*entities.User, error)
	}
	BankRepo interface {
		FindByBankCode(ctx context.Context, db database.QueryExecer, bankCode string) (*entities.Bank, error)
	}
	BankBranchRepo interface {
		FindByBankBranchCodeAndBank(ctx context.Context, db database.QueryExecer, bankBranchCode, bankID string) (*entities.BankBranch, error)
	}
	BankAccountRepo interface {
		FindByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.BankAccount, error)
		Upsert(ctx context.Context, db database.QueryExecer, bankAccounts ...*entities.BankAccount) error
	}
	StudentPaymentDetailActionLogRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.StudentPaymentDetailActionLog) error
	}
	UnleashClient              unleashclient.ClientInstance
	MasterConfigurationService MasterConfigurationService
	Env                        string
}

func NewOpenAPIModifierService(
	logger zap.SugaredLogger,
	jsm nats.JetStreamManagement,
	db database.Ext,
	serviceRepo *OpenAPIModifierServiceRepositories,
	unleashClient unleashclient.ClientInstance,
	mastermgmtConfigurationService MasterConfigurationService,
	env string,
) *OpenAPIModifierService {
	return &OpenAPIModifierService{
		logger:                            logger,
		JSM:                               jsm,
		DB:                                db,
		StudentPaymentDetailRepo:          serviceRepo.StudentPaymentDetailRepo,
		BillingAddressRepo:                serviceRepo.BillingAddressRepo,
		PrefectureRepo:                    serviceRepo.PrefectureRepo,
		UserRepo:                          serviceRepo.UserRepo,
		BankRepo:                          serviceRepo.BankRepo,
		BankBranchRepo:                    serviceRepo.BankBranchRepo,
		BankAccountRepo:                   serviceRepo.BankAccountRepo,
		StudentPaymentDetailActionLogRepo: serviceRepo.StudentPaymentDetailActionLogRepo,
		UnleashClient:                     unleashClient,
		MasterConfigurationService:        mastermgmtConfigurationService,
		Env:                               env,
	}
}
