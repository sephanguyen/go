package invoicemgmt

import (
	"context"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *suite) createMigratedInvoiceOfBillItem(ctx context.Context, invoiceStatus, invoiceReference string) error {
	var err error
	ctx, err = s.createMigrationStudentWithBillItem(ctx, float64(10.00), invoiceReference)
	if err != nil {
		return err
	}

	time.Sleep(invoiceConst.KafkaSyncSleepDuration) // wait for kafka sync

	err = InsertEntities(
		StepStateFromContext(ctx),
		s.EntitiesCreator.CreateMigratedInvoice(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceStatus),
	)

	if err != nil {
		return err
	}

	return nil
}

func getHeaderTitles(entityNameStr string) ([]string, error) {
	var headerTitles []string
	switch entityNameStr {
	case invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY.String():
		headerTitles = []string{
			"payment_csv_id",
			"payment_id",
			"invoice_id",
			"payment_method",
			"payment_status",
			"due_date",
			"expiry_date",
			"payment_date",
			"student_id",
			"payment_sequence_number",
			"is_exported",
			"created_at",
			"result_code",
			"amount",
			"reference",
		}
	case invoice_pb.DataMigrationEntityName_INVOICE_ENTITY.String():
		headerTitles = []string{
			"invoice_csv_id",
			"invoice_id",
			"student_id",
			"type",
			"status",
			"sub_total",
			"total",
			"created_at",
			"invoice_sequence_number",
			"is_exported",
			"reference1",
			"reference2",
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "entity name not supported")
	}

	return headerTitles, nil
}

func (s *suite) createMigrationStudentWithBillItem(ctx context.Context, finalPrice float64, invoiceReference string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error
	// Create a student if there is no existing student ID in step state
	if stepState.StudentID == "" {
		ctx, err = s.createMigrationStudent(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		// wait for kafka sync of bob entities that are needed for inserting fatima entities
		time.Sleep(invoiceConst.KafkaSyncSleepDuration)
	}
	err = InsertEntities(
		StepStateFromContext(ctx),
		s.EntitiesCreator.CreateTax(ctx, s.FatimaDBTrace),
		s.EntitiesCreator.CreateBillingSchedule(ctx, s.FatimaDBTrace, true),
		s.EntitiesCreator.CreateBillingSchedulePeriod(ctx, s.FatimaDBTrace),
		s.EntitiesCreator.CreateProduct(ctx, s.FatimaDBTrace),
		s.EntitiesCreator.CreateOrder(ctx, s.FatimaDBTrace, payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(), false),
		s.EntitiesCreator.CreateStudentProduct(ctx, s.FatimaDBTrace),
		s.EntitiesCreator.CreateMigratedBillItem(ctx, s.FatimaDBTrace, finalPrice, invoiceReference),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createMigrationStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentID := idutil.ULIDNow()

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateMigrationStudent(ctx, s.BobDBTrace, studentID),
		s.EntitiesCreator.CreateUserAccessPathForStudent(ctx, s.BobDBTrace),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
