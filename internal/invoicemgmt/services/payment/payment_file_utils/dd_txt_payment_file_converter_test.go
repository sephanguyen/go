package paymentfileutils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDirectDebitTextPaymentFileConverter_ConvertFromBytesToPaymentFile(t *testing.T) {
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
					"%v\n%v\n%v\n%v",
					"19100000004819株式会社ケーイーシー                              11040162南都銀行           150生駒支店           10192714                 ",
					"21162ﾅﾝﾄｷﾞﾝｺ        125ﾅﾗｼﾔｸｼﾖ            32305030ｱﾝﾉｳﾝﾍﾟｲｼﾞ                    00000010001787                 0        ",
					"8000001000000001000000000000000000000000000000000000000                                                                 ",
					"9                                                                                                                      ",
				),
			),
			expectedResult: &PaymentFile{
				DirectDebitFile: &DirectDebitFile{
					Header: &DirectDebitFileHeaderRecord{
						DataCategory:     1,
						TypeCode:         91,
						CodeCategory:     0,
						ConsignorCode:    4819,
						ConsignorName:    "株式会社ケーイーシー",
						DepositDate:      1104,
						BankNumber:       162,
						BankName:         "南都銀行",
						BankBranchNumber: 150,
						BankBranchName:   "生駒支店",
						DepositItems:     1,
						AccountNumber:    "0192714",
						DummyFiller:      "",
					},
					Data: []*DirectDebitFileDataRecord{
						{
							DataCategory:            2,
							DepositBankNumber:       1162,
							DepositBankName:         "ﾅﾝﾄｷﾞﾝｺ",
							DepositBankBranchNumber: 125,
							DepositBankBranchName:   "ﾅﾗｼﾔｸｼﾖ",
							DummyFiller1:            "",
							DepositItems:            3,
							AccountNumber:           "2305030",
							AccountOwnerName:        "ｱﾝﾉｳﾝﾍﾟｲｼﾞ",
							DepositAmount:           1000,
							NewCustomerCode:         "1",
							CustomerNumber:          "787",
							ResultCode:              "0",
							DummyFiller2:            "",
						},
					},
					Trailer: &DirectDebitFileTrailerRecord{
						DataCategory:      8,
						TotalTransactions: 1,
						TotalAmount:       1000,
						TransferredNumber: 0,
						TransferredAmount: 0,
						FailedNumber:      0,
						FailedAmount:      0,
						DummyFiller:       "",
					},
					End: &DirectDebitFileEndRecord{
						DataCategory: 9,
						DummyFiller:  "",
					},
				},
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "Convert with pure EN character",
			ctx:  ctx,
			fileContent: []byte(
				fmt.Sprintf(
					"%v\n%v\n%v\n%v",
					"19100000074632consignor-name-test-01GZFVHCXWTP5K9ARVB011033892partner-bank-na352partner-bank-br11442322                 ",
					"25299bank-name-phone436bank-branch-pho    11234567bank-account-holder-01GZFVHGWC00000010000193                 0        ",
					"8000001000000001000000000000000000000000000000000000000                                                                 ",
					"9                                                                                                                       ",
				),
			),
			expectedResult: &PaymentFile{
				DirectDebitFile: &DirectDebitFile{
					Header: &DirectDebitFileHeaderRecord{
						DataCategory:     1,
						TypeCode:         91,
						CodeCategory:     0,
						ConsignorCode:    74632,
						ConsignorName:    "consignor-name-test-01GZFVHCXWTP5K9ARVB0",
						DepositDate:      1103,
						BankNumber:       3892,
						BankName:         "partner-bank-na",
						BankBranchNumber: 352,
						BankBranchName:   "partner-bank-br",
						DepositItems:     1,
						AccountNumber:    "1442322",
						DummyFiller:      "",
					},
					Data: []*DirectDebitFileDataRecord{
						{
							DataCategory:            2,
							DepositBankNumber:       5299,
							DepositBankName:         "bank-name-phone",
							DepositBankBranchNumber: 436,
							DepositBankBranchName:   "bank-branch-pho",
							DummyFiller1:            "",
							DepositItems:            1,
							AccountNumber:           "1234567",
							AccountOwnerName:        "bank-account-holder-01GZFVHGWC",
							DepositAmount:           1000,
							NewCustomerCode:         "0",
							CustomerNumber:          "193",
							ResultCode:              "0",
							DummyFiller2:            "",
						},
					},
					Trailer: &DirectDebitFileTrailerRecord{
						DataCategory:      8,
						TotalTransactions: 1,
						TotalAmount:       1000,
						TransferredNumber: 0,
						TransferredAmount: 0,
						FailedNumber:      0,
						FailedAmount:      0,
						DummyFiller:       "",
					},
					End: &DirectDebitFileEndRecord{
						DataCategory: 9,
						DummyFiller:  "",
					},
				},
			},
			setup: func(ctx context.Context) {},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			converter := &DirectDebitTextPaymentFileConverter{}
			result, err := converter.ConvertFromBytesToPaymentFile(testCase.ctx, testCase.fileContent)

			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResult, result)
		})
	}
}
