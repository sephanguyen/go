package openapisvc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type HttpReponseTest struct {
	Data    interface{} `json:"data,omitempty"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
}

func TestOpenAPIModifierService_DirectDebitSetPaymentMethod(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID        = "user-id"
		unitTestEndpoint = "/unit-test"
	)

	mockDB := new(mock_database.Ext)
	mockTx := &mock_database.Tx{}
	mockUserRepo := &mock_repositories.MockUserRepo{}
	mockBankRepo := &mock_repositories.MockBankRepo{}
	mockBankBranchRepo := &mock_repositories.MockBankBranchRepo{}
	mockBillingAddressRepo := &mock_repositories.MockBillingAddressRepo{}
	mockStudentPaymentDetailRepo := &mock_repositories.MockStudentPaymentDetailRepo{}
	mockBankAccountRepo := &mock_repositories.MockBankAccountRepo{}
	mockStudentPaymentDetailActionLogRepo := &mock_repositories.MockStudentPaymentDetailActionLogRepo{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	s := &OpenAPIModifierService{
		DB:                                mockDB,
		UserRepo:                          mockUserRepo,
		BankRepo:                          mockBankRepo,
		BankBranchRepo:                    mockBankBranchRepo,
		BillingAddressRepo:                mockBillingAddressRepo,
		StudentPaymentDetailRepo:          mockStudentPaymentDetailRepo,
		BankAccountRepo:                   mockBankAccountRepo,
		UnleashClient:                     mockUnleashClient,
		StudentPaymentDetailActionLogRepo: mockStudentPaymentDetailActionLogRepo,
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// initial http request for ctx
	httpReq := &http.Request{
		Body: ioutil.NopCloser(bytes.NewReader([]byte{})),
	}
	c.Request = httpReq

	testError := errors.New("test-error")
	testUser := &entities.User{
		UserID: database.Text("test"),
	}
	testBank := &entities.Bank{
		BankID:   database.Text("1"),
		BankCode: database.Text("test-bank-code"),
	}

	testBankBranch := &entities.BankBranch{
		BankID:         database.Text("1"),
		BankBranchCode: database.Text("test-bank-branch-code"),
		BankBranchID:   database.Text("1"),
	}

	testBillingAddress := &entities.BillingAddress{
		BillingAddressID:       database.Text("1"),
		StudentPaymentDetailID: database.Text("1"),
	}

	testBankAccount := &entities.BankAccount{
		BankAccountID:          database.Text("1"),
		StudentPaymentDetailID: database.Text("1"),
	}

	testStudentPaymentDetail := &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text("1"),
		StudentID:              database.Text("1"),
		PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
	}

	testcases := []TestCase{
		{
			name: "happy case no existing bank account upsert convenience store",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    20000,
				Message: "success",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testBankBranch, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(testBillingAddress, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testStudentPaymentDetail, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBankAccountRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case with existing bank account upsert convenience store",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    20000,
				Message: "success",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testBankBranch, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(testBillingAddress, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testStudentPaymentDetail, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testBankAccount, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBankAccountRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case with existing bank account upsert direct debit",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    20000,
				Message: "success",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testBankBranch, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(testBillingAddress, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testStudentPaymentDetail, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testBankAccount, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBankAccountRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case no existing bank account upsert direct debit",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    20000,
				Message: "success",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testBankBranch, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(testBillingAddress, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testStudentPaymentDetail, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBankAccountRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed case missing external user id",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"bank_code": "",
								"bank_branch_code": "",
								"bank_account_number": "",
								"bank_account_holder": "",
								"bank_account_type": 0,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "external_user_id is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case missing bank code",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_branch_code": "",
								"bank_account_number": "",
								"bank_account_holder": "",
								"bank_account_type": 0,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "bank_code is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case missing bank branch code",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "",
								"bank_account_number": "",
								"bank_account_holder": "",
								"bank_account_type": 0,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "bank_branch_code is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case missing bank account number",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "",
								"bank_branch_code": "",
								"bank_account_holder": "",
								"bank_account_type": 0,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "bank_account_number is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case missing bank account holder",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "",
								"bank_branch_code": "",
								"bank_account_number": "",
								"bank_account_type": 0,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "bank_account_holder is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case missing bank account number",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "",
								"bank_branch_code": "",
								"bank_account_holder": "",
								"bank_account_type": 0,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "bank_account_number is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case missing bank account type",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "",
								"bank_branch_code": "",
								"bank_account_holder": "",
								"bank_account_number": "",
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "bank_account_type is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case missing is verified",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "",
								"bank_branch_code": "",
								"bank_account_holder": "",
								"bank_account_number": "",
								"bank_account_type": 0
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "is_verified is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case empty external user id verified false",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"",
								"bank_code": "",
								"bank_branch_code": "",
								"bank_account_holder": "",
								"bank_account_number": "",
								"bank_account_type": 0,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "external_user_id is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case empty external user id verified true",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"",
								"bank_code": "",
								"bank_branch_code": "",
								"bank_account_holder": "",
								"bank_account_number": "",
								"bank_account_type": 0,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "external_user_id is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case empty bank code verified true",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "",
								"bank_branch_code": "",
								"bank_account_holder": "",
								"bank_account_number": "",
								"bank_account_type": 0,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "bank_code is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case empty bank code verified false but there's bank branch code",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "",
								"bank_branch_code": "test",
								"bank_account_holder": "",
								"bank_account_number": "",
								"bank_account_type": 0,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "bank_code is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case empty bank branch code verified true",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "",
								"bank_account_holder": "",
								"bank_account_number": "",
								"bank_account_type": 0,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "bank_branch_code is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case empty bank account number verified true",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "",
								"bank_account_holder": "",
								"bank_account_type": 0,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "bank_account_number is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case empty bank account holder verified true",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "",
								"bank_account_type": 0,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40001,
				Message: "bank_account_holder is mandatory",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case bank account number is not 7 digits",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "123456",
								"bank_account_holder": "test",
								"bank_account_type": 0,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40004,
				Message: "bank_account_number is invalid",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case bank account number contains non numeric character",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "a2c4e6g",
								"bank_account_holder": "TEST",
								"bank_account_type": 0,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40004,
				Message: "bank_account_number is invalid",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case bank account holder not valid format",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "eXAMPLE - ｱ BRANCH - ｢123｣ ()  ﾟ ﾞ . ﾆ",
								"bank_account_type": 0,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40004,
				Message: "bank_account_holder is invalid",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case bank account type not valid format",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 3,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40004,
				Message: "bank_account_type is invalid",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case bank account type not valid format with unverified status",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 3,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    40004,
				Message: "bank_account_type is invalid",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "failed case external user id not existing when verified is true",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "failed case external user id not existing when verified is false",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "failed case bank code existing when verified is true",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "failed case bank code existing when verified is false",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "failed case bank branch existing when verified is true",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "failed case bank branch existing when verified is false",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "failed case no billing address record",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testBankBranch, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "failed error finding bank account",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testBankBranch, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(testBillingAddress, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testStudentPaymentDetail, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "failed no existing bank account upsert student payment detail convenience store payment method",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testBankBranch, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(testBillingAddress, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testStudentPaymentDetail, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed no existing bank account upsert student payment detail direct debit payment method",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testBankBranch, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(testBillingAddress, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testStudentPaymentDetail, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed existing bank account upsert student payment detail direct debit payment method",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testBankBranch, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(testBillingAddress, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testStudentPaymentDetail, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testBankAccount, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed existing bank account upsert student payment detail convenience store payment method",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testBankBranch, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(testBillingAddress, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testStudentPaymentDetail, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testBankAccount, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed existing bank account upsert bank account",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testBankBranch, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(testBillingAddress, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testStudentPaymentDetail, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testBankAccount, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBankAccountRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "failed no existing bank account upsert bank account",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": false
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testBankBranch, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(testBillingAddress, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testStudentPaymentDetail, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBankAccountRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "error on creating payment detail action log",
			ctx:  c.Request.Context(),
			PayloadByte: []byte(`{
						"student_bank_info":
							{
								"external_user_id":"test",
								"bank_code": "test-bank-code",
								"bank_branch_code": "test-bank-branch-code",
								"bank_account_number": "1234567",
								"bank_account_holder": "TEST",
								"bank_account_type": 1,
								"is_verified": true
							}
					}`),
			expectedResp: &HttpReponseTest{
				Code:    50000,
				Message: "internal error",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSetDirectDebitFeatureFlag, mock.Anything).Once().Return(true, nil)
				mockUserRepo.On("FindByUserExternalID", ctx, mockDB, mock.Anything).Once().Return(testUser, nil)
				mockBankRepo.On("FindByBankCode", ctx, mockDB, mock.Anything).Once().Return(testBank, nil)
				mockBankBranchRepo.On("FindByBankBranchCodeAndBank", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(testBankBranch, nil)
				mockBillingAddressRepo.On("FindByUserID", ctx, mockDB, mock.Anything).Once().Return(testBillingAddress, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testStudentPaymentDetail, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(testBankAccount, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockStudentPaymentDetailRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBankAccountRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockStudentPaymentDetailActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			httpReq := &http.Request{
				Body: ioutil.NopCloser(bytes.NewReader(testCase.PayloadByte)),
			}
			c.Request = httpReq
			s.UpsertStudentBankAccountInfo(c)

			body, err := ioutil.ReadAll(w.Body)
			if err != nil {
				panic(err)
			}

			var res HttpReponseTest
			err = json.Unmarshal(body, &res)
			if err != nil {
				panic(err)
			}

			assert.Equal(t, compareHttpResponse(testCase.expectedResp.(*HttpReponseTest), &res), true)

			mock.AssertExpectationsForObjects(t, mockDB, mockUserRepo, mockBankRepo, mockBankBranchRepo, mockBillingAddressRepo, mockStudentPaymentDetailRepo, mockBankAccountRepo, mockUnleashClient, mockStudentPaymentDetailActionLogRepo)
		})
	}
}

func compareHttpResponse(expectedResp *HttpReponseTest, actualResp *HttpReponseTest) bool {
	if expectedResp.Code != actualResp.Code {
		log.Fatalf("response code expected: %v but got %v", expectedResp.Code, actualResp.Code)
		return false
	}

	if expectedResp.Message != actualResp.Message {
		log.Fatalf("response message expected: %v but got %v", expectedResp.Message, actualResp.Message)
		return false
	}

	return true
}
