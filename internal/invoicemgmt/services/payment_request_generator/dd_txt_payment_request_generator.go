package generator

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/filestorage"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type bankHeaderRecord struct {
	DataCategory     string
	TypeCode         string
	CodeCategory     string
	ConsignorCode    string
	ConsignorName    string
	DepositDate      string
	BankNumber       string
	BankName         string
	BankBranchNumber string
	BankBranchName   string
	DepositItems     string
	AccountNumber    string
	Dummy            string
}

type bankDataRecord struct {
	DataCategory            string
	DepositBankNumber       string
	DepositBankName         string
	DepositBankBranchNumber string
	DepositBankBranchName   string
	Dummy1                  string
	DepositItems            string
	AccountNumber           string
	AccountOwnerName        string
	DepositAmount           string
	NewCustomerCode         string
	CustomerNumber          string
	ResultCode              string
	Dummy2                  string
}

type bankTrailerRecord struct {
	DataCategory      string
	TotalTransactions string
	TotalAmount       string
	TransferredNumber string
	TransferredAmount string
	FailedNumber      string
	FailedAmount      string
	Dummy             string
}

type bankEndRecord struct {
	DataCategory string
	Dummy        string
}

type paymentFileDetailsPerBank struct {
	dataMap        []*dataMap
	maxRecordLimit int32
}

type DDtxtPaymentRequestGenerator struct {
	*BasePaymentRequestGenerator
	Req                  *invoice_pb.CreatePaymentRequestRequest
	BulkPaymentRequestID string

	paymentAndFileAssocs []paymentAndFileAssoc
	bankDataMap          map[string]*paymentFileDetailsPerBank

	EnableEncodePaymentRequestFiles      bool
	EnableFormatPaymentRequestFileFields bool
}

func (g *DDtxtPaymentRequestGenerator) ValidateData(ctx context.Context) error {
	// Fetch payment and invoice
	paymentInvoice, err := g.getPaymentInvoice(ctx, g.Req.PaymentIds)
	if err != nil {
		return err
	}

	// Validate payment and invoice and return list of student IDs
	studentIDs, err := g.getListOfStudentsFromPaymentInvoice(paymentInvoice, g.Req.PaymentMethod.String(), 20, 10)
	if err != nil {
		return err
	}

	// Map the students to its bank account detail
	studentBankAccountMap, bankBranchIDs, err := g.getStudentBankAccountMap(ctx, studentIDs)
	if err != nil {
		return err
	}

	// Get the mappings of bank branch to bank and partner bank
	branchRelatedBankMap, err := g.getRelatedBankMap(ctx, bankBranchIDs)
	if err != nil {
		return err
	}

	bankAccountNumbers := []string{}
	for _, bankAccount := range studentBankAccountMap {
		bankAccountNumbers = append(bankAccountNumbers, bankAccount.BankAccount.BankAccountNumber.String)
	}

	existingAccountNumber, studentCCMap, err := g.getCustomerCodeMaps(ctx, studentIDs, bankAccountNumbers)
	if err != nil {
		return err
	}

	bankDataMap := make(map[string]*paymentFileDetailsPerBank)
	bankMaxRecordMap := make(map[string]int32)
	for _, e := range paymentInvoice {

		// If invoice total is negative, do not include the payment
		if isNegative(e.Invoice.Total) {
			continue
		}

		// Validate the bank and get the related bank of the student bank branch
		studentBank, relatedBank, err := g.validateAndGetRelatedBank(studentBankAccountMap, branchRelatedBankMap, e.Payment.StudentID.String)
		if err != nil {
			return err
		}

		d := &dataMap{
			Payment:                       e.Payment,
			Invoice:                       e.Invoice,
			BankAccount:                   studentBank.BankAccount,
			AccountNumberWithCustomerCode: existingAccountNumber,
		}

		if _, ok := bankDataMap[relatedBank.PartnerBank.BankName.String]; ok {
			bankDataMap[relatedBank.PartnerBank.BankName.String].dataMap = append(bankDataMap[relatedBank.PartnerBank.BankName.String].dataMap, d)
		} else {
			bankDataMap[relatedBank.PartnerBank.BankName.String] = &paymentFileDetailsPerBank{
				dataMap:        []*dataMap{d},
				maxRecordLimit: relatedBank.PartnerBank.RecordLimit.Int,
			}
		}

		bankMaxRecordMap[relatedBank.PartnerBank.BankName.String] = relatedBank.PartnerBank.RecordLimit.Int

		if cc, ok := studentCCMap[e.Payment.StudentID.String]; ok {
			d.NewCustomerCodeHistory = cc
		}
	}

	g.bankDataMap = bankDataMap

	return nil
}

