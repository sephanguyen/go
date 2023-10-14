package downloader

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/filestorage"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
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

type DirectDebitTXTPaymentFileDownloader struct {
	*BasePaymentFileDownloader
	PaymentFileID       string
	filePaymentDataList []*FilePaymentDataMap
}

func (d *DirectDebitTXTPaymentFileDownloader) ValidateData(ctx context.Context) error {
	// Get the payments and its invoices associated in a payment file
	filePaymentInvoices, err := d.getAndValidateFilePaymentInvoice(ctx, d.PaymentFileID)
	if err != nil {
		return err
	}
	// Validate payment and invoice and return list of student IDs
	studentIDs, err := d.getListOfStudentsFromPaymentInvoice(filePaymentInvoices, invoice_pb.PaymentMethod_DIRECT_DEBIT.String(), 20, 10)
	if err != nil {
		return err
	}
	// Map the students to its bank account detail
	studentBankAccountMap, bankBranchIDs, err := d.getStudentBankAccountMap(ctx, studentIDs)
	if err != nil {
		return err
	}
	// // Get the mappings of bank branch to bank and partner bank
	branchRelatedBankMap, err := d.getRelatedBankMap(ctx, bankBranchIDs)
	if err != nil {
		return err
	}

	studentCCMap, err := d.getCustomerCodes(ctx, studentIDs)
	if err != nil {
		return err
	}

	dataMapList := []*FilePaymentDataMap{}

	// wrap in transaction since there's creating of new customer code
	err = database.ExecInTx(ctx, d.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		for _, e := range filePaymentInvoices {
			// Validate the bank and get the related bank of the student bank branch
			studentBank, relatedBank, err := d.validateAndGetRelatedBank(studentBankAccountMap, branchRelatedBankMap, e.Payment.StudentID.String)
			if err != nil {
				return err
			}
			var newCustomerCodeHistory *entities.NewCustomerCodeHistory

			if _, ok := studentCCMap[e.Payment.StudentID.String]; !ok {
				// if no existing new customer code history for student, create one
				newCustomerCodeCreated, err := d.createNewCustomerCodeHistoryIfNotExist(ctx, tx, e.Payment.StudentID.String, studentBank.BankAccount.BankAccountNumber.String)
				if err != nil {
					return status.Error(codes.Internal, err.Error())
				}
				newCustomerCodeHistory = newCustomerCodeCreated
			} else {
				for _, cc := range studentCCMap[e.Payment.StudentID.String] {
					if cc.BankAccountNumber == studentBank.BankAccount.BankAccountNumber {
						// assign the customer code history
						newCustomerCodeHistory = cc
					}
				}
				if newCustomerCodeHistory == nil {
					// student has only other new customer code history records for other bank account number in the future
					// if no existing new customer code history for student on that bank account number, create one
					newCustomerCodeCreated, err := d.createNewCustomerCodeHistoryIfNotExist(ctx, tx, e.Payment.StudentID.String, studentBank.BankAccount.BankAccountNumber.String)

					if err != nil {
						return status.Error(codes.Internal, err.Error())
					}

					newCustomerCodeHistory = newCustomerCodeCreated
				}
			}

			dataMapList = append(dataMapList, &FilePaymentDataMap{
				Payment:                e.Payment,
				Invoice:                e.Invoice,
				StudentBankDetails:     studentBank, // Bank Account Details
				StudentRelatedBank:     relatedBank, // Bank, Bank Branch and Partner Bank
				NewCustomerCodeHistory: newCustomerCodeHistory,
			})
		}
		return nil
	})

	if err != nil {
		return err
	}

	d.filePaymentDataList = dataMapList

	return nil
}

func (d *DirectDebitTXTPaymentFileDownloader) GetByteContent(ctx context.Context) ([]byte, error) {
	byteData, err := GenTxtBankContent(d.filePaymentDataList)
	if err != nil {
		return nil, err
	}

	return byteData, nil
}

func (d *DirectDebitTXTPaymentFileDownloader) getStudentBankAccountMap(ctx context.Context, studentIDs []string) (map[string]*entities.StudentBankDetailsMap, []string, error) {
	studentBankAccountDetails, err := d.StudentPaymentDetailRepo.FindStudentBankDetailsByStudentIDs(ctx, d.DB, studentIDs)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, fmt.Sprintf("d.StudentPaymentDetailRepo.FindStudentBankDetailsByStudentIDs err: %v", err))
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

