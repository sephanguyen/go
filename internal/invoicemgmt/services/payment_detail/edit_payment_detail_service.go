package payment_detail

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"go.uber.org/zap"
)

type EditPaymentDetailService struct {
	invoice_pb.UnimplementedEditPaymentDetailServiceServer

	Logger *zap.SugaredLogger
	DB     database.Ext

	PrefectureRepo                    PrefectureRepo
	BillingAddressRepo                BillingAddressRepo
	StudentPaymentDetailRepo          StudentPaymentDetailRepo
	BankAccountRepo                   BankAccountRepo
	BankRepo                          BankRepo
	BankBranchRepo                    BankBranchRepo
	StudentPaymentDetailActionLogRepo StudentPaymentDetailActionLogRepo
}

type StudentPaymentDetailRepo interface {
	FindByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.StudentPaymentDetail, error)
	FindByID(ctx context.Context, db database.QueryExecer, studentPaymentDetailID string) (*entities.StudentPaymentDetail, error)
	Upsert(ctx context.Context, db database.QueryExecer, studentPaymentDetail ...*entities.StudentPaymentDetail) error
	SoftDelete(ctx context.Context, db database.QueryExecer, studentPaymentDetailIDs ...string) error
}

type BillingAddressRepo interface {
	FindByUserID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.BillingAddress, error)
	FindByID(ctx context.Context, db database.QueryExecer, id string) (*entities.BillingAddress, error)
	Upsert(ctx context.Context, db database.QueryExecer, billingAddress ...*entities.BillingAddress) error
	SoftDelete(ctx context.Context, db database.QueryExecer, billingAddressIDs ...string) error
}

type PrefectureRepo interface {
	FindByPrefectureCode(ctx context.Context, db database.QueryExecer, prefectureCode string) (*entities.Prefecture, error)
}

type BankAccountRepo interface {
	FindByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.BankAccount, error)
	FindByID(ctx context.Context, db database.QueryExecer, id string) (*entities.BankAccount, error)
	Upsert(ctx context.Context, db database.QueryExecer, bankAccounts ...*entities.BankAccount) error
}

type BankRepo interface {
	FindByID(ctx context.Context, db database.QueryExecer, bankID string) (*entities.Bank, error)
}

type BankBranchRepo interface {
	FindByID(ctx context.Context, db database.QueryExecer, bankBranchID string) (*entities.BankBranch, error)
}

type StudentPaymentDetailActionLogRepo interface {
	Create(ctx context.Context, db database.QueryExecer, e *entities.StudentPaymentDetailActionLog) error
}
