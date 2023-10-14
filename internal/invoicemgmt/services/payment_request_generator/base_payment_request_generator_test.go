package generator

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
)

func genMockDataMap(count int) []*dataMap {
	e := []*dataMap{}

	for i := 0; i < count; i++ {
		id := idutil.ULIDNow()

		e = append(e, &dataMap{
			Payment: &entities.Payment{
				PaymentID: database.Text(id),
			},
			Invoice: &entities.Invoice{
				InvoiceID: database.Text(id),
			},
		})
	}

	return e
}

func Test_generatePaymentAndFileAssocs(t *testing.T) {

	type testCase struct {
		dataMap               []*dataMap
		maximumPaymentPerFile int
		fileName              string
		fileExtension         string
		name                  string

		expected []paymentAndFileAssoc
	}

	testFileName := "test-file-name"
	txtExtension := "txt"

	mockDataMap := genMockDataMap(3)

	testCases := []testCase{
		{
			name:                  "",
			dataMap:               mockDataMap,
			maximumPaymentPerFile: 10,
			fileName:              testFileName,
			fileExtension:         txtExtension,
			expected: []paymentAndFileAssoc{
				{
					DataMap:        mockDataMap,
					TotalFileCount: 1,
					FileName:       fmt.Sprintf("%s.%s", testFileName, txtExtension),
				},
			},
		},
		{
			dataMap:               mockDataMap,
			maximumPaymentPerFile: 2,
			fileName:              testFileName,
			fileExtension:         txtExtension,
			expected: []paymentAndFileAssoc{
				{
					DataMap:        mockDataMap,
					TotalFileCount: 2,
					FileName:       fmt.Sprintf("%s_1of2.%s", testFileName, txtExtension),
				},
				{
					DataMap:        mockDataMap,
					TotalFileCount: 2,
					FileName:       fmt.Sprintf("%s_2of2.%s", testFileName, txtExtension),
				},
			},
		},
	}

	for _, tc := range testCases {

		res := generatePaymentAndFileAssocs(tc.dataMap, tc.maximumPaymentPerFile, tc.fileName, tc.fileExtension)

		if len(res) != len(tc.expected) {
			t.Errorf("Expecting length of result to be %d got %d", len(tc.expected), len(res))
		}

		for i, pf := range tc.expected {

			assert.Equal(t, pf.FileName, res[i].FileName)
			assert.Equal(t, pf.TotalFileCount, res[i].TotalFileCount)
			assert.Equal(t, pf.FileName, res[i].FileName)

		}

	}

}

func Test_generateBulkPaymentRequestFileEntity(t *testing.T) {

	type testCase struct {
		paymentFileAssoc     paymentAndFileAssoc
		bilkPaymentRequestID string

		expected *entities.BulkPaymentRequestFile
	}

	bulkPaymentRequestID := "bulk-payment-request-id-1"
	bulkPaymentRequestFileID := "bulk-payment-request-file-id-1"
	testFileName := "test-file-name"

	parentbulkPaymentRequestFileID := "bulk-payment-request-file-id-1"
	bulkPaymentRequestFileID2 := "bulk-payment-request-file-id-2"

	e1 := new(entities.BulkPaymentRequestFile)
	database.AllNullEntity(e1)

	_ = multierr.Combine(
		e1.BulkPaymentRequestID.Set(bulkPaymentRequestID),
		e1.BulkPaymentRequestFileID.Set(bulkPaymentRequestFileID),
		e1.FileName.Set(testFileName),
		e1.FileSequenceNumber.Set(1),
		e1.TotalFileCount.Set(1),
		e1.IsDownloaded.Set(false),
	)

	e2 := new(entities.BulkPaymentRequestFile)
	database.AllNullEntity(e2)

	_ = multierr.Combine(
		e2.BulkPaymentRequestID.Set(bulkPaymentRequestID),
		e2.BulkPaymentRequestFileID.Set(bulkPaymentRequestFileID2),
		e2.FileName.Set(testFileName),
		e2.FileSequenceNumber.Set(1),
		e2.TotalFileCount.Set(1),
		e2.IsDownloaded.Set(false),
		e2.ParentPaymentRequestFileID.Set(parentbulkPaymentRequestFileID),
	)

	testCases := []testCase{
		{
			paymentFileAssoc: paymentAndFileAssoc{
				TotalFileCount:       1,
				FileName:             testFileName,
				FileSequenceNumber:   1,
				PaymentRequestFileID: bulkPaymentRequestFileID,
			},
			bilkPaymentRequestID: bulkPaymentRequestID,
			expected:             e1,
		},
		{
			paymentFileAssoc: paymentAndFileAssoc{
				TotalFileCount:             1,
				FileName:                   testFileName,
				FileSequenceNumber:         1,
				PaymentRequestFileID:       bulkPaymentRequestFileID2,
				ParentPaymentRequestFileID: parentbulkPaymentRequestFileID,
			},
			bilkPaymentRequestID: bulkPaymentRequestID,
			expected:             e2,
		},
	}

	for _, tc := range testCases {

		res, err := generateBulkPaymentRequestFileEntity(tc.bilkPaymentRequestID, tc.paymentFileAssoc)
		if err != nil {
			t.Error("Error occured", err)
		}

		assert.Equal(t, tc.expected.BulkPaymentRequestID.String, res.BulkPaymentRequestID.String)
		assert.Equal(t, tc.expected.BulkPaymentRequestFileID.String, res.BulkPaymentRequestFileID.String)
		assert.Equal(t, tc.expected.ParentPaymentRequestFileID.String, res.ParentPaymentRequestFileID.String)
	}

}

