package invoicemgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func (s *suite) thereIsAnEventCreateStudentRequestWithUserAddressInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if strings.TrimSpace(stepState.StudentID) == "" {
		_, err := s.thereAreExistingStudents(ctx, 1)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	// create prefecture
	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreatePrefecture(ctx, s.BobDBTrace),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	randomIDString := idutil.ULIDNow()

	stepState.Request = &upb.EvtUser{
		Message: &upb.EvtUser_CreateStudent_{
			CreateStudent: &upb.EvtUser_CreateStudent{
				StudentId:        stepState.StudentID,
				StudentFirstName: stepState.CurrentStudentFirstName,
				StudentLastName:  stepState.CurrentStudentLastName,
				SchoolId:         stepState.SchoolID,
				UserAddress: &upb.UserAddress{
					PostalCode:   fmt.Sprintf("postal-code-%v", randomIDString),
					Prefecture:   stepState.PrefectureID,
					City:         fmt.Sprintf("city-%v", randomIDString),
					FirstStreet:  fmt.Sprintf("street-1-%v", randomIDString),
					SecondStreet: fmt.Sprintf("street-2-%v", randomIDString),
				},
			},
		},
	}
	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) yasuoSendTheCreateStudentEventRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	msg := stepState.Request.(*upb.EvtUser)
	data, err := proto.Marshal(msg)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	_, err = s.JSM.TracedPublish(contextWithToken(ctx), "nats.TracedPublish", constants.SubjectUserCreated, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// wait for processing
	time.Sleep(1 * time.Second)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentPaymentDetailRecordIsSuccessfullyCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentPaymentDetail := &entities.StudentPaymentDetail{}
	query := fmt.Sprintf("SELECT payment_method FROM %s WHERE student_id = $1 and resource_path = $2", studentPaymentDetail.TableName())
	// adding try do on selecting records as sometimes it becomes flaky and has error no rows in result set
	if err := try.Do(func(attempt int) (bool, error) {
		err := database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, stepState.StudentID, stepState.ResourcePath).ScanOne(studentPaymentDetail)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("error on selecting student payment detail record: %w for student: %v", err, stepState.StudentID)
		}

		if err == nil && studentPaymentDetail != nil {
			return false, nil
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("error on student payment detail created")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if studentPaymentDetail.PaymentMethod.String != invoice_pb.PaymentMethod_CONVENIENCE_STORE.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expecting payment method: %v but got: %v", invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(), studentPaymentDetail.PaymentMethod.String)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentBillingAddressRecordIsSuccessfullyCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	billingAddress := &entities.BillingAddress{}
	fields, _ := billingAddress.FieldMap()
	prefecture := &entities.Prefecture{}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1 and resource_path = $2", strings.Join(fields, ","), billingAddress.TableName())
	prefectureQuery := `SELECT prefecture_id FROM prefecture WHERE prefecture_code = $1`

	if err := try.Do(func(attempt int) (bool, error) {
		err := database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, stepState.StudentID, stepState.ResourcePath).ScanOne(billingAddress)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("error on selecting billing address record: %w for student: %v", err, stepState.StudentID)
		}
		// get prefecture id using prefecture code
		err = database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, prefectureQuery, billingAddress.PrefectureCode.String).ScanOne(prefecture)
		if err != nil {
			return false, fmt.Errorf("error on finding prefecture by id: %v", err)
		}

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("error on selecting prefecture record: %w for student: %v", err, stepState.StudentID)
		}

		if err == nil && billingAddress != nil && prefecture != nil {
			return false, nil
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("error on billing address created")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentBillingAddress = billingAddress
	stepState.CurrentPrefecture = prefecture

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) billingAddressIsTheSameAsUserAddress(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	userAddressRequestData := stepState.Request.(*upb.EvtUser).Message.(*upb.EvtUser_CreateStudent_).CreateStudent.UserAddress

	if stepState.CurrentBillingAddress == nil || stepState.CurrentPrefecture == nil {
		return StepStateToContext(ctx, stepState), errors.New("there is no current billing address or prefecture created")
	}

	err := validateBillingAddressInfoFromStudentEvent(userAddressRequestData, &BillingAddressInfo{
		PrefectureID: stepState.CurrentPrefecture.ID.String,
		City:         stepState.CurrentBillingAddress.City.String,
		Street1:      stepState.CurrentBillingAddress.Street1.String,
		Street2:      stepState.CurrentBillingAddress.Street2.String,
	})

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) noStudentPaymentDetailRecordCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	query := `SELECT count(*) FROM student_payment_detail WHERE student_id = $1 and resource_path = $2`

	if err := try.Do(func(attempt int) (bool, error) {
		var count int
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, query, stepState.StudentID, stepState.ResourcePath).Scan(&count)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("error on selecting student payment detail record: %w for student: %v", err, stepState.StudentID)
		}

		if err != nil && errors.Is(err, pgx.ErrNoRows) {
			// retry to fully check if there will be created student payment detail record on delay
			return true, err
		}

		if count >= 1 {
			return false, fmt.Errorf("error expected 0 student payment detail record but got: %v", count)
		}

		if count == 0 {
			return false, nil
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, err
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) noStudentBillingAddressRecordCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	query := `SELECT count(*) FROM billing_address WHERE user_id = $1 and resource_path = $2`

	if err := try.Do(func(attempt int) (bool, error) {
		var count int
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, query, stepState.StudentID, stepState.ResourcePath).Scan(&count)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("error on selecting billing address record: %w for student: %v", err, stepState.StudentID)
		}

		if err != nil && errors.Is(err, pgx.ErrNoRows) {
			// retry to fully check if there will be created billing address record on delay
			return true, err
		}

		if count >= 1 {
			return false, fmt.Errorf("error expected 0 billing address record but got: %v", count)
		}

		if count == 0 {
			return false, nil
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, err
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
