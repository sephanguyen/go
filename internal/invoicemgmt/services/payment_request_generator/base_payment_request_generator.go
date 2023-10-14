package generator

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/filestorage"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	maximumPaymentPerCSV = 1000
	maximumPaymentPerTXT = 2000
	csvFormat            = "csv"
	txtFormat            = "txt"
	fileNameTimeFormat   = "20060102"
)

type dataMap struct {
	Payment                       *entities.Payment
	Invoice                       *entities.Invoice
	NewCustomerCodeHistory        *entities.NewCustomerCodeHistory
	BankAccount                   *entities.BankAccount
	AccountNumberWithCustomerCode map[string]struct{}

	StudentBillingInfo *entities.StudentBillingDetailsMap
	StudentBankDetails *entities.StudentBankDetailsMap
	StudentRelatedBank *entities.BankRelationMap
	BillItemDetails    []*entities.InvoiceBillItemMap
	InvoiceAdjustments []*entities.InvoiceAdjustment
}

type paymentInvoiceMap struct {
	Payment *entities.Payment
	Invoice *entities.Invoice
}

// paymentAndFileAssoc used to identify the association of list of payments to a file
// Basically it defines what list of payments are included in a file
// This also contains information about the file
type paymentAndFileAssoc struct {
	DataMap                    []*dataMap
	TotalFileCount             int
	FileSequenceNumber         int
	FileName                   string
	PaymentRequestFileID       string
	ParentPaymentRequestFileID string
}

type PaymentRequestGenerator interface {
	ValidateData(ctx context.Context) error
	ValidateDataV2(ctx context.Context) error
	PlanPaymentAndFileAssociation(ctx context.Context) error
	SavePaymentAndFileAssociation(ctx context.Context) error
	SaveAndUploadPaymentFile(ctx context.Context) error
	SaveAndUploadPaymentFileV2(ctx context.Context) error
}

type BasePaymentRequestGenerator struct {
	DB            database.Ext
	Logger        zap.SugaredLogger
	UnleashClient unleashclient.ClientInstance
	Env           string

	PaymentRepo interface {
		FindPaymentInvoiceByIDs(ctx context.Context, db database.QueryExecer, paymentIDs []string) ([]*entities.PaymentInvoiceMap, error)
		UpdateIsExportedByPaymentRequestFileID(ctx context.Context, db database.QueryExecer, filePaymentID string, isExported bool) error
		UpdateIsExportedByPaymentIDs(ctx context.Context, db database.QueryExecer, paymentIDs []string, isExported bool) error
	}
	InvoiceRepo interface {
		UpdateIsExportedByPaymentRequestFileID(ctx context.Context, db database.QueryExecer, filePaymentID string, isExported bool) error
		UpdateIsExportedByInvoiceIDs(ctx context.Context, db database.QueryExecer, invoiceIDs []string, isExported bool) error
	}
	BulkPaymentRequestFileRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequestFile) (string, error)
	}
	BulkPaymentRequestFilePaymentRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequestFilePayment) (string, error)
	}
	PartnerConvenienceStoreRepo interface {
		FindOne(ctx context.Context, db database.QueryExecer) (*entities.PartnerConvenienceStore, error)
	}
	StudentPaymentDetailRepo interface {
		FindStudentBillingByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.StudentBillingDetailsMap, error)
		FindStudentBankDetailsByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.StudentBankDetailsMap, error)
	}
	BankBranchRepo interface {
		FindRelatedBankOfBankBranches(ctx context.Context, db database.QueryExecer, branchIDs []string) ([]*entities.BankRelationMap, error)
	}
	NewCustomerCodeHistoryRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.NewCustomerCodeHistory) error
		FindByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.NewCustomerCodeHistory, error)
		FindByAccountNumbers(ctx context.Context, db database.QueryExecer, bankAccountNumbers []string) ([]*entities.NewCustomerCodeHistory, error)
		UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.NewCustomerCodeHistory, fieldsToUpdate []string) error
	}
	PrefectureRepo interface {
		FindAll(ctx context.Context, db database.QueryExecer) ([]*entities.Prefecture, error)
	}
	BulkPaymentRepo interface {
		UpdateBulkPaymentStatusByIDs(ctx context.Context, db database.QueryExecer, status string, bulkPaymentIDs []string) error
	}
	BillItemRepo interface {
		FindInvoiceBillItemMapByInvoiceIDs(ctx context.Context, db database.QueryExecer, invoiceIDs []string) ([]*entities.InvoiceBillItemMap, error)
	}
	InvoiceAdjustmentRepo interface {
		FindByInvoiceIDs(ctx context.Context, db database.QueryExecer, invoiceIDs []string) ([]*entities.InvoiceAdjustment, error)
	}
	Validator        utils.PaymentRequestValidator
	FileStorage      filestorage.FileStorage
	TempFileCreator  utils.ITempFileCreator
	StringNormalizer *utils.StringNormalizer
}

