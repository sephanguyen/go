package downloader

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/filestorage"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentFileDownloader interface {
	ValidateData(ctx context.Context) error
	GetByteContent(ctx context.Context) ([]byte, error)
	GetByteContentV2(ctx context.Context) ([]byte, error)
}

type FilePaymentDataMap struct {
	Payment                *entities.Payment
	Invoice                *entities.Invoice
	StudentBillingInfo     *entities.StudentBillingDetailsMap
	StudentBankDetails     *entities.StudentBankDetailsMap // Bank Account Details
	StudentRelatedBank     *entities.BankRelationMap       // Bank Branch, Bank and Partner Bank
	NewCustomerCodeHistory *entities.NewCustomerCodeHistory
}

type BasePaymentFileDownloader struct {
	DB                                database.Ext
	Logger                            zap.SugaredLogger
	BulkPaymentRequestFilePaymentRepo interface {
		FindPaymentInvoiceByRequestFileID(ctx context.Context, db database.QueryExecer, id string) ([]*entities.FilePaymentInvoiceMap, error)
	}
	BulkPaymentRequestFileRepo interface {
		FindByPaymentFileID(ctx context.Context, db database.QueryExecer, id string) (*entities.BulkPaymentRequestFile, error)
	}
	PartnerConvenienceStoreRepo interface {
		FindOne(ctx context.Context, db database.QueryExecer) (*entities.PartnerConvenienceStore, error)
	}
	PartnerBankRepo interface {
		FindOne(ctx context.Context, db database.QueryExecer) (*entities.PartnerBank, error)
	}
	StudentPaymentDetailRepo interface {
		FindStudentBillingByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.StudentBillingDetailsMap, error)
		FindStudentBankDetailsByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.StudentBankDetailsMap, error)
	}
	BankBranchRepo interface {
		FindRelatedBankOfBankBranches(ctx context.Context, db database.QueryExecer, branchIDs []string) ([]*entities.BankRelationMap, error)
	}
	NewCustomerCodeHistoryRepo interface {
		FindByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.NewCustomerCodeHistory, error)
		Create(ctx context.Context, db database.QueryExecer, e *entities.NewCustomerCodeHistory) error
	}
	PrefectureRepo interface {
		FindAll(ctx context.Context, db database.QueryExecer) ([]*entities.Prefecture, error)
	}
	Validator       *utils.PaymentRequestValidator
	FileStorage     filestorage.FileStorage
	TempFileCreator utils.ITempFileCreator
}

func (d *BasePaymentFileDownloader) getAndValidateFilePaymentInvoice(ctx context.Context, paymentFileID string) ([]*entities.FilePaymentInvoiceMap, error) {
	// Get the payments and its invoices associated in a payment file
	filePaymentInvoices, err := d.BulkPaymentRequestFilePaymentRepo.FindPaymentInvoiceByRequestFileID(ctx, d.DB, paymentFileID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("d.BulkPaymentRequestFilePaymentRepo.FindPaymentInvoiceByRequestFileID err: %v", err))
	}

	if len(filePaymentInvoices) == 0 {
		return nil, status.Error(codes.Internal, "There is no associated payment in the request file")
	}

	return filePaymentInvoices, nil
}

func (d *BasePaymentFileDownloader) getListOfStudentsFromPaymentInvoice(filePaymentInvoice []*entities.FilePaymentInvoiceMap, paymentMethod string, maxPaymentSequenceDigit int, maxTotalAmountDigit int) ([]string, error) {
	studentIDs := make([]string, len(filePaymentInvoice))
	for i, e := range filePaymentInvoice {
		if e.Payment.PaymentMethod.String != paymentMethod {
			return nil, status.Error(codes.Internal, "the payment method is not equal to the given payment method parameter")
		}

		if !e.Payment.IsExported.Bool {
			return nil, status.Error(codes.Internal, "payment isExported field should be true")
		}

		// Check if the payment sequence number digit length exceeds the requirement
		paymentSeqNumStr := strconv.Itoa(int(e.Payment.PaymentSequenceNumber.Int))
		if len(paymentSeqNumStr) > maxPaymentSequenceDigit {
			return nil, status.Error(codes.Internal, "the payment sequence number length exceeds the limit")
		}

		err := multierr.Combine(
			d.Validator.ValidateInvoice(e.Invoice, true, maxTotalAmountDigit),
		)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		studentIDs[i] = e.Payment.StudentID.String
	}

	return studentIDs, nil
}

func (d *BasePaymentFileDownloader) getObjectandFileName(ctx context.Context, paymentFileID string) (objectName string, fileName string, err error) {
	paymentRequestFile, err := d.BulkPaymentRequestFileRepo.FindByPaymentFileID(ctx, d.DB, paymentFileID)
	if err != nil {
		return "", "", fmt.Errorf("d.BulkPaymentRequestFileRepo.FindByPaymentFileID err: %v", err)
	}

	objectName = fmt.Sprintf("%s-%s", paymentRequestFile.BulkPaymentRequestID.String, paymentRequestFile.FileName.String)
	objectName = d.FileStorage.FormatObjectName(objectName)

	return objectName, paymentRequestFile.FileName.String, nil
}

func (d *BasePaymentFileDownloader) downloadToTempFile(ctx context.Context, paymentRequestFileID string, contentType filestorage.ContentType) (tf *utils.TempFile, err error) {
	objectName, fileName, err := d.getObjectandFileName(ctx, paymentRequestFileID)
	if err != nil {
		return nil, err
	}

	tf, err = d.TempFileCreator.CreateTempFile(fileName)
	if err != nil {
		return nil, err
	}

	err = d.FileStorage.DownloadFile(ctx, filestorage.FileToDownloadInfo{
		ObjectName:          objectName,
		DestinationPathName: tf.ObjectPath,
		ContentType:         contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("d.FileStorage.DownloadFile err: %v", err)
	}

	return tf, nil
}

func (d *BasePaymentFileDownloader) closeAndCleanupFile(tf *utils.TempFile) {
	err := tf.Close()
	if err != nil {
		d.Logger.Warn("Error on closing the temporary file", err)
	}

	err = tf.CleanUp()
	if err != nil {
		d.Logger.Warn("Error on cleaning up temporary file", err)
	}
}
