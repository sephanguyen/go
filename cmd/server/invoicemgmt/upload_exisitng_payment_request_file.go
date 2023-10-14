package invoicemgmt

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/configurations"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/filestorage"
	downloader "github.com/manabie-com/backend/internal/invoicemgmt/services/payment_file_downloader"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

type PaymentRequestEntities struct {
	BulkPaymentRequest     *entities.BulkPaymentRequest
	BulkPaymentRequestFile *entities.BulkPaymentRequestFile
}

func init() {
	bootstrap.RegisterJob("invoicemgmt_upload_existing_payment_request_file", RunUploadExistingPaymentRequestFile)
}

func RunUploadExistingPaymentRequestFile(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	db := rsc.DB()
	sugaredLogger := rsc.Logger().Sugar()

	// Apply this job in selected organizations where invoice was implemented
	// - Manabie -2147483648
	// - KEC -2147483642
	// - E2E-Tokyo -2147483639
	// - KEC-Demo -2147483635
	organizationIDs := []string{"-2147483648", "-2147483642", "-2147483639", "-2147483635"}
	return UploadExistingPaymentRequestFile(ctx, rsc.Storage(), db, sugaredLogger, organizationIDs)
}

func UploadExistingPaymentRequestFile(
	ctx context.Context,
	storageConfig *configs.StorageConfig,
	db *database.DBTrace,
	sugaredLogger *zap.SugaredLogger,
	organizationIDs []string,
) error {
	repos := initRepositories()

	uploader := &ExistingPaymentRequestFileUploader{
		Logger:                            sugaredLogger,
		DB:                                db,
		StorageConfig:                     storageConfig,
		BulkPaymentRequestFilePaymentRepo: repos.BulkPaymentRequestFilePaymentRepo,
		BulkPaymentRequestFileRepo:        repos.BulkPaymentRequestFileRepo,
		PartnerConvenienceStoreRepo:       repos.PartnerConvenienceStoreRepo,
		PartnerBankRepo:                   repos.PartnerBankRepo,
		StudentPaymentDetailRepo:          repos.StudentPaymentDetailRepo,
		BankBranchRepo:                    repos.BankBranchRepo,
		NewCustomerCodeHistoryRepo:        repos.NewCustomerCodeHistoryRepo,
		PrefectureRepo:                    repos.PrefectureRepo,
	}

	errorList := []string{}
	for _, orgID := range organizationIDs {
		sugaredLogger.Infof("started the upload script in org %v", orgID)

		// Set the orgID in the context
		tenantContext := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: orgID,
				UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			},
		})

		// Get the user_id from an invoice_schedule record
		// It is important to have an existing invoice_schedule in an org in order to successfully query the tables with access control
		query := `SELECT user_id FROM invoice_schedule WHERE resource_path = $1 LIMIT 1`
		var userID string

		err := db.QueryRow(tenantContext, query, orgID).Scan(&userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				sugaredLogger.Infof("script cannot continue since there is no invoice schedule present in organization %v", orgID)
				continue
			}

			errorList = append(errorList, fmt.Sprintf("db.QueryRow error: %v organization_id: %v", err, orgID))
			continue
		}

		// Assign the user_id to Manabie claims
		claims := interceptors.JWTClaimsFromContext(tenantContext)
		if claims != nil {
			claims.Manabie.UserID = userID
			tenantContext = interceptors.ContextWithJWTClaims(tenantContext, claims)
		}

		// Get all payment request files
		bulkPaymentEntities, err := getBulkPaymentRequestAndFile(tenantContext, db)
		if err != nil {
			errorList = append(errorList, fmt.Sprintf("getBulkPaymentRequestAndFile error: %v organization_id: %v", err, orgID))
			continue
		}

		if len(bulkPaymentEntities) == 0 {
			sugaredLogger.Infof("there is no payment request file with null file_url in organization %v", orgID)
			continue
		}

		// Run the upload logic script
		err = uploader.UploadPaymentRequestFile(tenantContext, bulkPaymentEntities)
		if err != nil {
			errorList = append(errorList, fmt.Sprintf("uploader.UploadPaymentRequestFile error: %v organization_id: %v", err, orgID))
		}

		sugaredLogger.Infof("finished the upload script in org %v", orgID)
	}

	if len(errorList) != 0 {
		return errors.New(strings.Join(errorList, ","))
	}

	return nil
}