// nolint:unused
func (g *BasePaymentRequestGenerator) associatePaymentToFile(ctx context.Context, bulkPaymetRequestID string, paymentMethod string, paymentFileAssocs []paymentAndFileAssoc) error {
	err := database.ExecInTx(ctx, g.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		for _, paymentFileAssoc := range paymentFileAssocs {
			e, err := generateBulkPaymentRequestFileEntity(bulkPaymetRequestID, paymentFileAssoc)
			if err != nil {
				return err
			}

			// Create the payment request file entity
			requestFileID, err := g.BulkPaymentRequestFileRepo.Create(ctx, tx, e)
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestFileRepo.Create err: %v", err))
			}

			for _, data := range paymentFileAssoc.DataMap {
				// Associate the payments to payment request file
				err := g.savePaymentRequestFilePayment(ctx, tx, requestFileID, data.Payment)
				if err != nil {
					return err
				}

				// Insert or update customer code if direct debit
				if paymentMethod == invoice_pb.PaymentMethod_DIRECT_DEBIT.String() {
					_, err := g.upsertCustomerCode(ctx, tx, data.Payment.StudentID.String, data)
					if err != nil {
						return err
					}
				}
			}

			err = g.PaymentRepo.UpdateIsExportedByPaymentRequestFileID(ctx, tx, requestFileID, true)
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("g.PaymentRepo.UpdateIsExportedByPaymentRequestFileID err: %v", err))
			}

			err = g.InvoiceRepo.UpdateIsExportedByPaymentRequestFileID(ctx, tx, requestFileID, true)
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("g.InvoiceRepo.UpdateIsExportedByPaymentRequestFileID err: %v", err))
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (g *BasePaymentRequestGenerator) savePaymentRequestFilePayment(ctx context.Context, db database.QueryExecer, requestFileID string, payment *entities.Payment) error {
	e := new(entities.BulkPaymentRequestFilePayment)
	database.AllNullEntity(e)

	err := multierr.Combine(
		e.BulkPaymentRequestFileID.Set(requestFileID),
		e.PaymentID.Set(payment.PaymentID.String),
	)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("multierr.Combine: %v", err))
	}

	_, err = g.BulkPaymentRequestFilePaymentRepo.Create(ctx, db, e)
	if err != nil {
		// Check for UNIQUE constraint error. The payment_id column has UNIQUE constraint.
		// We can check using this error if the payment already belongs to a file instead of creating another query to check it.
		if strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			return status.Error(codes.Internal, fmt.Sprintf("Payment with ID %s already exists in a payment request file", payment.PaymentID.String))
		}

		return status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestFilePaymentRepo.Create err: %v", err))
	}

	return nil
}

func (g *BasePaymentRequestGenerator) createNewCustomerCodeHistory(ctx context.Context, db database.QueryExecer, studentID string, accountNumber string, customerCode string) (*entities.NewCustomerCodeHistory, error) {
	e := &entities.NewCustomerCodeHistory{}
	database.AllNullEntity(e)
	err := multierr.Combine(
		e.StudentID.Set(studentID),
		e.BankAccountNumber.Set(accountNumber),
		e.NewCustomerCode.Set(customerCode),
	)
	if err != nil {
		return nil, err
	}

	err = g.NewCustomerCodeHistoryRepo.Create(ctx, db, e)
	if err != nil {
		return nil, fmt.Errorf("g.NewCustomerCodeHistoryRepo.Create err %v", err)
	}

	return e, nil
}