func (g *DDtxtPaymentRequestGenerator) PlanPaymentAndFileAssociation(ctx context.Context) error {
	// Create different payment file for different banks
	for bankName, details := range g.bankDataMap {
		baseFileName := fmt.Sprintf(
			"Payment_DD_%s_%s",
			g.Req.DirectDebitDates.DueDate.AsTime().Format(fileNameTimeFormat),
			bankName,
		)

		shouldSetMaxRecordLimit, err := g.UnleashClient.IsFeatureEnabled(constant.EnableBulkAddValidatePh2, g.Env)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableKECFeedbackPh1, err))
		}

		maxContent := maximumPaymentPerTXT
		if shouldSetMaxRecordLimit {
			maxContent = int(details.maxRecordLimit)
		}

		g.paymentAndFileAssocs = append(g.paymentAndFileAssocs, generatePaymentAndFileAssocs(details.dataMap, maxContent, baseFileName, txtFormat)...)
	}

	return nil
}

func (g *DDtxtPaymentRequestGenerator) SavePaymentAndFileAssociation(ctx context.Context) error {
	err := g.associatePaymentToFile(ctx, g.BulkPaymentRequestID, invoice_pb.PaymentMethod_DIRECT_DEBIT.String(), g.paymentAndFileAssocs)
	if err != nil {
		return err
	}

	return nil
}

func (g *DDtxtPaymentRequestGenerator) getStudentBankAccountMap(ctx context.Context, studentIDs []string) (map[string]*entities.StudentBankDetailsMap, []string, error) {
	studentBankAccountDetails, err := g.StudentPaymentDetailRepo.FindStudentBankDetailsByStudentIDs(ctx, g.DB, studentIDs)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, fmt.Sprintf("g.StudentPaymentDetailRepo.FindStudentBankDetailsByStudentIDs err: %v", err))
	}

	studentBankAccountMap := make(map[string]*entities.StudentBankDetailsMap)
	bankBranchIDs := []string{}
	for _, e := range studentBankAccountDetails {

		// Only the first bank account will be used if student has multiple
		_, ok := studentBankAccountMap[e.StudentPaymentDetail.StudentID.String]
		if ok {
			continue
		}

		studentBankAccountMap[e.StudentPaymentDetail.StudentID.String] = e
		bankBranchIDs = append(bankBranchIDs, e.BankAccount.BankBranchID.String)
	}

	return studentBankAccountMap, bankBranchIDs, nil
}

func (g *DDtxtPaymentRequestGenerator) getRelatedBankMap(ctx context.Context, bankBranchIDs []string) (map[string]*entities.BankRelationMap, error) {
	relatedBankOfBankBranch, err := g.BankBranchRepo.FindRelatedBankOfBankBranches(ctx, g.DB, bankBranchIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("g.BankBranchRepo.FindRelatedBankOfBankBranches err: %v", err))
	}

	branchRelatedBankMap := make(map[string]*entities.BankRelationMap)
	for _, e := range relatedBankOfBankBranch {

		// If bank branch has multiple partner bank, use the default partner bank
		// If bank branch has only 1 mapped partner bank, use it even is_default is false
		// If no default partner bank in multiple partner bank, use the last partner bank

		existingBank, exists := branchRelatedBankMap[e.BankBranch.BankBranchID.String]
		if exists {

			// If the existing partner bank in branchRelatedBankMap is already the default, continue
			if existingBank.PartnerBank.IsDefault.Bool {
				continue
			}

			// Check if the partner bank is default
			if e.PartnerBank.IsDefault.Bool {
				branchRelatedBankMap[e.BankBranch.BankBranchID.String] = e
				continue
			}
		}

		branchRelatedBankMap[e.BankBranch.BankBranchID.String] = e
	}

	return branchRelatedBankMap, nil
}

