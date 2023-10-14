package paymentfileutils

import (
	"context"
)

type PaymentFileConverter interface {
	ConvertFromBytesToPaymentFile(ctx context.Context, file []byte) (*PaymentFile, error)
}

type PaymentFile struct {
	DirectDebitFile      *DirectDebitFile
	ConvenienceStoreFile *ConvenienceStoreFile
}

const (
	DataTypeHeaderRecord  = 1
	DataTypeDataRecord    = 2
	DataTypeTrailerRecord = 8
	DataTypeEndRecord     = 9
)
