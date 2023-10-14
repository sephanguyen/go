package invoicesvc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type studentBillItemDetails struct {
	IDs      []int32
	Total    int32
	SubTotal float32
}

type orgError struct {
	Org string
	Err error
}

type studentError struct {
	studentID string
	err       string
}

func (s *InvoiceModifierService) InvoiceScheduleChecker(ctx context.Context, req *invoice_pb.InvoiceScheduleCheckerRequest) (*invoice_pb.InvoiceScheduleCheckerResponse, error) {
	if err := s.validateInvoiceCheckerRequest(req); err != nil {
		return nil, err
	}

	startTime := time.Now().UTC()
	orgs, err := s.OrganizationRepo.GetOrganizations(ctx, s.DB)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.OrganizationRepo.GetOrganizations error: %v", err))
	}

	errorMap := make(map[string]error)

	orgErrChan := make(chan orgError, len(orgs))

	var wg sync.WaitGroup

	for _, org := range orgs {
		wg.Add(1)

		go func(o *entities.Organization) {
			defer wg.Done()

			tenantContext := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: o.OrganizationID.String,
					UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
				},
			})

			err := s.checkScheduledInvoice(tenantContext, startTime, req.InvoiceDate.AsTime().UTC())
			orgErrChan <- orgError{
				Org: o.Name.String,
				Err: err,
			}
		}(org)
	}

	go func() {
		wg.Wait()
		close(orgErrChan)
	}()

	for item := range orgErrChan {
		if item.Err != nil {
			errorMap[item.Org] = item.Err
		}
	}

	if len(errorMap) != 0 {
		tenantErrs := genTenantErrorStr(orgs, errorMap)
		return nil, status.Error(codes.Internal, tenantErrs)
	}

	return &invoice_pb.InvoiceScheduleCheckerResponse{
		Successful: true,
	}, nil
}

func (s *InvoiceModifierService) checkScheduledInvoice(ctx context.Context, startTime, invoiceDate time.Time) error {
	// Get the scheduled invoice with SCHEDULED status on the given date
	useKECFeedbackPh1, err := s.UnleashClient.IsFeatureEnabled(constant.EnableKECFeedbackPh1, s.Env)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableKECFeedbackPh1, err))
	}

	var (
		scheduledInvoice *entities.InvoiceSchedule
	)

	if useKECFeedbackPh1 {
		scheduledInvoice, err = s.InvoiceScheduleRepo.GetByStatusAndScheduledDate(ctx, s.DB, invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_SCHEDULED.String(), invoiceDate)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("s.InvoiceScheduleRepo.GetByStatusAndScheduledDate error: %v", err)
		}
	} else {
		scheduledInvoice, err = s.InvoiceScheduleRepo.GetByStatusAndInvoiceDate(ctx, s.DB, invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_SCHEDULED.String(), invoiceDate)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("s.InvoiceScheduleRepo.GetByStatusAndInvoiceDate error: %v", err)
		}
	}

	// If there is no scheduled invoice, no need to process
	if scheduledInvoice == nil {
		resourcePath, _ := interceptors.ResourcePathFromContext(ctx)
		log.Printf("CheckScheduledInvoice on %v: there is no scheduled invoice on the given date %v on organization %s \n", time.Now().UTC(), invoiceDate, resourcePath)

		return nil
	}

	// Inject the user ID to context
	claims := interceptors.JWTClaimsFromContext(ctx)
	if claims != nil {
		claims.Manabie.UserID = scheduledInvoice.UserID.String
		ctx = interceptors.ContextWithJWTClaims(ctx, claims)
	}

	return s.processScheduledInvoice(ctx, scheduledInvoice, startTime)
}

