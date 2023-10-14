package generator

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *DDtxtPaymentRequestGenerator) GenTxtBankContentV2(filePaymentDataList []*dataMap) ([]byte, error) {
	var b bytes.Buffer
	// write header record to buffer
	err := g.writeFileHeaderRecordV2(&b, filePaymentDataList)
	if err != nil {
		return nil, err
	}

	// write data records to buffer
	totalAmount, err := g.writeFileDataRecordsV2(&b, filePaymentDataList)
	if err != nil {
		return nil, err
	}

	// write trailer record to buffer
	err = writeFileTrailerRecordV2(&b, len(filePaymentDataList), totalAmount)
	if err != nil {
		return nil, err
	}
	// write trailer record to buffer
	err = writeFileEndRecordV2(&b)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (g *DDtxtPaymentRequestGenerator) writeFileHeaderRecordV2(b *bytes.Buffer, filePaymentInvoiceMap []*dataMap) error {
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

	halfWidthConsignorCode := g.StringNormalizer.ToHalfWidth(studentPartnerBank.ConsignorCode.String)
	halfWidthConsignorName := g.StringNormalizer.ToHalfWidth(studentPartnerBank.ConsignorName.String)
	halfWidthBankNumber := g.StringNormalizer.ToHalfWidth(studentPartnerBank.BankNumber.String)
	halfWidthBankName := g.StringNormalizer.ToHalfWidth(studentPartnerBank.BankName.String)
	halfWidthBakBranchNumber := g.StringNormalizer.ToHalfWidth(studentPartnerBank.BankBranchNumber.String)
	halfWidthBankBranchName := g.StringNormalizer.ToHalfWidth(studentPartnerBank.BankBranchName.String)
	halfWidthAccountNumber := g.StringNormalizer.ToHalfWidth(studentPartnerBank.AccountNumber.String)

	headerRecord := bankHeaderRecord{
		DataCategory:     "1",
		TypeCode:         "91",
		CodeCategory:     "0",
		ConsignorCode:    utils.AddPrefixStringWithLimit(halfWidthConsignorCode, "0", 10),
		ConsignorName:    utils.AddSuffixStringWithLimit(halfWidthConsignorName, " ", 40),
		DepositDate:      dueDateStr,
		BankNumber:       utils.AddSuffixStringWithLimit(halfWidthBankNumber, "0", 4),
		BankName:         utils.AddSuffixStringWithLimit(halfWidthBankName, " ", 15),
		BankBranchNumber: utils.AddPrefixStringWithLimit(halfWidthBakBranchNumber, "0", 3),
		BankBranchName:   utils.AddSuffixStringWithLimit(halfWidthBankBranchName, " ", 15),
		DepositItems:     utils.AddSuffixStringWithLimit(depositItems, " ", 1),
		AccountNumber:    utils.AddPrefixStringWithLimit(halfWidthAccountNumber, "0", 7),
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

func (g *DDtxtPaymentRequestGenerator) writeFileDataRecordsV2(b *bytes.Buffer, filePaymentInvoiceMap []*dataMap) (int64, error) {
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

		halfWidthBankCode := g.StringNormalizer.ToHalfWidth(studentRelatedBank.Bank.BankCode.String)
		halfWidthBankNamePhonetic := g.StringNormalizer.ToHalfWidth(studentRelatedBank.Bank.BankNamePhonetic.String)
		halfWidthBankBranchCode := g.StringNormalizer.ToHalfWidth(studentRelatedBank.BankBranch.BankBranchCode.String)
		halfWidthBankBranchPhoneticName := g.StringNormalizer.ToHalfWidth(studentRelatedBank.BankBranch.BankBranchPhoneticName.String)
		halfWidthBankAccountNumber := g.StringNormalizer.ToHalfWidth(studentBankAccount.BankAccount.BankAccountNumber.String)
		halfWidthBankAccountHolder := g.StringNormalizer.ToHalfWidth(studentBankAccount.BankAccount.BankAccountHolder.String)

		dataRecord := bankDataRecord{
			DataCategory:            "2",
			DepositBankNumber:       utils.AddPrefixStringWithLimit(halfWidthBankCode, "0", 4),
			DepositBankName:         utils.AddSuffixStringWithLimit(halfWidthBankNamePhonetic, " ", 15),
			DepositBankBranchNumber: utils.AddPrefixStringWithLimit(halfWidthBankBranchCode, "0", 3),
			DepositBankBranchName:   utils.AddSuffixStringWithLimit(halfWidthBankBranchPhoneticName, " ", 15),
			Dummy1:                  utils.AddPrefixString("", " ", 4), // 4 spaces
			DepositItems:            utils.LimitString(depositItems, 1),
			AccountNumber:           utils.AddPrefixStringWithLimit(halfWidthBankAccountNumber, "0", 7),
			AccountOwnerName:        utils.AddSuffixStringWithLimit(halfWidthBankAccountHolder, " ", 30),
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

func writeFileTrailerRecordV2(b *bytes.Buffer, totalTransactions int, totalAmount int64) error {
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

func writeFileEndRecordV2(b *bytes.Buffer) error {
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