func (g *DDtxtPaymentRequestGenerator) validateAndGetRelatedBank(
	studentBankAccountMap map[string]*entities.StudentBankDetailsMap,
	branchRelatedBankMap map[string]*entities.BankRelationMap,
	studentID string,
) (*entities.StudentBankDetailsMap, *entities.BankRelationMap, error) {
	// Validate if each student in paymentInvoice have billing details
	studentBank, ok := studentBankAccountMap[studentID]
	if !ok {
		return nil, nil, status.Error(codes.Internal, "There is a student that do not have bank account")
	}

	// Currently it on use the first billing address of a student.
	// If we will support multiple bank account, we can add a condition to select here
	relatedBank, ok := branchRelatedBankMap[studentBank.BankAccount.BankBranchID.String]
	if !ok {
		return nil, nil, status.Error(codes.Internal, "Student bank has no related bank")
	}

	err := multierr.Combine(
		g.Validator.ValidateStudentPaymentDetail(studentBank.StudentPaymentDetail),
		g.Validator.ValidateBankAccount(studentBank.BankAccount),
		g.Validator.ValidateBankBranch(relatedBank.BankBranch),
		g.Validator.ValidateBank(relatedBank.Bank),
		g.Validator.ValidatePartnerBank(relatedBank.PartnerBank),
	)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	return studentBank, relatedBank, nil
}

func (g *DDtxtPaymentRequestGenerator) getCustomerCodeMaps(ctx context.Context, studentIDs []string, bankAccountNumber []string) (map[string]struct{}, map[string]*entities.NewCustomerCodeHistory, error) {
	existingAccountCC, err := g.NewCustomerCodeHistoryRepo.FindByAccountNumbers(ctx, g.DB, bankAccountNumber)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, fmt.Sprintf("g.NewCustomerCodeHistoryRepo.FindByAccountNumbers err :%v", err))
	}

	existingAccountNumber := make(map[string]struct{})
	for _, cc := range existingAccountCC {
		existingAccountNumber[cc.BankAccountNumber.String] = struct{}{}
	}

	studentCC, err := g.NewCustomerCodeHistoryRepo.FindByStudentIDs(ctx, g.DB, studentIDs)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, fmt.Sprintf("g.NewCustomerCodeHistoryRepo.FindByStudentIDs err :%v", err))
	}

	studentCCMap := make(map[string]*entities.NewCustomerCodeHistory)
	for _, e := range studentCC {
		studentCCMap[e.StudentID.String] = e
	}

	return existingAccountNumber, studentCCMap, nil
}

func (g *DDtxtPaymentRequestGenerator) ValidateDataV2(ctx context.Context) error {
	// Fetch payment and invoice
	paymentInvoice, err := g.getPaymentInvoice(ctx, g.Req.PaymentIds)
	if err != nil {
		return err
	}

	// Validate payment and invoice and return list of student IDs
	studentIDs, err := g.getListOfStudentsFromPaymentInvoice(paymentInvoice, invoice_pb.PaymentMethod_DIRECT_DEBIT.String(), 20, 10)
	if err != nil {
		return err
	}

	// Map the students to its bank account detail
	studentBankAccountMap, bankBranchIDs, err := g.getStudentBankAccountMap(ctx, studentIDs)
	if err != nil {
		return err
	}

	// Get the mappings of bank branch to bank and partner bank
	branchRelatedBankMap, err := g.getRelatedBankMap(ctx, bankBranchIDs)
	if err != nil {
		return err
	}

	bankAccountNumbers := []string{}
	for _, bankAccount := range studentBankAccountMap {
		bankAccountNumbers = append(bankAccountNumbers, bankAccount.BankAccount.BankAccountNumber.String)
	}

	existingAccountNumber, studentCCMap, err := g.getCustomerCodeMaps(ctx, studentIDs, bankAccountNumbers)
	if err != nil {
		return err
	}

	bankDataMap := make(map[string]*paymentFileDetailsPerBank)
	for _, e := range paymentInvoice {
		// If invoice total is negative, do not include the payment
		if isNegative(e.Invoice.Total) {
			continue
		}

		// Validate the bank and get the related bank of the student bank branch
		studentBank, relatedBank, err := g.validateAndGetRelatedBank(studentBankAccountMap, branchRelatedBankMap, e.Payment.StudentID.String)
		if err != nil {
			return err
		}

		d := &dataMap{
			Payment:                       e.Payment,
			Invoice:                       e.Invoice,
			BankAccount:                   studentBank.BankAccount,
			AccountNumberWithCustomerCode: existingAccountNumber,
			StudentBankDetails:            studentBank,
			StudentRelatedBank:            relatedBank,
		}

		if _, ok := bankDataMap[relatedBank.PartnerBank.BankName.String]; ok {
			bankDataMap[relatedBank.PartnerBank.BankName.String].dataMap = append(bankDataMap[relatedBank.PartnerBank.BankName.String].dataMap, d)
		} else {
			bankDataMap[relatedBank.PartnerBank.BankName.String] = &paymentFileDetailsPerBank{
				dataMap:        []*dataMap{d},
				maxRecordLimit: relatedBank.PartnerBank.RecordLimit.Int,
			}
		}

		if cc, ok := studentCCMap[e.Payment.StudentID.String]; ok {
			d.NewCustomerCodeHistory = cc
		}
	}

	g.bankDataMap = bankDataMap

	return nil
}

