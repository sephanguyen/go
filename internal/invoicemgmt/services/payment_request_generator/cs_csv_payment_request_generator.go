package generator

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
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

type CreatePaymentRequestFlags struct {
	EnableOptionalValidationInPaymentRequest bool
	EnableBillingMessageInCSVMessages        bool
	UseKECFeedbackPh1                        bool
}

type CScsvPaymentRequestGenerator struct {
	*BasePaymentRequestGenerator
	Req                                      *invoice_pb.CreatePaymentRequestRequest
	BulkPaymentRequestID                     string
	UseKECFeedbackPh1                        bool
	RequestDate                              time.Time
	EnableOptionalValidationInPaymentRequest bool
	EnableBillingMessageInCSVMessages        bool
	EnableEncodePaymentRequestFiles          bool
	EnableFormatPaymentRequestFileFields     bool

	partnerCS                    *entities.PartnerConvenienceStore
	paymentAndFileAssocs         []paymentAndFileAssoc
	dataMapList                  []*dataMap
	prefectureCodeWithNameMapped map[string]string
}

func (g *CScsvPaymentRequestGenerator) ValidateData(ctx context.Context) error {
	// Validate the partner convenience store
	err := g.validatePartnerCS(ctx)
	if err != nil {
		return err
	}

	// Fetch payment and invoice
	paymentInvoice, err := g.getPaymentInvoice(ctx, g.Req.PaymentIds)
	if err != nil {
		return err
	}

	// Validate payment and invoice and return list of student IDs
	studentIDs, err := g.getListOfStudentsFromPaymentInvoice(paymentInvoice, g.Req.PaymentMethod.String(), 17, 7)
	if err != nil {
		return err
	}

	// Create student billing map
	studentBillingMap, err := g.getStudentBillingMap(ctx, studentIDs)
	if err != nil {
		return err
	}

	dataMapList := []*dataMap{}
	for _, e := range paymentInvoice {
		// If invoice total is negative, do not include the payment
		if isNegative(e.Invoice.Total) {
			continue
		}

		// Validate the billing address and payment detail of student
		if err := g.validateStudentBillingDetails(studentBillingMap, e.Payment.StudentID.String); err != nil {
			return err
		}

		dataMapList = append(dataMapList, &dataMap{
			Payment: e.Payment,
			Invoice: e.Invoice,
		})
	}

	g.dataMapList = dataMapList

	return nil
}

func (g *CScsvPaymentRequestGenerator) PlanPaymentAndFileAssociation(ctx context.Context) error {
	baseFileName := fmt.Sprintf(
		"Payment_CS_%sto%s",
		g.Req.ConvenienceStoreDates.DueDateFrom.AsTime().Format(fileNameTimeFormat),
		g.Req.ConvenienceStoreDates.DueDateUntil.AsTime().Format(fileNameTimeFormat),
	)

	if g.UseKECFeedbackPh1 {
		baseFileName = fmt.Sprintf(
			"Payment_CS_created_date_%s",
			g.RequestDate.Format(fileNameTimeFormat),
		)
	}

	g.paymentAndFileAssocs = generatePaymentAndFileAssocs(g.dataMapList, maximumPaymentPerCSV, baseFileName, csvFormat)

	return nil
}

func (g *CScsvPaymentRequestGenerator) SavePaymentAndFileAssociation(ctx context.Context) error {
	err := g.associatePaymentToFile(ctx, g.BulkPaymentRequestID, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(), g.paymentAndFileAssocs)
	if err != nil {
		return err
	}

	return nil
}

func (g *CScsvPaymentRequestGenerator) validatePartnerCS(ctx context.Context) error {
	partnerCS, err := g.PartnerConvenienceStoreRepo.FindOne(ctx, g.DB)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return status.Error(codes.Internal, "Partner has no associated convenience store")
		}

		return status.Error(codes.Internal, fmt.Sprintf("g.InvoiceService.PartnerConvenienceStoreRepo.FindOne err: %v", err))
	}
	err = g.Validator.ValidatePartnerConvenienceStore(partnerCS)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (g *CScsvPaymentRequestGenerator) getStudentBillingMap(ctx context.Context, studentIDs []string) (map[string]*entities.StudentBillingDetailsMap, error) {
	studentBillingDetails, err := g.StudentPaymentDetailRepo.FindStudentBillingByStudentIDs(ctx, g.DB, studentIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("g.StudentPaymentDetailRepo.FindStudentBillingByStudentIDs err: %v", err))
	}

	studentBillingMap := make(map[string]*entities.StudentBillingDetailsMap)
	for _, e := range studentBillingDetails {
		// Only the first bank account will be used if student has multiple
		_, ok := studentBillingMap[e.StudentPaymentDetail.StudentID.String]
		if ok {
			continue
		}
		studentBillingMap[e.StudentPaymentDetail.StudentID.String] = e
	}

	return studentBillingMap, nil
}

