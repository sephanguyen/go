package paymentfileutils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvenienceStoreCSVPaymentFileConverter_ConvertFromBytesToPaymentFile(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	type testCase struct {
		name           string
		fileContent    []byte
		ctx            context.Context
		expectedResult *PaymentFile
		expectedErr    error
		setup          func(ctx context.Context)
	}

	testCases := []testCase{
		{
			name: "Convert with JP character",
			ctx:  ctx,
			fileContent: []byte(
				fmt.Sprintf(
					"%v\n%v",
					"種別,収納日,収納時刻,バーコード情報,ユーザー使用欄１,ユーザー使用欄２,印紙フラグ,金額,コンビニ本部コード,コンビニ名,コンビニ店舗コード,振込日,データ作成日",
					"02,20230504,944,EAN91929023219420000000000037554399999999999999,00000,0000375543,0,12345,10,セブン－イレブン,13233,0,20230504",
				),
			),
			expectedResult: &PaymentFile{
				ConvenienceStoreFile: &ConvenienceStoreFile{
					DataRecord: []*ConvenienceStoreFileDataRecord{
						{
							Category:                   "02",
							DateOfReceipt:              20230504,
							TimeOfReceipt:              944,
							BarcodeInformation:         "EAN91929023219420000000000037554399999999999999",
							CodeForUser1:               "00000",
							CodeForUser2:               "0000375543",
							RevenueStamp:               0,
							Amount:                     12345,
							ConvenienceStoreHQCode:     10,
							ConvenienceStoreName:       "セブン－イレブン",
							ConvenienceStoreBranchCode: 13233,
							TransferredDate:            0,
							CreatedDate:                20230504,
						},
					},
				},
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "Convert with EN character",
			ctx:  ctx,
			fileContent: []byte(
				fmt.Sprintf(
					"%v\n%v",
					"種別,収納日,収納時刻,バーコード情報,ユーザー使用欄１,ユーザー使用欄２,印紙フラグ,金額,コンビニ本部コード,コンビニ名,コンビニ店舗コード,振込日,データ作成日",
					"02,20230504,944,EAN91929023219420000000000037554399999999999999,00000,0000375543,0,12345,10,MiniGo,13233,0,20230504",
				),
			),
			expectedResult: &PaymentFile{
				ConvenienceStoreFile: &ConvenienceStoreFile{
					DataRecord: []*ConvenienceStoreFileDataRecord{
						{
							Category:                   "02",
							DateOfReceipt:              20230504,
							TimeOfReceipt:              944,
							BarcodeInformation:         "EAN91929023219420000000000037554399999999999999",
							CodeForUser1:               "00000",
							CodeForUser2:               "0000375543",
							RevenueStamp:               0,
							Amount:                     12345,
							ConvenienceStoreHQCode:     10,
							ConvenienceStoreName:       "MiniGo",
							ConvenienceStoreBranchCode: 13233,
							TransferredDate:            0,
							CreatedDate:                20230504,
						},
					},
				},
			},
			setup: func(ctx context.Context) {},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			converter := &ConvenienceStoreCSVPaymentFileConverter{}
			result, err := converter.ConvertFromBytesToPaymentFile(testCase.ctx, testCase.fileContent)

			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResult, result)
		})
	}
}
