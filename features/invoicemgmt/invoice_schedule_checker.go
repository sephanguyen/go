package invoicemgmt

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	userConstant "github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) theOrganizationsHaveStudentWithBillItems(ctx context.Context, orgs string, studentCount int, billItemCount int, billStatus string) (context.Context, error) {

	stepState := StepStateFromContext(ctx)

	// A map of organization and the number of student with bill items
	orgList := strings.Split(orgs, ",")
	for _, org := range orgList {
		stepState.OrganizationStudentNumberMap[strings.TrimSpace(org)] = studentCount
	}

	originalLocationID := stepState.LocationID

	// Create the student and bill items for each organization
	for org, studentCount := range stepState.OrganizationStudentNumberMap {

		// get the location of org and assign to stepState.LocationID
		query := `
			SELECT grap.location_id
			FROM  granted_role gr
			INNER JOIN role r
				ON gr.role_id = r.role_id
			INNER JOIN granted_role_access_path grap
				ON gr.granted_role_id = grap.granted_role_id
			WHERE r.role_name = $1 AND r.resource_path = $2
			LIMIT 1
		`

		var locationID string
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, query, userConstant.RoleSchoolAdmin, org).Scan(&locationID)
		if err == nil {
			stepState.LocationID = locationID
		} else {
			stepState.LocationID = originalLocationID
		}

		// Reset the Student ID list in stepState to prevent other student from org to be mapped in different org
		stepState.StudentIds = []string{}

		// This setting of context is necessary to switch the context and resource path
		ctx = s.setResourcePathAndClaims(ctx, org)

		ctx, err = s.thereAreStudentsThatHasBillItemWithStatusAndType(ctx, studentCount, billItemCount, billStatus, payment_pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.thereAreStudentsThatHasBillItemWithStatus err %w", err)
		}
		stepState.OrganizationStudentListMap[org] = stepState.StudentIds
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsScheduledInvoicetoBeRunAtDayForTheseOrganizations(ctx context.Context, day int, orgs string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// If the feature toggle of KEC Feedback is off, set the invoiceDate to the given day and set the scheduledDate to invoiceDate + 1 day
	// If the feature toggle is off, set the scheduledDate to the given day and set the invoiceDate to scheduledDate - 1 day
	invoiceDate := time.Now().AddDate(0, 0, day)
	scheduledDate := invoiceDate.Add(24 * time.Hour)
	if isFeatureToggleEnabled(s.UnleashSuite.UnleashSrvAddr, s.UnleashSuite.UnleashLocalAdminAPIKey, constant.EnableKECFeedbackPh1) {
		scheduledDate = time.Now().AddDate(0, 0, day)
		invoiceDate = scheduledDate.Add(-24 * time.Hour)
	}

	stepState.CutoffDate = invoiceDate

	// Create the invoice schedule for each organization
	for _, org := range s.getOrgList(orgs) {
		// This setting of context is necessary to switch the context and resource path
		ctx = s.setResourcePathAndClaims(ctx, org)

		err := InsertEntities(
			StepStateFromContext(ctx),
			s.EntitiesCreator.CreateInvoiceSchedule(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceDate, scheduledDate, invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_SCHEDULED.String()),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Error on creating InvoiceSchedule: %w", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInvoiceScheduledCheckerEndpointWasCalledAtDay(ctx context.Context, day int) (context.Context, error) {
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	stepState := StepStateFromContext(ctx)

	invoiceDate := time.Now().AddDate(0, 0, day)

	req := &invoice_pb.InvoiceScheduleCheckerRequest{
		InvoiceDate: timestamppb.New(invoiceDate),
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewInternalServiceClient(s.InvoiceMgmtConn).InvoiceScheduleChecker(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInvoiceScheduledCheckerEndpointWasCalledAtDayConcurrently(ctx context.Context, day int) (context.Context, error) {
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	stepState := StepStateFromContext(ctx)

	invoiceDate := time.Now().AddDate(0, 0, day)

	req := &invoice_pb.InvoiceScheduleCheckerRequest{
		InvoiceDate: timestamppb.New(invoiceDate),
	}

	type invoiceScheduleRes struct {
		Response *invoice_pb.InvoiceScheduleCheckerResponse
		Err      error
	}

	stepState.ConcurrentCount = 3
	var wg sync.WaitGroup
	ch := make(chan invoiceScheduleRes, stepState.ConcurrentCount)

	for i := 0; i < stepState.ConcurrentCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := invoice_pb.NewInternalServiceClient(s.InvoiceMgmtConn).InvoiceScheduleChecker(contextWithToken(ctx), req)
			ch <- invoiceScheduleRes{
				Response: resp,
				Err:      err,
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	responseList := []*invoice_pb.InvoiceScheduleCheckerResponse{}
	errorList := []error{}
	for res := range ch {
		if res.Err != nil {
			errorList = append(errorList, res.Err)
			continue
		}
		responseList = append(responseList, res.Response)
	}

	stepState.Response = responseList
	stepState.ErrorList = errorList

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) onlyOneResponseHasOKStatusAndOthersHaveError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	responses, ok := stepState.Response.([]*invoice_pb.InvoiceScheduleCheckerResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting the type to be []*invoice_pb.InvoiceScheduleCheckerResponse got %T", responses)
	}

	countOfSuccessful := 0
	for _, r := range responses {
		if r != nil {
			countOfSuccessful++
		}
	}

	// check if the number of successful is only one
	if countOfSuccessful != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting number of successful request to be only one got %v", countOfSuccessful)
	}

	// check if all errors are only related to UNIQUE constraint error
	for _, err := range stepState.ErrorList {
		if !strings.Contains(err.Error(), "history already exists or another process is currently running") {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting unique constraint error got %v", err)
		}
	}

	// check if the length of error is equal to the number of request minus one (the successful one)
	if len(stepState.ErrorList) != stepState.ConcurrentCount-1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting the error count to be %v got %v", stepState.ConcurrentCount-1, stepState.ErrorList)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreCorrectNumberOfStudentInvoiceGeneratedInOrganization(ctx context.Context, orgs string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, org := range s.getOrgList(orgs) {

		// This setting of context is necessary to switch the context and resource path
		ctx = s.setResourcePathAndClaims(ctx, org)

		actualCount, err := s.countGeneratedInvoiceOfStudents(ctx, org)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.countGeneratedInvoiceOfStudents %v", err)
		}

		studentNumber := stepState.OrganizationStudentNumberMap[org]
		if actualCount != int64(studentNumber) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting number of generated invoice to be %d got %d in org %s. StudentIDs: %v", studentNumber, actualCount, org, stepState.OrganizationStudentListMap[org])
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theScheduledInvoiceStatusIsUpdatedTo(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for org := range stepState.OrganizationStudentNumberMap {

		// This setting of context is necessary to switch the context and resource path
		ctx = s.setResourcePathAndClaims(ctx, org)

		invoiceSchedule := &entities.InvoiceSchedule{}
		fields, _ := invoiceSchedule.FieldMap()
		query := fmt.Sprintf("SELECT %s FROM %s WHERE invoice_schedule_id = $1", strings.Join(fields, ","), invoiceSchedule.TableName())
		err := database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, stepState.InvoiceScheduleID).ScanOne(invoiceSchedule)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Error on selecting invoice schedule: %w", err)
		}

		if invoiceSchedule.Status.String != status {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting status to be %s got %s", status, invoiceSchedule.Status.String)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) onlyStudentBilledBillItemsAreInvoiced(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for org := range stepState.OrganizationStudentNumberMap {
		// This setting of context is necessary to switch the context and resource path
		ctx = s.setResourcePathAndClaims(ctx, org)
		var count int
		invoiceBillItem := &entities.InvoiceBillItem{}
		query := fmt.Sprintf("SELECT COUNT(invoice_bill_item_id) FROM %s WHERE past_billing_status != $1 AND resource_path = $2", invoiceBillItem.TableName())
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, query, payment_pb.BillingStatus_BILLING_STATUS_BILLED.String(), org).Scan(&count)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Error on selecting count invoice bill item: %w", err)
		}

		if count > 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting count of other bill items that are not BILLED in status to be 0 got %d", count)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aHistoryOfScheduledInvoiceWasSaved(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for org := range stepState.OrganizationStudentNumberMap {

		// This setting of context is necessary to switch the context and resource path
		ctx = s.setResourcePathAndClaims(ctx, org)

		invoiceScheduleHistory := &entities.InvoiceScheduleHistory{}
		fields, _ := invoiceScheduleHistory.FieldMap()
		query := fmt.Sprintf("SELECT %s FROM %s WHERE invoice_schedule_id = $1", strings.Join(fields, ","), invoiceScheduleHistory.TableName())
		err := database.Select(ctx, s.InvoiceMgmtPostgresDBTrace, query, stepState.InvoiceScheduleID).ScanOne(invoiceScheduleHistory)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Error on selecting invoice schedule: %w", err)
		}

		if invoiceScheduleHistory.InvoiceScheduleID.String != stepState.InvoiceScheduleID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("There is a problem with InvoiceScheduleHistory")
		}

		stepState.OrganizationInvoiceHistoryMap[org] = invoiceScheduleHistory.InvoiceScheduleHistoryID.String
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreNoInvoiceScheduledStudentWasSaved(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for org := range stepState.OrganizationStudentNumberMap {

		// This setting of context is necessary to switch the context and resource path
		ctx = s.setResourcePathAndClaims(ctx, org)

		stmt := "SELECT COUNT(*) as count FROM invoice_schedule_student WHERE invoice_schedule_history_id = $1"

		var actualCount int64
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, stepState.OrganizationInvoiceHistoryMap[org]).Scan(&actualCount)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error on counting the invoice %v", err)
		}

		if actualCount != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting number of student with scheduled invoice error to be 0 got %d", actualCount)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreInvoiceScheduledStudentSaved(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for org := range stepState.OrganizationStudentNumberMap {

		// This setting of context is necessary to switch the context and resource path
		ctx = s.setResourcePathAndClaims(ctx, org)

		studentIDs := stepState.OrganizationStudentListMap[org]

		stmt := "SELECT COUNT(*) as count FROM invoice_schedule_student WHERE student_id = ANY($1)"

		var actualCount int64
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, studentIDs).Scan(&actualCount)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error on counting the invoice %v", err)
		}

		studentLength := len(studentIDs)
		if actualCount != int64(studentLength) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting number of student with scheduled invoice error to be %d got %d", studentLength, actualCount)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) setResourcePathAndClaims(ctx context.Context, org string) context.Context {
	stepState := StepStateFromContext(ctx)

	claim := interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: org,
			DefaultRole:  bobEntities.UserGroupSchoolAdmin,
			UserGroup:    bobEntities.UserGroupSchoolAdmin,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, &claim)

	s.StepState.ResourcePath = org

	return StepStateToContext(ctx, stepState)
}

