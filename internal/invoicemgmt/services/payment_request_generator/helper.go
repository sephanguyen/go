package generator

import (
	"encoding/json"
	"fmt"

	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	orderEntities "github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgtype"
)

type billItemMessage struct {
	Description string
	Amount      float64
}

func getBillItemDescriptionAndAmount(billItemDetails *entities.InvoiceBillItemMap) (string, float64, error) {
	billItemName := ""
	if billItemDetails.BillingItemDescription.Status == pgtype.Present {
		var billItemDesc orderEntities.BillingItemDescription
		err := json.Unmarshal(billItemDetails.BillingItemDescription.Bytes, &billItemDesc)
		if err != nil {
			return "", 0, err
		}

		billItemName = billItemDesc.ProductName
	}

	billingAmount := billItemDetails.FinalPrice
	if billItemDetails.AdjustmentPrice.Status == pgtype.Present {
		billingAmount = billItemDetails.AdjustmentPrice
		billItemName = constant.AdjustmentBillingKeyword + " " + billItemName
	}

	amount, err := utils.GetFloat64ExactValueAndDecimalPlaces(billingAmount, "2")
	if err != nil {
		return "", 0, err
	}

	return billItemName, amount, nil
}

func genOverallBillingDescAmount(dataMap *dataMap) ([]*billItemMessage, error) {
	overallBillItemMessages := make([]*billItemMessage, 0)

	// Generate for bill item
	for _, b := range dataMap.BillItemDetails {
		description, amount, err := getBillItemDescriptionAndAmount(b)
		if err != nil {
			return nil, err
		}

		overallBillItemMessages = append(overallBillItemMessages, &billItemMessage{Description: description, Amount: amount})
	}

	// Generate for invoice adjustment
	for _, ia := range dataMap.InvoiceAdjustments {
		amount, err := utils.GetFloat64ExactValueAndDecimalPlaces(ia.Amount, "2")
		if err != nil {
			return nil, err
		}

		overallBillItemMessages = append(overallBillItemMessages, &billItemMessage{Description: ia.Description.String, Amount: amount})
	}

	return overallBillItemMessages, nil
}

func genFilteredBillingDescAmount(overallBillItemMessages []*billItemMessage) []*billItemMessage {
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

	return filteredBillingMessage
}

func formatPaymentRequestCurrency(amount float64, maxCount int) (string, error) {
	formattedCurrency := utils.FormatCurrency(amount) + "円"
	if len([]rune(formattedCurrency)) > maxCount {
		return "", fmt.Errorf("%s reached the max length %d", formattedCurrency, maxCount)
	}

	return formattedCurrency, nil
}

func genBillingMessageSlice(filteredBillingMessage []*billItemMessage) ([]string, error) {
	message := []string{}

	for _, m := range filteredBillingMessage {
		desc, amount := "", ""
		if m != nil {
			desc = utils.LimitString(m.Description, 24)
			formattedCurrency, err := formatPaymentRequestCurrency(m.Amount, 24)
			if err != nil {
				return nil, err
			}

			amount = utils.AddPrefixStringWithLimit(formattedCurrency, " ", 24)
		}

		message = append(message, desc, amount)
	}

	return message, nil
}

func (g *CScsvPaymentRequestGenerator) genBillingMessageSliceV2(filteredBillingMessage []*billItemMessage) ([]string, error) {
	message := []string{}

	for _, m := range filteredBillingMessage {
		desc, amount := "", ""
		if m != nil {
			formattedCurrency, err := formatPaymentRequestCurrency(m.Amount, 24)
			if err != nil {
				return nil, err
			}

			// Convert description to full width
			fullWidthDescription := g.StringNormalizer.ToFullWidth(m.Description)

			// Convert amount to full width
			fullWidthAmount := g.StringNormalizer.ToFullWidth(formattedCurrency)

			desc = utils.LimitString(fullWidthDescription, 24)
			amount = utils.AddPrefixStringWithLimit(fullWidthAmount, " ", 24)
		}

		message = append(message, desc, amount)
	}

	return message, nil
}
