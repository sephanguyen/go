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

func (s *suite) theInvalidLeavingReasonLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportLeavingReasonRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportLeavingReasonResponse)
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

func convertToLeavingReasonTypeString(i int) string {
	leavingReasonType := []string{pb.LeavingReasonType_LEAVING_REASON_TYPE_WITHDRAWAL.String(), pb.LeavingReasonType_LEAVING_REASON_TYPE_GRADUATE.String(), pb.LeavingReasonType_LEAVING_REASON_TYPE_LOA.String()}
	return leavingReasonType[i-1]
}

func (s *suite) theValidLeavingReasonLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allLeavingReasons, err := s.selectAllLeavingReasons(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// we should use map for allAccoutingCategories but it leads to some more code and not many items in
	// stepState.ValidCsvRows and allAccoutingCategories, so we can do like below to make it simple
	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		name := rowSplit[1]
		leavingReasonType, err := strconv.Atoi(strings.TrimSpace(rowSplit[2]))

		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		remarks := rowSplit[3]
		isArchived, err := strconv.ParseBool(rowSplit[4])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		found := false
		for _, e := range allLeavingReasons {
			if e.Name.String == name && e.LeavingReasonType.Get() == convertToLeavingReasonTypeString(leavingReasonType) && e.Remark.String == remarks && e.IsArchived.Bool == isArchived {
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

func (s *suite) theImportLeavingReasonTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	allLeavingReasons, err := s.selectAllLeavingReasons(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		name := rowSplit[1]
		leavingReasonType, err := strconv.Atoi(strings.TrimSpace(rowSplit[2]))

		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		remarks := rowSplit[3]
		isArchived, err := strconv.ParseBool(rowSplit[4])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		found := false
		for _, e := range allLeavingReasons {
			if e.Name.String == name && e.LeavingReasonType.Get() == convertToLeavingReasonTypeString(leavingReasonType) && e.Remark.String == remarks && e.IsArchived.Bool == isArchived {
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

func (s *suite) importingLeavingReason(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.PaymentConn).
		ImportLeavingReason(contextWithToken(ctx), stepState.Request.(*pb.ImportLeavingReasonRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anLeavingReasonValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomeLeavingReasons(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	validRow1 := fmt.Sprintf(",Cat %s,1,Remarks %s,1", idutil.ULIDNow(), idutil.ULIDNow())
	validRow2 := fmt.Sprintf(",Cat %s,1,Remarks %s,1", idutil.ULIDNow(), idutil.ULIDNow())

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportLeavingReasonRequest{
			Payload: []byte(fmt.Sprintf(`leaving_reason_id,name,leaving_reason_type,remark,is_archived
			%s
			%s`, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anLeavingReasonValidRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomeLeavingReasons(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	existingLeavingReasons, err := s.selectAllLeavingReasons(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	validRow1 := fmt.Sprintf(",Cat %s,1,Remarks %s,1", idutil.ULIDNow(), idutil.ULIDNow())
	validRow2 := fmt.Sprintf(",Cat %s,1,Remarks %s,1", idutil.ULIDNow(), idutil.ULIDNow())
	validRow3 := fmt.Sprintf(",Cat %s,2,,1", idutil.ULIDNow())
	validRow4 := fmt.Sprintf("%s,Cat %s,1,Remarks %s,0", existingLeavingReasons[0].LeavingReasonID.String, idutil.ULIDNow(), idutil.ULIDNow())
	invalidEmptyRow1 := fmt.Sprintf(",Cat %s,1,Remarks %s,", idutil.ULIDNow(), idutil.ULIDNow())
	invalidEmptyRow2 := fmt.Sprintf("%s,Cat %s,2,Remarks %s,", existingLeavingReasons[1].LeavingReasonID.String, idutil.ULIDNow(), idutil.ULIDNow())
	invalidValueRow1 := fmt.Sprintf(",Cat %s,2,Remarks %s,Archived", idutil.ULIDNow(), idutil.ULIDNow())
	invalidValueRow2 := fmt.Sprintf("%s,Cat %s,1,Remarks %s,Archived", existingLeavingReasons[2].LeavingReasonID.String, idutil.ULIDNow(), idutil.ULIDNow())

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportLeavingReasonRequest{
			Payload: []byte(fmt.Sprintf(`leaving_reason_id,name,leaving_reason_type,remark,is_archived
			%s
			%s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportLeavingReasonRequest{
			Payload: []byte(fmt.Sprintf(`leaving_reason_id,name,leaving_reason_type,remark,is_archived
			%s
			%s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportLeavingReasonRequest{
			Payload: []byte(fmt.Sprintf(`leaving_reason_id,name,leaving_reason_type,remark,is_archived
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

func (s *suite) anLeavingReasonInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportLeavingReasonRequest{}
	case "header only":
		stepState.Request = &pb.ImportLeavingReasonRequest{
			Payload: []byte(`leaving_reason_id,name,leaving_reason_type,remark,is_archived`),
		}
	case "number of column is not equal 5":
		stepState.Request = &pb.ImportLeavingReasonRequest{
			Payload: []byte(`leaving_reason_id,name,leaving_reason_type,remark
			1,Cat 1,1,Remarks 1
			2,Cat 2,2,Remarks 2
			3,Cat 3,1,Remarks 3`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportLeavingReasonRequest{
			Payload: []byte(`leaving_reason_id,name,leaving_reason_type,remark,is_archived
			1,Cat 1,1,Remarks 1
			2,Cat 2,2,Remarks 2
			3,Cat 3,1,Remarks 3`),
		}
	case "wrong leaving_reason_id column name in header":
		stepState.Request = &pb.ImportLeavingReasonRequest{
			Payload: []byte(`Number,name,leaving_reason_type,remark,is_archived
			1,Cat 1,1,Remarks 1,0
			2,Cat 2,2,Remarks 2,0
			3,Cat 3,1,Remarks 3,0`),
		}
	case "wrong name column name in header":
		stepState.Request = &pb.ImportLeavingReasonRequest{
			Payload: []byte(`leaving_reason_id,Naming,leaving_reason_type,remark,is_archived
			1,Cat 1,1,Remarks 1,0
			2,Cat 2,2,Remarks 2,0
			3,Cat 3,1,Remarks 3,0`),
		}
	case "wrong leaving_reason_type column name in header":
		stepState.Request = &pb.ImportLeavingReasonRequest{
			Payload: []byte(`leaving_reason_id,Naming,Leaving_reason_type,remark,is_archived
			1,Cat 1,1,Remarks 1,0
			2,Cat 2,2,Remarks 2,0
			3,Cat 3,1,Remarks 3,0`),
		}
	case "wrong remarks column name in header":
		stepState.Request = &pb.ImportLeavingReasonRequest{
			Payload: []byte(`leaving_reason_id,name,leaving_reason_type,Description,is_archived
			1,Cat 1,1,Remarks 1,0
			2,Cat 2,2,Remarks 2,0
			3,Cat 3,1,Remarks 3,0`),
		}
	case "wrong is_archived column name in header":
		stepState.Request = &pb.ImportLeavingReasonRequest{
			Payload: []byte(`leaving_reason_id,name,leaving_reason_type,remark,IsArchived
			1,Cat 1,1,Remarks 1,0
			2,Cat 2,2,Remarks 2,0
			3,Cat 3,1,Remarks 3,0`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllLeavingReasons(ctx context.Context) ([]*entities.LeavingReason, error) {
	allEntities := []*entities.LeavingReason{}
	stmt :=
		`
		SELECT
			leaving_reason_id,
			name,
			leaving_reason_type,
		    remark,
			is_archived
		FROM
			leaving_reason
		`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query leaving_reason")
	}

	defer rows.Close()
	for rows.Next() {
		e := &entities.LeavingReason{}
		err := rows.Scan(
			&e.LeavingReasonID,
			&e.Name,
			&e.LeavingReasonType,
			&e.Remark,
			&e.IsArchived,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan leaving reason")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *suite) insertSomeLeavingReasons(ctx context.Context) error {
	for i := 0; i < 3; i++ {
		randomStr := idutil.ULIDNow()
		name := database.Text("Cat " + randomStr)
		leavingReasonType := database.Text("1")
		remarks := database.Text("Remark " + randomStr)
		isArchived := database.Bool(rand.Int()%2 == 0)
		stmt := `INSERT INTO leaving_reason
		(leaving_reason_id, name, leaving_reason_type,  remark, is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now())`

		_, err := s.FatimaDBTrace.Exec(ctx, stmt, randomStr, name, leavingReasonType, remarks, isArchived)
		if err != nil {
			return fmt.Errorf("cannot insert leaving_reason, err: %s", err)
		}
	}
	return nil
}
