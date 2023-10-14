package downloader

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/filestorage"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type csHeaderRecord struct {
	RecordCategory   string
	DataCategory     string
	CreatedDate      string
	ManufacturerCode string
	CompanyCode      string
	ShopCode         string
}

type csInvoiceControlRecord struct {
	RecordCategory   string
	DataCategory     string
	CompanyName      string
	CompanyTelNumber string
	PostalCode       string
	Address1         string
	Address2         string
}

type csInvoiceRecord struct {
	RecordCategory    string
	DataCategory      string
	Code              string
	CreatedDate       string
	DeadlineOfPayment string
	PostalCode        string
	Address1          string
	Address2          string
	ContactInfo       string
	Name              string
	Amount            string
	RevenueStampFlag  string
	WrittenDeadline   string
}

type csEndRecord struct {
	RecordCategory string
	NumberOfRecord string
	TotalAmount    string
}

type ConvenienceStoreCSVPaymentFileDownloader struct {
	*BasePaymentFileDownloader
	PaymentFileID                            string
	UseKECFeedbackPh1                        bool
	EnableOptionalValidationInPaymentRequest bool

	partnerCS                    *entities.PartnerConvenienceStore
	filePaymentDataList          []*FilePaymentDataMap
	prefectureCodeWithNameMapped map[string]string
}

func (d *ConvenienceStoreCSVPaymentFileDownloader) ValidateData(ctx context.Context) error {
	// Fetch the convenience store
	partnerCS, err := d.PartnerConvenienceStoreRepo.FindOne(ctx, d.DB)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return status.Error(codes.Internal, "Partner has no associated convenience store")
		}
		return status.Error(codes.Internal, fmt.Sprintf("d.PartnerConvenienceStoreRepo.FindOne err: %v", err))
	}

	err = d.Validator.ValidatePartnerConvenienceStore(partnerCS)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	// fetch all prefectures to be mapped on billing address detail
	prefecturesMapped, err := d.getPrefectureCodeWithNameMapped(ctx)

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// Get the payments and its invoices associated in a payment file
	filePaymentInvoices, err := d.getAndValidateFilePaymentInvoice(ctx, d.PaymentFileID)
	if err != nil {
		return err
	}

	// Validate payment and invoice and return list of student IDs
	studentIDs, err := d.getListOfStudentsFromPaymentInvoice(filePaymentInvoices, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(), 17, 7)
	if err != nil {
		return err
	}

	// Create student billing map
	studentBillingMap, err := d.getStudentBillingMap(ctx, studentIDs)
	if err != nil {
		return err
	}

	dataMapList := []*FilePaymentDataMap{}

	for _, e := range filePaymentInvoices {
		// Validate the billing address and payment detail of student
		studentbillingInfo, err := d.validateStudentBillingDetails(studentBillingMap, e.Payment.StudentID.String)
		if err != nil {
			return err
		}

		dataMapList = append(dataMapList, &FilePaymentDataMap{
			Payment:            e.Payment,
			Invoice:            e.Invoice,
			StudentBillingInfo: studentbillingInfo,
		})
	}

	d.partnerCS = partnerCS
	d.filePaymentDataList = dataMapList
	d.prefectureCodeWithNameMapped = prefecturesMapped

	return nil
}

func (d *ConvenienceStoreCSVPaymentFileDownloader) GetByteContent(ctx context.Context) ([]byte, error) {
	csvData, err := GenCSVData(d.partnerCS, d.filePaymentDataList, d.prefectureCodeWithNameMapped, d.UseKECFeedbackPh1)
	if err != nil {
		return nil, err
	}

	bytesData, err := GenCSVBytesFromData(csvData)
	if err != nil {
		return nil, err
	}

	return bytesData, nil
}

func GenCSVData(partnerCS *entities.PartnerConvenienceStore, filePaymentDataList []*FilePaymentDataMap, prefectureCodeWithNameMapped map[string]string, useKECFeedbackPh1 bool) ([][]string, error) {
	curentDateJST, err := GetTimeInJST(time.Now().UTC())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("GetTimeInJST err: %v", err))
	}
	createdDateStr := curentDateJST.Format("20060102")

	csvData := [][]string{}
	totalAmount := int64(0)

	for _, filePaymentData := range filePaymentDataList {
		headerRecordSlice := createHeaderRecord(partnerCS, createdDateStr)
		invoiceControlRecordSlice := createInvoiceControlRecord(partnerCS)

		exactTotal, err := GetFloat64ExactValueAndDecimalPlaces(filePaymentData.Invoice.Total, "2")
		if err != nil {
			return nil, err
		}

		totalAmount += int64(exactTotal)

		invoiceRecordSlice, err := createInvoiceRecords(filePaymentData, createdDateStr, prefectureCodeWithNameMapped, useKECFeedbackPh1)
		if err != nil {
			return nil, err
		}

		message1RecordSlice := createMessage1Record(partnerCS)
		message2RecordSlice := createMessage2Record(partnerCS)
		message3RecordSlice := createMessage3Record(partnerCS)

		csvData = append(csvData, headerRecordSlice)
		csvData = append(csvData, invoiceControlRecordSlice)
		csvData = append(csvData, invoiceRecordSlice)
		csvData = append(csvData, message1RecordSlice)
		csvData = append(csvData, message2RecordSlice)
		csvData = append(csvData, message3RecordSlice)
	}

	// to get the total number of record; get the sum of
	// - 3 * number of record (this includes the header, invoice control and invoice record)
	// - 1 (this is the end record)
	totalNumberOfRecord := (3 * len(filePaymentDataList)) + 1
	endRecordSlice, err := createEndRecord(totalNumberOfRecord, totalAmount)
	if err != nil {
		return nil, err
	}
	csvData = append(csvData, endRecordSlice)

	return csvData, nil
}