func (g *CScsvPaymentRequestGenerator) validateStudentBillingDetails(studentBillingMap map[string]*entities.StudentBillingDetailsMap, studentID string) error {
	// Validate if each student in paymentInvoice have billing details
	studentBilling, ok := studentBillingMap[studentID]
	if !ok {
		return status.Error(codes.Internal, "There is a student that does not have billing details")
	}

	// Currently it on use the first billing address of a student.
	// If we will support multiple billing address, we can add a condition to select here
	err := multierr.Combine(
		g.Validator.ValidateStudentPaymentDetail(studentBilling.StudentPaymentDetail),
		g.Validator.ValidateBillingAddress(studentBilling.BillingAddress, &utils.FeatureFlags{
			EnableOptionalValidationInPaymentRequest: g.EnableOptionalValidationInPaymentRequest,
		}),
	)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (g *CScsvPaymentRequestGenerator) validateStudentBillingDetailsV2(studentBillingMap map[string]*entities.StudentBillingDetailsMap, studentID string) (*entities.StudentBillingDetailsMap, error) {
	// Validate if each student in paymentInvoice have billing details
	studentBilling, ok := studentBillingMap[studentID]
	if !ok {
		return nil, status.Error(codes.Internal, "There is a student that does not have billing details")
	}
	// Currently it on use the first billing address of a student.
	// If we will support multiple billing address, we can add a condition to select here
	err := multierr.Combine(
		g.Validator.ValidateStudentPaymentDetail(studentBilling.StudentPaymentDetail),
		g.Validator.ValidateBillingAddress(studentBilling.BillingAddress, &utils.FeatureFlags{
			EnableOptionalValidationInPaymentRequest: g.EnableOptionalValidationInPaymentRequest,
		}),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return studentBilling, nil
}

func (g *CScsvPaymentRequestGenerator) ValidateDataV2(ctx context.Context) error {
	// Validate the partner convenience store
	partnerCS, err := g.PartnerConvenienceStoreRepo.FindOne(ctx, g.DB)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return status.Error(codes.Internal, "Partner has no associated convenience store")
		}
		return status.Error(codes.Internal, fmt.Sprintf("g.PartnerConvenienceStoreRepo.FindOne err: %v", err))
	}

	err = g.Validator.ValidatePartnerConvenienceStore(partnerCS)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// fetch all prefectures to be mapped on billing address detail
	prefecturesMapped, err := g.getPrefectureCodeWithNameMapped(ctx)

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// Fetch payment and invoice
	paymentInvoice, err := g.getPaymentInvoice(ctx, g.Req.PaymentIds)
	if err != nil {
		return err
	}

	// Validate payment and invoice and return list of student IDs
	studentIDs, err := g.getListOfStudentsFromPaymentInvoice(paymentInvoice, g.Req.PaymentMethod.String(), 17, 7)
	if err != nil {
		return err
	}

	// get student billing map
	studentBillingMap, err := g.getStudentBillingMap(ctx, studentIDs)
	if err != nil {
		return err
	}

	invoiceBillItemDetails := make(map[string][]*entities.InvoiceBillItemMap)
	invoiceAdjustmentMap := make(map[string][]*entities.InvoiceAdjustment)

	if g.EnableBillingMessageInCSVMessages {
		invoiceIDs := make([]string, len(paymentInvoice))
		for _, e := range paymentInvoice {
			invoiceIDs = append(invoiceIDs, e.Invoice.InvoiceID.String)
		}

		invoiceBillItemDetails, err = g.getInvoiceBillItemMap(ctx, invoiceIDs)
		if err != nil {
			return err
		}

		invoiceAdjustmentMap, err = g.getInvoiceAdjustmentsMap(ctx, invoiceIDs)
		if err != nil {
			return err
		}
	}

	dataMapList := []*dataMap{}
	for _, e := range paymentInvoice {
		// If invoice total is negative, do not include the payment
		if isNegative(e.Invoice.Total) {
			continue
		}
		// Validate the billing address and payment detail of student
		studentbillingInfo, err := g.validateStudentBillingDetailsV2(studentBillingMap, e.Payment.StudentID.String)
		if err != nil {
			return err
		}

		dm := &dataMap{
			Payment:            e.Payment,
			Invoice:            e.Invoice,
			StudentBillingInfo: studentbillingInfo,
		}

		if billItemDetails, ok := invoiceBillItemDetails[e.Invoice.InvoiceID.String]; ok {
			dm.BillItemDetails = billItemDetails
		}

		if invoiceAdjustments, ok := invoiceAdjustmentMap[e.Invoice.InvoiceID.String]; ok {
			dm.InvoiceAdjustments = invoiceAdjustments
		}

		dataMapList = append(dataMapList, dm)
	}
	g.partnerCS = partnerCS
	g.dataMapList = dataMapList
	g.prefectureCodeWithNameMapped = prefecturesMapped

	return nil
}

