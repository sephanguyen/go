package invoicemgmt

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/unleash"
	invoicePackageConst "github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	downloader "github.com/manabie-com/backend/internal/invoicemgmt/services/payment_file_downloader"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	orderEntities "github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func chunkSliceofSliceOfString(slices [][]string, chunkSize int) [][][]string {
	var chunks [][][]string

	for i := 0; i < len(slices); i += chunkSize {
		end := i + chunkSize

		if end > len(slices) {
			end = len(slices)
		}

		chunks = append(chunks, slices[i:end])
	}

	return chunks
}

// nolint:unparam
func isIndexesEmpty(s []string, indexes []int, name string) error {
	for _, index := range indexes {
		if len(s[index]) != 0 {
			return fmt.Errorf("%s should be empty", name)
		}
	}

	return nil
}

func isEqual(actual, expected string, name string) error {
	if actual != expected {
		return fmt.Errorf("%s should be equal to %s. given: %s", name, actual, expected)
	}
	return nil
}

func isEqualInt(actual, expected int, name string) error {
	if actual != expected {
		return fmt.Errorf("%s should be equal to %v. given: %v", name, actual, expected)
	}
	return nil
}

func isReachedTheMaxLength(value string, maxLen int, name string) error {
	if len([]rune(value)) > maxLen {
		return fmt.Errorf("%s length should not be greater than \"%d\"", name, maxLen)
	}
	return nil
}

func isEqualLength(value string, l int, name string) error {
	if len([]rune(value)) != l {
		return fmt.Errorf("%s length should be \"%d\"", name, l)
	}
	return nil
}

func isCorrectDateFormat(value string, format string, name string) error {
	_, err := time.Parse(format, value)
	if err != nil {
		return fmt.Errorf("the %s uses wrong date format", name)
	}
	return nil
}

func validateCSVHeaderRecord(partnerCS *entities.PartnerConvenienceStore, line []string) error {
	if err := multierr.Combine(
		isEqual(line[0], "1", "record category"),
		isEqual(line[1], "1", "data category"),
		isCorrectDateFormat(line[2], "20060102", "created date"),
		isEqualLength(line[3], 6, "manufacturer code"),
		isEqualLength(line[4], 5, "company code"),
		isReachedTheMaxLength(line[5], 6, "shop code"),
		isIndexesEmpty(line, []int{6, 7, 8, 9, 10, 11, 12}, "filter columns"),
	); err != nil {
		return err
	}

	// validate the partnerCS values
	if err := multierr.Combine(
		isEqual(line[3], utils.LimitString(strconv.Itoa(int(partnerCS.ManufacturerCode.Int)), 6), "manufacturer code"),
		isEqual(line[4], utils.LimitString(strconv.Itoa(int(partnerCS.CompanyCode.Int)), 5), "company code"),
		isEqual(line[5], utils.LimitString(partnerCS.ShopCode.String, 6), "shop code"),
	); err != nil {
		return err
	}

	return nil
}

func validateCSVInvoiceControlRecord(partnerCS *entities.PartnerConvenienceStore, line []string) error {
	if err := multierr.Combine(
		isEqual(line[0], "1", "record category"),
		isEqual(line[1], "3", "data category"),
		isReachedTheMaxLength(line[2], 20, "company name"),
		isReachedTheMaxLength(line[3], 15, "company tel number"),
		isReachedTheMaxLength(line[4], 8, "postal code"),
		isReachedTheMaxLength(line[5], 20, "address1"),
		isReachedTheMaxLength(line[6], 20, "address2"),
		isIndexesEmpty(line, []int{7, 8, 9, 10, 11, 12}, "filter columns"),
	); err != nil {
		return err
	}

	// validate the partnerCS values
	if err := multierr.Combine(
		isEqual(line[2], utils.LimitString(partnerCS.CompanyName.String, 20), "company name"),
		isEqual(line[3], utils.LimitString(partnerCS.CompanyTelNumber.String, 15), "company tel number"),
		isEqual(line[4], utils.LimitString(partnerCS.PostalCode.String, 8), "postal code"),
		isEqual(line[5], utils.LimitString(partnerCS.Address1.String, 20), "address1"),
		isEqual(line[6], utils.LimitString(partnerCS.Address2.String, 20), "address2"),
	); err != nil {
		return err
	}

	return nil
}