func Test_getRelatedBankMap(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockBankBranchRepo := new(mock_repositories.MockBankBranchRepo)

	ddGenerator := &DDtxtPaymentRequestGenerator{
		BasePaymentRequestGenerator: &BasePaymentRequestGenerator{
			BankBranchRepo: mockBankBranchRepo,
		},
	}

	bankBranchID := "test1"

	type testCase struct {
		name         string
		ctx          context.Context
		expectedResp map[string]*entities.BankRelationMap
		setup        func(ctx context.Context)
	}

	testCases := []testCase{
		{
			name: "Bank branch only have 1 partner bank with default is false",
			ctx:  ctx,
			expectedResp: map[string]*entities.BankRelationMap{
				bankBranchID: {
					PartnerBank: &entities.PartnerBank{
						PartnerBankID: database.Text("test1"),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.BankRelationMap{
					{
						PartnerBank: &entities.PartnerBank{
							PartnerBankID: database.Text("test1"),
							IsDefault:     database.Bool(false),
						},
						BankBranch: &entities.BankBranch{
							BankBranchID: database.Text(bankBranchID),
						},
					},
				}, nil)
			},
		},
		{
			name: "Bank branch have two partner bank with default",
			ctx:  ctx,
			expectedResp: map[string]*entities.BankRelationMap{
				bankBranchID: {
					PartnerBank: &entities.PartnerBank{
						PartnerBankID: database.Text("test2"),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.BankRelationMap{
					{
						PartnerBank: &entities.PartnerBank{
							PartnerBankID: database.Text("test1"),
							IsDefault:     database.Bool(false),
						},
						BankBranch: &entities.BankBranch{
							BankBranchID: database.Text(bankBranchID),
						},
					},
					{
						PartnerBank: &entities.PartnerBank{
							PartnerBankID: database.Text("test2"),
							IsDefault:     database.Bool(true),
						},
						BankBranch: &entities.BankBranch{
							BankBranchID: database.Text(bankBranchID),
						},
					},
				}, nil)
			},
		},
		{
			name: "Bank branch have two partner bank with default as first returned partner bank",
			ctx:  ctx,
			expectedResp: map[string]*entities.BankRelationMap{
				bankBranchID: {
					PartnerBank: &entities.PartnerBank{
						PartnerBankID: database.Text("test1"),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.BankRelationMap{
					{
						PartnerBank: &entities.PartnerBank{
							PartnerBankID: database.Text("test1"),
							IsDefault:     database.Bool(true),
						},
						BankBranch: &entities.BankBranch{
							BankBranchID: database.Text(bankBranchID),
						},
					},
					{
						PartnerBank: &entities.PartnerBank{
							PartnerBankID: database.Text("test2"),
							IsDefault:     database.Bool(false),
						},
						BankBranch: &entities.BankBranch{
							BankBranchID: database.Text(bankBranchID),
						},
					},
				}, nil)
			},
		},
		{
			name: "Bank branch have multiple partner bank with no default",
			ctx:  ctx,
			expectedResp: map[string]*entities.BankRelationMap{
				bankBranchID: {
					PartnerBank: &entities.PartnerBank{
						PartnerBankID: database.Text("test3"),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockBankBranchRepo.On("FindRelatedBankOfBankBranches", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.BankRelationMap{
					{
						PartnerBank: &entities.PartnerBank{
							PartnerBankID: database.Text("test1"),
							IsDefault:     database.Bool(false),
						},
						BankBranch: &entities.BankBranch{
							BankBranchID: database.Text(bankBranchID),
						},
					},
					{
						PartnerBank: &entities.PartnerBank{
							PartnerBankID: database.Text("test2"),
							IsDefault:     database.Bool(false),
						},
						BankBranch: &entities.BankBranch{
							BankBranchID: database.Text(bankBranchID),
						},
					},
					{
						PartnerBank: &entities.PartnerBank{
							PartnerBankID: database.Text("test3"),
							IsDefault:     database.Bool(false),
						},
						BankBranch: &entities.BankBranch{
							BankBranchID: database.Text(bankBranchID),
						},
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := ddGenerator.getRelatedBankMap(testCase.ctx, []string{bankBranchID})
			if err != nil {
				fmt.Println(err)
			}

			actual := response[bankBranchID]
			expected := testCase.expectedResp[bankBranchID]

			assert.Equal(t, actual.PartnerBank.PartnerBankID.String, expected.PartnerBank.PartnerBankID.String)
		})

	}

}