func (g *CScsvPaymentRequestGenerator) getPrefectureCodeWithNameMapped(ctx context.Context) (map[string]string, error) {
	// using find all and mapped all prefectures than looping billing address and finding prefecture for performance wise
	prefectures, err := g.PrefectureRepo.FindAll(ctx, g.DB)
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

func (g *CScsvPaymentRequestGenerator) SaveAndUploadPaymentFileV2(ctx context.Context) error {
	err := database.ExecInTx(ctx, g.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		for _, paymentFileAssoc := range g.paymentAndFileAssocs {
			// convert to data bytes
			byteContent, err := g.GetByteContent(ctx, paymentFileAssoc.DataMap)
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("cannot convert byte content: %v on file name: %v with bulk payment request id: %v", err, paymentFileAssoc.FileName, g.BulkPaymentRequestID))
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
					ContentType: filestorage.ContentTypeCSV,
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

func (g *CScsvPaymentRequestGenerator) GetByteContent(ctx context.Context, dataMapList []*dataMap) ([]byte, error) {
	csvData, err := g.GenCSVData(g.partnerCS, dataMapList, g.prefectureCodeWithNameMapped, &CreatePaymentRequestFlags{
		UseKECFeedbackPh1:                        g.UseKECFeedbackPh1,
		EnableOptionalValidationInPaymentRequest: g.EnableOptionalValidationInPaymentRequest,
		EnableBillingMessageInCSVMessages:        g.EnableBillingMessageInCSVMessages,
	})
	if err != nil {
		return nil, err
	}

	bytesData, err := GenCSVBytesFromData(csvData)
	if err != nil {
		return nil, err
	}

	return bytesData, nil
}

func (g *CScsvPaymentRequestGenerator) GenCSVData(partnerCS *entities.PartnerConvenienceStore, filePaymentDataList []*dataMap, prefectureCodeWithNameMapped map[string]string, flags *CreatePaymentRequestFlags) ([][]string, error) {
	if g.EnableFormatPaymentRequestFileFields {
		return g.GenCSVDataV2(partnerCS, filePaymentDataList, prefectureCodeWithNameMapped, flags)
	}

	curentDateJST, err := GetTimeInJST(time.Now().UTC())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("GetTimeInJST err: %v", err))
	}
	createdDateStr := curentDateJST.Format("20060102")
	csvData := [][]string{}
	totalAmount := int64(0)

	for _, filePaymentData := range filePaymentDataList {
		headerRecordSlice := createFileHeaderRecord(partnerCS, createdDateStr)
		invoiceControlRecordSlice := createFileInvoiceControlRecord(partnerCS)
		exactTotal, err := GetFloat64ExactValueAndDecimalPlaces(filePaymentData.Invoice.Total, "2")
		if err != nil {
			return nil, err
		}

		totalAmount += int64(exactTotal)
		invoiceRecordSlice, err := createFileInvoiceRecords(filePaymentData, createdDateStr, prefectureCodeWithNameMapped, flags.UseKECFeedbackPh1)
		if err != nil {
			return nil, err
		}

		billItemMessage, err := g.createBillItemDetailMessageRecords(filePaymentData)
		if err != nil {
			return nil, err
		}

		message1RecordSlice := createFileMessage1Record(partnerCS)
		csvData = append(csvData, headerRecordSlice)
		csvData = append(csvData, invoiceControlRecordSlice)
		csvData = append(csvData, invoiceRecordSlice)
		csvData = append(csvData, message1RecordSlice)

		if flags.EnableBillingMessageInCSVMessages {
			csvData = append(csvData, billItemMessage...)
		} else {
			message2RecordSlice := createFileMessage2Record(partnerCS)
			message3RecordSlice := createFileMessage3Record(partnerCS)
			csvData = append(csvData, message2RecordSlice)
			csvData = append(csvData, message3RecordSlice)
		}
	}

	// to get the total number of record; get the sum of
	// - 3 * number of record (this includes the header, invoice control and invoice record)
	// - 1 (this is the end record)
	totalNumberOfRecord := (3 * len(filePaymentDataList)) + 1
	endRecordSlice, err := createFileEndRecord(totalNumberOfRecord, totalAmount)
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

func createFileHeaderRecord(partnerCS *entities.PartnerConvenienceStore, createdDateStr string) []string {
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

func createFileInvoiceControlRecord(partnerCS *entities.PartnerConvenienceStore) []string {
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

func createFileInvoiceRecords(filePaymentData *dataMap, createdDateStr string, prefectureCodeWithNameMapped map[string]string, useKECFeedbackPh1 bool) ([]string, error) {
	paymentSeqNumStr := strconv.Itoa(int(filePaymentData.Payment.PaymentSequenceNumber.Int))
	// check if the length of payment sequence number exceeds the requirement
	if len(paymentSeqNumStr) > 17 {
		return nil, status.Error(codes.Internal, "The payment sequence number length exceeds the limit")
	}

	// Set the due date in JST
	// In the future we may need get the date time for other countries
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

func createFileMessage1Record(partnerCS *entities.PartnerConvenienceStore) []string {
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

func createFileMessage2Record(partnerCS *entities.PartnerConvenienceStore) []string {
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

func createFileMessage3Record(partnerCS *entities.PartnerConvenienceStore) []string {
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

func (g *CScsvPaymentRequestGenerator) createBillItemDetailMessageRecords(dataMap *dataMap) ([][]string, error) {
	// Generate overall bill item and adjustment messages
	overallBillItemMessages, err := genOverallBillingDescAmount(dataMap)
	if err != nil {
		return nil, err
	}

	// Generate bill item messages
	filteredBillingMessage := genFilteredBillingDescAmount(overallBillItemMessages)

	messages := make([][]string, 2)

	// Assign bill item messages in message 9 - 16
	messages[0] = []string{
		utils.LimitString("3", 1),
		utils.LimitString("7", 1),
		utils.LimitString("2", 1),
	}
	firstBillingMsg, err := genBillingMessageSlice(filteredBillingMessage[:4])
	if err != nil {
		return nil, err
	}
	messages[0] = append(messages[0], firstBillingMsg...)
	messages[0] = append(messages[0], "", "")

	paymentAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(dataMap.Payment.Amount, "2")
	if err != nil {
		return nil, err
	}

	invoiceAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(dataMap.Invoice.Total, "2")
	if err != nil {
		return nil, err
	}

	// Assign bill item messages in message 17 - 24
	messages[1] = []string{
		utils.LimitString("3", 1),
		utils.LimitString("7", 1),
		utils.LimitString("3", 1),
	}
	secondBillingMsg, err := genBillingMessageSlice(filteredBillingMessage[4:6])
	if err != nil {
		return nil, err
	}
	messages[1] = append(messages[1], secondBillingMsg...)

	formattedInvoiceAmount, err := formatPaymentRequestCurrency(invoiceAmount, 24)
	if err != nil {
		return nil, err
	}

	formattedPaymentAmount, err := formatPaymentRequestCurrency(paymentAmount, 24)
	if err != nil {
		return nil, err
	}

	messages[1] = append(messages[1],
		utils.LimitString("合計", 24),
		utils.AddPrefixStringWithLimit(formattedInvoiceAmount, " ", 24),
		utils.LimitString("今回ご請求分", 24),
		utils.AddPrefixStringWithLimit(formattedPaymentAmount, " ", 24),
		"",
		"",
	)

	return messages, nil
}

func createFileEndRecord(totalRecord int, totalAmount int64) ([]string, error) {
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

func (g *CScsvPaymentRequestGenerator) SaveAndUploadPaymentFile(ctx context.Context) error {
	err := database.ExecInTx(ctx, g.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		for _, paymentFileAssoc := range g.paymentAndFileAssocs {
			// convert to data bytes
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
					ContentType: filestorage.ContentTypeCSV,
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

			// upload to cloud storage
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