func (d *DirectDebitTXTPaymentFileDownloader) getRelatedBankMap(ctx context.Context, bankBranchIDs []string) (map[string]*entities.BankRelationMap, error) {
	relatedBankOfBankBranch, err := d.BankBranchRepo.FindRelatedBankOfBankBranches(ctx, d.DB, bankBranchIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("d.BankBranchRepo.FindRelatedBankOfBankBranches err: %v", err))
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

func (d *DirectDebitTXTPaymentFileDownloader) validateAndGetRelatedBank(
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
		d.Validator.ValidateStudentPaymentDetail(studentBank.StudentPaymentDetail),
		d.Validator.ValidateBankAccount(studentBank.BankAccount),
		d.Validator.ValidateBankBranch(relatedBank.BankBranch),
		d.Validator.ValidateBank(relatedBank.Bank),
		d.Validator.ValidatePartnerBank(relatedBank.PartnerBank),
	)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	return studentBank, relatedBank, nil
}

func (d *DirectDebitTXTPaymentFileDownloader) getCustomerCodes(ctx context.Context, studentIDs []string) (map[string][]*entities.NewCustomerCodeHistory, error) {
	studentCC, err := d.NewCustomerCodeHistoryRepo.FindByStudentIDs(ctx, d.DB, studentIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("d.NewCustomerCodeHistoryRepo.FindByStudentIDs err :%v", err))
	}

	studentCCMap := make(map[string][]*entities.NewCustomerCodeHistory)
	for _, e := range studentCC {
		if cc, ok := studentCCMap[e.StudentID.String]; ok {
			studentCCMap[e.StudentID.String] = append(cc, e)
		} else {
			studentCCMap[e.StudentID.String] = []*entities.NewCustomerCodeHistory{e}
		}
	}

	return studentCCMap, nil
}

func (d *DirectDebitTXTPaymentFileDownloader) createNewCustomerCodeHistoryIfNotExist(ctx context.Context, db database.QueryExecer, studentID string, accountNumber string) (*entities.NewCustomerCodeHistory, error) {
	e := &entities.NewCustomerCodeHistory{}
	database.AllNullEntity(e)
	err := multierr.Combine(
		e.StudentID.Set(studentID),
		e.BankAccountNumber.Set(accountNumber),
		e.NewCustomerCode.Set("1"),
	)
	if err != nil {
		return nil, err
	}
	err = d.NewCustomerCodeHistoryRepo.Create(ctx, db, e)
	if err != nil {
		return nil, fmt.Errorf("g.NewCustomerCodeHistoryRepo.Create err %v", err)
	}

	return e, nil
}

func GenTxtBankContent(filePaymentDataList []*FilePaymentDataMap) ([]byte, error) {
	var b bytes.Buffer
	// write header record to buffer
	err := writeHeaderRecord(&b, filePaymentDataList)
	if err != nil {
		return nil, err
	}

	// write data records to buffer
	totalAmount, err := writeDataRecords(&b, filePaymentDataList)
	if err != nil {
		return nil, err
	}

	// write trailer record to buffer
	err = writeTrailerRecord(&b, len(filePaymentDataList), totalAmount)
	if err != nil {
		return nil, err
	}
	// write trailer record to buffer
	err = writeEndRecord(&b)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func writeHeaderRecord(b *bytes.Buffer, filePaymentInvoiceMap []*FilePaymentDataMap) error {
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

func writeDataRecords(b *bytes.Buffer, filePaymentInvoiceMap []*FilePaymentDataMap) (int64, error) {
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

func writeTrailerRecord(b *bytes.Buffer, totalTransactions int, totalAmount int64) error {
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

func writeEndRecord(b *bytes.Buffer) error {
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

func writeWhiteSpace(b *bytes.Buffer) error {
	_, err := b.WriteString("\n")
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Error on writing white space err: %v", err))
	}

	return nil
}

func (d *DirectDebitTXTPaymentFileDownloader) GetByteContentV2(ctx context.Context) ([]byte, error) {
	tempFile, err := d.downloadToTempFile(ctx, d.PaymentFileID, filestorage.ContentTypeTXT)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer d.closeAndCleanupFile(tempFile)

	// Set the offset to 0 to read from the beginning of the file
	_, err = tempFile.File.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("csvFile.Seek error: %v", err)
	}

	// Read the file and get the byte content
	bytes, err := os.ReadFile(tempFile.ObjectPath)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("os.ReadFile err: %v", err))
	}

	return bytes, nil
}