func (g *DDtxtPaymentRequestGenerator) SaveAndUploadPaymentFileV2(ctx context.Context) error {
	err := database.ExecInTx(ctx, g.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		for _, paymentFileAssoc := range g.paymentAndFileAssocs {
			// check and upsert first for new customer code to be mapped
			for j, data := range paymentFileAssoc.DataMap {
				newCustomerCodeHistory, err := g.upsertCustomerCode(ctx, tx, data.Payment.StudentID.String, data)
				if err != nil {
					return err
				}

				paymentFileAssoc.DataMap[j].NewCustomerCodeHistory = newCustomerCodeHistory
			}

			byteContent, err := g.GetByteContent(ctx, paymentFileAssoc.DataMap)
			if err != nil {
				return err
			}

			content := byteContent
			if g.EnableEncodePaymentRequestFiles {
				// Encode to Shift JIS
				content, err = utils.EncodeByteToShiftJIS(byteContent)
				if err != nil {
					return status.Error(codes.Internal, err.Error())
				}
			}

			objectUploader, err := utils.NewObjectUploader(
				g.FileStorage,
				g.TempFileCreator,
				&utils.ObjectInfo{
					ObjectName:  fmt.Sprintf("%s-%s", g.BulkPaymentRequestID, paymentFileAssoc.FileName),
					ByteContent: content,
					ContentType: filestorage.ContentTypeTXT,
				},
			)
			defer func() {
				err := objectUploader.Close()
				if err != nil {
					g.Logger.Warn(err)
				}
			}()
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("cannot initialize file storage upload: %v on file name: %v with bulk payment request id: %v", err, paymentFileAssoc.FileName, g.BulkPaymentRequestID))
			}

			e, err := generateBulkPaymentRequestFileEntityV2(g.BulkPaymentRequestID, objectUploader.GetDownloadFileURL(), paymentFileAssoc)
			if err != nil {
				return err
			}

			// Create the payment request file entity
			_, err = g.BulkPaymentRequestFileRepo.Create(ctx, tx, e)
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("s.BulkPaymentRequestFileRepo.Create err: %v", err))
			}

			paymentIDs := make([]string, len(paymentFileAssoc.DataMap))
			invoiceIDs := make([]string, len(paymentFileAssoc.DataMap))
			bulkPaymentIDs := make([]string, 0)

			for i, data := range paymentFileAssoc.DataMap {
				// entities' ids to update
				paymentIDs[i] = data.Payment.PaymentID.String
				invoiceIDs[i] = data.Invoice.InvoiceID.String
				// check if payment is belong to a bulk payment and not yet mapped
				if data.Payment.BulkPaymentID.Status == pgtype.Present && !slices.Contains(bulkPaymentIDs, data.Payment.BulkPaymentID.String) {
					bulkPaymentIDs = append(bulkPaymentIDs, data.Payment.BulkPaymentID.String)
				}
			}

			if err := g.InvoiceRepo.UpdateIsExportedByInvoiceIDs(ctx, tx, invoiceIDs, true); err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("error InvoiceRepo UpdateIsExportedByInvoiceIDs: %v", err))
			}

			if err := g.PaymentRepo.UpdateIsExportedByPaymentIDs(ctx, tx, paymentIDs, true); err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("error PaymentRepo UpdateIsExportedByPaymentIDs: %v", err))
			}

			if len(bulkPaymentIDs) > 0 {
				if err := g.BulkPaymentRepo.UpdateBulkPaymentStatusByIDs(ctx, tx, invoice_pb.BulkPaymentStatus_BULK_PAYMENT_EXPORTED.String(), bulkPaymentIDs); err != nil {
					return status.Error(codes.Internal, fmt.Sprintf("error BulkPaymentRepo UpdateBulkPaymentStatusByIDs: %v", err))
				}
			}

			// upload to gcloud storage
			if err := objectUploader.DoUploadFile(ctx); err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("objectUploader.DoUploadFile error: %v with object name: %v", err, objectUploader.GetFormattedObjectName()))
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (g *DDtxtPaymentRequestGenerator) SaveAndUploadPaymentFile(ctx context.Context) error {
	err := database.ExecInTx(ctx, g.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		for _, paymentFileAssoc := range g.paymentAndFileAssocs {
			// check and upsert first for new customer code to be mapped
			for j, data := range paymentFileAssoc.DataMap {
				newCustomerCodeHistory, err := g.upsertCustomerCode(ctx, tx, data.Payment.StudentID.String, data)
				if err != nil {
					return err
				}

				paymentFileAssoc.DataMap[j].NewCustomerCodeHistory = newCustomerCodeHistory
			}

			byteContent, err := g.GetByteContent(ctx, paymentFileAssoc.DataMap)
			if err != nil {
				g.Logger.Warnf("cannot convert byte content: %v on file name: %v", err, paymentFileAssoc.FileName)
			}
			// initialize file URL to be empty
			var fileURL string

			objectUploader, err := utils.NewObjectUploader(
				g.FileStorage,
				g.TempFileCreator,
				&utils.ObjectInfo{
					ObjectName:  fmt.Sprintf("%s-%s", g.BulkPaymentRequestID, paymentFileAssoc.FileName),
					ByteContent: byteContent,
					ContentType: filestorage.ContentTypeTXT,
				},
			)
			defer func() {
				err := objectUploader.Close()
				if err != nil {
					g.Logger.Warn(err)
				}
			}()
			if err != nil {
				g.Logger.Warnf("cannot initialize file storage upload: %v on file name: %v with bulk payment request id: %v", err, paymentFileAssoc.FileName, g.BulkPaymentRequestID)
			} else {
				fileURL = objectUploader.GetDownloadFileURL()
			}

			e, err := generateBulkPaymentRequestFileEntityV2(g.BulkPaymentRequestID, fileURL, paymentFileAssoc)
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
			}

			err = g.PaymentRepo.UpdateIsExportedByPaymentRequestFileID(ctx, tx, requestFileID, true)
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("g.PaymentRepo.UpdateIsExportedByPaymentRequestFileID err: %v", err))
			}

			err = g.InvoiceRepo.UpdateIsExportedByPaymentRequestFileID(ctx, tx, requestFileID, true)
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("g.InvoiceRepo.UpdateIsExportedByPaymentRequestFileID err: %v", err))
			}

			// upload to gcloud storage
			if err := objectUploader.DoUploadFile(ctx); err != nil {
				g.Logger.Warnf("objectUploader.DoUploadFile error: %v with object name: %v", err, objectUploader.GetFormattedObjectName())
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (g *DDtxtPaymentRequestGenerator) GetByteContent(ctx context.Context, dataMapList []*dataMap) ([]byte, error) {
	byteData, err := g.GenTxtBankContent(dataMapList)
	if err != nil {
		return nil, err
	}

	return byteData, nil
}

func (g *DDtxtPaymentRequestGenerator) GenTxtBankContent(filePaymentDataList []*dataMap) ([]byte, error) {
	if g.EnableFormatPaymentRequestFileFields {
		return g.GenTxtBankContentV2(filePaymentDataList)
	}

	var b bytes.Buffer
	// write header record to buffer
	err := writeFileHeaderRecord(&b, filePaymentDataList)
	if err != nil {
		return nil, err
	}

	// write data records to buffer
	totalAmount, err := writeFileDataRecords(&b, filePaymentDataList)
	if err != nil {
		return nil, err
	}

	// write trailer record to buffer
	err = writeFileTrailerRecord(&b, len(filePaymentDataList), totalAmount)
	if err != nil {
		return nil, err
	}
	// write trailer record to buffer
	err = writeFileEndRecord(&b)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func writeFileHeaderRecord(b *bytes.Buffer, filePaymentInvoiceMap []*dataMap) error {
	// Use the first payment to get the due date
	// All payments in the file have the same due date
	dueDateJST, err := GetTimeInJST(filePaymentInvoiceMap[0].Payment.PaymentDueDate.Time)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("getCurrentTimeInJST err: %v", err))
	}
	dueDateStr := dueDateJST.Format("0102")

	// A file consist of the same partner bank
	studentPartnerBank := filePaymentInvoiceMap[0].StudentRelatedBank.PartnerBank

	// get the equivalent deposit item value for partner bank which is already validated
	var depositItems string
	for index, value := range constant.PartnerBankDepositItems {
		if value == studentPartnerBank.DepositItems.String {
			depositItems = strconv.Itoa(index)
		}
	}

	headerRecord := bankHeaderRecord{
		DataCategory:     "1",
		TypeCode:         "91",
		CodeCategory:     "0",
		ConsignorCode:    utils.AddPrefixStringWithLimit(studentPartnerBank.ConsignorCode.String, "0", 10),
		ConsignorName:    utils.AddSuffixStringWithLimit(studentPartnerBank.ConsignorName.String, " ", 40),
		DepositDate:      dueDateStr,
		BankNumber:       utils.AddSuffixStringWithLimit(studentPartnerBank.BankNumber.String, "0", 4),
		BankName:         utils.AddSuffixStringWithLimit(studentPartnerBank.BankName.String, " ", 15),
		BankBranchNumber: utils.AddPrefixStringWithLimit(studentPartnerBank.BankBranchNumber.String, "0", 3),
		BankBranchName:   utils.AddSuffixStringWithLimit(studentPartnerBank.BankBranchName.String, " ", 15),
		DepositItems:     utils.AddSuffixStringWithLimit(depositItems, " ", 1),
		AccountNumber:    utils.AddPrefixStringWithLimit(studentPartnerBank.AccountNumber.String, "0", 7),
		Dummy:            utils.AddPrefixString("", " ", 17), // 17 spaces
	}

	content := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s",
		headerRecord.DataCategory,
		headerRecord.TypeCode,
		headerRecord.CodeCategory,
		headerRecord.ConsignorCode,
		headerRecord.ConsignorName,
		headerRecord.DepositDate,
		headerRecord.BankNumber,
		headerRecord.BankName,
		headerRecord.BankBranchNumber,
		headerRecord.BankBranchName,
		headerRecord.DepositItems,
		headerRecord.AccountNumber,
		headerRecord.Dummy,
	)

	_, err = b.WriteString(content)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Error on writing header record err: %v", err))
	}

	err = writeWhiteSpace(b)
	if err != nil {
		return err
	}

	return nil
}

