package paymentfileutils

import (
	"context"
	"time"
)

type DirectDebitTextPaymentFileValidator struct {
	*BasePaymentFileValidator
	PaymentDate time.Time
}

func (t *DirectDebitTextPaymentFileValidator) Validate(ctx context.Context, paymentFile *PaymentFile) (*PaymentValidationResult, error) {
	file := paymentFile.DirectDebitFile

	records := make([]*GenericPaymentFileRecord, 0)
	for _, dataRecord := range file.Data {
		record := &GenericPaymentFileRecord{
			Amount:        dataRecord.DepositAmount,
			PaymentDate:   &t.PaymentDate,
			PaymentNumber: dataRecord.CustomerNumber,
			ResultCode:    dataRecord.ResultCode,
		}

		records = append(records, record)
	}

	genericFile := &GenericPaymentFile{
		GenericPaymentData:     records,
		PaymentMethod:          DirectDebit,
		TransferredTotalAmount: float64(file.Trailer.TransferredAmount),
		TransferredNumber:      file.Trailer.TransferredNumber,
		FailedTotalAmount:      float64(file.Trailer.FailedAmount),
		FailedNumber:           file.Trailer.FailedNumber,
	}

	return t.GenericValidate(ctx, genericFile)
}