func (s *suite) thereIsNoScheduledInvoiceToBeRunAtDay(ctx context.Context, day int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	invoiceDate := time.Now().AddDate(0, 0, day)

	for org := range stepState.OrganizationStudentNumberMap {
		// This setting of context is necessary to switch the context and resource path
		ctx = s.setResourcePathAndClaims(ctx, org)

		scheduledInvoiceRepo := &repositories.InvoiceScheduleRepo{}
		scheduledInvoice, err := scheduledInvoiceRepo.GetByStatusAndInvoiceDate(
			ctx,
			s.InvoiceMgmtPostgresDBTrace,
			invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_SCHEDULED.String(),
			invoiceDate,
		)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("err scheduledInvoiceRepo.GetByStatusAndInvoiceDate: %w", err)
		}

		if scheduledInvoice != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting no scheduled invoice got 1")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreNoStudentsInvoiceGeneratedInOrganizations(ctx context.Context, orgs string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, org := range s.getOrgList(orgs) {

		// This setting of context is necessary to switch the context and resource path
		ctx = s.setResourcePathAndClaims(ctx, org)

		actualCount, err := s.countGeneratedInvoiceOfOrganization(ctx, org, invoice_pb.InvoiceType_SCHEDULED.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("s.countGeneratedInvoiceOfOrganization %v", err)
		}

		if actualCount != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting number of generated invoice to be 0 got %d in org %s", actualCount, org)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) countGeneratedInvoiceOfStudents(ctx context.Context, org string) (int64, error) {
	stepState := StepStateFromContext(ctx)

	stmt := "SELECT invoice_id FROM invoice WHERE student_id = ANY($1) AND status = $2"

	// This setting of context is necessary to switch the context and resource path
	ctx = s.setResourcePathAndClaims(ctx, org)

	var students pgtype.TextArray
	_ = students.Set(stepState.OrganizationStudentListMap[org])
	rows, err := s.InvoiceMgmtPostgresDBTrace.Query(ctx, stmt, students, invoice_pb.InvoiceStatus_DRAFT.String())
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	stepState.InvoiceIDs = []string{}
	for rows.Next() {
		var (
			invoiceID string
		)
		err := rows.Scan(&invoiceID)
		if err != nil {
			return 0, fmt.Errorf("row.Scan: %w", err)
		}

		stepState.InvoiceIDs = append(stepState.InvoiceIDs, invoiceID)
	}

	return int64(len(stepState.InvoiceIDs)), nil
}