func validateCSVInvoiceControlRecordV2(partnerCS *entities.PartnerConvenienceStore, line []string) error {
	if err := multierr.Combine(
		isEqual(line[0], "1", "record category"),
		isEqual(line[1], "3", "data category"),
		isReachedTheMaxLength(line[2], 20, "company name"),
		isReachedTheMaxLength(line[3], 15, "company tel number"),
		isReachedTheMaxLength(line[4], 8, "postal code"),
		isReachedTheMaxLength(line[5], 20, "address1"),
		isReachedTheMaxLength(line[6], 20, "address2"),
		isIndexesEmpty(line, []int{7, 8, 9, 10, 11, 12}, "filter columns"),
	); err != nil {
		return err
	}

	n, err := utils.NewStringNormalizer()
	if err != nil {
		return err
	}

	// validate the partnerCS values
	return multierr.Combine(
		isEqual(line[2], utils.LimitString(n.ToFullWidth(partnerCS.CompanyName.String), 20), "company name"),
		isEqual(line[3], utils.LimitString(n.ToHalfWidth(partnerCS.CompanyTelNumber.String), 15), "company tel number"),
		isEqual(line[4], utils.LimitString(n.ToHalfWidth(partnerCS.PostalCode.String), 8), "postal code"),
		isEqual(line[5], utils.LimitString(n.ToFullWidth(partnerCS.Address1.String), 20), "address1"),
		isEqual(line[6], utils.LimitString(n.ToFullWidth(partnerCS.Address2.String), 20), "address2"),
	)
}

type FilePaymentDataMap struct {
	Payment                *entities.Payment
	Invoice                *entities.Invoice
	StudentBillingInfo     *entities.StudentBillingDetailsMap
	StudentBankDetails     *entities.StudentBankDetailsMap //Bank Account Details
	StudentRelatedBank     *entities.BankRelationMap       //Bank Branch, Bank and Partner Bank
	NewCustomerCodeHistory *entities.NewCustomerCodeHistory
	BillItemDetails        []*entities.InvoiceBillItemMap
	InvoiceAdjustments     []*entities.InvoiceAdjustment
}

func validateCSVInvoiceRecord(dataMapList *FilePaymentDataMap, line []string, prefectureMapped map[string]string, unleashSuite *unleash.Suite) error {
	if err := multierr.Combine(
		isEqual(line[0], "3", "record category"),
		isEqual(line[1], "1", "data category"),
		isEqual(line[2], utils.AddPrefixStringWithLimit(strconv.Itoa(int(dataMapList.Payment.PaymentSequenceNumber.Int)), "0", 17), "code"),
		isCorrectDateFormat(line[3], "20060102", "created date"),
		isReachedTheMaxLength(line[5], 8, "postal code"),
		isReachedTheMaxLength(line[6], 20, "address1"),
		isReachedTheMaxLength(line[7], 20, "address2"),
		isReachedTheMaxLength(line[8], 25, "contact info"),
		isReachedTheMaxLength(line[9], 20, "name"),
		isReachedTheMaxLength(line[10], 7, "amount"),
		isReachedTheMaxLength(line[11], 1, "revenue stamp flag"),
		isCorrectDateFormat(line[12], "20060102", "written deadline"),
	); err != nil {
		return err
	}

	// validate the deadline of payment
	if line[4] != "99999999" {
		dueDateJST, _ := downloader.GetTimeInJST(dataMapList.Payment.PaymentDueDate.Time)
		dueDateStr := dueDateJST.Format("20060102")
		if err := multierr.Combine(
			isCorrectDateFormat(line[3], "20060102", "deadline of payment"),
			isEqual(line[3], dueDateStr, "deadline of payment"),
		); err != nil {
			return err
		}
	}

	expiryDateJST, _ := downloader.GetTimeInJST(dataMapList.Payment.PaymentExpiryDate.Time)
	expiryDateStr := expiryDateJST.Format("20060102")
	exactTotal, err := downloader.GetFloat64ExactValueAndDecimalPlaces(dataMapList.Invoice.Total, "2")
	if err != nil {
		return err
	}

	// validate the payment invoice
	if err := multierr.Combine(
		isEqual(line[10], utils.LimitString(strconv.FormatInt(int64(exactTotal), 10), 7), "amount"),
		isEqual(line[12], expiryDateStr, "written deadline"),
	); err != nil {
		return err
	}

	// validate the billing address of student
	// associated student billing information
	studentBillingAddressInfo := dataMapList.StudentBillingInfo

	prefectureName, ok := prefectureMapped[studentBillingAddressInfo.BillingAddress.PrefectureCode.String]
	if !ok {
		return fmt.Errorf("student billing details on payment student id: %v has error on prefecture code: %v", studentBillingAddressInfo.StudentPaymentDetail.StudentID.String, studentBillingAddressInfo.BillingAddress.PrefectureCode)
	}

	expectedPayerName := utils.LimitString(studentBillingAddressInfo.StudentPaymentDetail.PayerName.String, 20)
	if isFeatureToggleEnabled(unleashSuite.UnleashSrvAddr, unleashSuite.UnleashLocalAdminAPIKey, invoicePackageConst.EnableKECFeedbackPh1) {
		expectedPayerName = utils.LimitString(studentBillingAddressInfo.StudentPaymentDetail.PayerName.String, 16) + "・保護者"
	}

	if err := multierr.Combine(
		isEqual(line[5], utils.LimitString(studentBillingAddressInfo.BillingAddress.PostalCode.String, 8), "postal code"),
		isEqual(line[6], utils.LimitString(fmt.Sprintf("%s %s %s %s", prefectureName, studentBillingAddressInfo.BillingAddress.City.String, studentBillingAddressInfo.BillingAddress.Street1.String, studentBillingAddressInfo.BillingAddress.Street2.String), 20), "address1"),
		isEqual(line[7], utils.LimitString("", 20), "address2"),
		isEqual(line[8], utils.LimitString(studentBillingAddressInfo.StudentPaymentDetail.PayerPhoneNumber.String, 25), "contact info"),
		isEqual(line[9], expectedPayerName, "name"),
	); err != nil {
		return err
	}

	return nil
}

