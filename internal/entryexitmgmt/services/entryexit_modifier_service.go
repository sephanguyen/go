package services

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/entryexitmgmt/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type MasterConfigurationService interface {
	GetConfigurations(context.Context, *mpb.GetConfigurationsRequest, ...grpc.CallOption) (*mpb.GetConfigurationsResponse, error)
}

type IStudentQRRepo interface {
	Create(ctx context.Context, db database.QueryExecer, e *entities.StudentQR) error
	Upsert(ctx context.Context, db database.QueryExecer, e *entities.StudentQR) error
	FindByID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.StudentQR, error)
	DeleteByStudentID(ctx context.Context, db database.QueryExecer, studentID string) error
}

type IStudentEntryExitRecordsRepo interface {
	Create(ctx context.Context, db database.QueryExecer, e *entities.StudentEntryExitRecords) error
	Update(ctx context.Context, db database.QueryExecer, e *entities.StudentEntryExitRecords) error
	GetLatestRecordByID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.StudentEntryExitRecords, error)
	LockAdvisoryByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (bool, error)
	UnLockAdvisoryByStudentID(ctx context.Context, db database.QueryExecer, studentID string) error
	SoftDeleteByID(ctx context.Context, db database.QueryExecer, id pgtype.Int4) error
	RetrieveRecordsByStudentID(ctx context.Context, db database.QueryExecer, filter repositories.RetrieveEntryExitRecordFilter) ([]*entities.StudentEntryExitRecords, error)
}

type IEntryExitQueueRepo interface {
	Create(ctx context.Context, db database.QueryExecer, e *entities.EntryExitQueue) error
}

type DBandRepo struct {
	DB                          database.Ext
	StudentQrRepo               IStudentQRRepo
	StudentEntryExitRecordsRepo IStudentEntryExitRecordsRepo
	EntryExitQueueRepo          IEntryExitQueueRepo
}

type Repositories struct {
	StudentQRRepo               IStudentQRRepo
	StudentEntryExitRecordsRepo IStudentEntryExitRecordsRepo
	EntryExitQueueRepo          IEntryExitQueueRepo
	StudentRepo                 *repositories.StudentRepo
	StudentParentRepo           *repositories.StudentParentRepo
	UserRepo                    *repositories.UserRepo
}

type Libraries struct {
	Logger                zap.SugaredLogger
	JSM                   nats.JetStreamManagement
	DB                    database.Ext
	UploadServiceSelector *UploadServiceSelector
}

type QREncryptionSecretKeys struct {
	EncryptionKey         string
	EncryptionKeyTokyo    string
	EncryptionKeySynersia string
}

type EntryExitModifierService struct {
	logger                      zap.SugaredLogger
	JSM                         nats.JetStreamManagement
	DB                          database.Ext
	encryptSecretKeyV2          string
	encryptSecretKeyTokyoV2     string
	encryptSecretKeySynersiaV2  string
	StudentQRRepo               IStudentQRRepo
	StudentEntryExitRecordsRepo IStudentEntryExitRecordsRepo
	EntryExitQueueRepo          IEntryExitQueueRepo
	StudentRepo                 interface {
		FindByID(context.Context, database.QueryExecer, string) (*entities.Student, error)
	}
	StudentParentRepo interface {
		GetParentIDsByStudentID(context.Context, database.QueryExecer, string) ([]string, error)
	}
	UserRepo interface {
		FindByID(context.Context, database.QueryExecer, string) (*entities.User, error)
	}
	CryptV2                    Crypt
	UploadServiceSelector      *UploadServiceSelector
	retryQuerySleep            time.Duration
	retryNotificationSleep     time.Duration
	MasterConfigurationService MasterConfigurationService
}

func NewEntryExitModifierService(
	libraries *Libraries,
	repositories *Repositories,
	encryptionSecretKey *QREncryptionSecretKeys,
	mastermgmtConfigurationService MasterConfigurationService,
) *EntryExitModifierService {
	return &EntryExitModifierService{
		logger:                      libraries.Logger,
		JSM:                         libraries.JSM,
		DB:                          libraries.DB,
		UploadServiceSelector:       libraries.UploadServiceSelector,
		StudentQRRepo:               repositories.StudentQRRepo,
		StudentEntryExitRecordsRepo: repositories.StudentEntryExitRecordsRepo,
		EntryExitQueueRepo:          repositories.EntryExitQueueRepo,
		StudentRepo:                 repositories.StudentRepo,
		StudentParentRepo:           repositories.StudentParentRepo,
		UserRepo:                    repositories.UserRepo,
		CryptV2:                     &CryptV2{},
		encryptSecretKeyV2:          encryptionSecretKey.EncryptionKey,
		encryptSecretKeyTokyoV2:     encryptionSecretKey.EncryptionKeyTokyo,
		encryptSecretKeySynersiaV2:  encryptionSecretKey.EncryptionKeySynersia,
		retryQuerySleep:             500 * time.Millisecond,
		retryNotificationSleep:      1 * time.Second,
		MasterConfigurationService:  mastermgmtConfigurationService,
	}
}
