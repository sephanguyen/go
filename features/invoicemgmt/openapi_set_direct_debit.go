package invoicemgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	http_port "github.com/manabie-com/backend/internal/invoicemgmt/services/http"
	openapisvc "github.com/manabie-com/backend/internal/invoicemgmt/services/open_api"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) thisStudentIsIncludedOnBankOpenAPIPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &UpsertStudentBankRequestInfo{}

	payload := s.generateBankAPIPayload(ctx)

	err := json.Unmarshal(payload, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreExistingBankAndBankBranch(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateBank(ctx, s.InvoiceMgmtPostgresDBTrace, false),
		s.EntitiesCreator.CreateBankBranch(ctx, s.InvoiceMgmtPostgresDBTrace, false),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateBankAPIPayload(ctx context.Context) []byte {
	return genSetDirectDebitPayload(genDefaultStudentBankRequestInfoProfile(ctx))
}

func (s *suite) bankInfoOfTheStudentWasUpsertedSuccessfullyByOpenAPI(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp := s.Response.(http_port.Response)
	if resp.Code != 20000 {
		return ctx, fmt.Errorf("message: %s, code: %d", resp.Message, resp.Code)
	}

	bankrequestInfo := stepState.Request.(*UpsertStudentBankRequestInfo)

	bankAccount, err := (&repositories.BankAccountRepo{}).FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "BankAccountRepo.FindByStudentID")
	}

	if bankAccount.BankAccountHolder.String != bankrequestInfo.StudentBankRequestInfo.BankAccountHolder.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err on bank account holder expected: %v but got %v", bankrequestInfo.StudentBankRequestInfo.BankAccountHolder.String, bankAccount.BankAccountHolder.String)
	}

	if bankAccount.BankID.String != stepState.BankID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err on bank id expected: %v but got %v", stepState.BankID, bankAccount.BankID.String)
	}

	if bankAccount.BankBranchID.String != stepState.BankBranchID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err on bank branch id expected: %v but got %v", stepState.BankBranchID, bankAccount.BankBranchID.String)
	}

	if bankAccount.BankAccountNumber.String != bankrequestInfo.StudentBankRequestInfo.BankAccountNumber.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err on bank account number expected: %v but got %v", bankrequestInfo.StudentBankRequestInfo.BankAccountNumber.String, bankAccount.BankAccountNumber.String)
	}

	bankAccountType := invoice_pb.BankAccountType_value[bankAccount.BankAccountType.String]

	if bankAccountType != bankrequestInfo.StudentBankRequestInfo.BankAccountType.Int {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err on bank account type expected: %d but got %d", bankrequestInfo.StudentBankRequestInfo.BankAccountType, bankAccountType)
	}

	if bankAccount.IsVerified.Bool != bankrequestInfo.StudentBankRequestInfo.IsVerified.Bool {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err on bank bank account number expected: %t but got %t", bankrequestInfo.StudentBankRequestInfo.IsVerified.Bool, bankAccount.IsVerified.Bool)
	}

	studentPaymentDetail, err := (&repositories.StudentPaymentDetailRepo{}).FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "StudentPaymentDetailRepo.FindByStudentID")
	}

	switch bankAccount.IsVerified.Bool {
	case true:
		if studentPaymentDetail.PaymentMethod.String != invoice_pb.PaymentMethod_DIRECT_DEBIT.String() {
			return StepStateToContext(ctx, stepState), errors.New("payment method should be direct debit")
		}
	default:
		if studentPaymentDetail.PaymentMethod.String != invoice_pb.PaymentMethod_CONVENIENCE_STORE.String() {
			return StepStateToContext(ctx, stepState), errors.New("payment method should be convenience store")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminSubmitsTheBankOpenAPIPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	jsonPayload, err := json.Marshal(stepState.Request)
	if err != nil {
		return ctx, errors.Wrap(err, "json.Marshal")
	}

	rsc := bootstrap.NewResources().WithLoggerC(&configs.CommonConfig{Name: "gandalf", Environment: "local", ActualEnvironment: "local", Organization: "manabie"})

	url := fmt.Sprintf(`http://%s%s`, rsc.GetHTTPAddress("invoicemgmt"), constant.StudentBankInfoEndpoint)

	bodyBytes, err := s.makeHTTPRequest(http.MethodPut, url, jsonPayload)
	if err != nil {
		return ctx, errors.Wrap(err, "s.makeHTTPRequest")
	}

	if bodyBytes == nil {
		return ctx, fmt.Errorf("body is nil")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receivesFailedResponseCodeFromOpenAPI(ctx context.Context, resultCode int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp := s.Response.(http_port.Response)
	if resp.Code != resultCode {
		return ctx, fmt.Errorf("err on response result code expected: %d but got %d", resultCode, resp.Code)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentIsIncludedOnBankOpenAPIInvalidPayload(ctx context.Context, conditions string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &UpsertStudentBankRequestInfo{}

	defaultStudentBankReqInfo := genDefaultStudentBankRequestInfoProfile(ctx)

	var payload []byte
	switch conditions {
	// conditions for other scenarios are already been tested on unit test
	case "internal error":
		stepState.BankCode = "test-not-exist"
		payload = s.generateBankAPIPayload(ctx)
	case "missing field":
		defaultStudentBankReqInfo.ExternalUserID = database.Text(`""`)
		payload = genSetDirectDebitPayload(defaultStudentBankReqInfo)
	case "invalid public key":
		stepState.BankOpenAPIPublicKey = "test-public-key"
		payload = s.generateBankAPIPayload(ctx)
	case "invalid private key":
		stepState.BankOpenAPIPrivateKey = "test-private-key"
		payload = s.generateBankAPIPayload(ctx)
	case "invalid account number with verified account":
		defaultStudentBankReqInfo.BankAccountNumber = database.Text(`"abcd567"`)
		defaultStudentBankReqInfo.IsVerified = database.Bool(true)
		payload = genSetDirectDebitPayload(defaultStudentBankReqInfo)
	case "invalid account number with unverified account":
		defaultStudentBankReqInfo.BankAccountNumber = database.Text(`"sadakuu"`)
		defaultStudentBankReqInfo.IsVerified = database.Bool(false)
		payload = genSetDirectDebitPayload(defaultStudentBankReqInfo)
	case "invalid account type with verified account":
		defaultStudentBankReqInfo.BankAccountType = database.Int4(-3)
		defaultStudentBankReqInfo.IsVerified = database.Bool(false)
		payload = genSetDirectDebitPayload(defaultStudentBankReqInfo)
	case "invalid account type with unverified account":
		defaultStudentBankReqInfo.BankAccountType = database.Int4(5)
		defaultStudentBankReqInfo.IsVerified = database.Bool(true)
		payload = genSetDirectDebitPayload(defaultStudentBankReqInfo)
	}

	err := json.Unmarshal(payload, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminAlreadySetupAnAPIUser(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.setupAPIKey(ctx)
	if err != nil {
		return ctx, errors.Wrap(err, "s.setupAPIKey")
	}

	if err := try.Do(func(attempt int) (bool, error) {
		var count int
		stmt := `
			SELECT count(r.role_id) FROM user_group_member ugm
				INNER JOIN granted_role gt ON ugm.user_group_id = gt.user_group_id
				INNER JOIN role r ON gt.role_id = r.role_id
			WHERE ugm.user_id = $1
				AND gt.deleted_at IS NULL
				AND ugm.deleted_at IS NULL
				AND r.deleted_at IS NULL
				AND r.role_name = 'OpenAPI'
			`
		row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, stepState.CurrentUserID)
		err := row.Scan(&count)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}
		if count != 0 {
			return false, nil
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("role open api not found for current user id: %v", stepState.CurrentUserID)
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func genDefaultStudentBankRequestInfoProfile(ctx context.Context) *openapisvc.StudentBankRequestInfoProfile {
	stepState := StepStateFromContext(ctx)

	return &openapisvc.StudentBankRequestInfoProfile{
		ExternalUserID:    database.Text(fmt.Sprintf(`"external-%v"`, stepState.StudentID)),
		BankCode:          database.Text(fmt.Sprintf(`"%v"`, stepState.BankCode)),
		BankBranchCode:    database.Text(fmt.Sprintf(`"%v"`, stepState.BankBranchCode)),
		BankAccountNumber: database.Text(fmt.Sprintf(`"%v"`, "1234567")),
		BankAccountHolder: database.Text(`"TESTHOLDER"`),
		BankAccountType:   database.Int4(1),
		IsVerified:        database.Bool(true),
	}
}

func genSetDirectDebitPayload(req *openapisvc.StudentBankRequestInfoProfile) []byte {
	return []byte(fmt.Sprintf(`{"student_bank_info": {"external_user_id": %s,"bank_code": %s,"bank_branch_code": %s,"bank_account_number": %s,"bank_account_holder": %s,"bank_account_type": %d,"is_verified": %t}}`, req.ExternalUserID.String, req.BankCode.String, req.BankBranchCode.String, req.BankAccountNumber.String, req.BankAccountHolder.String, req.BankAccountType.Int, req.IsVerified.Bool))
}