func validateCSVInvoiceRecordV2(dataMapList *FilePaymentDataMap, line []string, prefectureMapped map[string]string, unleashSuite *unleash.Suite) error {
	if err := multierr.Combine(
		isEqual(line[0], "3", "record category"),
		isEqual(line[1], "1", "data category"),
		isEqual(line[2], utils.AddPrefixStringWithLimit(strconv.Itoa(int(dataMapList.Payment.PaymentSequenceNumber.Int)), "0", 17), "code"),
		isCorrectDateFormat(line[3], "20060102", "created date"),
		isReachedTheMaxLength(line[5], 8, "postal code"),
		isReachedTheMaxLength(line[6], 20, "address1"),
		isReachedTheMaxLength(line[7], 20, "address2"),
		isReachedTheMaxLength(line[8], 25, "contact info"),
		isReachedTheMaxLength(line[9], 20, "name"),
		isReachedTheMaxLength(line[10], 7, "amount"),
		isReachedTheMaxLength(line[11], 1, "revenue stamp flag"),
		isCorrectDateFormat(line[12], "20060102", "written deadline"),
	); err != nil {
		return err
	}

	// validate the deadline of payment
	if line[4] != "99999999" {
		dueDateJST, _ := downloader.GetTimeInJST(dataMapList.Payment.PaymentDueDate.Time)
		dueDateStr := dueDateJST.Format("20060102")
		if err := multierr.Combine(
			isCorrectDateFormat(line[3], "20060102", "deadline of payment"),
			isEqual(line[3], dueDateStr, "deadline of payment"),
		); err != nil {
			return err
		}
	}

	expiryDateJST, _ := downloader.GetTimeInJST(dataMapList.Payment.PaymentExpiryDate.Time)
	expiryDateStr := expiryDateJST.Format("20060102")
	exactTotal, err := downloader.GetFloat64ExactValueAndDecimalPlaces(dataMapList.Invoice.Total, "2")
	if err != nil {
		return err
	}

	// validate the payment invoice
	if err := multierr.Combine(
		isEqual(line[10], utils.LimitString(strconv.FormatInt(int64(exactTotal), 10), 7), "amount"),
		isEqual(line[12], expiryDateStr, "written deadline"),
	); err != nil {
		return err
	}

	// validate the billing address of student
	// associated student billing information
	studentBillingAddressInfo := dataMapList.StudentBillingInfo

	prefectureName, ok := prefectureMapped[studentBillingAddressInfo.BillingAddress.PrefectureCode.String]
	if !ok {
		return fmt.Errorf("student billing details on payment student id: %v has error on prefecture code: %v", studentBillingAddressInfo.StudentPaymentDetail.StudentID.String, studentBillingAddressInfo.BillingAddress.PrefectureCode)
	}

	expectedPayerName := utils.LimitString(studentBillingAddressInfo.StudentPaymentDetail.PayerName.String, 20)
	if isFeatureToggleEnabled(unleashSuite.UnleashSrvAddr, unleashSuite.UnleashLocalAdminAPIKey, invoicePackageConst.EnableKECFeedbackPh1) {
		expectedPayerName = utils.LimitString(studentBillingAddressInfo.StudentPaymentDetail.PayerName.String, 16) + "・保護者"
	}

	n, err := utils.NewStringNormalizer()
	if err != nil {
		return err
	}

	return multierr.Combine(
		isEqual(line[5], utils.LimitString(n.ToHalfWidth(studentBillingAddressInfo.BillingAddress.PostalCode.String), 8), "postal code"),
		isEqual(line[6], utils.LimitString(n.ToFullWidth(fmt.Sprintf("%s %s %s %s", prefectureName, studentBillingAddressInfo.BillingAddress.City.String, studentBillingAddressInfo.BillingAddress.Street1.String, studentBillingAddressInfo.BillingAddress.Street2.String)), 20), "address1"),
		isEqual(line[7], utils.LimitString("", 20), "address2"),
		isEqual(line[8], utils.LimitString(n.ToFullWidth(studentBillingAddressInfo.StudentPaymentDetail.PayerPhoneNumber.String), 25), "contact info"),
		isEqual(line[9], n.ToFullWidth(expectedPayerName), "name"),
	)
}

