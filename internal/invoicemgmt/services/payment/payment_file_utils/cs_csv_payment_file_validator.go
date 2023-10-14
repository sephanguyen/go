package paymentfileutils

import (
	"context"
	"fmt"
	"time"
)

type ConvenienceStoreCSVPaymentFileValidator struct {
	*BasePaymentFileValidator
}

func (t *ConvenienceStoreCSVPaymentFileValidator) Validate(ctx context.Context, paymentFile *PaymentFile) (*PaymentValidationResult, error) {
	file := paymentFile.ConvenienceStoreFile
	records := make([]*GenericPaymentFileRecord, 0)

	for lineNo, dataRecord := range file.DataRecord {
		var transferredDate time.Time
		var receiveDate time.Time
		var err error

		if dataRecord.TransferredDate != 0 {
			transferredDateStr := fmt.Sprintf("%v", dataRecord.TransferredDate)
			if len(transferredDateStr) != 8 {
				return nil, fmt.Errorf("invalid transferred date value (%v) at line %v", transferredDateStr, lineNo+1)
			}

			transferredDate, err = time.Parse("20060102", transferredDateStr)
			if err != nil {
				return nil, fmt.Errorf("error parsing transferred date to string: %s has error: %v", transferredDateStr, err)
			}

			receiveDateStr := fmt.Sprintf("%v", dataRecord.DateOfReceipt)
			if len(transferredDateStr) != 8 {
				return nil, fmt.Errorf("invalid receive date value (%v) at line %v", receiveDateStr, lineNo+1)
			}

			receiveDate, err = time.Parse("20060102", receiveDateStr)
			if err != nil {
				return nil, fmt.Errorf("error parsing receive date to string: %s has error: %v", receiveDateStr, err)
			}
		}

		record := &GenericPaymentFileRecord{
			Amount:        dataRecord.Amount,
			PaymentDate:   &transferredDate,
			ValidatedDate: &receiveDate,
			PaymentNumber: dataRecord.CodeForUser2,
			ResultCode:    dataRecord.Category,
			CreatedDate:   dataRecord.CreatedDate,
		}

		records = append(records, record)
	}

	genericFile := &GenericPaymentFile{
		GenericPaymentData: records,
		PaymentMethod:      ConvenienceStore,
	}

	return t.GenericValidate(ctx, genericFile)
}
