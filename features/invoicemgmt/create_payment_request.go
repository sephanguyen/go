package invoicemgmt

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/common"
	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) partnerHasExistingConvenienceStore(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	partnerCSRepo := &repositories.PartnerConvenienceStoreRepo{}
	_, err := partnerCSRepo.FindOne(ctx, s.InvoiceMgmtPostgresDBTrace)
	if err == nil {
		return StepStateToContext(ctx, stepState), nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("partnerCSRepo.FindOne err: %v", err)
	}

	// Only create partner convenience store if there is no existing for partner
	err = InsertEntities(
		StepStateFromContext(ctx),
		s.EntitiesCreator.CreatePartnerConvenienceStore(ctx, s.InvoiceMgmtPostgresDBTrace),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminIsAtCreatePaymentRequestModal(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request := &invoice_pb.CreatePaymentRequestRequest{
		PaymentIds: stepState.PaymentIDs,
	}
	stepState.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminChoosesAsPaymentMethod(ctx context.Context, method string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var paymentMethod invoice_pb.PaymentMethod

	switch method {
	case "CONVENIENCE STORE":
		paymentMethod = invoice_pb.PaymentMethod_CONVENIENCE_STORE
	case "DIRECT DEBIT":
		paymentMethod = invoice_pb.PaymentMethod_DIRECT_DEBIT
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("payment method %s is not supported", method)
	}

	req := stepState.Request.(*invoice_pb.CreatePaymentRequestRequest)
	req.PaymentMethod = paymentMethod

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminAddsPaymentDueDateFromAndPaymentDueDateUntil(ctx context.Context, from, until int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	dueDateFrom := time.Now().AddDate(0, 0, from)
	dueDateUntil := time.Now().AddDate(0, 0, until)

	req := stepState.Request.(*invoice_pb.CreatePaymentRequestRequest)
	req.ConvenienceStoreDates = &invoice_pb.CreatePaymentRequestRequest_ConvenieceStoreDates{}

	req.ConvenienceStoreDates.DueDateFrom = timestamppb.New(dueDateFrom)
	req.ConvenienceStoreDates.DueDateUntil = timestamppb.New(dueDateUntil)

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

// nolint:unparam
func (s *suite) adminAddsPaymentDueDate(ctx context.Context, due int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	dueDate := time.Now().AddDate(0, 0, due)

	req := stepState.Request.(*invoice_pb.CreatePaymentRequestRequest)
	req.DirectDebitDates = &invoice_pb.CreatePaymentRequestRequest_DirectDebitDates{}

	req.DirectDebitDates.DueDate = timestamppb.New(dueDate)

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminClicksSaveCreatePaymentRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*invoice_pb.CreatePaymentRequestRequest)

	t, err := utils.GetTimeInLocation(time.Now(), utils.CountryJp)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = t
	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).CreatePaymentRequest(contextWithToken(ctx), req)
	if stepState.ResponseErr == nil {
		stepState.BulkPaymentRequestID = stepState.Response.(*invoice_pb.CreatePaymentRequestResponse).BulkPaymentRequestId
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thePaymentsAreAssociatedToAPaymentRequestFile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `
		SELECT 
			f.bulk_payment_request_file_id,
			COUNT(payment_id) AS payment_count
		FROM bulk_payment_request_file_payment fp
		INNER JOIN bulk_payment_request_file f
			ON f.bulk_payment_request_file_id = fp.bulk_payment_request_file_id
		WHERE f.bulk_payment_request_id = $1 AND f.resource_path = $2
		GROUP BY f.bulk_payment_request_file_id
	`
	rows, err := s.InvoiceMgmtPostgresDBTrace.Query(ctx, query, stepState.BulkPaymentRequestID, stepState.ResourcePath)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error on selecting payment files err: %v", err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("rows.Err() err: %v", err)
	}

	fileIDs := make([]string, 0)
	totalCount := 0

	for rows.Next() {
		var fileID string
		var count int
		if err := rows.Scan(&fileID, &count); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("row.Scan() err: %v", err)
		}

		totalCount += count
		fileIDs = append(fileIDs, fileID)
	}

	// Check if the count from the query is equal to the number of payments
	if totalCount != len(stepState.PaymentIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting number payment associated to a file to be %d got %d", len(stepState.PaymentIDs), totalCount)
	}

	// Can be use for other test step
	stepState.PaymentRequestFileIDs = fileIDs

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereArePaymentFileWithCorrectFileNameSavedOnDatabase(ctx context.Context, count int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(stepState.PaymentRequestFileIDs) != count {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting number payment file to be %d got %d", count, len(stepState.PaymentRequestFileIDs))
	}

	for _, fileID := range stepState.PaymentRequestFileIDs {
		// Fetch the payment file entity
		query := `
			SELECT
				bulk_payment_request_file_id,
				file_name,
				file_sequence_number,
				total_file_count
			FROM bulk_payment_request_file
			WHERE bulk_payment_request_file_id = $1
			AND resource_path = $2
		`
		e := &entities.BulkPaymentRequestFile{}
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, query, fileID, stepState.ResourcePath).Scan(&e.BulkPaymentRequestFileID, &e.FileName, &e.FileSequenceNumber, &e.TotalFileCount)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error on selecting payment file err: %v", err)
		}

		fileName, err := s.generateExpectedPaymentRequestFileName(ctx, e, s.InvoiceMgmtPostgresDBTrace)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if fileName != e.FileName.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting payment file name to be %v got %v", fileName, e.FileName.String)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thePaymentsAndInvoicesIsExportedFieldWasSetTo(ctx context.Context, isExportedStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	isExported := isExportedStr == "true"

	paymentQuery := "SELECT COUNT(*) FROM payment WHERE payment_id = ANY($1) AND is_exported = ($2)"
	invoiceQuery := "SELECT COUNT(*) FROM invoice WHERE invoice_id = ANY($1) AND is_exported = ($2)"

	var paymentCount int64
	err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, paymentQuery, stepState.PaymentIDs, isExported).Scan(&paymentCount)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error on counting the payment %v", err)
	}

	if int64(len(stepState.PaymentIDs)) != paymentCount {
		return StepStateToContext(ctx, stepState), fmt.Errorf("there are payments that were not exported")
	}

	var invoiceCount int64
	err = s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, invoiceQuery, stepState.InvoiceIDs, isExported).Scan(&invoiceCount)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error on counting the invoice %v", err)
	}

	if int64(len(stepState.InvoiceIDs)) != invoiceCount {
		return StepStateToContext(ctx, stepState), fmt.Errorf("there are invoices that were not exported")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aPaymentIsAlreadyExported(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(stepState.PaymentIDs) == 0 {
		return StepStateToContext(ctx, stepState), errors.New("no existing payment IDs")
	}

	paymentRepo := &repositories.PaymentRepo{}

	payment, err := paymentRepo.FindByPaymentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.PaymentIDs[0])
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("paymentRepo.FindByPaymentID err: %v", err)
	}

	payment.IsExported = database.Bool(true)

	err = paymentRepo.UpdateWithFields(ctx, s.InvoiceMgmtPostgresDBTrace, payment, []string{"is_exported"})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("paymentRepo.UpdateWithFields err: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentHasPaymentDetailAndBillingAddress(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	resChan := make(chan error, len(stepState.StudentIds))
	var wg sync.WaitGroup

	for _, id := range stepState.StudentIds {
		wg.Add(1)
		go func(studentID string) {
			// Create new instance of step state
			newStepState := &common.StepState{}
			newStepState.ResourcePath = stepState.ResourcePath
			newStepState.LocationID = stepState.LocationID
			newStepState.StudentID = studentID

			newCtx, cancel := context.WithCancel(ctx)

			// Inject the new step state to new context to prevent data race when used concurrently
			newCtx = StepStateToContext(newCtx, newStepState)

			defer cancel()
			defer wg.Done()

			err := InsertEntities(
				newStepState,
				s.EntitiesCreator.CreatePrefecture(ctx, s.BobDBTrace),
				s.EntitiesCreator.CreateStudentPaymentDetail(newCtx, s.InvoiceMgmtPostgresDBTrace, invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(), studentID),
				s.EntitiesCreator.CreateBillingAddress(newCtx, s.InvoiceMgmtPostgresDBTrace),
			)

			resChan <- err
		}(id)
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	for err := range resChan {
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAnExistingBankMappedToPartnerBank(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreatePartnerBank(ctx, s.InvoiceMgmtPostgresDBTrace, true),
		s.EntitiesCreator.CreateBank(ctx, s.InvoiceMgmtPostgresDBTrace, false),
		s.EntitiesCreator.CreateBankMapping(ctx, s.InvoiceMgmtPostgresDBTrace),
		s.EntitiesCreator.CreateBankBranch(ctx, s.InvoiceMgmtPostgresDBTrace, false),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsHasPaymentAndBankAccountDetail(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.createStudentBankAccount(ctx, stepState.StudentIds, stepState.BankBranchID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreBanksMappedToDifferentPartnerBank(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Each bank is mapped to one partner bank
	for i := 0; i < 2; i++ {
		ctx, err := s.thereIsAnExistingBankMappedToPartnerBank(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.BankBranchIDs = append(stepState.BankBranchIDs, stepState.BankBranchID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsHaveBankAccountInEitherOfTheBanks(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(stepState.StudentIds) < 2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("this step requires 2 or more students")
	}

	if len(stepState.BankBranchIDs) < 2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("this step requires 2 bank branch IDs")
	}

	// Divide the students. The first half will use the first bank, the second half will use the second bank
	middleIndex := len(stepState.StudentIds) / 2

	list1 := stepState.StudentIds[:middleIndex]
	list2 := stepState.StudentIds[middleIndex:]

	ctx, err := s.createStudentBankAccount(ctx, list1, stepState.BankBranchIDs[0])
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.createStudentBankAccount(ctx, list2, stepState.BankBranchIDs[1])
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsHaveNewCustomerCodeHistoryRecord(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ccRepo := &repositories.NewCustomerCodeHistoryRepo{}

	cc, err := ccRepo.FindByStudentIDs(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentIds)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("ccRepo.FindByStudentIDs err: %v", err)
	}

	if len(cc) != len(stepState.StudentIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting number of customer code history record to be %d got %d", len(stepState.StudentIds), len(cc))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateExpectedPaymentRequestFileName(ctx context.Context, e *entities.BulkPaymentRequestFile, db *database.DBTrace) (string, error) {
	stepState := StepStateFromContext(ctx)
	fileNameTimeFormat := "20060102"

	var (
		fileName      string
		fileExtension string
	)

	req := stepState.Request.(*invoice_pb.CreatePaymentRequestRequest)

	// Generate the expected file name based on due dates, sequence number and total file count
	switch req.PaymentMethod.String() {
	case invoice_pb.PaymentMethod_CONVENIENCE_STORE.String():
		dueDateFrom := req.ConvenienceStoreDates.DueDateFrom.AsTime().Format(fileNameTimeFormat)
		dueDateUntil := req.ConvenienceStoreDates.DueDateUntil.AsTime().Format(fileNameTimeFormat)

		fileName = fmt.Sprintf("Payment_CS_%sto%s", dueDateFrom, dueDateUntil)
		if isFeatureToggleEnabled(s.UnleashSuite.UnleashSrvAddr, s.UnleashSuite.UnleashLocalAdminAPIKey, constant.EnableKECFeedbackPh1) {
			fileName = fmt.Sprintf(
				"Payment_CS_created_date_%s",
				stepState.RequestSentAt.Format(fileNameTimeFormat),
			)
		}

		fileExtension = "csv"
	case invoice_pb.PaymentMethod_DIRECT_DEBIT.String():
		dueDate := req.DirectDebitDates.DueDate.AsTime().Format(fileNameTimeFormat)
		bankName := ""

		// Get one student ID from the payment
		var studentID string
		studentPaymentQuery := `
			SELECT 
				p.student_id
			FROM bulk_payment_request_file_payment fp
			INNER JOIN payment p
				ON p.payment_id = fp.payment_id
			WHERE bulk_payment_request_file_id = $1
			AND p.resource_path = $2
		`
		err := db.QueryRow(ctx, studentPaymentQuery, e.BulkPaymentRequestFileID.String, stepState.ResourcePath).Scan(&studentID)
		if err != nil {
			return "", fmt.Errorf("error on selecting payment file err: %v", err)
		}

		// Get the student bank details
		studentPaymentDetailRepo := &repositories.StudentPaymentDetailRepo{}
		studentBankAccountDetails, err := studentPaymentDetailRepo.FindStudentBankDetailsByStudentIDs(ctx, db, []string{studentID})
		if err != nil {
			return "", fmt.Errorf("error on selecting student bank detail err: %v", err)
		}

		if len(studentBankAccountDetails) == 0 {
			return "", fmt.Errorf("student %s has no bank details", studentID)
		}

		// Get the related partner bank
		bankAccount := studentBankAccountDetails[0].BankAccount
		bankBranchRepo := &repositories.BankBranchRepo{}
		relatedBankOfBankBranch, err := bankBranchRepo.FindRelatedBankOfBankBranches(ctx, db, []string{bankAccount.BankBranchID.String})
		if err != nil {
			return "", fmt.Errorf("bankBranchRepo.FindRelatedBankOfBankBranches err: %v", err)
		}

		// Loop to be the same in business logic
		for _, e := range relatedBankOfBankBranch {
			bankName = e.PartnerBank.BankName.String
		}

		fileName = fmt.Sprintf("Payment_DD_%s_%s", dueDate, bankName)
		fileExtension = "txt"
	}

	if e.TotalFileCount.Int > 1 {
		fileName = fmt.Sprintf("%s_%dof%d", fileName, e.FileSequenceNumber.Int, e.TotalFileCount.Int)
	}

	fileName = fmt.Sprintf("%s.%s", fileName, fileExtension)

	return fileName, nil
}

func (s *suite) thisPartnerBankRecordLimitIs(ctx context.Context, recordLimit int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := "UPDATE partner_bank SET record_limit = $1 WHERE partner_bank_id = $2 AND resource_path = $3"

	_, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt, recordLimit, s.PartnerBankID, s.ResourcePath)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error updating partner bank record limit: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}