func GenCSVBytesFromData(csvData [][]string) ([]byte, error) {
	// write the CSV data to buffer
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	err := writer.WriteAll(csvData)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("writer.WriteAll err: %v", writer.Error()))
	}

	writer.Flush()

	if writer.Error() != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error on writing CSV err: %v", writer.Error()))
	}

	return buffer.Bytes(), nil
}

func createHeaderRecord(partnerCS *entities.PartnerConvenienceStore, createdDateStr string) []string {
	manufactureCodeStr := strconv.Itoa(int(partnerCS.ManufacturerCode.Int))
	companyCodeStr := strconv.Itoa(int(partnerCS.CompanyCode.Int))

	headerRecord := csHeaderRecord{
		RecordCategory:   utils.LimitString("1", 1),
		DataCategory:     utils.LimitString("1", 1),
		CreatedDate:      utils.LimitString(createdDateStr, 8),
		ManufacturerCode: utils.LimitString(manufactureCodeStr, 6),
		CompanyCode:      utils.LimitString(companyCodeStr, 5),
		ShopCode:         utils.LimitString(partnerCS.ShopCode.String, 6),
	}

	headerRecordSlice := []string{
		headerRecord.RecordCategory,
		headerRecord.DataCategory,
		headerRecord.CreatedDate,
		headerRecord.ManufacturerCode,
		headerRecord.CompanyCode,
		headerRecord.ShopCode,
		"",
		"",
		"",
		"",
		"",
		"",
		"",
	}

	return headerRecordSlice
}

func createInvoiceControlRecord(partnerCS *entities.PartnerConvenienceStore) []string {
	invoiceControlRecord := csInvoiceControlRecord{
		RecordCategory:   utils.LimitString("1", 1),
		DataCategory:     utils.LimitString("3", 1),
		CompanyName:      utils.LimitString(partnerCS.CompanyName.String, 20),
		CompanyTelNumber: utils.LimitString(partnerCS.CompanyTelNumber.String, 15),
		PostalCode:       utils.LimitString(partnerCS.PostalCode.String, 8),
		Address1:         utils.LimitString(partnerCS.Address1.String, 20),
		Address2:         utils.LimitString(partnerCS.Address2.String, 20),
	}

	invoiceControlRecordSlice := []string{
		invoiceControlRecord.RecordCategory,
		invoiceControlRecord.DataCategory,
		invoiceControlRecord.CompanyName,
		invoiceControlRecord.CompanyTelNumber,
		invoiceControlRecord.PostalCode,
		invoiceControlRecord.Address1,
		invoiceControlRecord.Address2,
		"",
		"",
		"",
		"",
		"",
		"",
	}

	return invoiceControlRecordSlice
}