func validateCSVMessageRecord(partnerCS *entities.PartnerConvenienceStore, line []string, offset int) error {
	if err := multierr.Combine(
		isEqual(line[0], "3", "record category"),
		isEqual(line[1], "7", "data category"),
		isEqual(line[2], strconv.Itoa(offset), "message code"),
		func() error {
			for _, i := range []int{3, 4, 5, 6, 7, 8, 9, 10} {
				if err := isReachedTheMaxLength(line[i], 24, "message"); err != nil {
					return err
				}
			}
			return nil
		}(),

		isIndexesEmpty(line, []int{11, 12}, "filter columns"),
	); err != nil {
		return err
	}

	var messages []string
	switch offset {
	case 1:
		messages = []string{
			partnerCS.Message1.String,
			partnerCS.Message2.String,
			partnerCS.Message3.String,
			partnerCS.Message4.String,
			partnerCS.Message5.String,
			partnerCS.Message6.String,
			partnerCS.Message7.String,
			partnerCS.Message8.String,
		}
	case 2:
		messages = []string{
			partnerCS.Message9.String,
			partnerCS.Message10.String,
			partnerCS.Message11.String,
			partnerCS.Message12.String,
			partnerCS.Message13.String,
			partnerCS.Message14.String,
			partnerCS.Message15.String,
			partnerCS.Message16.String,
		}
	case 3:
		messages = []string{
			partnerCS.Message17.String,
			partnerCS.Message18.String,
			partnerCS.Message19.String,
			partnerCS.Message20.String,
			partnerCS.Message21.String,
			partnerCS.Message22.String,
			partnerCS.Message23.String,
			partnerCS.Message24.String,
		}
	default:
		return fmt.Errorf("there is something wrong with message offset %d", offset)
	}

	// validate the partner CS values
	if err := multierr.Combine(
		isEqual(line[3], utils.LimitString(messages[0], 24), fmt.Sprintf("message-%d", ((offset-1)*8)+1)),
		isEqual(line[4], utils.LimitString(messages[1], 24), fmt.Sprintf("message-%d", ((offset-1)*8)+2)),
		isEqual(line[5], utils.LimitString(messages[2], 24), fmt.Sprintf("message-%d", ((offset-1)*8)+3)),
		isEqual(line[6], utils.LimitString(messages[3], 24), fmt.Sprintf("message-%d", ((offset-1)*8)+4)),
		isEqual(line[7], utils.LimitString(messages[4], 24), fmt.Sprintf("message-%d", ((offset-1)*8)+5)),
		isEqual(line[8], utils.LimitString(messages[5], 24), fmt.Sprintf("message-%d", ((offset-1)*8)+6)),
		isEqual(line[9], utils.LimitString(messages[6], 24), fmt.Sprintf("message-%d", ((offset-1)*8)+7)),
		isEqual(line[10], utils.LimitString(messages[7], 24), fmt.Sprintf("message-%d", ((offset-1)*8)+8)),
	); err != nil {
		return err
	}

	return nil
}

type billItemMessage struct {
	Description string
	Amount      float64
}

func genBillingMessages(dataMap *FilePaymentDataMap) ([]*billItemMessage, error) {
	overallBillItemMessages := make([]*billItemMessage, 0)

	// Generate for bill item
	for _, b := range dataMap.BillItemDetails {
		billItemName := ""
		if b.BillingItemDescription.Status == pgtype.Present {
			var billItemDesc orderEntities.BillingItemDescription
			err := json.Unmarshal(b.BillingItemDescription.Bytes, &billItemDesc)
			if err != nil {
				return nil, err
			}

			billItemName = billItemDesc.ProductName
		}

		billingAmount := b.FinalPrice
		if b.AdjustmentPrice.Status == pgtype.Present {
			billingAmount = b.AdjustmentPrice
			billItemName = invoicePackageConst.AdjustmentBillingKeyword + " " + billItemName
		}

		amount, err := utils.GetFloat64ExactValueAndDecimalPlaces(billingAmount, "2")
		if err != nil {
			return nil, err
		}

		overallBillItemMessages = append(overallBillItemMessages, &billItemMessage{Description: billItemName, Amount: amount})
	}

	// Generate for invoice adjustment
	for _, ia := range dataMap.InvoiceAdjustments {
		amount, err := utils.GetFloat64ExactValueAndDecimalPlaces(ia.Amount, "2")
		if err != nil {
			return nil, err
		}

		overallBillItemMessages = append(overallBillItemMessages, &billItemMessage{Description: ia.Description.String, Amount: amount})
	}

	filteredBillingMessage := make([]*billItemMessage, 6)

	switch {
	case len(overallBillItemMessages) > 6:
		// Assign the first 5 bill item messages to first 5 messages
		copy(filteredBillingMessage, overallBillItemMessages[:5])

		// Get the total remaining amount of bill item message
		var remainingAmount float64
		for _, b := range overallBillItemMessages[5:] {
			remainingAmount += b.Amount
		}

		// The last message contains the remaining amount of bill item
		filteredBillingMessage[5] = &billItemMessage{
			Description: "その他",
			Amount:      remainingAmount,
		}
	default:
		// Here, it is expected that the length of billing item message is less than or equal to 6
		// Since the list billItemMessages is already initialized, if the length is less than 6, other message will contain empty string
		copy(filteredBillingMessage, overallBillItemMessages)
	}

	return filteredBillingMessage, nil
}