func (g *BasePaymentRequestGenerator) upsertCustomerCode(ctx context.Context, db database.QueryExecer, studentID string, dataMap *dataMap) (*entities.NewCustomerCodeHistory, error) {
	// Student do not have customer code
	customerCode := "0"
	if dataMap.NewCustomerCodeHistory == nil {
		// Check if the student bank account already exists in customer code then set the customer code to 0
		_, ok := dataMap.AccountNumberWithCustomerCode[dataMap.BankAccount.BankAccountNumber.String]
		if !ok {
			customerCode = "1"
		}

		newCustomerCode, err := g.createNewCustomerCodeHistory(ctx, db, studentID, dataMap.BankAccount.BankAccountNumber.String, customerCode)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return newCustomerCode, nil
	}

	// Do not update if the customer code is already 0
	if dataMap.NewCustomerCodeHistory.NewCustomerCode.String == "0" {
		return dataMap.NewCustomerCodeHistory, nil
	}

	dataMap.NewCustomerCodeHistory.NewCustomerCode = database.Text("0")
	err := g.NewCustomerCodeHistoryRepo.UpdateWithFields(ctx, db, dataMap.NewCustomerCodeHistory, []string{"new_customer_code"})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("g.NewCustomerCodeHistoryRepo.UpdateWithFields err %v", err))
	}

	return dataMap.NewCustomerCodeHistory, nil
}

func generateBulkPaymentRequestFileEntity(bulkPaymentRequestID string, paymentFileAssoc paymentAndFileAssoc) (*entities.BulkPaymentRequestFile, error) {
	e := new(entities.BulkPaymentRequestFile)
	database.AllNullEntity(e)

	if err := multierr.Combine(
		e.BulkPaymentRequestID.Set(bulkPaymentRequestID),
		e.BulkPaymentRequestFileID.Set(paymentFileAssoc.PaymentRequestFileID),
		e.FileName.Set(paymentFileAssoc.FileName),
		e.FileSequenceNumber.Set(paymentFileAssoc.FileSequenceNumber),
		e.TotalFileCount.Set(paymentFileAssoc.TotalFileCount),
		e.IsDownloaded.Set(false),
	); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("multierr.Combine: %v", err))
	}

	// If there is a given parent file ID, assign it to the entity
	if strings.TrimSpace(paymentFileAssoc.ParentPaymentRequestFileID) != "" {
		_ = e.ParentPaymentRequestFileID.Set(paymentFileAssoc.ParentPaymentRequestFileID)
	}

	return e, nil
}

func generateBulkPaymentRequestFileEntityV2(bulkPaymentRequestID, downloadFileURL string, paymentFileAssoc paymentAndFileAssoc) (*entities.BulkPaymentRequestFile, error) {
	e := new(entities.BulkPaymentRequestFile)
	database.AllNullEntity(e)

	if err := multierr.Combine(
		e.BulkPaymentRequestID.Set(bulkPaymentRequestID),
		e.BulkPaymentRequestFileID.Set(paymentFileAssoc.PaymentRequestFileID),
		e.FileName.Set(paymentFileAssoc.FileName),
		e.FileSequenceNumber.Set(paymentFileAssoc.FileSequenceNumber),
		e.TotalFileCount.Set(paymentFileAssoc.TotalFileCount),
		e.IsDownloaded.Set(false),
		e.FileURL.Set(downloadFileURL),
	); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("multierr.Combine: %v", err))
	}

	// If there is a given parent file ID, assign it to the entity
	if strings.TrimSpace(paymentFileAssoc.ParentPaymentRequestFileID) != "" {
		_ = e.ParentPaymentRequestFileID.Set(paymentFileAssoc.ParentPaymentRequestFileID)
	}

	return e, nil
}

func generatePaymentAndFileAssocs(dataMap []*dataMap, maximumPaymentPerFile int, fileName string, fileExtension string) []paymentAndFileAssoc {
	dataMapChunk := chunkDataMapList(dataMap, maximumPaymentPerFile)
	totalFileCount := len(dataMapChunk)

	paymentAndFileAssocs := make([]paymentAndFileAssoc, totalFileCount)
	parentID := ""

	for i, d := range dataMapChunk {
		paymentFileID := idutil.ULIDNow()
		newFileName := addSequenceToFileName(fileName, totalFileCount, i+1)

		paymentAndFileAssocs[i] = paymentAndFileAssoc{
			DataMap:                    d,
			TotalFileCount:             totalFileCount,
			FileName:                   fmt.Sprintf("%s.%s", newFileName, fileExtension),
			PaymentRequestFileID:       paymentFileID,
			ParentPaymentRequestFileID: parentID,
			FileSequenceNumber:         i + 1,
		}

		// Set the parentID here so the value will be empty for parent file
		if i == 0 {
			parentID = paymentFileID
		}
	}

	return paymentAndFileAssocs
}

func addSequenceToFileName(fileName string, totalFileCount int, sequenceNumber int) string {
	isMultiFile := totalFileCount > 1

	if isMultiFile {
		fileName = fmt.Sprintf("%s_%dof%d", fileName, sequenceNumber, totalFileCount)
	}

	return fileName
}