func (s *InvoiceModifierService) processScheduledInvoice(ctx context.Context, scheduledInvoice *entities.InvoiceSchedule, startTime time.Time) error {
	// Create initial history
	scheduledHistoryID, err := s.saveInitialScheduleHistory(ctx, s.DB, scheduledInvoice.InvoiceScheduleID.String, time.Now().UTC())
	if err != nil {
		if strings.Contains(err.Error(), "\"invoice_schedule_history_invoice_schedule_id_key\" (SQLSTATE 23505)") {
			return fmt.Errorf("invoice schedule %s history already exists or another process is currently running", scheduledInvoice.InvoiceScheduleID.String)
		}
		return err
	}

	// Get bill items with BILLED and PENDING status
	billedBillItems, err := s.BillItemRepo.FindByStatuses(ctx, s.DB, []string{
		payment_pb.BillingStatus_BILLING_STATUS_BILLED.String(),
	})
	if err != nil {
		return fmt.Errorf("s.BillItemRepo.FindByStatuses error: %v", err)
	}

	enableReviewOrderChecking, err := s.UnleashClient.IsFeatureEnabled(constant.EnableReviewOrderChecking, s.Env)
	if err != nil {
		return fmt.Errorf("s.UnleashClient.IsFeatureEnabled err: %v", err)
	}

	billItemsAlreadyReviewed := make([]*entities.BillItem, 0)
	for _, billedBillItem := range billedBillItems {
		// Exclude bill items that were created after the cutoff date
		if billedBillItem.CreatedAt.Time == scheduledInvoice.InvoiceDate.Time || billedBillItem.CreatedAt.Time.After(scheduledInvoice.InvoiceDate.Time) {
			continue
		}

		// Exclude bill items that are not yet reviewed
		if !billedBillItem.IsReviewed.Bool && enableReviewOrderChecking {
			continue
		}

		billItemsAlreadyReviewed = append(billItemsAlreadyReviewed, billedBillItem)
	}

	log.Printf("number of bill items skipped/have review required tag: %v\n", len(billedBillItems)-len(billItemsAlreadyReviewed))

	failedInvoices := 0
	studentErrors := []*studentError{}

	// Map the bill items to the student
	studentBillItemMap, studentErrors := s.getStudentBillItemDetailsMap(billItemsAlreadyReviewed, studentErrors)
	failedInvoices += len(studentErrors)

	// Generate the invoice of students if there are students with billed item
	resp := &invoice_pb.GenerateInvoicesResponse{
		InvoicesData: []*invoice_pb.GenerateInvoicesResponse_InvoicesData{},
		Errors:       []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError{},
	}
	if len(studentBillItemMap) != 0 {
		createInvoiceReq := s.getGenerateInvoiceRequest(studentBillItemMap)
		log.Printf("number of students that have bill_items: %v number of invoice request %v", len(studentBillItemMap), len(createInvoiceReq.Invoices))

		resp, err = s.GenerateInvoices(ctx, createInvoiceReq)
		if err != nil {
			return fmt.Errorf("s.GenerateInvoices error: %v", err)
		}
	}

	enableRetryFailedInvoiceSchedule, err := s.UnleashClient.IsFeatureEnabled(constant.EnableRetryFailedInvoiceSchedule, s.Env)
	if err != nil {
		return fmt.Errorf("s.UnleashClient.IsFeatureEnabled err: %v", err)
	}

	if enableRetryFailedInvoiceSchedule {
		s.logger.Debug("Auto Retrying Failed Invoices")
		var newResErr []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError

		// get failed invoices before saving invoice history
		handledResErr, unhandledResErr := s.splitResponseErrors(resp)
		newResErr = append(newResErr, unhandledResErr...)
		loopCount := 1
		maxLoopCount := 50

		// Loop and reprocess only invoices with known errors due to invoice seq contention/unique constraints
		for len(handledResErr) > 0 && loopCount <= maxLoopCount {
			s.logger.Debug("LoopLogic: Loop count: ", loopCount)
			s.logger.Debug("LoopLogic: Length of handledResErr: ", len(handledResErr))

			invoiceDetails, err := s.genResErrInvoiceDetail(handledResErr)
			if err != nil {
				return fmt.Errorf("s.genResErrInvoiceDetail error: %v", err)
			}

			resp, err = s.GenerateInvoices(ctx, &invoice_pb.GenerateInvoicesRequest{
				Invoices: invoiceDetails,
			})
			if err != nil {
				return fmt.Errorf("s.GenerateInvoices error: %v", err)
			}

			handledResErr, unhandledResErr = s.splitResponseErrors(resp)
			newResErr = append(newResErr, unhandledResErr...)
			loopCount++
		}

		// Add separated res Err if over allowed
		if loopCount > maxLoopCount {
			newResErr = append(newResErr, handledResErr...)
		}

		// replace resp.Errors with the new compiled errors
		resp.Errors = newResErr

		s.logger.Debug("LoopLogic: Done Looping: Total Loop: ", loopCount)
	} else {
		log.Printf("LoopLogic: Skipped Retry logic")
	}

	// Add a retry for saving the history and updating the invoice schedule
	err = utils.DoWithMaxRetry(func(attempt int) (bool, error) {
		err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
			// Update the invoice schedule history total students, number of failed invoices and execution end date
			failedInvoices += len(resp.Errors)
			err = s.updateInvoiceScheduleHistory(ctx, tx, scheduledHistoryID, len(studentBillItemMap), failedInvoices)
			if err != nil {
				return err
			}

			status := invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_COMPLETED.String()
			if len(resp.Errors) != 0 {
				status = invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_INCOMPLETE.String()

				for _, e := range resp.Errors {
					studentErrors = append(studentErrors, &studentError{studentID: e.InvoiceDetail.StudentId, err: e.Error})
				}
			}

			if len(studentErrors) != 0 {
				// Save the detailed error of history if there are error in generating invoice
				err = s.saveStudentHistoryError(ctx, tx, scheduledHistoryID, studentErrors)
				if err != nil {
					return fmt.Errorf("s.saveStudentHistoryError error: %v", err)
				}
			}

			// Update the scheduled invoice status to COMPLETED
			scheduledInvoice.Status = database.Text(status)
			err = s.InvoiceScheduleRepo.Update(ctx, tx, scheduledInvoice)
			if err != nil {
				return fmt.Errorf("s.InvoiceScheduleRepo.Update error: %v", err)
			}

			return nil
		})

		if err == nil {
			return false, nil
		}

		log.Printf("Retrying saving history. Attempt: %d \n", attempt)
		return attempt < 10, err
	}, 10)

	if err != nil {
		return err
	}

	return nil
}