func createInvoiceRecords(filePaymentData *FilePaymentDataMap, createdDateStr string, prefectureCodeWithNameMapped map[string]string, useKECFeedbackPh1 bool) ([]string, error) {
	paymentSeqNumStr := strconv.Itoa(int(filePaymentData.Payment.PaymentSequenceNumber.Int))
	// check if the length of payment sequence number exceeds the requirement
	if len(paymentSeqNumStr) > 17 {
		return nil, status.Error(codes.Internal, "The payment sequence number length exceeds the limit")
	}

	// Set the due date in JST
	dueDateStr := "99999999"
	if !filePaymentData.Payment.PaymentDueDate.Time.IsZero() {
		dueDateJST, err := GetTimeInJST(filePaymentData.Payment.PaymentDueDate.Time)
		if err != nil {
			return nil, err
		}
		dueDateStr = dueDateJST.Format("20060102")
	}

	// Set the expiry date in JST
	expireDateJST, err := GetTimeInJST(filePaymentData.Payment.PaymentExpiryDate.Time)
	if err != nil {
		return nil, err
	}
	expireDateStr := expireDateJST.Format("20060102")

	// Get the exact total amount of invoice
	exactTotal, err := GetFloat64ExactValueAndDecimalPlaces(filePaymentData.Invoice.Total, "2")
	if err != nil {
		return nil, err
	}
	exactTotalStr := strconv.FormatInt(int64(exactTotal), 10)

	// validate the length of total amount
	if len(exactTotalStr) > 7 {
		return nil, status.Error(codes.Internal, "The invoice total length exceeds the limit")
	}
	// associated student billing information
	studentBillingAddressInfo := filePaymentData.StudentBillingInfo

	prefectureName, ok := prefectureCodeWithNameMapped[studentBillingAddressInfo.BillingAddress.PrefectureCode.String]
	if !ok {
		return nil, status.Error(codes.Internal, fmt.Errorf("student %v with billing details prefecture code %v that does not match prefecture records", studentBillingAddressInfo.StudentPaymentDetail.StudentID.String, studentBillingAddressInfo.BillingAddress.PrefectureCode.String).Error())
	}

	payerName := utils.LimitString(studentBillingAddressInfo.StudentPaymentDetail.PayerName.String, 20)
	if useKECFeedbackPh1 {
		payerName = utils.LimitString(studentBillingAddressInfo.StudentPaymentDetail.PayerName.String, 16) + "・保護者"
	}

	invoiceRecord := csInvoiceRecord{
		RecordCategory:    utils.LimitString("3", 1),
		DataCategory:      utils.LimitString("1", 1),
		Code:              utils.AddPrefixStringWithLimit(paymentSeqNumStr, "0", 17),
		CreatedDate:       utils.LimitString(createdDateStr, 8),
		DeadlineOfPayment: utils.LimitString(dueDateStr, 8),
		PostalCode:        utils.LimitString(studentBillingAddressInfo.BillingAddress.PostalCode.String, 8),
		Address1:          utils.LimitString(fmt.Sprintf("%s %s %s %s", prefectureName, studentBillingAddressInfo.BillingAddress.City.String, studentBillingAddressInfo.BillingAddress.Street1.String, studentBillingAddressInfo.BillingAddress.Street2.String), 20),
		Address2:          utils.LimitString("", 20),
		ContactInfo:       utils.LimitString(studentBillingAddressInfo.StudentPaymentDetail.PayerPhoneNumber.String, 25),
		Name:              payerName,
		Amount:            utils.LimitString(exactTotalStr, 7),
		RevenueStampFlag:  utils.LimitString("", 1), // always no data
		WrittenDeadline:   utils.LimitString(expireDateStr, 8),
	}

	return []string{
		invoiceRecord.RecordCategory,
		invoiceRecord.DataCategory,
		invoiceRecord.Code,
		invoiceRecord.CreatedDate,
		invoiceRecord.DeadlineOfPayment,
		invoiceRecord.PostalCode,
		invoiceRecord.Address1,
		invoiceRecord.Address2,
		invoiceRecord.ContactInfo,
		invoiceRecord.Name,
		invoiceRecord.Amount,
		invoiceRecord.RevenueStampFlag,
		invoiceRecord.WrittenDeadline,
	}, nil
}

func createMessage1Record(partnerCS *entities.PartnerConvenienceStore) []string {
	message1RecordSlice := []string{
		utils.LimitString("3", 1),
		utils.LimitString("7", 1),
		utils.LimitString("1", 1),
		utils.LimitString(partnerCS.Message1.String, 24),
		utils.LimitString(partnerCS.Message2.String, 24),
		utils.LimitString(partnerCS.Message3.String, 24),
		utils.LimitString(partnerCS.Message4.String, 24),
		utils.LimitString(partnerCS.Message5.String, 24),
		utils.LimitString(partnerCS.Message6.String, 24),
		utils.LimitString(partnerCS.Message7.String, 24),
		utils.LimitString(partnerCS.Message8.String, 24),
		"",
		"",
	}

	return message1RecordSlice
}

func createMessage2Record(partnerCS *entities.PartnerConvenienceStore) []string {
	message2RecordSlice := []string{
		utils.LimitString("3", 1),
		utils.LimitString("7", 1),
		utils.LimitString("2", 1),
		utils.LimitString(partnerCS.Message9.String, 24),
		utils.LimitString(partnerCS.Message10.String, 24),
		utils.LimitString(partnerCS.Message11.String, 24),
		utils.LimitString(partnerCS.Message12.String, 24),
		utils.LimitString(partnerCS.Message13.String, 24),
		utils.LimitString(partnerCS.Message14.String, 24),
		utils.LimitString(partnerCS.Message15.String, 24),
		utils.LimitString(partnerCS.Message16.String, 24),
		"",
		"",
	}

	return message2RecordSlice
}

