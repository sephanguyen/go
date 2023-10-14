package invoicemgmt

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	invoice_common "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/common"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) anExistingStudentWithBillingOrBankAccountInfo(ctx context.Context, billingInfo string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	const billingAddress = "billing address"

	// Use createStudent to add access path to student (for security filter)
	ctx, err := s.createStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	studentID := stepState.StudentID
	stepState.StudentIds = append(stepState.StudentIds, studentID)

	switch billingInfo {
	case billingAddress, "billing address and bank account":
		err := InsertEntities(
			stepState,
			s.EntitiesCreator.CreatePrefecture(ctx, s.BobDBTrace),
			s.EntitiesCreator.CreateBank(ctx, s.InvoiceMgmtPostgresDBTrace, false),
			s.EntitiesCreator.CreateBankBranch(ctx, s.InvoiceMgmtPostgresDBTrace, false),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		request := aValidUpsertStudentPaymentInfo(studentID, "", "", stepState.PrefectureCode, "", stepState.BankID, stepState.BankBranchID)

		if billingInfo == billingAddress {
			request.BankAccountInfo = nil
		}
		stepState.Request = request
		if _, err := s.callUpsertBillingInfoForStudent(ctx, "school admin", request); err != nil {
			return StepStateToContext(ctx, stepState), errors.Wrap(err, "callUpsertBillingInfoForStudent")
		}
		if err := s.validStudentPaymentDetail(ctx, s.InvoiceMgmtDB, request); err != nil {
			return StepStateToContext(ctx, stepState), errors.Wrap(err, "validStudentPaymentDetail")
		}
		if err := validatePaymentMethod(ctx, s.InvoiceMgmtPostgresDBTrace, request, entities.PaymentMethodConvenienceStore); err != nil {
			return StepStateToContext(ctx, stepState), errors.Wrap(err, "validatePaymentMethod")
		}
		if err := validBillingAddress(ctx, s.InvoiceMgmtDB, request); err != nil {
			return StepStateToContext(ctx, stepState), errors.Wrap(err, "validBillingAddress")
		}
	case "non existing student payment detail":
		stepState.StudentPaymentDetailID = idutil.ULIDNow()
	case "no billing address detail":
		err := InsertEntities(
			stepState,
			s.EntitiesCreator.UpsertStudentPaymentDetail(ctx, s.InvoiceMgmtPostgresDBTrace, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func aValidUpsertStudentPaymentInfo(studentID string, studentPaymentDetailID string, billingAddressID string, prefectureCode string, bankAccountID string, bankID string, bankBranchID string) *invoice_pb.UpsertStudentPaymentInfoRequest {
	randomInt, _ := rand.Int(rand.Reader, big.NewInt(2))
	request := &invoice_pb.UpsertStudentPaymentInfoRequest{
		StudentId: studentID,
		BillingInfo: &invoice_pb.BillingInformation{
			StudentPaymentDetailId: studentPaymentDetailID,
			PayerName:              fmt.Sprintf("%s-payer_name", studentID),
			PayerPhoneNumber:       fmt.Sprintf("%s-payer_phone_number", studentID),
			BillingAddress: &invoice_pb.BillingAddress{
				BillingAddressId: billingAddressID,
				PostalCode:       fmt.Sprintf("%s-postal_code", studentID),
				PrefectureCode:   prefectureCode,
				City:             fmt.Sprintf("%s-city", studentID),
				Street1:          fmt.Sprintf("%s-street_1", studentID),
				Street2:          fmt.Sprintf("%s-street_2", studentID),
			},
		},
		BankAccountInfo: &invoice_pb.BankAccountInformation{
			BankAccountId:     bankAccountID,
			BankId:            bankID,
			BankBranchId:      bankBranchID,
			BankAccountHolder: fmt.Sprintf("%s-EXAMPLE BANK - ｱ BRANCH - ｢123｣ ()  ﾟ ﾞ . ﾆ", studentID),
			BankAccountNumber: "1234567",
			BankAccountType:   invoice_pb.BankAccountType(randomInt.Int64()),
			IsVerified:        false,
		},
	}
	return request
}

func (s *suite) callUpsertBillingInfoForStudent(ctx context.Context, userCallAPI string, request *invoice_pb.UpsertStudentPaymentInfoRequest) (*invoice_pb.UpsertStudentPaymentInfoResponse, error) {
	ctx, err := s.signedAsAccount(ctx, userCallAPI)
	if err != nil {
		return nil, err
	}
	return invoice_pb.NewEditPaymentDetailServiceClient(s.InvoiceMgmtConn).UpsertStudentPaymentInfo(contextWithToken(ctx), request)
}

func adjustUpsertStudentPaymentRequest(adjustType string, validRequest *invoice_pb.UpsertStudentPaymentInfoRequest) (*invoice_pb.UpsertStudentPaymentInfoRequest, error) {
	switch adjustType {
	case "is valid":
		break
	case "missing student id":
		validRequest.StudentId = ""
	case "missing billing information":
		validRequest.BillingInfo = nil
	case "missing both billing information and bank account":
		validRequest.BillingInfo = nil
		validRequest.BankAccountInfo = nil
	case "missing payer name":
		validRequest.BillingInfo.PayerName = ""
	case "missing payer phone number":
		validRequest.BillingInfo.PayerPhoneNumber = ""
	case "missing billing address":
		validRequest.BillingInfo.BillingAddress = nil
	case "missing postal code":
		validRequest.BillingInfo.BillingAddress.PostalCode = ""
	case "missing prefecture code":
		validRequest.BillingInfo.BillingAddress.PrefectureCode = ""
	case "has non-exist prefecture code":
		validRequest.BillingInfo.BillingAddress.PrefectureCode = idutil.ULIDNow()
	case "missing city":
		validRequest.BillingInfo.BillingAddress.City = ""
	case "missing street 1":
		validRequest.BillingInfo.BillingAddress.Street1 = ""
	case "missing street 2":
		validRequest.BillingInfo.BillingAddress.Street2 = ""
	case "has non-exist payment detail id":
		validRequest.BillingInfo.StudentPaymentDetailId = idutil.ULIDNow()
	case "has non-exist billing address id":
		validRequest.BillingInfo.BillingAddress.BillingAddressId = idutil.ULIDNow()
	case "is valid and has verified status":
		validRequest.BankAccountInfo.IsVerified = true
	case "is valid and has unverified status":
		validRequest.BankAccountInfo.IsVerified = false
	case "has non-exist bank id and verified status":
		validRequest.BankAccountInfo.BankId = idutil.ULIDNow()
		validRequest.BankAccountInfo.IsVerified = true
	case "has non-exist bank id and unverified status":
		validRequest.BankAccountInfo.BankId = idutil.ULIDNow()
		validRequest.BankAccountInfo.IsVerified = false
	case "missing bank id and has verified status":
		validRequest.BankAccountInfo.BankId = ""
		validRequest.BankAccountInfo.IsVerified = true
	case "missing bank id and has unverified status":
		validRequest.BankAccountInfo.BankId = ""
		validRequest.BankAccountInfo.IsVerified = false
	case "has non-exist bank branch id and verified status":
		validRequest.BankAccountInfo.BankBranchId = idutil.ULIDNow()
		validRequest.BankAccountInfo.IsVerified = true
	case "has non-exist bank branch id and unverified status":
		validRequest.BankAccountInfo.BankBranchId = idutil.ULIDNow()
		validRequest.BankAccountInfo.IsVerified = false
	case "missing bank branch id and has verified status":
		validRequest.BankAccountInfo.BankBranchId = ""
		validRequest.BankAccountInfo.IsVerified = true
	case "missing bank branch id and has unverified status":
		validRequest.BankAccountInfo.BankBranchId = ""
		validRequest.BankAccountInfo.IsVerified = false
	case "missing bank account holder and has verified status":
		validRequest.BankAccountInfo.BankAccountHolder = ""
		validRequest.BankAccountInfo.IsVerified = true
	case "has invalid alphabet in bank account holder and has verified status":
		validRequest.BankAccountInfo.BankAccountHolder = "abcdefghijklmnopqrstuvwxyz"
		validRequest.BankAccountInfo.IsVerified = true
	case "has invalid kana in bank account holder and has verified status":
		validRequest.BankAccountInfo.BankAccountHolder = "ｧｨｩｪｫｯｬｭｮ"
		validRequest.BankAccountInfo.IsVerified = true
	case "has invalid symbol in bank account holder and has verified status":
		validRequest.BankAccountInfo.BankAccountHolder = `~!@#$%^&*+=[]{}|;:'",<>./?`
		validRequest.BankAccountInfo.IsVerified = true
	case "missing bank account holder and has unverified status":
		validRequest.BankAccountInfo.BankAccountHolder = ""
		validRequest.BankAccountInfo.IsVerified = false
	case "has invalid alphabet in bank account holder and has unverified status":
		validRequest.BankAccountInfo.BankAccountHolder = "abcdefghijklmnopqrstuvwxyz"
		validRequest.BankAccountInfo.IsVerified = false
	case "has invalid kana in bank account holder and has unverified status":
		validRequest.BankAccountInfo.BankAccountHolder = "ｧｨｩｪｫｯｬｭｮ"
		validRequest.BankAccountInfo.IsVerified = false
	case "has invalid symbol in bank account holder and has unverified status":
		validRequest.BankAccountInfo.BankAccountHolder = `~!@#$%^&*+=[]{}|;:'",<>./?`
		validRequest.BankAccountInfo.IsVerified = false
	case "missing bank account number and has verified status":
		validRequest.BankAccountInfo.BankAccountNumber = ""
		validRequest.BankAccountInfo.IsVerified = true
	case "missing bank account number and has unverified status":
		validRequest.BankAccountInfo.BankAccountNumber = ""
		validRequest.BankAccountInfo.IsVerified = false
	case "missing bank account type and has verified status":
		validRequest.BankAccountInfo.BankAccountType = invoice_pb.BankAccountType(999)
		validRequest.BankAccountInfo.IsVerified = true
	case "missing bank account type and has unverified status":
		validRequest.BankAccountInfo.BankAccountType = invoice_pb.BankAccountType(999)
		validRequest.BankAccountInfo.IsVerified = false
	default:
		return nil, errors.New("doesn't support this type of request")
	}
	return validRequest, nil
}

func (s *suite) createABillingInformationThatForExistingStudent(ctx context.Context, user, typeOfInfo string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreatePrefecture(ctx, s.BobDBTrace),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	request := aValidUpsertStudentPaymentInfo(stepState.StudentID, "", "", stepState.PrefectureCode, "", "", "")
	request.BankAccountInfo = nil
	request, err = adjustUpsertStudentPaymentRequest(typeOfInfo, request)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Request = request

	ctx, err = s.signedAsAccount(ctx, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewEditPaymentDetailServiceClient(s.InvoiceMgmtConn).UpsertStudentPaymentInfo(contextWithToken(ctx), request)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createABankAccountThatForExistingStudent(ctx context.Context, user, typeOfInfo string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateBank(ctx, s.InvoiceMgmtPostgresDBTrace, false),
		s.EntitiesCreator.CreateBankBranch(ctx, s.InvoiceMgmtPostgresDBTrace, false),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	request := aValidUpsertStudentPaymentInfo(stepState.StudentID, "", "", "", "", stepState.BankID, stepState.BankBranchID)
	request.BillingInfo = nil
	request, err = adjustUpsertStudentPaymentRequest(typeOfInfo, request)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Request = request

	ctx, err = s.signedAsAccount(ctx, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewEditPaymentDetailServiceClient(s.InvoiceMgmtConn).UpsertStudentPaymentInfo(contextWithToken(ctx), request)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateWithNewBillingInformationThatForExistingStudent(ctx context.Context, user, typeOfInfo string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreatePrefecture(ctx, s.BobDBTrace),
	)

	studentID := stepState.StudentID
	studentPaymentDetail, err := (&repositories.StudentPaymentDetailRepo{}).FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, studentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "StudentPaymentDetailRepo.FindByID")
	}

	billingAddress, err := (&repositories.BillingAddressRepo{}).FindByUserID(ctx, s.InvoiceMgmtPostgresDBTrace, studentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "BillingAddressRepo.FindByID")
	}

	updateRequest := aValidUpsertStudentPaymentInfo(studentID, studentPaymentDetail.StudentPaymentDetailID.String, billingAddress.BillingAddressID.String, stepState.PrefectureCode, "", "", "")
	updateRequest.BankAccountInfo = nil
	now := time.Now()
	updateRequest.BillingInfo.PayerName += fmt.Sprintf("-updated_at_%d", now.UnixMilli())
	updateRequest.BillingInfo.PayerPhoneNumber += fmt.Sprintf("-updated_at_%d", now.UnixMilli())
	updateRequest.BillingInfo.BillingAddress.PostalCode += fmt.Sprintf("-updated_at_%d", now.UnixMilli())
	updateRequest.BillingInfo.BillingAddress.City += fmt.Sprintf("-updated_at_%d", now.UnixMilli())
	updateRequest.BillingInfo.BillingAddress.Street1 += fmt.Sprintf("-updated_at_%d", now.UnixMilli())
	updateRequest.BillingInfo.BillingAddress.Street2 += fmt.Sprintf("-updated_at_%d", now.UnixMilli())

	updateRequest, err = adjustUpsertStudentPaymentRequest(typeOfInfo, updateRequest)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = updateRequest

	stepState.Response, stepState.ResponseErr = s.callUpsertBillingInfoForStudent(ctx, user, updateRequest)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateWithNewBankAccountThatForExistingStudent(ctx context.Context, user, typeOfInfo string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentID := stepState.StudentID

	bankAccount, err := (&repositories.BankAccountRepo{}).FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, studentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "BankAccountRepo.FindByStudentID")
	}

	err = InsertEntities(
		stepState,
		s.EntitiesCreator.CreateBank(ctx, s.InvoiceMgmtPostgresDBTrace, false),
		s.EntitiesCreator.CreateBankBranch(ctx, s.InvoiceMgmtPostgresDBTrace, false),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	updateRequest := aValidUpsertStudentPaymentInfo(studentID, "", "", "", bankAccount.BankAccountID.String, stepState.BankID, stepState.BankBranchID)
	updateRequest.BillingInfo = nil

	now := time.Now()
	updateRequest.BankAccountInfo.BankAccountHolder += fmt.Sprintf("-UPDATED-AT-%d", now.UnixMilli())
	updateRequest.BankAccountInfo.BankAccountNumber = "7654321"
	for bankAccountTypeInt, bankAccountTypeString := range invoice_pb.BankAccountType_name {
		if updateRequest.BankAccountInfo.BankAccountType.String() != bankAccountTypeString {
			updateRequest.BankAccountInfo.BankAccountType = invoice_pb.BankAccountType(bankAccountTypeInt)
		}
	}

	updateRequest, err = adjustUpsertStudentPaymentRequest(typeOfInfo, updateRequest)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Request = updateRequest

	stepState.Response, stepState.ResponseErr = s.callUpsertBillingInfoForStudent(ctx, user, updateRequest)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) billingInformation(ctx context.Context, result string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request := stepState.Request.(*invoice_pb.UpsertStudentPaymentInfoRequest)

	switch result {
	case "failed to create", "failed to update":
		if stepState.ResponseErr == nil {
			return StepStateToContext(ctx, stepState), errors.New("expected error in response but actual is nil")
		}
	case "successfully created", "successfully updated":
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected no error in response but actual there is error: %w", stepState.ResponseErr)
		}
		if !stepState.Response.(*invoice_pb.UpsertStudentPaymentInfoResponse).Successful {
			return StepStateToContext(ctx, stepState), errors.New("expected true status in response but actual is false")
		}
		if err := s.validStudentPaymentDetail(ctx, s.InvoiceMgmtPostgresDBTrace, request); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if err := validBillingAddress(ctx, s.InvoiceMgmtPostgresDBTrace, request); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		expectedPaymentMethod := entities.PaymentMethodConvenienceStore
		if request.BankAccountInfo != nil && request.BankAccountInfo.IsVerified {
			expectedPaymentMethod = entities.PaymentMethodDirectDebit
		}
		if err := validatePaymentMethod(ctx, s.InvoiceMgmtPostgresDBTrace, request, expectedPaymentMethod); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) bankAccount(ctx context.Context, result string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request := stepState.Request.(*invoice_pb.UpsertStudentPaymentInfoRequest)

	switch result {
	case "failed to create", "failed to update":
		if stepState.ResponseErr == nil {
			return StepStateToContext(ctx, stepState), errors.New("expected error in response but actual is nil")
		}
	case "successfully created", "successfully updated":
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected no error in response but actual there is error: %w", stepState.ResponseErr)
		}
		if !stepState.Response.(*invoice_pb.UpsertStudentPaymentInfoResponse).Successful {
			return StepStateToContext(ctx, stepState), errors.New("expected true status in response but actual is false")
		}
		if err := validBankAccount(ctx, s.InvoiceMgmtPostgresDBTrace, request); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		expectedPaymentMethod := entities.PaymentMethodConvenienceStore
		if request.BankAccountInfo.IsVerified {
			expectedPaymentMethod = entities.PaymentMethodDirectDebit
		}
		if err := validatePaymentMethod(ctx, s.InvoiceMgmtPostgresDBTrace, request, expectedPaymentMethod); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validStudentPaymentDetail(ctx context.Context, db database.QueryExecer, request *invoice_pb.UpsertStudentPaymentInfoRequest) error {
	stepState := StepStateFromContext(ctx)
	studentPaymentDetail, err := (&repositories.StudentPaymentDetailRepo{}).FindByStudentID(ctx, db, request.StudentId)
	stepState.StudentPaymentDetailID = studentPaymentDetail.StudentPaymentDetailID.String

	if err != nil {
		return errors.Wrap(err, "StudentPaymentDetailRepo.FindByStudentID")
	}

	if studentPaymentDetail == nil {
		return errors.New("`expected student payment detail is inserted  actual is not`")
	}

	if request.BillingInfo.StudentPaymentDetailId == "" {
		//assertion for create mode
		if studentPaymentDetail.StudentPaymentDetailID.String == "" {
			return errors.New(`expected student payment detail has id after created but actual is empty`)
		}
	} else {
		//assertion for update mode
		if studentPaymentDetail.StudentPaymentDetailID.String != request.BillingInfo.StudentPaymentDetailId {
			return fmt.Errorf(`expected student payment detail id: "%s" but actual is: %s`, request.BillingInfo.StudentPaymentDetailId, studentPaymentDetail.StudentPaymentDetailID.String)
		}
	}

	switch {
	case studentPaymentDetail.PayerName.String != request.BillingInfo.PayerName:
		return fmt.Errorf(`expected payer name: "%s" but actual is: %s`, request.BillingInfo.PayerName, studentPaymentDetail.PayerName.String)
	case studentPaymentDetail.PayerPhoneNumber.String != request.BillingInfo.PayerPhoneNumber:
		return fmt.Errorf(`expected payer phone number: "%s" but actual is: %s`, request.BillingInfo.PayerPhoneNumber, studentPaymentDetail.PayerPhoneNumber.String)
	case studentPaymentDetail.DeletedAt.Status != pgtype.Null:
		return fmt.Errorf(`expected deletedAt is null but actual is: %s`, studentPaymentDetail.DeletedAt.Time)
	}

	return nil
}

func validBillingAddress(ctx context.Context, db database.QueryExecer, request *invoice_pb.UpsertStudentPaymentInfoRequest) error {
	billingAddress, err := (&repositories.BillingAddressRepo{}).FindByUserID(ctx, db, request.StudentId)
	if err != nil {
		return errors.Wrap(err, "StudentPaymentDetailRepo.FindByUserID")
	}

	if billingAddress == nil {
		return errors.New("`expected billing address is inserted  actual is not`")
	}

	if request.BillingInfo.StudentPaymentDetailId == "" {
		//assertion for create mode
		if billingAddress.StudentPaymentDetailID.String == "" {
			return errors.New(`expected student payment detail has id after created but actual is empty`)
		}
	} else {
		//assertion for update mode
		if billingAddress.StudentPaymentDetailID.String != request.BillingInfo.StudentPaymentDetailId {
			return errors.New(fmt.Sprintf(`expected student payment detail id: "%s" but actual is: %s`, request.BillingInfo.StudentPaymentDetailId, billingAddress.StudentPaymentDetailID.String))
		}
	}

	switch {
	case billingAddress.PostalCode.String != request.BillingInfo.BillingAddress.PostalCode:
		return errors.Errorf("expected postal code: %s but actual is: %s", request.BillingInfo.BillingAddress.PostalCode, billingAddress.PostalCode.String)
	case billingAddress.PrefectureCode.String != request.BillingInfo.BillingAddress.PrefectureCode:
		return errors.Errorf("expected prefecture code: %s but actual is: %s", request.BillingInfo.BillingAddress.PrefectureCode, billingAddress.PrefectureCode.String)
	case billingAddress.City.String != request.BillingInfo.BillingAddress.City:
		return errors.Errorf("expected city: %s but actual is: %s", request.BillingInfo.BillingAddress.City, billingAddress.City.String)
	case billingAddress.Street1.String != request.BillingInfo.BillingAddress.Street1:
		return errors.Errorf("expected street 1: %s but actual is: %s", request.BillingInfo.BillingAddress.Street1, billingAddress.Street1.String)
	case billingAddress.Street2.String != request.BillingInfo.BillingAddress.Street2:
		return errors.Errorf("expected street 2: %s but actual is: %s", request.BillingInfo.BillingAddress.Street2, billingAddress.Street2.String)
	case billingAddress.DeletedAt.Status != pgtype.Null:
		return errors.Errorf("expected deletedAt is null but actual is: %s", billingAddress.DeletedAt.Time)
	}

	return nil
}

func validatePaymentMethod(ctx context.Context, db database.QueryExecer, request *invoice_pb.UpsertStudentPaymentInfoRequest, expectedPaymentMethod string) error {
	studentPaymentDetail, err := (&repositories.StudentPaymentDetailRepo{}).FindByStudentID(ctx, db, request.StudentId)
	if err != nil {
		return errors.Wrap(err, "StudentPaymentDetailRepo.FindByStudentID")
	}

	if studentPaymentDetail == nil {
		return errors.New("`expected student payment detail is inserted  actual is not`")
	}

	if studentPaymentDetail.PaymentMethod.String != expectedPaymentMethod {
		return errors.New(fmt.Sprintf(`expected student payment method is "%s" but actual is: %s`, expectedPaymentMethod, studentPaymentDetail.PaymentMethod.String))
	}
	return nil
}

func validBankAccount(ctx context.Context, db database.QueryExecer, request *invoice_pb.UpsertStudentPaymentInfoRequest) error {
	bankAccount, err := (&repositories.BankAccountRepo{}).FindByStudentID(ctx, db, request.StudentId)
	if err != nil {
		return errors.Wrap(err, "BankAccountRepo.FindByStudentID")
	}

	if bankAccount == nil {
		return errors.New("`expected bank account is inserted  actual is not`")
	}

	if request.BillingInfo == nil {
		//assertion for create mode
		if bankAccount.StudentPaymentDetailID.String == "" {
			return errors.New(`expected student payment detail has id after created but actual is empty`)
		}
	} else {
		//assertion for update mode
		if bankAccount.StudentPaymentDetailID.String != request.BillingInfo.StudentPaymentDetailId {
			return errors.New(fmt.Sprintf(`expected student payment detail id: "%s" but actual is: %s`, request.BillingInfo.StudentPaymentDetailId, bankAccount.StudentPaymentDetailID.String))
		}
	}

	switch {
	case bankAccount.BankID.String != request.BankAccountInfo.BankId:
		return errors.New(fmt.Sprintf(`expected bank id: "%s" but actual is: %s`, request.BankAccountInfo.BankId, bankAccount.BankID.String))
	case bankAccount.BankBranchID.String != request.BankAccountInfo.BankBranchId:
		return errors.New(fmt.Sprintf(`expected bank branch id: "%s" but actual is: %s`, request.BankAccountInfo.BankBranchId, bankAccount.BankBranchID.String))
	case bankAccount.BankAccountHolder.String != request.BankAccountInfo.BankAccountHolder:
		return errors.New(fmt.Sprintf(`expected bank account holder: "%s" but actual is: %s`, request.BankAccountInfo.BankAccountHolder, bankAccount.BankAccountHolder.String))
	case bankAccount.BankAccountNumber.String != request.BankAccountInfo.BankAccountNumber:
		return errors.New(fmt.Sprintf(`expected bank account number: "%s" but actual is: %s`, request.BankAccountInfo.BankAccountNumber, bankAccount.BankAccountNumber.String))
	case bankAccount.BankAccountType.String != invoice_pb.BankAccountType_name[int32(request.BankAccountInfo.BankAccountType)]:
		return errors.New(fmt.Sprintf(`expected bank account type: "%s" but actual is: %s`, invoice_pb.BankAccountType_name[int32(request.BankAccountInfo.BankAccountType)], bankAccount.BankAccountType.String))
	case bankAccount.IsVerified.Bool != request.BankAccountInfo.IsVerified:
		return errors.New(fmt.Sprintf(`expected is verified: "%v" but actual is: %v`, request.BankAccountInfo.IsVerified, bankAccount.IsVerified.Bool))
	case bankAccount.DeletedAt.Status != pgtype.Null:
		return errors.New(fmt.Sprintf(`expected deletedAt is null but actual is: %s`, bankAccount.DeletedAt.Time))
	}

	return nil
}

func (s *suite) thisStudentBankAccountIsVerified(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	bankAccount, err := (&repositories.BankAccountRepo{}).FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "BankAccountRepo.FindByStudentID")
	}

	studentPaymentDetail, err := (&repositories.StudentPaymentDetailRepo{}).FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "StudentPaymentDetailRepo.FindByID")
	}

	billingAddress, err := (&repositories.BillingAddressRepo{}).FindByUserID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "BillingAddressRepo.FindByID")
	}

	request := aValidUpsertStudentPaymentInfo(stepState.StudentID,
		studentPaymentDetail.StudentPaymentDetailID.String, billingAddress.BillingAddressID.String, billingAddress.PrefectureCode.String,
		bankAccount.BankAccountID.String, stepState.BankID, stepState.BankBranchID)

	request.BankAccountInfo.IsVerified = true
	if _, err := s.callUpsertBillingInfoForStudent(ctx, "school admin", request); err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "callUpsertBillingInfoForStudent")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theStudentDefaultPaymentMethodWasSetTo(ctx context.Context, paymentMethodStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentPaymentDetail, err := (&repositories.StudentPaymentDetailRepo{}).FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "StudentPaymentDetailRepo.FindByID")
	}

	if studentPaymentDetail.PaymentMethod.String != paymentMethodStr {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting payment method to be %s got %s", paymentMethodStr, studentPaymentDetail.PaymentMethod.String)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) requestUpdatesOnStudentPaymentInfoWithInformation(ctx context.Context, updateAction, updateInfo string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request, err := s.retrieveRequestBasedOnUpdateAction(ctx, updateAction)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	updateInfoSlice := strings.Split(updateInfo, "-")
	now := time.Now()
	for _, updateInfoStr := range updateInfoSlice {
		switch updateInfoStr {
		case "PostalCode":
			request.BillingInfo.BillingAddress.PostalCode = fmt.Sprintf("postal_code_%d", now.UnixMilli())
		case "City":
			request.BillingInfo.BillingAddress.City = fmt.Sprintf("city_%d", now.UnixMilli())
		case "Street1":
			request.BillingInfo.BillingAddress.Street1 = fmt.Sprintf("street1_%d", now.UnixMilli())
		case "PrefectureCode":
			err := InsertEntities(
				stepState,
				s.EntitiesCreator.CreatePrefecture(ctx, s.InvoiceMgmtPostgresDBTrace),
			)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			request.BillingInfo.BillingAddress.PrefectureCode = stepState.PrefectureCode
		case "PayerName":
			request.BillingInfo.PayerName = fmt.Sprintf("payer_name_%d", now.UnixMilli())
		case "PayerPhone":
			request.BillingInfo.PayerPhoneNumber = fmt.Sprintf("payer_phone_number_%d", now.UnixMilli())
		case "BankId":
			err := InsertEntities(
				stepState,
				s.EntitiesCreator.CreateBank(ctx, s.InvoiceMgmtPostgresDBTrace, false),
			)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			request.BankAccountInfo.BankId = stepState.BankID
		case "BankBranchId":
			err := InsertEntities(
				stepState,
				s.EntitiesCreator.CreateBankBranch(ctx, s.InvoiceMgmtPostgresDBTrace, false),
			)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			request.BankAccountInfo.BankBranchId = stepState.BankBranchID
		case "Verified":
			// update first the existing bank verified status initially to unverified
			ctx, err = s.updateWithNewBankAccountThatForExistingStudent(ctx, "school admin", "is valid and has unverified status")
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			request.BankAccountInfo.IsVerified = true
		case "NotVerified":
			// update first the existing bank verified status initially to verified
			ctx, err = s.updateWithNewBankAccountThatForExistingStudent(ctx, "school admin", "is valid and has verified status")
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			request.BankAccountInfo.IsVerified = false
		case "BankAccountType":
			randomInt, _ := rand.Int(rand.Reader, big.NewInt(2))
			request.BankAccountInfo.BankAccountType = invoice_pb.BankAccountType(randomInt.Int64())
		case "BankAccountHolder":
			request.BankAccountInfo.BankAccountHolder = fmt.Sprintf("%d-EXAMPLE BANK - ｱ BRANCH - ｢123｣ ()  ﾟ ﾞ . ﾆ", now.UnixMilli())
		case "Street2":
			request.BillingInfo.BillingAddress.Street2 = fmt.Sprintf("street2_%d", now.UnixMilli())
		}
	}

	stepState.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminUpdatesTheStudentPaymentInformation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = invoice_pb.NewEditPaymentDetailServiceClient(s.InvoiceMgmtConn).UpsertStudentPaymentInfo(contextWithToken(ctx), stepState.Request.(*invoice_pb.UpsertStudentPaymentInfoRequest))
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentPaymentInformationUpdatedSuccessfullyWithActionLogRecord(ctx context.Context, updateAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentPaymentDetail, err := (&repositories.StudentPaymentDetailRepo{}).FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "StudentPaymentDetailRepo.FindByID")
	}

	// retrieve student payment detail action log
	studentPaymentDetailActionLog := &entities.StudentPaymentDetailActionLog{}

	fields, _ := studentPaymentDetailActionLog.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_payment_detail_id = $1 AND action = $2 ORDER BY created_at DESC LIMIT 1", strings.Join(fields, ","), studentPaymentDetailActionLog.TableName())

	err = database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, studentPaymentDetail.StudentPaymentDetailID.String, updateAction).ScanOne(studentPaymentDetailActionLog)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error on selecting student payment detail action log: %w", err)
	}

	if strings.TrimSpace(studentPaymentDetailActionLog.UserID.String) == "" {
		return StepStateToContext(ctx, stepState), errors.New("error student payment detail action log empty user id")
	}

	if strings.TrimSpace(studentPaymentDetailActionLog.UserID.String) != stepState.CurrentUserID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected %v student payment detail action log user id but got: %v", stepState.CurrentUserID, strings.TrimSpace(studentPaymentDetailActionLog.UserID.String))
	}

	actionDetailJSONData := studentPaymentDetailActionLog.ActionDetail.Bytes

	var result StudentPaymentActionDetailLogType
	err = json.Unmarshal(actionDetailJSONData, &result)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error cannot convert student payment detail action log json to interface: %w", err)
	}

	// check action log values as checking of updated fields are already in other scenarios
	ctx, err = s.validateStudentPaymentDetailActionLogDetail(ctx, &result, updateAction)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) noStudentPaymentDetailActionLogRecorded(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentPaymentDetail, err := (&repositories.StudentPaymentDetailRepo{}).FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "StudentPaymentDetailRepo.FindByID")
	}
	// retrieve student payment detail action log
	studentPaymentDetailActionLog := &entities.StudentPaymentDetailActionLog{}

	fields, _ := studentPaymentDetailActionLog.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_payment_detail_id = $1", strings.Join(fields, ","), studentPaymentDetailActionLog.TableName())

	err = database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, studentPaymentDetail.StudentPaymentDetailID.String).ScanOne(studentPaymentDetailActionLog)
	// error should contains no rows in result set
	if err == nil && !errors.Is(err, pgx.ErrNoRows) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected no student payment detail action log on student: %v", stepState.StudentID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminUpdatesStudentPaymentInformationWithSameInformation(ctx context.Context, updateAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	request, err := s.retrieveRequestBasedOnUpdateAction(ctx, updateAction)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewEditPaymentDetailServiceClient(s.InvoiceMgmtConn).UpsertStudentPaymentInfo(contextWithToken(ctx), request)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrieveRequestBasedOnUpdateAction(ctx context.Context, updateAction string) (*invoice_pb.UpsertStudentPaymentInfoRequest, error) {
	stepState := StepStateFromContext(ctx)
	var (
		bankAccount    *entities.BankAccount
		billingAddress *entities.BillingAddress
		err            error
	)

	request, ok := stepState.Request.(*invoice_pb.UpsertStudentPaymentInfoRequest)
	if !ok {
		return nil, fmt.Errorf("the request should be type *invoice_pb.UpsertStudentPaymentInfoRequest got %T", request)
	}

	switch updateAction {
	case invoice_common.StudentPaymentDetailAction_UPDATED_BILLING_AND_BANK_DETAILS.String():
		billingAddress, err = (&repositories.BillingAddressRepo{}).FindByUserID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
		if err != nil {
			return nil, errors.Wrap(err, "BillingAddressRepo.FindByID")
		}

		bankAccount, err = (&repositories.BankAccountRepo{}).FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
		if err != nil {
			return nil, errors.Wrap(err, "BankAccountRepo.FindByStudentID")
		}

		request.BillingInfo.StudentPaymentDetailId = billingAddress.StudentPaymentDetailID.String
		request.BillingInfo.BillingAddress.BillingAddressId = billingAddress.BillingAddressID.String
		request.BankAccountInfo.BankAccountId = bankAccount.BankAccountID.String

	case invoice_common.StudentPaymentDetailAction_UPDATED_BANK_DETAILS.String(), invoice_common.StudentPaymentDetailAction_UPDATED_PAYMENT_METHOD.String():
		request.BillingInfo = nil
		bankAccount, err = (&repositories.BankAccountRepo{}).FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
		if err != nil {
			return nil, errors.Wrap(err, "BankAccountRepo.FindByStudentID")
		}
		request.BankAccountInfo.BankAccountId = bankAccount.BankAccountID.String
	case invoice_common.StudentPaymentDetailAction_UPDATED_BILLING_DETAILS.String():
		request.BankAccountInfo = nil

		billingAddress, err = (&repositories.BillingAddressRepo{}).FindByUserID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
		if err != nil {
			return nil, errors.Wrap(err, "BillingAddressRepo.FindByID")
		}

		request.BillingInfo.StudentPaymentDetailId = billingAddress.StudentPaymentDetailID.String
		request.BillingInfo.BillingAddress.BillingAddressId = billingAddress.BillingAddressID.String
	default:
		return nil, errors.New("invalid update info on student payment")
	}

	return request, nil
}