func chunkDataMapList(slice []*dataMap, chunkSize int) [][]*dataMap {
	var chunks [][]*dataMap

	if chunkSize <= 0 {
		return [][]*dataMap{slice}
	}

	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

func isNegative(totalAmount pgtype.Numeric) bool {
	return totalAmount.Int.Cmp(big.NewInt(0)) == -1
}

func (g *BasePaymentRequestGenerator) getPaymentInvoice(ctx context.Context, paymentIDs []string) ([]*entities.PaymentInvoiceMap, error) {
	paymentInvoice, err := g.PaymentRepo.FindPaymentInvoiceByIDs(ctx, g.DB, paymentIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("g.PaymentRepo.FindPaymentInvoiceByIDs err: %v", err))
	}

	if len(paymentInvoice) != len(paymentIDs) {
		return nil, status.Error(codes.Internal, "There are payments that does not exist")
	}

	return paymentInvoice, nil
}

func (g *BasePaymentRequestGenerator) getListOfStudentsFromPaymentInvoice(
	paymentInvoice []*entities.PaymentInvoiceMap,
	paymentMethod string,
	maxPaymentSequenceDigit int,
	maxTotalAmountDigit int,
) ([]string, error) {
	studentIDs := make([]string, len(paymentInvoice))
	for i, e := range paymentInvoice {
		err := multierr.Combine(
			g.Validator.ValidatePayment(e.Payment, paymentMethod, false, maxPaymentSequenceDigit),
			g.Validator.ValidateInvoice(e.Invoice, false, maxTotalAmountDigit),
		)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		studentIDs[i] = e.Payment.StudentID.String
	}

	return studentIDs, nil
}

func (g *BasePaymentRequestGenerator) getInvoiceAdjustmentsMap(
	ctx context.Context,
	invoiceIDs []string,
) (map[string][]*entities.InvoiceAdjustment, error) {
	m := make(map[string][]*entities.InvoiceAdjustment)

	// Find bill item by invoice IDs
	invoiceAdjustments, err := g.InvoiceAdjustmentRepo.FindByInvoiceIDs(ctx, g.DB, invoiceIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("g.InvoiceAdjustmentRepo.FindByInvoiceIDs err: %v", err))
	}

	// Map invoices with bill items
	for _, e := range invoiceAdjustments {
		adjustmentList, ok := m[e.InvoiceID.String]
		if ok {
			m[e.InvoiceID.String] = append(adjustmentList, e)
		} else {
			m[e.InvoiceID.String] = []*entities.InvoiceAdjustment{e}
		}
	}

	return m, nil
}

func (g *BasePaymentRequestGenerator) getInvoiceBillItemMap(
	ctx context.Context,
	invoiceIDs []string,
) (map[string][]*entities.InvoiceBillItemMap, error) {
	m := make(map[string][]*entities.InvoiceBillItemMap)

	// Find bill item by invoice IDs
	invoiceBillItemMap, err := g.BillItemRepo.FindInvoiceBillItemMapByInvoiceIDs(ctx, g.DB, invoiceIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("g.BillItemRepo.FindInvoiceBillItemMapByInvoiceIDs err: %v", err))
	}

	// Map invoices with bill items
	for _, e := range invoiceBillItemMap {
		billItemList, ok := m[e.InvoiceID.String]
		if ok {
			m[e.InvoiceID.String] = append(billItemList, e)
		} else {
			m[e.InvoiceID.String] = []*entities.InvoiceBillItemMap{e}
		}
	}

	return m, nil
}

func GetTimeInJST(t time.Time) (time.Time, error) {
	timezone := "Asia/Tokyo"
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}

	return t.In(location), nil
}

func GetFloat64ExactValueAndDecimalPlaces(amount pgtype.Numeric, decimal string) (float64, error) {
	var floatAmount float64
	err := amount.AssignTo(&floatAmount)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, err.Error())
	}

	getExactValueWithDecimalPlaces, err := strconv.ParseFloat(fmt.Sprintf("%."+decimal+"f", floatAmount), 64)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, err.Error())
	}

	return getExactValueWithDecimalPlaces, nil
}

func writeWhiteSpace(b *bytes.Buffer) error {
	_, err := b.WriteString("\n")
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Error on writing white space err: %v", err))
	}

	return nil
}

type fileStorageUploadInfo struct {
	ObjectName      string
	PathName        string
	DownloadFileURL string
	TemporaryFile   *utils.TempFile
}