func (s *InvoiceModifierService) validateInvoiceCheckerRequest(req *invoice_pb.InvoiceScheduleCheckerRequest) error {
	if req.InvoiceDate == nil {
		return status.Error(codes.InvalidArgument, "invalid InvoiceDate value")
	}

	return nil
}

func (s *InvoiceModifierService) getGenerateInvoiceRequest(studentBillItemMap map[string]*studentBillItemDetails) *invoice_pb.GenerateInvoicesRequest {
	invoiceRequests := make([]*invoice_pb.GenerateInvoiceDetail, len(studentBillItemMap))

	index := 0
	for studentID, billItemDetails := range studentBillItemMap {
		// Make sure there is a bill item in each student
		if len(billItemDetails.IDs) == 0 {
			continue
		}

		invoiceRequests[index] = &invoice_pb.GenerateInvoiceDetail{
			StudentId:   studentID,
			BillItemIds: billItemDetails.IDs,
			Total:       billItemDetails.Total,
			SubTotal:    billItemDetails.SubTotal,
			InvoiceType: invoice_pb.InvoiceType_SCHEDULED,
		}

		index++
	}

	return &invoice_pb.GenerateInvoicesRequest{
		Invoices: invoiceRequests,
	}
}

func (s *InvoiceModifierService) getStudentBillItemDetailsMap(billedBillItems []*entities.BillItem, studentErrors []*studentError) (map[string]*studentBillItemDetails, []*studentError) {
	// Map the bill items to the student
	studentBillItemMap := make(map[string]*studentBillItemDetails)
	for _, billItem := range billedBillItems {
		exactPriceWithDecimalPlaces, err := getBillItemPrice(billItem)
		if err != nil {
			studentErrors = append(studentErrors, &studentError{
				studentID: billItem.StudentID.String,
				err:       err.Error(),
			})
			continue
		}

		if billItemDetails, ok := studentBillItemMap[billItem.StudentID.String]; ok {
			billItemDetails.IDs = append(billItemDetails.IDs, billItem.BillItemSequenceNumber.Int)
			billItemDetails.Total += int32(exactPriceWithDecimalPlaces)
			billItemDetails.SubTotal += float32(exactPriceWithDecimalPlaces)
		} else {
			studentBillItemMap[billItem.StudentID.String] = &studentBillItemDetails{
				IDs:      []int32{billItem.BillItemSequenceNumber.Int},
				Total:    int32(exactPriceWithDecimalPlaces),
				SubTotal: float32(exactPriceWithDecimalPlaces),
			}
		}
	}

	return studentBillItemMap, studentErrors
}

func (s *InvoiceModifierService) updateInvoiceScheduleHistory(
	ctx context.Context,
	db database.QueryExecer,
	scheduleInvoiceHistoryID string,
	studentWithBillItemsCount int,
	studentWithFailedInvoice int,
) error {
	invoiceScheduleHistory := &entities.InvoiceScheduleHistory{}
	database.AllNullEntity(invoiceScheduleHistory)

	err := multierr.Combine(
		invoiceScheduleHistory.InvoiceScheduleHistoryID.Set(scheduleInvoiceHistoryID),
		invoiceScheduleHistory.NumberOfFailedInvoices.Set(studentWithFailedInvoice),
		invoiceScheduleHistory.TotalStudents.Set(studentWithBillItemsCount),
		invoiceScheduleHistory.ExecutionEndDate.Set(time.Now().UTC()),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine error: %v", err)
	}

	err = s.InvoiceScheduleHistoryRepo.UpdateWithFields(ctx, db, invoiceScheduleHistory, []string{"number_of_failed_invoices", "total_students", "execution_end_date"})
	if err != nil {
		return fmt.Errorf("s.InvoiceScheduleHistoryRepo.UpdateWithFields error: %v", err)
	}

	return nil
}