func validateCSVMessageRecordWithBillingMessage(partnerCS *entities.PartnerConvenienceStore, dataMap *FilePaymentDataMap, line []string, offset int) error {
	if err := multierr.Combine(
		isEqual(line[0], "3", "record category"),
		isEqual(line[1], "7", "data category"),
		isEqual(line[2], strconv.Itoa(offset), "message code"),
		func() error {
			for _, i := range []int{3, 4, 5, 6, 7, 8, 9, 10} {
				if err := isReachedTheMaxLength(line[i], 24, "message"); err != nil {
					return err
				}
			}
			return nil
		}(),

		isIndexesEmpty(line, []int{11, 12}, "filter columns"),
	); err != nil {
		return err
	}

	billMessages, err := genBillingMessages(dataMap)
	if err != nil {
		return err
	}

	switch offset {
	case 1:
		if err := multierr.Combine(
			isEqual(line[3], utils.LimitString(partnerCS.Message1.String, 24), fmt.Sprintf("message-%d", ((offset-1)*8)+1)),
			isEqual(line[4], utils.LimitString(partnerCS.Message2.String, 24), fmt.Sprintf("message-%d", ((offset-1)*8)+2)),
			isEqual(line[5], utils.LimitString(partnerCS.Message3.String, 24), fmt.Sprintf("message-%d", ((offset-1)*8)+3)),
			isEqual(line[6], utils.LimitString(partnerCS.Message4.String, 24), fmt.Sprintf("message-%d", ((offset-1)*8)+4)),
			isEqual(line[7], utils.LimitString(partnerCS.Message5.String, 24), fmt.Sprintf("message-%d", ((offset-1)*8)+5)),
			isEqual(line[8], utils.LimitString(partnerCS.Message6.String, 24), fmt.Sprintf("message-%d", ((offset-1)*8)+6)),
			isEqual(line[9], utils.LimitString(partnerCS.Message7.String, 24), fmt.Sprintf("message-%d", ((offset-1)*8)+7)),
			isEqual(line[10], utils.LimitString(partnerCS.Message8.String, 24), fmt.Sprintf("message-%d", ((offset-1)*8)+8)),
		); err != nil {
			return err
		}
	case 2:
		curDescLine, curAmountLine := 3, 4
		for i := 0; i < len(billMessages[:4]); i++ {
			desc, amount := "", ""
			if billMessages[i] != nil {
				desc = utils.LimitString(billMessages[i].Description, 24)
				amount = utils.AddPrefixStringWithLimit(utils.FormatCurrency(billMessages[i].Amount)+"円", " ", 24)
			}

			if err := multierr.Combine(
				isEqual(line[curDescLine], desc, fmt.Sprintf("billing description %d", i+1)),
				isEqual(line[curAmountLine], amount, fmt.Sprintf("billing amount %d", i+1)),
			); err != nil {
				return err
			}

			curDescLine += 2
			curAmountLine += 2
		}
	case 3:
		paymentAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(dataMap.Payment.Amount, "2")
		if err != nil {
			return err
		}
		invoiceAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(dataMap.Invoice.Total, "2")
		if err != nil {
			return err
		}

		curDescLine, curAmountLine := 3, 4
		for i := 4; i < len(billMessages[4:6])+4; i++ {
			desc, amount := "", ""
			if billMessages[i] != nil {
				desc = utils.LimitString(billMessages[i].Description, 24)
				amount = utils.AddPrefixStringWithLimit(utils.FormatCurrency(billMessages[i].Amount)+"円", " ", 24)
			}

			if err := multierr.Combine(
				isEqual(line[curDescLine], desc, fmt.Sprintf("billing description %d", i+1)),
				isEqual(line[curAmountLine], amount, fmt.Sprintf("billing amount %d", i+1)),
			); err != nil {
				return err
			}

			curDescLine += 2
			curAmountLine += 2
		}

		if err := multierr.Combine(
			isEqual(line[7], utils.LimitString("合計", 24), "invoice amount label"),
			isEqual(line[8], utils.AddPrefixStringWithLimit(utils.FormatCurrency(invoiceAmount)+"円", " ", 24), "invoice amount"),
			isEqual(line[9], utils.LimitString("今回ご請求分", 24), "payment amount label"),
			isEqual(line[10], utils.AddPrefixStringWithLimit(utils.FormatCurrency(paymentAmount)+"円", " ", 24), "payment amount"),
		); err != nil {
			return err
		}
	default:
		return fmt.Errorf("there is something wrong with message offset %d", offset)
	}

	return nil
}

