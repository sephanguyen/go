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

func (s *suite) theInvalidBillingScheduleLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportBillingScheduleRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportBillingScheduleResponse)
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

func (s *suite) theValidBillingScheduleLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allBillingSchedules, err := s.selectAllBillingSchedules(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// we should use map for allAccoutingCategories but it leads to some more code and not many items in
	// stepState.ValidCsvRows and allAccoutingCategories, so we can do like below to make it simple
	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		name := rowSplit[1]
		remarks := rowSplit[2]
		isArchived, err := strconv.ParseBool(rowSplit[3])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		found := false
		for _, e := range allBillingSchedules {
			if e.Name.String == name && e.Remarks.String == remarks && e.IsArchived.Bool == isArchived {
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

func (s *suite) anBillingScheduleValidRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomeBillingSchedules(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	existingBillingSchedules, err := s.selectAllBillingSchedules(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	validRow1 := fmt.Sprintf(",Cat %s,Remarks %s,1", idutil.ULIDNow(), idutil.ULIDNow())
	validRow2 := fmt.Sprintf(",Cat %s,Remarks %s,1", idutil.ULIDNow(), idutil.ULIDNow())
	validRow3 := fmt.Sprintf(",Cat %s,,1", idutil.ULIDNow())
	validRow4 := fmt.Sprintf("%s,Cat %s,Remarks %s,0", existingBillingSchedules[0].BillingScheduleID.String, idutil.ULIDNow(), idutil.ULIDNow())
	invalidEmptyRow1 := fmt.Sprintf(",Cat %s,Remarks %s,", idutil.ULIDNow(), idutil.ULIDNow())
	invalidEmptyRow2 := fmt.Sprintf("%s,Cat %s,Remarks %s,", existingBillingSchedules[1].BillingScheduleID.String, idutil.ULIDNow(), idutil.ULIDNow())
	invalidValueRow1 := fmt.Sprintf(",Cat %s,Remarks %s,Archived", idutil.ULIDNow(), idutil.ULIDNow())
	invalidValueRow2 := fmt.Sprintf("%s,Cat %s,Remarks %s,Archived", existingBillingSchedules[2].BillingScheduleID.String, idutil.ULIDNow(), idutil.ULIDNow())

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportBillingScheduleRequest{
			Payload: []byte(fmt.Sprintf(`billing_schedule_id,name,remarks,is_archived
			%s
			%s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportBillingScheduleRequest{
			Payload: []byte(fmt.Sprintf(`billing_schedule_id,name,remarks,is_archived
			%s
			%s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportBillingScheduleRequest{
			Payload: []byte(fmt.Sprintf(`billing_schedule_id,name,remarks,is_archived
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

func (s *suite) importingBillingSchedule(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.PaymentConn).
		ImportBillingSchedule(contextWithToken(ctx), stepState.Request.(*pb.ImportBillingScheduleRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anBillingScheduleValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomeBillingSchedules(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	validRow1 := fmt.Sprintf(",Cat %s,Remarks %s,1", idutil.ULIDNow(), idutil.ULIDNow())
	validRow2 := fmt.Sprintf(",Cat %s,Remarks %s,1", idutil.ULIDNow(), idutil.ULIDNow())
	stepState.ValidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportBillingScheduleRequest{
			Payload: []byte(fmt.Sprintf(`billing_schedule_id,name,remarks,is_archived
			%s
			%s`, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anBillingScheduleInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportBillingScheduleRequest{}
	case "header only":
		stepState.Request = &pb.ImportBillingScheduleRequest{
			Payload: []byte(`billing_schedule_id,name,remarks,is_archived`),
		}
	case "number of column is not equal 4":
		stepState.Request = &pb.ImportBillingScheduleRequest{
			Payload: []byte(`billing_schedule_id,name,remarks
			1,Cat 1,Remarks 1
			2,Cat 2,Remarks 2
			3,Cat 3,Remarks 3`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportBillingScheduleRequest{
			Payload: []byte(`billing_schedule_id,name,remarks,is_archived
			1,Cat 1,Remarks 1
			2,Cat 2,Remarks 2
			3,Cat 3,Remarks 3`),
		}
	case "wrong billing_schedule_id column name in header":
		stepState.Request = &pb.ImportBillingScheduleRequest{
			Payload: []byte(`Number,name,remarks,is_archived
			1,Cat 1,Remarks 1,0
			2,Cat 2,Remarks 2,0
			3,Cat 3,Remarks 3,0`),
		}
	case "wrong name column name in header":
		stepState.Request = &pb.ImportBillingScheduleRequest{
			Payload: []byte(`billing_schedule_id,Naming,remarks,is_archived
			1,Cat 1,Remarks 1,0
			2,Cat 2,Remarks 2,0
			3,Cat 3,Remarks 3,0`),
		}
	case "wrong remarks column name in header":
		stepState.Request = &pb.ImportBillingScheduleRequest{
			Payload: []byte(`billing_schedule_id,name,Description,is_archived
			1,Cat 1,Remarks 1,0
			2,Cat 2,Remarks 2,0
			3,Cat 3,Remarks 3,0`),
		}
	case "wrong is_archived column name in header":
		stepState.Request = &pb.ImportBillingScheduleRequest{
			Payload: []byte(`billing_schedule_id,name,remarks,IsArchived
			1,Cat 1,Remarks 1,0
			2,Cat 2,Remarks 2,0
			3,Cat 3,Remarks 3,0`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllBillingSchedules(ctx context.Context) ([]*entities.BillingSchedule, error) {
	allEntities := []*entities.BillingSchedule{}
	stmt :=
		`
		SELECT
			billing_schedule_id,
			name,
		    remarks,
			is_archived
		FROM
			billing_schedule
		`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query billing_schedule")
	}

	defer rows.Close()
	for rows.Next() {
		e := &entities.BillingSchedule{}
		err := rows.Scan(
			&e.BillingScheduleID,
			&e.Name,
			&e.Remarks,
			&e.IsArchived,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan billing schedule")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *suite) insertSomeBillingSchedules(ctx context.Context) error {
	for i := 0; i < 3; i++ {
		randomStr := idutil.ULIDNow()
		name := database.Text("Cat " + randomStr)
		remarks := database.Text("Remarks " + randomStr)
		isArchived := database.Bool(rand.Int()%2 == 0)
		stmt := `INSERT INTO billing_schedule
		(billing_schedule_id, name, remarks, is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, now(), now())`
		_, err := s.FatimaDBTrace.Exec(ctx, stmt, randomStr, name, remarks, isArchived)
		if err != nil {
			return fmt.Errorf("cannot insert billing schedule, err: %s", err)
		}
	}
	return nil
}

func (s *suite) theImportBillingScheduleTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	allBillingSchedules, err := s.selectAllBillingSchedules(ctx)
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
		remarks := rowSplit[2]
		isArchived, err := strconv.ParseBool(rowSplit[3])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		for _, e := range allBillingSchedules {
			if e.Name.String == name && e.Remarks.String == remarks && e.IsArchived.Bool == isArchived {
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