type ExistingPaymentRequestFileUploader struct {
	Logger                            *zap.SugaredLogger
	DB                                *database.DBTrace
	StorageConfig                     *configs.StorageConfig
	BulkPaymentRequestFilePaymentRepo *repositories.BulkPaymentRequestFilePaymentRepo
	BulkPaymentRequestFileRepo        *repositories.BulkPaymentRequestFileRepo
	PartnerConvenienceStoreRepo       *repositories.PartnerConvenienceStoreRepo
	PartnerBankRepo                   *repositories.PartnerBankRepo
	StudentPaymentDetailRepo          *repositories.StudentPaymentDetailRepo
	BankBranchRepo                    *repositories.BankBranchRepo
	NewCustomerCodeHistoryRepo        *repositories.NewCustomerCodeHistoryRepo
	PrefectureRepo                    *repositories.PrefectureRepo
}

func (u *ExistingPaymentRequestFileUploader) UploadPaymentRequestFile(ctx context.Context, bulkPaymentEntities []*PaymentRequestEntities) error {
	basePaymentFileDownloader := &downloader.BasePaymentFileDownloader{
		DB:                                u.DB,
		Logger:                            *u.Logger,
		BulkPaymentRequestFilePaymentRepo: u.BulkPaymentRequestFilePaymentRepo,
		BulkPaymentRequestFileRepo:        u.BulkPaymentRequestFileRepo,
		PartnerConvenienceStoreRepo:       u.PartnerConvenienceStoreRepo,
		PartnerBankRepo:                   u.PartnerBankRepo,
		StudentPaymentDetailRepo:          u.StudentPaymentDetailRepo,
		BankBranchRepo:                    u.BankBranchRepo,
		NewCustomerCodeHistoryRepo:        u.NewCustomerCodeHistoryRepo,
		PrefectureRepo:                    u.PrefectureRepo,
		Validator:                         &utils.PaymentRequestValidator{},
	}

	tempFileCreator := &utils.TempFileCreator{TempDirPattern: constant.InvoicemgmtTemporaryDir}
	fileStorage, err := u.getFileStorageInstance()
	if err != nil {
		return err
	}

	errorList := []error{}
	for _, entities := range bulkPaymentEntities {
		err = u.uploadPaymentFile(ctx, entities, basePaymentFileDownloader, fileStorage, tempFileCreator)
		if err != nil {
			u.Logger.Warn(err.Error())
			errorList = append(errorList, err)
		}
	}

	if len(errorList) != 0 {
		return fmt.Errorf("there are %v error/s during uploading the payment file. check the logs for more information", len(errorList))
	}

	return nil
}

