package payment

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/pkg/errors"
)

func (s *suite) theInvalidBillingSchedulePeriodLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportBillingSchedulePeriodRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportBillingSchedulePeriodResponse)
	for _, row := range stepState.InvalidCsvRows {
		found := false
		for _, e := range resp.Errors {
			if strings.TrimSpace(reqSplit[e.RowNumber-1]) == row {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid line is not returned in response")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidBillingSchedulePeriodLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allBillingSchedulePeriods, err := s.selectAllBillingSchedulePeriods(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// we should use map for allBillingSchedulePeriods but it leads to some more code and not many items in
	// stepState.ValidCsvRows and allBillingSchedulePeriods, so we can do like below to make it simple
	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		name := rowSplit[1]
		billingScheduleID := rowSplit[2]
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		startDate, err := parseToDate(rowSplit[3])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		endDate, err := parseToDate(rowSplit[4])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		billingDate, err := parseToDate(rowSplit[5])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		remarks := rowSplit[6]
		isArchived, err := strconv.ParseBool(rowSplit[7])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		found := false
		for _, e := range allBillingSchedulePeriods {
			if e.Name.String == name && e.BillingScheduleID.String == billingScheduleID && e.StartDate.Time.Equal(startDate) && e.EndDate.Time.Equal(endDate) && e.BillingDate.Time.Equal(billingDate) && e.Remarks.String == remarks && e.IsArchived.Bool == isArchived {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingBillingSchedulePeriod(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	client := pb.NewImportMasterDataServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = client.ImportBillingSchedulePeriod(
		contextWithToken(ctx),
		stepState.Request.(*pb.ImportBillingSchedulePeriodRequest),
	)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anBillingSchedulePeriodValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomeBillingSchedulePeriods(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	validRow1 := fmt.Sprintf(
		",Cat %s,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks %s,0",
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)
	validRow2 := fmt.Sprintf(
		",Cat %s,1,2021-12-07,2021-12-08,2021-12-09,Remarks %s,1",
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(fmt.Sprintf(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,remarks,is_archived
			%s
			%s`, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anBillingSchedulePeriodInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{}
	case "header only":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,remarks,is_archived`),
		}
	case "number of column is not equal 8":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,remarks
			1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1
			2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2
			3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,remarks,is_archived
			1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1
			2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2
			3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3`),
		}
	case "wrong billing_schedule_period_id column name in header":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(`Number,name,billing_schedule_id,start_date,end_date,billing_date,remarks,is_archived
			1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
			2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
			3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
		}
	case "wrong name column name in header":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(`billing_schedule_period_id,Naming,billing_schedule_id,start_date,end_date,billing_date,remarks,is_archived
			1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
			2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
			3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
		}
	case "wrong billing_schedule_id column name in header":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(`billing_schedule_period_id,name,BillingScheduleID,start_date,end_date,billing_date,remarks,is_archived
			1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
			2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
			3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
		}
	case "wrong start_date column name in header":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,StartDate,end_date,billing_date,remarks,is_archived
			1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
			2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
			3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
		}
	case "wrong end_date column name in header":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,start_date,EndDate,billing_date,remarks,is_archived
			1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
			2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
			3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
		}
	case "wrong billing_date column name in header":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,BillingDate,remarks,is_archived
			1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
			2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
			3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
		}
	case "wrong remarks column name in header":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,Description,is_archived
			1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
			2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
			3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
		}
	case "wrong is_archived column name in header":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,remarks,IsArchived
			1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
			2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
			3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllBillingSchedulePeriods(ctx context.Context) ([]*entities.BillingSchedulePeriod, error) {
	allEntities := []*entities.BillingSchedulePeriod{}
	stmt :=
		`
		SELECT
			billing_schedule_period_id,
			name,
			billing_schedule_id,
			start_date,
			end_date,
			billing_date,
			remarks,
			is_archived
		FROM
			billing_schedule_period
		`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query billing_schedule_period")
	}

	defer rows.Close()
	for rows.Next() {
		e := &entities.BillingSchedulePeriod{}
		err := rows.Scan(
			&e.BillingSchedulePeriodID,
			&e.Name,
			&e.BillingScheduleID,
			&e.StartDate,
			&e.EndDate,
			&e.BillingDate,
			&e.Remarks,
			&e.IsArchived,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan billing schedule period")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *suite) insertSomeBillingSchedulePeriods(ctx context.Context) error {
	err := s.insertSomeBillingSchedules(ctx)
	if err != nil {
		return err
	}
	allBillingSchedules, err := s.selectAllBillingSchedules(ctx)
	if err != nil {
		return err
	}

	for i := 0; i < 3; i++ {
		uniqueID := idutil.ULIDNow()
		name := database.Text("Cat " + uniqueID)
		billingScheduleID := allBillingSchedules[i].BillingScheduleID
		remarks := database.Text("Remarks " + uniqueID)
		isArchived := database.Bool(rand.Int()%2 == 0)
		stmt := `INSERT INTO billing_schedule_period
		(billing_schedule_period_id, name, billing_schedule_id, start_date, end_date, billing_date, remarks, is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, now(), now(), now(), $4, $5, now(), now())`
		_, err := s.FatimaDBTrace.Exec(ctx, stmt, uniqueID, name, billingScheduleID, remarks, isArchived)
		if err != nil {
			return fmt.Errorf("cannot insert billing schedule period, err: %s", err)
		}
	}
	return nil
}

func (s *suite) anBillingSchedulePeriodValidRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomeBillingSchedulePeriods(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	existingBillingSchedulePeriods, err := s.selectAllBillingSchedulePeriods(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	validRow1 := fmt.Sprintf(
		",Cat %s,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks %s,0",
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)
	validRow2 := fmt.Sprintf(
		",Cat %s,1,2021-12-07,2021-12-08,2021-12-09,Remarks %s,1",
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)
	validRow3 := fmt.Sprintf(
		",Cat %s,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,,1",
		idutil.ULIDNow(),
	)
	validRow4 := fmt.Sprintf(
		"%s,Cat %s,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks %s,1",
		existingBillingSchedulePeriods[0].BillingSchedulePeriodID.String,
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)
	invalidEmptyRow1 := fmt.Sprintf(
		",Cat %s,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks %s,",
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)
	invalidEmptyRow2 := fmt.Sprintf(
		"%s,Cat %s,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks %s,",
		existingBillingSchedulePeriods[1].BillingSchedulePeriodID.String,
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)
	invalidValueRow1 := fmt.Sprintf(
		",Cat %s,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks %s,Archived",
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)
	invalidValueRow2 := fmt.Sprintf(
		"%s,Cat %s,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks %s,Archived",
		existingBillingSchedulePeriods[2].BillingSchedulePeriodID.String,
		idutil.ULIDNow(),
		idutil.ULIDNow(),
	)

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(fmt.Sprintf(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,remarks,is_archived
			%s
			%s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(fmt.Sprintf(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,remarks,is_archived
			%s
			%s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportBillingSchedulePeriodRequest{
			Payload: []byte(fmt.Sprintf(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,remarks,is_archived
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s`, validRow1, validRow2, validRow3, validRow4, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3, validRow4}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theImportBillingSchedulePeriodTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	allBillingSchedulePeriods, err := s.selectAllBillingSchedulePeriods(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	for _, row := range stepState.ValidCsvRows {
		found := false

		rowSplit := strings.Split(row, ",")
		name := rowSplit[1]
		billingScheduleID := rowSplit[2]
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		startDate, err := parseToDate(rowSplit[3])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		endDate, err := parseToDate(rowSplit[4])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		billingDate, err := parseToDate(rowSplit[5])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		remarks := rowSplit[6]
		isArchived, err := strconv.ParseBool(rowSplit[7])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		for _, e := range allBillingSchedulePeriods {
			if e.Name.String == name && e.BillingScheduleID.String == billingScheduleID && e.StartDate.Time.Equal(startDate) && e.EndDate.Time.Equal(endDate) && e.BillingDate.Time.Equal(billingDate) && e.Remarks.String == remarks && e.IsArchived.Bool == isArchived {
				found = true
				break
			}
		}
		if found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