func (s *suite) countGeneratedInvoiceOfOrganization(ctx context.Context, org string, invoiceType string) (int64, error) {
	stmt := "SELECT COUNT(*) as count FROM invoice WHERE status = $1 AND resource_path = $2 AND type = $3"

	// This setting of context is necessary to switch the context and resource path
	ctx = s.setResourcePathAndClaims(ctx, org)

	var actualCount int64
	err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, "DRAFT", org, invoiceType).Scan(&actualCount)
	if err != nil {
		return 0, fmt.Errorf("error on counting the invoice %v", err)
	}

	return actualCount, nil
}

func (s *suite) theOrganizationsHaveStudentWithBillItemsWithError(ctx context.Context, orgs string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// A map of organization and the number of student with bill items
	// Please do not use these organizations in other test. This step in inserting bill item directly to invoicemgmt. This may cause
	// conflict in syncing bill item
	orgList := strings.Split(orgs, ",")
	for _, org := range orgList {
		stepState.OrganizationStudentNumberMap[strings.TrimSpace(org)] = 2
	}

	// Create the student and bill items for each organization
	for org, studentCount := range stepState.OrganizationStudentNumberMap {

		// This setting of context is necessary to switch the context and resource path
		ctx = s.setResourcePathAndClaims(ctx, org)
		for i := 0; i < studentCount; i++ {

			studentID := idutil.ULIDNow()
			err := InsertEntities(
				StepStateFromContext(ctx),
				s.EntitiesCreator.CreateStudent(ctx, s.BobDBTrace, studentID),
				s.EntitiesCreator.WaitForKafkaSync(invoiceConst.KafkaSyncSleepDuration), // wait for kafka sync
				s.EntitiesCreator.CreateBillItemOnInvoicemgmt(ctx, s.InvoiceMgmtPostgresDBTrace, payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
			)

			stepState.OrganizationStudentListMap[org] = append(stepState.OrganizationStudentListMap[org], stepState.StudentID)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error on creating bill item on invoicemgmt DB %w", err)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aBillItemOfTheseOrganizationHasAdjustmentPrice(ctx context.Context, orgs string, adjustmentPrice int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Create the invoice schedule for each organization
	for _, org := range s.getOrgList(orgs) {
		// This setting of context is necessary to switch the context and resource path
		ctx = s.setResourcePathAndClaims(ctx, org)

		stmt := `
			UPDATE bill_item SET adjustment_price = $1, bill_type = $2 WHERE bill_item_sequence_number = (
				SELECT bill_item_sequence_number FROM bill_item 
				WHERE billing_status = 'BILLING_STATUS_BILLED'
				AND resource_path = $3
				LIMIT 1
			) AND resource_path = $3
		`
		_, err := s.FatimaDBTrace.Exec(ctx, stmt, adjustmentPrice, payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String(), s.ResourcePath)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error updating bill item: %v", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getOrgList(orgs string) []string {

	orgList := []string{}

	for _, org := range strings.Split(orgs, ",") {
		orgList = append(orgList, strings.TrimSpace(org))
	}

	return orgList
}

func (s *suite) allBillItemWithReviewRequiredTagWasSkipped(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	for org, studentIDs := range stepState.OrganizationStudentListMap {
		billItemSequenceNumbers := []int32{}
		for _, studentID := range studentIDs {
			billItemIDs := stepState.StudentBillItemMap[studentID]
			billItemSequenceNumbers = append(billItemSequenceNumbers, billItemIDs...)
		}

		billItems, err := s.findBillItemsByBillItemSequenceNumbers(ctx, billItemSequenceNumbers, org)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("err findBillItemsByBillItemSequenceNumbers: %v", err)
		}

		reviewRequiredTagBillItems := []int32{}
		for _, billItem := range billItems {
			if !billItem.IsReviewed.Bool {
				reviewRequiredTagBillItems = append(reviewRequiredTagBillItems, billItem.BillItemSequenceNumber.Int)
			}
		}

		var arr pgtype.Int4Array
		_ = arr.Set(reviewRequiredTagBillItems)

		stmt := "SELECT COUNT(*) FROM invoice_bill_item WHERE bill_item_sequence_number = ANY($1) AND resource_path = $2"
		var count int
		err = s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, arr, org).Scan(&count)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if count != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting bill item with review required tag is skipped but got %v bill_items: %v org: %v", count, reviewRequiredTagBillItems, org)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) allBillItemCreatedAfterCutoffDateWasSkipped(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	for org, studentIDs := range stepState.OrganizationStudentListMap {
		billItemSequenceNumbers := []int32{}
		for _, studentID := range studentIDs {
			billItemIDs := stepState.StudentBillItemMap[studentID]
			billItemSequenceNumbers = append(billItemSequenceNumbers, billItemIDs...)
		}

		billItems, err := s.findBillItemsByBillItemSequenceNumbers(ctx, billItemSequenceNumbers, org)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("err findBillItemsByBillItemSequenceNumbers: %v", err)
		}

		invalidBillItems := []int32{}
		for _, billItem := range billItems {
			if billItem.CreatedAt.Time == stepState.CutoffDate || billItem.CreatedAt.Time.After(stepState.CutoffDate) {
				invalidBillItems = append(invalidBillItems, billItem.BillItemSequenceNumber.Int)
			}
		}

		var arr pgtype.Int4Array
		_ = arr.Set(invalidBillItems)

		stmt := "SELECT COUNT(*) FROM invoice_bill_item WHERE bill_item_sequence_number = ANY($1) AND resource_path = $2"
		var count int
		err = s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, arr, org).Scan(&count)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if count != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting bill item created after cutoff date is skipped but got %v bill_items: %v org: %v", count, invalidBillItems, org)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) billItemCreatedAfterTheCutoffDate(ctx context.Context, quantity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for studentID, billItemIDs := range stepState.StudentBillItemMap {
		if len(billItemIDs) > 1 {
			// Update the created_at of bill item before the cutoff date
			if err := s.updateBillItemCreatedAt(ctx, studentID, billItemIDs, stepState.CutoffDate.Add(-24*time.Hour)); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("err updating bill item created_at: %v", err)
			}

			switch quantity {
			case "one":
				// Only the second bill item will be updated
				if err := s.updateBillItemCreatedAt(ctx, studentID, []int32{billItemIDs[1]}, stepState.CutoffDate.Add(24*time.Hour)); err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("err updating bill item created_at: %v", err)
				}
			case "all":
				if err := s.updateBillItemCreatedAt(ctx, studentID, billItemIDs, stepState.CutoffDate.Add(24*time.Hour)); err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("err updating bill item created_at: %v", err)
				}
			}
		}
	}

	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	return StepStateToContext(ctx, stepState), nil
}