func (s *suite) thisStudentBillingAddressWasRemoved(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentID := stepState.StudentID
	billingAddress, err := (&repositories.BillingAddressRepo{}).FindByUserID(ctx, s.InvoiceMgmtPostgresDBTrace, studentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "BillingAddressRepo.FindByID")
	}

	err = multierr.Combine(
		billingAddress.PostalCode.Set(""),
		billingAddress.PrefectureCode.Set(""),
		billingAddress.City.Set(""),
		billingAddress.Street1.Set(""),
		billingAddress.Street2.Set(""),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error setting empty string for billing address: %v", err)
	}

	if err := (&repositories.BillingAddressRepo{}).Upsert(ctx, s.InvoiceMgmtPostgresDBTrace, billingAddress); err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "BillingAddressRepo.Upsert")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisStudentPaymentMethodWasRemoved(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentID := stepState.StudentID
	studentPaymentDetail, err := (&repositories.StudentPaymentDetailRepo{}).FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, studentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "StudentPaymentDetailRepo.FindByID")
	}

	err = multierr.Combine(
		studentPaymentDetail.PayerName.Set(""),
		studentPaymentDetail.PaymentMethod.Set(""),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error setting empty string for payment detail method and name: %v", err)
	}

	if err := (&repositories.StudentPaymentDetailRepo{}).Upsert(ctx, s.InvoiceMgmtPostgresDBTrace, studentPaymentDetail); err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "StudentPaymentDetailRepo.Upsert")
	}

	return StepStateToContext(ctx, stepState), nil
}
