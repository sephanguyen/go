package paymentfileutils

import (
	"context"
	"fmt"

	gocsv "github.com/gocarina/gocsv"
)

type ConvenienceStoreCSVPaymentFileConverter struct{}

type ConvenienceStoreFile struct {
	DataRecord []*ConvenienceStoreFileDataRecord
}

type ConvenienceStoreFileDataRecord struct {
	Category                   string `csv:"種別"`
	DateOfReceipt              int    `csv:"収納日"`
	TimeOfReceipt              int    `csv:"収納時刻"`
	BarcodeInformation         string `csv:"バーコード情報"`
	CodeForUser1               string `csv:"ユーザー使用欄１"`
	CodeForUser2               string `csv:"ユーザー使用欄２"`
	RevenueStamp               int    `csv:"印紙フラグ"`
	Amount                     int    `csv:"金額"`
	ConvenienceStoreHQCode     int    `csv:"コンビニ本部コード"`
	ConvenienceStoreName       string `csv:"コンビニ名"`
	ConvenienceStoreBranchCode int    `csv:"コンビニ店舗コード"`
	TransferredDate            int    `csv:"振込日"`
	CreatedDate                int    `csv:"データ作成日"`
}

func (t *ConvenienceStoreCSVPaymentFileConverter) ConvertFromBytesToPaymentFile(ctx context.Context, file []byte) (*PaymentFile, error) {
	var data []*ConvenienceStoreFileDataRecord

	err := gocsv.UnmarshalBytes(file, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling CSV file: %v", err)
	}

	paymentFile := &PaymentFile{
		ConvenienceStoreFile: &ConvenienceStoreFile{
			DataRecord: data,
		},
	}

	return paymentFile, nil
}
