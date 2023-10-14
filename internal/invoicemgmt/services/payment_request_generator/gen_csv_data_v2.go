package generator

import (
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *CScsvPaymentRequestGenerator) GenCSVDataV2(partnerCS *entities.PartnerConvenienceStore, filePaymentDataList []*dataMap, prefectureCodeWithNameMapped map[string]string, flags *CreatePaymentRequestFlags) ([][]string, error) {
	curentDateJST, err := GetTimeInJST(time.Now().UTC())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("GetTimeInJST err: %v", err))
	}
	createdDateStr := curentDateJST.Format("20060102")
	csvData := [][]string{}
	totalAmount := int64(0)

	for _, filePaymentData := range filePaymentDataList {
		headerRecordSlice := g.createFileHeaderRecordV2(partnerCS, createdDateStr)
		invoiceControlRecordSlice := g.createFileInvoiceControlRecordV2(partnerCS)

		exactTotal, err := GetFloat64ExactValueAndDecimalPlaces(filePaymentData.Invoice.Total, "2")
		if err != nil {
			return nil, err
		}

		totalAmount += int64(exactTotal)
		invoiceRecordSlice, err := g.createFileInvoiceRecordV2(filePaymentData, createdDateStr, prefectureCodeWithNameMapped, flags.UseKECFeedbackPh1)
		if err != nil {
			return nil, err
		}

		billItemMessage, err := g.createBillItemDetailMessageRecordsV2(filePaymentData)
		if err != nil {
			return nil, err
		}

		message1RecordSlice := g.createFileMessage1RecordV2(partnerCS)
		csvData = append(csvData, headerRecordSlice)
		csvData = append(csvData, invoiceControlRecordSlice)
		csvData = append(csvData, invoiceRecordSlice)
		csvData = append(csvData, message1RecordSlice)

		if flags.EnableBillingMessageInCSVMessages {
			csvData = append(csvData, billItemMessage...)
		} else {
			message2RecordSlice := g.createFileMessage2RecordV2(partnerCS)
			message3RecordSlice := g.createFileMessage3RecordV2(partnerCS)
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

func (g *CScsvPaymentRequestGenerator) createFileHeaderRecordV2(partnerCS *entities.PartnerConvenienceStore, createdDateStr string) []string {
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

func (g *CScsvPaymentRequestGenerator) createFileInvoiceControlRecordV2(partnerCS *entities.PartnerConvenienceStore) []string {
	fullWidthCompanyName := g.StringNormalizer.ToFullWidth(partnerCS.CompanyName.String)
	halfWidthCompanyTelNumber := g.StringNormalizer.ToHalfWidth(partnerCS.CompanyTelNumber.String)
	halfWidthPostalCode := g.StringNormalizer.ToHalfWidth(partnerCS.PostalCode.String)
	fullWidthAddress1 := g.StringNormalizer.ToFullWidth(partnerCS.Address1.String)
	fullWidthAddress2 := g.StringNormalizer.ToFullWidth(partnerCS.Address2.String)

	invoiceControlRecord := csInvoiceControlRecord{
		RecordCategory:   utils.LimitString("1", 1),
		DataCategory:     utils.LimitString("3", 1),
		CompanyName:      utils.LimitString(fullWidthCompanyName, 20),
		CompanyTelNumber: utils.LimitString(halfWidthCompanyTelNumber, 15),
		PostalCode:       utils.LimitString(halfWidthPostalCode, 8),
		Address1:         utils.LimitString(fullWidthAddress1, 20),
		Address2:         utils.LimitString(fullWidthAddress2, 20),
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

func (g *CScsvPaymentRequestGenerator) createFileInvoiceRecordV2(filePaymentData *dataMap, createdDateStr string, prefectureCodeWithNameMapped map[string]string, useKECFeedbackPh1 bool) ([]string, error) {
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

	halfWidthPostalCode := g.StringNormalizer.ToHalfWidth(studentBillingAddressInfo.BillingAddress.PostalCode.String)
	fullWidthAddress1 := g.StringNormalizer.ToFullWidth(fmt.Sprintf("%s %s %s %s", prefectureName, studentBillingAddressInfo.BillingAddress.City.String, studentBillingAddressInfo.BillingAddress.Street1.String, studentBillingAddressInfo.BillingAddress.Street2.String))
	fullWidthContactInfo := g.StringNormalizer.ToFullWidth(studentBillingAddressInfo.StudentPaymentDetail.PayerPhoneNumber.String)
	fullWidthName := g.StringNormalizer.ToFullWidth(payerName)

	invoiceRecord := csInvoiceRecord{
		RecordCategory:    utils.LimitString("3", 1),
		DataCategory:      utils.LimitString("1", 1),
		Code:              utils.AddPrefixStringWithLimit(paymentSeqNumStr, "0", 17),
		CreatedDate:       utils.LimitString(createdDateStr, 8),
		DeadlineOfPayment: utils.LimitString(dueDateStr, 8),
		PostalCode:        utils.LimitString(halfWidthPostalCode, 8),
		Address1:          utils.LimitString(fullWidthAddress1, 20),
		Address2:          utils.LimitString("", 20),
		ContactInfo:       utils.LimitString(fullWidthContactInfo, 25),
		Name:              fullWidthName,
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

func (g *CScsvPaymentRequestGenerator) getFullWidthMessages(messages []string) []string {
	fullWidthMessages := []string{}
	for _, m := range messages {
		fullWidthMessage := g.StringNormalizer.ToFullWidth(m)
		fullWidthMessages = append(fullWidthMessages, fullWidthMessage)
	}

	return fullWidthMessages
}

func (g *CScsvPaymentRequestGenerator) createFileMessage1RecordV2(partnerCS *entities.PartnerConvenienceStore) []string {
	fullWidthMessages := g.getFullWidthMessages([]string{
		partnerCS.Message1.String,
		partnerCS.Message2.String,
		partnerCS.Message3.String,
		partnerCS.Message4.String,
		partnerCS.Message5.String,
		partnerCS.Message6.String,
		partnerCS.Message7.String,
		partnerCS.Message8.String,
	})

	message1RecordSlice := []string{
		utils.LimitString("3", 1),
		utils.LimitString("7", 1),
		utils.LimitString("1", 1),
		utils.LimitString(fullWidthMessages[0], 24),
		utils.LimitString(fullWidthMessages[1], 24),
		utils.LimitString(fullWidthMessages[2], 24),
		utils.LimitString(fullWidthMessages[3], 24),
		utils.LimitString(fullWidthMessages[4], 24),
		utils.LimitString(fullWidthMessages[5], 24),
		utils.LimitString(fullWidthMessages[6], 24),
		utils.LimitString(fullWidthMessages[7], 24),
		"",
		"",
	}

	return message1RecordSlice
}

func (g *CScsvPaymentRequestGenerator) createFileMessage2RecordV2(partnerCS *entities.PartnerConvenienceStore) []string {
	fullWidthMessages := g.getFullWidthMessages([]string{
		partnerCS.Message9.String,
		partnerCS.Message10.String,
		partnerCS.Message11.String,
		partnerCS.Message12.String,
		partnerCS.Message13.String,
		partnerCS.Message14.String,
		partnerCS.Message15.String,
		partnerCS.Message16.String,
	})

	message1RecordSlice := []string{
		utils.LimitString("3", 1),
		utils.LimitString("7", 1),
		utils.LimitString("2", 1),
		utils.LimitString(fullWidthMessages[0], 24),
		utils.LimitString(fullWidthMessages[1], 24),
		utils.LimitString(fullWidthMessages[2], 24),
		utils.LimitString(fullWidthMessages[3], 24),
		utils.LimitString(fullWidthMessages[4], 24),
		utils.LimitString(fullWidthMessages[5], 24),
		utils.LimitString(fullWidthMessages[6], 24),
		utils.LimitString(fullWidthMessages[7], 24),
		"",
		"",
	}

	return message1RecordSlice
}

func (g *CScsvPaymentRequestGenerator) createFileMessage3RecordV2(partnerCS *entities.PartnerConvenienceStore) []string {
	fullWidthMessages := g.getFullWidthMessages([]string{
		partnerCS.Message17.String,
		partnerCS.Message18.String,
		partnerCS.Message19.String,
		partnerCS.Message20.String,
		partnerCS.Message21.String,
		partnerCS.Message22.String,
		partnerCS.Message23.String,
		partnerCS.Message24.String,
	})

	message1RecordSlice := []string{
		utils.LimitString("3", 1),
		utils.LimitString("7", 1),
		utils.LimitString("3", 1),
		utils.LimitString(fullWidthMessages[0], 24),
		utils.LimitString(fullWidthMessages[1], 24),
		utils.LimitString(fullWidthMessages[2], 24),
		utils.LimitString(fullWidthMessages[3], 24),
		utils.LimitString(fullWidthMessages[4], 24),
		utils.LimitString(fullWidthMessages[5], 24),
		utils.LimitString(fullWidthMessages[6], 24),
		utils.LimitString(fullWidthMessages[7], 24),
		"",
		"",
	}

	return message1RecordSlice
}

func (g *CScsvPaymentRequestGenerator) createBillItemDetailMessageRecordsV2(dataMap *dataMap) ([][]string, error) {
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
	firstBillingMsg, err := g.genBillingMessageSliceV2(filteredBillingMessage[:4])
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
	secondBillingMsg, err := g.genBillingMessageSliceV2(filteredBillingMessage[4:6])
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

	fullWidthInvoiceAmount := g.StringNormalizer.ToFullWidth(formattedInvoiceAmount)
	fullWidthPaymentAmount := g.StringNormalizer.ToFullWidth(formattedPaymentAmount)

	messages[1] = append(messages[1],
		utils.LimitString("合計", 24),
		utils.AddPrefixStringWithLimit(fullWidthInvoiceAmount, " ", 24),
		utils.LimitString("今回ご請求分", 24),
		utils.AddPrefixStringWithLimit(fullWidthPaymentAmount, " ", 24),
		"",
		"",
	)

	return messages, nil
}
