package invoicemgmt

import (
	"context"
	"fmt"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/protobuf/proto"
)

func (s *suite) thereIsAnExistingStudentWithBillingAddressAndPaymentDetail(ctx context.Context, billingAddressExistence, paymentDetailExistence string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	_, err := s.createStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	if paymentDetailExistence == "existing" {
		err = InsertEntities(
			stepState,
			s.EntitiesCreator.CreateStudentPaymentDetail(ctx, s.InvoiceMgmtPostgresDBTrace, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(), stepState.StudentID),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	if billingAddressExistence == "existing" {
		err = InsertEntities(
			stepState,
			s.EntitiesCreator.CreatePrefecture(ctx, s.BobDBTrace),
			s.EntitiesCreator.CreateBillingAddress(ctx, s.InvoiceMgmtPostgresDBTrace),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) yasuoSendTheUpdateStudentEventWith(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	firstName := stepState.CurrentStudentFirstName
	lastName := stepState.CurrentStudentLastName
	userAddress := &upb.UserAddress{
		PostalCode:   fmt.Sprintf("updated-postal-code-%v", stepState.StudentPaymentDetailID),
		Prefecture:   stepState.PrefectureID,
		City:         fmt.Sprintf("updated-city-%v", stepState.StudentPaymentDetailID),
		FirstStreet:  fmt.Sprintf("updated-street-1-%v", stepState.StudentPaymentDetailID),
		SecondStreet: fmt.Sprintf("updated-street-2-%v", stepState.StudentPaymentDetailID),
	}

	switch condition {
	case "complete updated user address info and updated payer name":
		firstName = fmt.Sprintf("updated-%v", stepState.CurrentStudentFirstName)
		lastName = fmt.Sprintf("updated-%v", stepState.CurrentStudentLastName)

		// create new prefecture
		err := InsertEntities(
			stepState,
			s.EntitiesCreator.CreatePrefecture(ctx, s.BobDBTrace),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		userAddress.Prefecture = stepState.PrefectureID
	case "updated payer name":
		firstName = fmt.Sprintf("updated-%v", stepState.CurrentStudentFirstName)
		lastName = fmt.Sprintf("updated-%v", stepState.CurrentStudentLastName)
	case "missing user address":
		userAddress = nil
	case "one important billing address field is empty":
		userAddress.City = ""
	case "all important billing address field is empty":
		userAddress = &upb.UserAddress{
			PostalCode:   "",
			Prefecture:   "",
			City:         "",
			SecondStreet: "",
		}
	}

	msg := &upb.EvtUser{
		Message: &upb.EvtUser_UpdateStudent_{
			UpdateStudent: &upb.EvtUser_UpdateStudent{
				StudentId:        stepState.StudentID,
				StudentFirstName: firstName,
				StudentLastName:  lastName,
				UserAddress:      userAddress,
			},
		},
	}

	stepState.Request = msg

	data, err := proto.Marshal(msg)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	_, err = s.JSM.TracedPublish(contextWithToken(ctx), "nats.TracedPublish", constants.SubjectUserUpdated, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// Wait for processing
	time.Sleep(1 * time.Second)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentPayerNameSuccessfullyUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentPaymentDetailRepo := &repositories.StudentPaymentDetailRepo{}
	studentPaymentDetail, err := studentPaymentDetailRepo.FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("studentPaymentDetailRepo.FindByStudentID err: %v", err)
	}

	msg := stepState.Request.(*upb.EvtUser)
	var studentName string
	switch msg.Message.(type) {
	case *upb.EvtUser_CreateStudent_:
		msgCreateEvt := msg.GetCreateStudent()
		studentName = fmt.Sprintf("%v %v", msgCreateEvt.StudentLastName, msgCreateEvt.StudentFirstName)
	case *upb.EvtUser_UpdateStudent_:
		msgUpdateEvt := msg.GetUpdateStudent()
		studentName = fmt.Sprintf("%v %v", msgUpdateEvt.StudentLastName, msgUpdateEvt.StudentFirstName)
	}

	if studentPaymentDetail.PayerName.String != studentName {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student_id: %v expecting payer name to be %v got %v", stepState.StudentID, studentName, studentPaymentDetail.PayerName.String)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentBillingAddressRecordSuccessfullyUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	billingAddressRepo := &repositories.BillingAddressRepo{}
	billingAddress, err := billingAddressRepo.FindByUserID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("billingAddressRepo.FindByUserID err: %v", err)
	}

	msg := stepState.Request.(*upb.EvtUser)
	updateMsg := msg.GetUpdateStudent()

	if billingAddress.City.String != updateMsg.UserAddress.City {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student_id: %v expecting city to be %v got %v", stepState.StudentID, updateMsg.UserAddress.City, billingAddress.City.String)
	}

	if billingAddress.Street1.String != updateMsg.UserAddress.FirstStreet {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student_id: %v expecting street1 to be %v got %v", stepState.StudentID, updateMsg.UserAddress.FirstStreet, billingAddress.Street1.String)
	}

	if billingAddress.Street2.String != updateMsg.UserAddress.SecondStreet {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student_id: %v expecting street2 to be %v got %v", stepState.StudentID, updateMsg.UserAddress.SecondStreet, billingAddress.Street2.String)
	}

	if billingAddress.PostalCode.String != updateMsg.UserAddress.PostalCode {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student_id: %v expecting postal code to be %v got %v", stepState.StudentID, updateMsg.UserAddress.PostalCode, billingAddress.PostalCode.String)
	}

	prefectureRepo := &repositories.PrefectureRepo{}
	prefecture, err := prefectureRepo.FindByPrefectureID(ctx, s.InvoiceMgmtPostgresDBTrace, updateMsg.UserAddress.Prefecture)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("prefectureRepo.FindByPrefectureID err: %v", err)
	}

	if billingAddress.PrefectureCode.String != prefecture.PrefectureCode.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student_id: %v expecting prefecture code to be %v got %v", stepState.StudentID, prefecture.PrefectureCode.String, billingAddress.PrefectureCode.String)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentPayerNameIsNotUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentPaymentDetailRepo := &repositories.StudentPaymentDetailRepo{}
	studentPaymentDetail, err := studentPaymentDetailRepo.FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("studentPaymentDetailRepo.FindByStudentID err: %v", err)
	}

	if studentPaymentDetail.PayerName.String != stepState.CurrentPayerName {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student_id: %v expecting payer name to be %v got %v", stepState.StudentID, stepState.CurrentPayerName, studentPaymentDetail.PayerName.String)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentPaymentMethodIncludingPayerNameSuccessfullyRemoved(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentPaymentDetailRepo := &repositories.StudentPaymentDetailRepo{}
	studentPaymentDetail, err := studentPaymentDetailRepo.FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("studentPaymentDetailRepo.FindByStudentID err: %v", err)
	}

	if studentPaymentDetail.PayerName.String != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student_id: %v expecting payer name to be removed got %v", stepState.StudentID, studentPaymentDetail.PaymentMethod.String)
	}

	if studentPaymentDetail.PaymentMethod.String != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student_id: %v expecting payment method to be removed got %v", stepState.StudentID, studentPaymentDetail.PaymentMethod.String)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentBillingAddressRecordSuccessfullyRemoved(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	billingAddressRepo := &repositories.BillingAddressRepo{}
	billingAddress, err := billingAddressRepo.FindByUserID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("billingAddressRepo.FindByUserID err: %v", err)
	}

	if billingAddress.City.String != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student_id: %v expecting city to be removed got %v", stepState.StudentID, billingAddress.City.String)
	}

	if billingAddress.Street1.String != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student_id: %v expecting street1 to be removed got %v", stepState.StudentID, billingAddress.Street1.String)
	}

	if billingAddress.Street2.String != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student_id: %v expecting street2 to be removed got %v", stepState.StudentID, billingAddress.Street2.String)
	}

	if billingAddress.PostalCode.String != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student_id: %v expecting postal code to be removed got %v", stepState.StudentID, billingAddress.PostalCode.String)
	}

	if billingAddress.PrefectureCode.String != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student_id: %v expecting prefecture code to be removed got %v", stepState.StudentID, billingAddress.PrefectureCode.String)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theDefaultPaymentMethodOfThisStudentIs(ctx context.Context, defaultPaymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if defaultPaymentMethod == "EMPTY" {
		defaultPaymentMethod = ""
	}

	stmt := `UPDATE student_payment_detail SET payment_method = $1 WHERE resource_path = $2 AND student_id = $3`
	if _, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt, defaultPaymentMethod, s.ResourcePath, stepState.StudentID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisStudentHasBankAccountWithVerificationStatus(ctx context.Context, verificationStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	isVerified := verificationStatus == "verified"

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateBankAccount(ctx, s.InvoiceMgmtPostgresDBTrace, isVerified),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