func createMessage3Record(partnerCS *entities.PartnerConvenienceStore) []string {
	message3RecordSlice := []string{
		utils.LimitString("3", 1),
		utils.LimitString("7", 1),
		utils.LimitString("3", 1),
		utils.LimitString(partnerCS.Message17.String, 24),
		utils.LimitString(partnerCS.Message18.String, 24),
		utils.LimitString(partnerCS.Message19.String, 24),
		utils.LimitString(partnerCS.Message20.String, 24),
		utils.LimitString(partnerCS.Message21.String, 24),
		utils.LimitString(partnerCS.Message22.String, 24),
		utils.LimitString(partnerCS.Message23.String, 24),
		utils.LimitString(partnerCS.Message24.String, 24),
		"",
		"",
	}

	return message3RecordSlice
}

func createEndRecord(totalRecord int, totalAmount int64) ([]string, error) {
	totalAmountStr := strconv.FormatInt(totalAmount, 10)
	totalNumberOfRecordStr := strconv.Itoa(totalRecord)
	if len(totalAmountStr) > 10 {
		return nil, status.Error(codes.Internal, "The sum of invoices length exceeds the limit")
	}
	if len(totalNumberOfRecordStr) > 8 {
		return nil, status.Error(codes.Internal, "The total number of records length exceeds the limit")
	}

	csEndRecord := csEndRecord{
		RecordCategory: utils.LimitString("9", 1),
		NumberOfRecord: utils.LimitString(totalNumberOfRecordStr, 8),
		TotalAmount:    utils.LimitString(totalAmountStr, 10),
	}

	endRecordSlice := []string{
		csEndRecord.RecordCategory,
		csEndRecord.NumberOfRecord,
		csEndRecord.TotalAmount,
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
	}

	return endRecordSlice, nil
}

func (d *ConvenienceStoreCSVPaymentFileDownloader) getStudentBillingMap(ctx context.Context, studentIDs []string) (map[string]*entities.StudentBillingDetailsMap, error) {
	studentBillingDetails, err := d.StudentPaymentDetailRepo.FindStudentBillingByStudentIDs(ctx, d.DB, studentIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("StudentPaymentDetailRepo.FindStudentBillingByStudentIDs err: %v", err))
	}

	if len(studentBillingDetails) == 0 {
		return nil, status.Error(codes.Internal, "No student billing detail records")
	}

	studentBillingMap := make(map[string]*entities.StudentBillingDetailsMap)
	for _, e := range studentBillingDetails {
		_, ok := studentBillingMap[e.StudentPaymentDetail.StudentID.String]
		if ok {
			continue
		}
		studentBillingMap[e.StudentPaymentDetail.StudentID.String] = e
	}
	return studentBillingMap, nil
}

func (d *ConvenienceStoreCSVPaymentFileDownloader) validateStudentBillingDetails(studentBillingMap map[string]*entities.StudentBillingDetailsMap, studentID string) (*entities.StudentBillingDetailsMap, error) {
	// Validate if each student in paymentInvoice have billing details
	studentBilling, ok := studentBillingMap[studentID]
	if !ok {
		return nil, status.Error(codes.Internal, "There is a student that does not have billing details")
	}

	// Currently it on use the first billing address of a student.
	// If we will support multiple billing address, we can add a condition to select here
	err := multierr.Combine(
		d.Validator.ValidateStudentPaymentDetail(studentBilling.StudentPaymentDetail),
		d.Validator.ValidateBillingAddress(studentBilling.BillingAddress, &utils.FeatureFlags{
			EnableOptionalValidationInPaymentRequest: d.EnableOptionalValidationInPaymentRequest,
		}),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return studentBilling, nil
}

func (d *ConvenienceStoreCSVPaymentFileDownloader) getPrefectureCodeWithNameMapped(ctx context.Context) (map[string]string, error) {
	// using find all and mapped all prefectures than looping billing address and finding prefecture for performance wise
	prefectures, err := d.PrefectureRepo.FindAll(ctx, d.DB)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("PrefectureRepo.FindAll err: %v", err))
	}

	if len(prefectures) == 0 {
		return nil, status.Error(codes.Internal, "No prefecture records")
	}

	prefectureMap := make(map[string]string)
	for _, e := range prefectures {
		_, ok := prefectureMap[e.PrefectureCode.String]
		if ok {
			continue
		}
		prefectureMap[e.PrefectureCode.String] = e.Name.String
	}
	return prefectureMap, nil
}

func (d *ConvenienceStoreCSVPaymentFileDownloader) GetByteContentV2(ctx context.Context) ([]byte, error) {
	tempFile, err := d.downloadToTempFile(ctx, d.PaymentFileID, filestorage.ContentTypeCSV)
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