func writeFileDataRecords(b *bytes.Buffer, filePaymentInvoiceMap []*dataMap) (int64, error) {
	dataRecords := []bankDataRecord{}
	totalAmount := int64(0)
	for _, filePaymentInvoice := range filePaymentInvoiceMap {
		exactTotal, err := GetFloat64ExactValueAndDecimalPlaces(filePaymentInvoice.Invoice.Total, "2")
		if err != nil {
			return 0, err
		}
		totalAmount += int64(exactTotal)

		paymentSeqNumStr := strconv.Itoa(int(filePaymentInvoice.Payment.PaymentSequenceNumber.Int))
		exactTotalStr := strconv.FormatInt(int64(exactTotal), 10)

		// validate the length of payment sequence number
		if len(paymentSeqNumStr) > 20 {
			return 0, status.Error(codes.Internal, "The payment sequence number length exceeds the limit")
		}

		// validate the length total amount
		if len(exactTotalStr) > 10 {
			return 0, status.Error(codes.Internal, "The invoice total length exceeds the limit")
		}

		// get the equivalent deposit item value for bank account type which is already validated
		studentRelatedBank := filePaymentInvoice.StudentRelatedBank
		studentBankAccount := filePaymentInvoice.StudentBankDetails
		var depositItems string
		for index, value := range constant.PartnerBankDepositItems {
			if value == studentBankAccount.BankAccount.BankAccountType.String {
				depositItems = strconv.Itoa(index)
			}
		}

		dataRecord := bankDataRecord{
			DataCategory:            "2",
			DepositBankNumber:       utils.AddPrefixStringWithLimit(studentRelatedBank.Bank.BankCode.String, "0", 4),
			DepositBankName:         utils.AddSuffixStringWithLimit(studentRelatedBank.Bank.BankNamePhonetic.String, " ", 15),
			DepositBankBranchNumber: utils.AddPrefixStringWithLimit(studentRelatedBank.BankBranch.BankBranchCode.String, "0", 3),
			DepositBankBranchName:   utils.AddSuffixStringWithLimit(studentRelatedBank.BankBranch.BankBranchPhoneticName.String, " ", 15),
			Dummy1:                  utils.AddPrefixString("", " ", 4), // 4 spaces
			DepositItems:            utils.LimitString(depositItems, 1),
			AccountNumber:           utils.AddPrefixStringWithLimit(studentBankAccount.BankAccount.BankAccountNumber.String, "0", 7),
			AccountOwnerName:        utils.AddSuffixStringWithLimit(studentBankAccount.BankAccount.BankAccountHolder.String, " ", 30),
			DepositAmount:           utils.AddPrefixStringWithLimit(exactTotalStr, "0", 10),
			NewCustomerCode:         utils.LimitString(filePaymentInvoice.NewCustomerCodeHistory.NewCustomerCode.String, 1),
			CustomerNumber:          utils.AddSuffixStringWithLimit(paymentSeqNumStr, " ", 20),
			ResultCode:              "0",                               // always 0
			Dummy2:                  utils.AddPrefixString("", " ", 8), // 8 spaces
		}

		dataRecords = append(dataRecords, dataRecord)
	}

	// write data record to buffer
	for _, dataRecord := range dataRecords {
		content := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s",
			dataRecord.DataCategory,
			dataRecord.DepositBankNumber,
			dataRecord.DepositBankName,
			dataRecord.DepositBankBranchNumber,
			dataRecord.DepositBankBranchName,
			dataRecord.Dummy1,
			dataRecord.DepositItems,
			dataRecord.AccountNumber,
			dataRecord.AccountOwnerName,
			dataRecord.DepositAmount,
			dataRecord.NewCustomerCode,
			dataRecord.CustomerNumber,
			dataRecord.ResultCode,
			dataRecord.Dummy2,
		)

		_, err := b.WriteString(content)
		if err != nil {
			return 0, status.Error(codes.Internal, fmt.Sprintf("Error on writing data record err: %v", err))
		}

		err = writeWhiteSpace(b)
		if err != nil {
			return 0, err
		}
	}

	return totalAmount, nil
}