func (s *InvoiceModifierService) saveInitialScheduleHistory(
	ctx context.Context,
	db database.QueryExecer,
	scheduleInvoiceID string,
	startTime time.Time,
) (string, error) {
	invoiceScheduleHistory := &entities.InvoiceScheduleHistory{}
	database.AllNullEntity(invoiceScheduleHistory)
	_ = multierr.Combine(
		invoiceScheduleHistory.InvoiceScheduleID.Set(scheduleInvoiceID),
		invoiceScheduleHistory.NumberOfFailedInvoices.Set(0),
		invoiceScheduleHistory.TotalStudents.Set(0),
		invoiceScheduleHistory.ExecutionStartDate.Set(startTime),
		invoiceScheduleHistory.ExecutionEndDate.Set(startTime),
	)
	historyID, err := s.InvoiceScheduleHistoryRepo.Create(ctx, db, invoiceScheduleHistory)
	if err != nil {
		return "", fmt.Errorf("s.InvoiceScheduleHistoryRepo.Create error: %v", err)
	}

	return historyID, nil
}

func (s *InvoiceModifierService) saveStudentHistoryError(
	ctx context.Context, db database.QueryExecer, historyID string, studentErrors []*studentError,
) error {
	studentWithErrors := make([]*entities.InvoiceScheduleStudent, len(studentErrors))
	for i, item := range studentErrors {
		e := &entities.InvoiceScheduleStudent{}
		database.AllNullEntity(e)

		err := multierr.Combine(
			e.InvoiceScheduleHistoryID.Set(historyID),
			e.StudentID.Set(item.studentID),
			e.ErrorDetails.Set("System Error"),
			e.ActualErrorDetails.Set(item.err),
		)
		if err != nil {
			return fmt.Errorf("multierr.Combine error: %v", err)
		}

		studentWithErrors[i] = e
	}

	err := s.InvoiceScheduleStudentRepo.CreateMultiple(ctx, db, studentWithErrors)
	if err != nil {
		return fmt.Errorf("s.InvoiceScheduleStudentRepo.CreateMultiple error: %v", err)
	}

	return nil
}

func genTenantErrorStr(orgs []*entities.Organization, errorMap map[string]error) string {
	// Re-arrange the sequence of errors so that it will be easy to test the error value
	errors := []string{}
	for _, org := range orgs {
		orgErr, ok := errorMap[org.Name.String]
		if ok {
			errors = append(errors, fmt.Sprintf("%s: %s", org.Name.String, orgErr.Error()))
		}
	}

	return fmt.Sprintf("Errors per tenant: %v", strings.Join(errors, ", "))
}

func (s *InvoiceModifierService) splitResponseErrors(resp *invoice_pb.GenerateInvoicesResponse) ([]*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError, []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError) {
	var handledResErr []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError
	var unhandledResErr []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError

	for _, errorDetail := range resp.Errors {
		if containsRetryError(errorDetail.Error) {
			handledResErr = append(handledResErr, errorDetail)
		} else {
			unhandledResErr = append(unhandledResErr, errorDetail)
		}
	}

	return handledResErr, unhandledResErr
}

func (s *InvoiceModifierService) genResErrInvoiceDetail(respErr []*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError) ([]*invoice_pb.GenerateInvoiceDetail, error) {
	if len(respErr) == 0 {
		return nil, fmt.Errorf("respErr can't be empty but got %v", respErr)
	}

	invoiceDetails := make([]*invoice_pb.GenerateInvoiceDetail, 0, len(respErr))
	for _, errorDetail := range respErr {
		invoiceDetails = append(invoiceDetails, errorDetail.InvoiceDetail)
	}

	return invoiceDetails, nil
}

func containsRetryError(message string) bool {
	errList := []string{"canceling statement due to statement timeout", "duplicate key value violates unique constraint"}
	for _, keyword := range errList {
		if strings.Contains(message, keyword) {
			return true
		}
	}
	return false
}