func (u *ExistingPaymentRequestFileUploader) uploadPaymentFile(
	ctx context.Context,
	entities *PaymentRequestEntities,
	basePaymentFileDownloader *downloader.BasePaymentFileDownloader,
	fileStorage filestorage.FileStorage,
	tempFileCreator *utils.TempFileCreator,
) error {
	// Get the byte content of the payment file
	byteContent, fileContent, err := u.getByteContent(ctx, entities, basePaymentFileDownloader)
	if err != nil {
		return err
	}

	objectUploader, err := utils.NewObjectUploader(
		fileStorage,
		tempFileCreator,
		&utils.ObjectInfo{
			ObjectName:  fmt.Sprintf("%s-%s", entities.BulkPaymentRequest.BulkPaymentRequestID.String, entities.BulkPaymentRequestFile.FileName.String),
			ByteContent: byteContent,
			ContentType: fileContent,
		},
	)
	defer func() {
		err := objectUploader.Close()
		if err != nil {
			u.Logger.Warn(err)
		}
	}()
	if err != nil {
		return err
	}

	if err := database.ExecInTx(ctx, u.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// Update file URL of payment request file
		err = entities.BulkPaymentRequestFile.FileURL.Set(objectUploader.GetDownloadFileURL())
		if err != nil {
			return err
		}
		err = u.BulkPaymentRequestFileRepo.UpdateWithFields(ctx, tx, entities.BulkPaymentRequestFile, []string{"updated_at", "file_url"})
		if err != nil {
			return fmt.Errorf("error BulkPaymentRequestFileRepo.UpdateWithFields err: %v", err)
		}

		// Upload to cloud storage
		if err := objectUploader.DoUploadFile(ctx); err != nil {
			return fmt.Errorf("error uploading the file in cloud storage err: %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (u *ExistingPaymentRequestFileUploader) getByteContent(
	ctx context.Context,
	entities *PaymentRequestEntities,
	basePaymentFileDownloader *downloader.BasePaymentFileDownloader,
) ([]byte, filestorage.ContentType, error) {
	var (
		fileDownloader downloader.PaymentFileDownloader
		fileContent    filestorage.ContentType
	)

	switch entities.BulkPaymentRequest.PaymentMethod.String {
	case invoice_pb.PaymentMethod_DIRECT_DEBIT.String():
		fileDownloader = &downloader.DirectDebitTXTPaymentFileDownloader{
			PaymentFileID:             entities.BulkPaymentRequestFile.BulkPaymentRequestFileID.String,
			BasePaymentFileDownloader: basePaymentFileDownloader,
		}
		fileContent = filestorage.ContentTypeTXT
	case invoice_pb.PaymentMethod_CONVENIENCE_STORE.String():
		fileDownloader = &downloader.ConvenienceStoreCSVPaymentFileDownloader{
			PaymentFileID:             entities.BulkPaymentRequestFile.BulkPaymentRequestFileID.String,
			BasePaymentFileDownloader: basePaymentFileDownloader,
		}
		fileContent = filestorage.ContentTypeCSV
	}

	// Validate the data also initialize the mapping of data
	err := fileDownloader.ValidateData(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("fileDownloader.ValidateData err: %v", err)
	}

	// Get the byte content of the payment request file
	byteContent, err := fileDownloader.GetByteContent(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("fileDownloader.GetByteContent err: %v", err)
	}

	return byteContent, fileContent, nil
}

func (u *ExistingPaymentRequestFileUploader) getFileStorageInstance() (filestorage.FileStorage, error) {
	fileStorageName := filestorage.GoogleCloudStorageService
	if strings.Contains(u.StorageConfig.Endpoint, "minio") {
		fileStorageName = filestorage.MinIOService
	}

	fileStorage, err := filestorage.GetFileStorage(fileStorageName, u.StorageConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to init %v file storage", fileStorageName)
	}

	return fileStorage, nil
}

func getBulkPaymentRequestAndFile(ctx context.Context, db *database.DBTrace) ([]*PaymentRequestEntities, error) {
	entityList := []*PaymentRequestEntities{}
	resourcePath, err := interceptors.ResourcePathFromContext(ctx)
	if err != nil {
		return nil, err
	}

	getPaymentRequestFilesQuery := `
		SELECT
			pr.bulk_payment_request_id,
			pr.payment_method,
			prf.bulk_payment_request_file_id,
			prf.file_name,
			prf.file_url,
			prf.file_sequence_number,
			prf.total_file_count
		FROM bulk_payment_request_file prf
		INNER JOIN bulk_payment_request pr
			ON pr.bulk_payment_request_id = prf.bulk_payment_request_id
		WHERE (prf.file_url IS NULL or prf.file_url = '')
			AND pr.error_details IS NULL
			AND prf.resource_path = $1
			AND prf.deleted_at IS NULL
			AND pr.deleted_at IS NULL
	`

	rows, err := db.Query(ctx, getPaymentRequestFilesQuery, resourcePath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		bulkPaymentRequest := &entities.BulkPaymentRequest{}
		database.AllNullEntity(bulkPaymentRequest)

		bulkPaymentRequestFile := &entities.BulkPaymentRequestFile{}
		database.AllNullEntity(bulkPaymentRequestFile)

		err := rows.Scan(
			&bulkPaymentRequest.BulkPaymentRequestID,
			&bulkPaymentRequest.PaymentMethod,
			&bulkPaymentRequestFile.BulkPaymentRequestFileID,
			&bulkPaymentRequestFile.FileName,
			&bulkPaymentRequestFile.FileURL,
			&bulkPaymentRequestFile.FileSequenceNumber,
			&bulkPaymentRequestFile.TotalFileCount,
		)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		entityList = append(entityList, &PaymentRequestEntities{
			BulkPaymentRequest:     bulkPaymentRequest,
			BulkPaymentRequestFile: bulkPaymentRequestFile,
		})
	}

	return entityList, nil
}