func writeFileTrailerRecord(b *bytes.Buffer, totalTransactions int, totalAmount int64) error {
	totalAmountStr := strconv.FormatInt(totalAmount, 10)
	totalTransactionsStr := strconv.Itoa(totalTransactions)
	if len(totalAmountStr) > 12 {
		return status.Error(codes.Internal, "The sum of invoices length exceeds the limit")
	}

	if len(totalTransactionsStr) > 6 {
		return status.Error(codes.Internal, "The total transactions length exceeds the limit")
	}

	trailerRecord := bankTrailerRecord{
		DataCategory:      "8",
		TotalTransactions: utils.AddPrefixStringWithLimit(totalTransactionsStr, "0", 6),
		TotalAmount:       utils.AddPrefixStringWithLimit(totalAmountStr, "0", 12),
		TransferredNumber: utils.AddPrefixString("", "0", 6),  // 6 zeroes
		TransferredAmount: utils.AddPrefixString("", "0", 12), // 12 zeroes
		FailedNumber:      utils.AddPrefixString("", "0", 6),  // 6 zeroes
		FailedAmount:      utils.AddPrefixString("", "0", 12), // 12 zeroes
		Dummy:             utils.AddPrefixString("", " ", 65), // 65 spaces
	}

	content := fmt.Sprintf("%s%s%s%s%s%s%s%s",
		trailerRecord.DataCategory,
		trailerRecord.TotalTransactions,
		trailerRecord.TotalAmount,
		trailerRecord.TransferredNumber,
		trailerRecord.TransferredAmount,
		trailerRecord.FailedNumber,
		trailerRecord.FailedAmount,
		trailerRecord.Dummy,
	)

	_, err := b.WriteString(content)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Error on writing trailer record err: %v", err))
	}

	err = writeWhiteSpace(b)
	if err != nil {
		return err
	}

	return nil
}

func writeFileEndRecord(b *bytes.Buffer) error {
	endRecord := bankEndRecord{
		DataCategory: "9",
		Dummy:        utils.AddPrefixString("", " ", 119), // 119 spaces
	}

	content := fmt.Sprintf("%s%s",
		endRecord.DataCategory,
		endRecord.Dummy,
	)

	_, err := b.WriteString(content)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Error on writing end record err: %v", err))
	}

	return nil
}