func validateCSVMessageRecordWithBillingMessageV2(partnerCS *entities.PartnerConvenienceStore, dataMap *FilePaymentDataMap, line []string, offset int) error {
	if err := multierr.Combine(
		isEqual(line[0], "3", "record category"),
		isEqual(line[1], "7", "data category"),
		isEqual(line[2], strconv.Itoa(offset), "message code"),
		func() error {
			for _, i := range []int{3, 4, 5, 6, 7, 8, 9, 10} {
				if err := isReachedTheMaxLength(line[i], 24, "message"); err != nil {
					return err
				}
			}
			return nil
		}(),

		isIndexesEmpty(line, []int{11, 12}, "filter columns"),
	); err != nil {
		return err
	}

	billMessages, err := genBillingMessages(dataMap)
	if err != nil {
		return err
	}

	n, err := utils.NewStringNormalizer()
	if err != nil {
		return err
	}

	switch offset {
	case 1:
		if err := multierr.Combine(
			isEqual(line[3], utils.LimitString(n.ToFullWidth(partnerCS.Message1.String), 24), fmt.Sprintf("message-%d", ((offset-1)*8)+1)),
			isEqual(line[4], utils.LimitString(n.ToFullWidth(partnerCS.Message2.String), 24), fmt.Sprintf("message-%d", ((offset-1)*8)+2)),
			isEqual(line[5], utils.LimitString(n.ToFullWidth(partnerCS.Message3.String), 24), fmt.Sprintf("message-%d", ((offset-1)*8)+3)),
			isEqual(line[6], utils.LimitString(n.ToFullWidth(partnerCS.Message4.String), 24), fmt.Sprintf("message-%d", ((offset-1)*8)+4)),
			isEqual(line[7], utils.LimitString(n.ToFullWidth(partnerCS.Message5.String), 24), fmt.Sprintf("message-%d", ((offset-1)*8)+5)),
			isEqual(line[8], utils.LimitString(n.ToFullWidth(partnerCS.Message6.String), 24), fmt.Sprintf("message-%d", ((offset-1)*8)+6)),
			isEqual(line[9], utils.LimitString(n.ToFullWidth(partnerCS.Message7.String), 24), fmt.Sprintf("message-%d", ((offset-1)*8)+7)),
			isEqual(line[10], utils.LimitString(n.ToFullWidth(partnerCS.Message8.String), 24), fmt.Sprintf("message-%d", ((offset-1)*8)+8)),
		); err != nil {
			return err
		}
	case 2:
		curDescLine, curAmountLine := 3, 4
		for i := 0; i < len(billMessages[:4]); i++ {
			desc, amount := "", ""
			if billMessages[i] != nil {
				desc = utils.LimitString(n.ToFullWidth(billMessages[i].Description), 24)
				amount = utils.AddPrefixStringWithLimit(n.ToFullWidth(utils.FormatCurrency(billMessages[i].Amount)+"円"), " ", 24)
			}

			if err := multierr.Combine(
				isEqual(line[curDescLine], desc, fmt.Sprintf("billing description %d", i+1)),
				isEqual(line[curAmountLine], amount, fmt.Sprintf("billing amount %d", i+1)),
			); err != nil {
				return err
			}

			curDescLine += 2
			curAmountLine += 2
		}
	case 3:
		paymentAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(dataMap.Payment.Amount, "2")
		if err != nil {
			return err
		}
		invoiceAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(dataMap.Invoice.Total, "2")
		if err != nil {
			return err
		}

		curDescLine, curAmountLine := 3, 4
		for i := 4; i < len(billMessages[4:6])+4; i++ {
			desc, amount := "", ""
			if billMessages[i] != nil {
				desc = utils.LimitString(n.ToFullWidth(billMessages[i].Description), 24)
				amount = utils.AddPrefixStringWithLimit(n.ToFullWidth(utils.FormatCurrency(billMessages[i].Amount)+"円"), " ", 24)
			}

			if err := multierr.Combine(
				isEqual(line[curDescLine], desc, fmt.Sprintf("billing description %d", i+1)),
				isEqual(line[curAmountLine], amount, fmt.Sprintf("billing amount %d", i+1)),
			); err != nil {
				return err
			}

			curDescLine += 2
			curAmountLine += 2
		}

		if err := multierr.Combine(
			isEqual(line[7], utils.LimitString("合計", 24), "invoice amount label"),
			isEqual(line[8], utils.AddPrefixStringWithLimit(n.ToFullWidth(utils.FormatCurrency(invoiceAmount)+"円"), " ", 24), "invoice amount"),
			isEqual(line[9], utils.LimitString("今回ご請求分", 24), "payment amount label"),
			isEqual(line[10], utils.AddPrefixStringWithLimit(n.ToFullWidth(utils.FormatCurrency(paymentAmount)+"円"), " ", 24), "payment amount"),
		); err != nil {
			return err
		}
	default:
		return fmt.Errorf("there is something wrong with message offset %d", offset)
	}

	return nil
}

func validateCSVEndRecord(paymentInvoices []*entities.PaymentInvoiceMap, line []string) error {
	if err := multierr.Combine(
		isEqual(line[0], "9", "record category"),
		isReachedTheMaxLength(line[1], 8, "number of record"),
		isReachedTheMaxLength(line[2], 10, "total amount"),
		isIndexesEmpty(line, []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, "filter columns"),
	); err != nil {
		return err
	}

	totalAmount := int64(0)
	for _, paymentInvoice := range paymentInvoices {
		exactTotal, err := downloader.GetFloat64ExactValueAndDecimalPlaces(paymentInvoice.Invoice.Total, "2")
		if err != nil {
			return err
		}
		totalAmount += int64(exactTotal)
	}

	// validate the values
	if err := multierr.Combine(
		isEqual(line[1], strconv.Itoa((3*len(paymentInvoices)+1)), "number of record"),
		isEqual(line[2], strconv.FormatInt(totalAmount, 10), "total amount"),
	); err != nil {
		return err
	}

	return nil
}

func validateBankTxtHeaderRecord(partnerBank *entities.PartnerBank, line string) error {
	lineRune := []rune(line)

	dataCategory := string(lineRune[0])
	typeCode := string(lineRune[1:3])
	codeCategory := string(lineRune[3])
	consignorCode := string(lineRune[4:14])
	consignorName := string(lineRune[14:54])
	depositDate := string(lineRune[54:58])
	bankNumber := string(lineRune[58:62])
	bankName := string(lineRune[62:77])
	bankBranchNumber := string(lineRune[77:80])
	bankBranchName := string(lineRune[80:95])
	depositItems := string(lineRune[95:96])
	accountNumber := string(lineRune[96:103])
	dummy := string(lineRune[103:120])

	if err := multierr.Combine(
		isEqual(dataCategory, "1", "data category"),
		isEqual(typeCode, "91", "type code"),
		isEqual(codeCategory, "0", "code category"),
		isCorrectDateFormat(depositDate, "0102", "deposit date"),
		isEqual(dummy, utils.AddPrefixString("", " ", 17), "dummy"),
	); err != nil {
		return err
	}

	// get the equivalent deposit item value
	var depositItemValue string
	for k, v := range invoicePackageConst.PartnerBankDepositItems {
		if v == partnerBank.DepositItems.String {
			depositItemValue = strconv.Itoa(k)
		}
	}

	// check partner bank values
	if err := multierr.Combine(
		isEqual(consignorCode, utils.AddPrefixStringWithLimit(partnerBank.ConsignorCode.String, "0", 10), "consignor code"),
		isEqual(consignorName, utils.AddSuffixStringWithLimit(partnerBank.ConsignorName.String, " ", 40), "consignor name"),
		isEqual(bankNumber, utils.AddSuffixStringWithLimit(partnerBank.BankNumber.String, "0", 4), "bank number"),
		isEqual(bankName, utils.AddSuffixStringWithLimit(partnerBank.BankName.String, " ", 15), "bank name"),
		isEqual(bankBranchNumber, utils.AddPrefixStringWithLimit(partnerBank.BankBranchNumber.String, "0", 3), "bank branch number"),
		isEqual(bankBranchName, utils.AddSuffixStringWithLimit(partnerBank.BankBranchName.String, " ", 15), "bank branch name"),
		isEqual(depositItems, utils.AddSuffixStringWithLimit(depositItemValue, " ", 1), "deposit items"),
		isEqual(accountNumber, utils.AddPrefixStringWithLimit(partnerBank.AccountNumber.String, "0", 7), "account number"),
	); err != nil {
		return err
	}

	return nil
}

func validateBankTxtDataRecord(filePaymentData *FilePaymentDataMap, line string) error {
	lineRune := []rune(line)

	dataCategory := string(line[0])
	depositBankNumber := string(lineRune[1:5])
	depositBankName := string(lineRune[5:20])
	depositBankBranchNumber := string(lineRune[20:23])
	depositBankBranchName := string(lineRune[23:38])
	dummy1 := string(lineRune[38:42])
	depositItems := string(lineRune[42:43])
	accountNumber := string(lineRune[43:50])
	accountOwnerName := string(lineRune[50:80])
	depositAmount := string(lineRune[80:90])
	newCustomerCode := string(lineRune[90:91])
	customerNumber := string(lineRune[91:111])
	resultCode := string(lineRune[111:112])
	dummy2 := string(lineRune[112:120])

	if err := multierr.Combine(
		isEqual(dataCategory, "2", "data category"),
		isEqual(dummy1, utils.AddPrefixString("", " ", 4), "dummy"),
		isEqual(dummy2, utils.AddPrefixString("", " ", 8), "dummy"),
		isEqual(resultCode, "0", "result code"),
	); err != nil {
		return err
	}

	studentRelatedBank := filePaymentData.StudentRelatedBank
	studentBankAccount := filePaymentData.StudentBankDetails

	var studentBankDepositItems string
	for index, value := range invoicePackageConst.PartnerBankDepositItems {
		if value == studentBankAccount.BankAccount.BankAccountType.String {
			studentBankDepositItems = strconv.Itoa(index)
		}
	}
	// customer number
	paymentSeqNumStr := strconv.Itoa(int(filePaymentData.Payment.PaymentSequenceNumber.Int))

	if err := multierr.Combine(
		isEqual(depositBankNumber, utils.AddPrefixStringWithLimit(studentRelatedBank.Bank.BankCode.String, "0", 4), "deposit bank number"),
		isEqual(depositBankName, utils.AddSuffixStringWithLimit(studentRelatedBank.Bank.BankNamePhonetic.String, " ", 15), "deposit bank name"),
		isEqual(depositBankBranchNumber, utils.AddPrefixStringWithLimit(studentRelatedBank.BankBranch.BankBranchCode.String, "0", 3), "deposit bank branch number"),
		isEqual(depositBankBranchName, utils.AddSuffixStringWithLimit(studentRelatedBank.BankBranch.BankBranchPhoneticName.String, " ", 15), "deposit bank branch name"),
		isEqual(depositItems, utils.LimitString(studentBankDepositItems, 1), "deposit bank account items"),
		isEqual(accountNumber, utils.AddPrefixStringWithLimit(studentBankAccount.BankAccount.BankAccountNumber.String, "0", 7), "bank account number"),
		isEqual(accountOwnerName, utils.AddSuffixStringWithLimit(studentBankAccount.BankAccount.BankAccountHolder.String, " ", 30), "bank account owner name"),
		isEqual(newCustomerCode, utils.LimitString(filePaymentData.NewCustomerCodeHistory.NewCustomerCode.String, 1), "new customer code"),
		isEqual(customerNumber, utils.AddSuffixStringWithLimit(paymentSeqNumStr, " ", 20), "customer number"),
	); err != nil {
		return err
	}

	// check the value of invoice
	exactTotal, err := downloader.GetFloat64ExactValueAndDecimalPlaces(filePaymentData.Invoice.Total, "2")
	if err != nil {
		return err
	}

	return isEqual(depositAmount, utils.AddPrefixStringWithLimit(strconv.FormatInt(int64(exactTotal), 10), "0", 10), "deposit amount")
}

func validateBankTxtTrailerRecord(paymentInvoices []*entities.PaymentInvoiceMap, line string) error {
	lineRune := []rune(line)

	dataCategory := string(lineRune[0])
	totalTransactions := string(lineRune[1:7])
	totalAmount := string(lineRune[7:19])
	transferredNumber := string(lineRune[19:25])
	transferredAmount := string(lineRune[25:37])
	failedNumber := string(lineRune[37:43])
	failedAmount := string(lineRune[43:55])
	dummy := string(lineRune[55:120])

	if err := multierr.Combine(
		isEqual(dataCategory, "8", "data category"),
		isEqual(transferredNumber, utils.AddPrefixString("", "0", 6), "transferred number"),
		isEqual(transferredAmount, utils.AddPrefixString("", "0", 12), "transferred amount"),
		isEqual(failedNumber, utils.AddPrefixString("", "0", 6), "failed number"),
		isEqual(failedAmount, utils.AddPrefixString("", "0", 12), "failed amount"),
		isEqual(dummy, utils.AddPrefixString("", " ", 65), "dummy"),
	); err != nil {
		return err
	}

	expectedTotalAmount := int64(0)
	for _, paymentInvoice := range paymentInvoices {
		exactTotal, err := downloader.GetFloat64ExactValueAndDecimalPlaces(paymentInvoice.Invoice.Total, "2")
		if err != nil {
			return err
		}
		expectedTotalAmount += int64(exactTotal)
	}

	// validate data
	if err := multierr.Combine(
		isEqual(totalTransactions, utils.AddPrefixStringWithLimit(strconv.Itoa(len(paymentInvoices)), "0", 6), "total transactions"),
		isEqual(totalAmount, utils.AddPrefixStringWithLimit(strconv.FormatInt(expectedTotalAmount, 10), "0", 12), "total amount"),
	); err != nil {
		return err
	}

	return nil
}

func validateBankTxtEndRecord(line string) error {
	lineRune := []rune(line)

	dataCategory := string(lineRune[0])
	dummy := string(lineRune[1:])

	if err := multierr.Combine(
		isEqual(dataCategory, "9", "data category"),
		isEqual(dummy, utils.AddPrefixString("", " ", 119), "dummy"),
	); err != nil {
		return err
	}

	return nil
}

func generatePaymentDateFormat(formatDateTime time.Time) string {
	return fmt.Sprintf("%v%02d%02d", formatDateTime.Year(), int(formatDateTime.Month()), formatDateTime.Day())
}
