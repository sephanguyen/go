package payment

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

func (s *suite) theInvalidDiscountLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportDiscountRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportDiscountResponse)
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

func (s *suite) theValidDiscountLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allDiscounts, err := s.selectAllDiscounts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// we should use map for allDiscounts but it leads to some more code and not many items in
	// stepState.ValidCsvRows and allDiscounts, so we can do like below to make it simple
	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		name := rowSplit[1]
		discountType, err := strconv.ParseInt(rowSplit[2], 10, 64)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		discountAmountType, err := strconv.ParseInt(rowSplit[3], 10, 64)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		discountAmountValue, err := strconv.ParseFloat(rowSplit[4], 64)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		discountAmountValueNumeric := pgtype.Numeric{}
		if err = discountAmountValueNumeric.Set(discountAmountValue); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		recurringValidDuration := 0
		if rowSplit[5] != "" {
			recurringValidDuration, err = strconv.Atoi(rowSplit[5])
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
		availableFrom, err := parseToDate(rowSplit[6])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		availableUntil, err := parseToDate(rowSplit[7])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		remarks := rowSplit[8]
		isArchived, err := strconv.ParseBool(rowSplit[9])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		found := false
		for _, e := range allDiscounts {
			if e.Name.String == name && e.DiscountType.String == pb.DiscountType_name[int32(discountType)] && e.DiscountAmountType.String == pb.DiscountAmountType_name[int32(discountAmountType)] &&
				isNumericEqual(e.DiscountAmountValue, discountAmountValueNumeric) && int(e.RecurringValidDuration.Int) == recurringValidDuration &&
				e.AvailableFrom.Time.Equal(availableFrom) && e.AvailableUntil.Time.Equal(availableUntil) && e.Remarks.String == remarks && e.IsArchived.Bool == isArchived {
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

func (s *suite) theImportDiscountTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	allDiscounts, err := s.selectAllDiscounts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		name := rowSplit[1]
		discountType, err := strconv.ParseInt(rowSplit[2], 10, 64)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		discountAmountType, err := strconv.ParseInt(rowSplit[3], 10, 64)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		discountAmountValue, err := strconv.ParseFloat(rowSplit[4], 64)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		discountAmountValueNumeric := pgtype.Numeric{}
		if err = discountAmountValueNumeric.Set(discountAmountValue); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		recurringValidDuration := 0
		if rowSplit[5] != "" {
			recurringValidDuration, err = strconv.Atoi(rowSplit[5])
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
		availableFrom, err := parseToDate(rowSplit[6])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		availableUntil, err := parseToDate(rowSplit[7])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		remarks := rowSplit[8]
		isArchived, err := strconv.ParseBool(rowSplit[9])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studentTagIDValidate := rowSplit[len(rowSplit)-2]
		parentTagIDValidate := rowSplit[len(rowSplit)-1]

		found := false
		for _, e := range allDiscounts {
			if e.Name.String == name && e.DiscountType.String == pb.DiscountType_name[int32(discountType)] && e.DiscountAmountType.String == pb.DiscountAmountType_name[int32(discountAmountType)] &&
				isNumericEqual(e.DiscountAmountValue, discountAmountValueNumeric) && int(e.RecurringValidDuration.Int) == recurringValidDuration &&
				e.AvailableFrom.Time.Equal(availableFrom) && e.AvailableUntil.Time.Equal(availableUntil) && e.Remarks.String == remarks && e.IsArchived.Bool == isArchived && e.StudentTagIDValidation.String == studentTagIDValidate && e.ParentTagIDValidation.String == parentTagIDValidate {
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

func isNumericEqual(n1, n2 pgtype.Numeric) bool {
	if n1.Status == n2.Status &&
		float64(n1.Int.Int64())*math.Pow10(int(n1.Exp)) == float64(n2.Int.Int64())*math.Pow10(int(n2.Exp)) {
		return true
	}
	return false
}

func (s *suite) importingDiscount(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.PaymentConn).
		ImportDiscount(contextWithToken(ctx), stepState.Request.(*pb.ImportDiscountRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anDiscountValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomeDiscounts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	validRow1 := fmt.Sprintf(",Discount %s,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,,,", idutil.ULIDNow())
	validRow2 := fmt.Sprintf(",Discount %s,2,1,12.25,2,2021-12-07,2022-10-07,Remarks,0,,,", idutil.ULIDNow())

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(fmt.Sprintf(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
			%s
			%s`, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anDiscountValidRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomeDiscounts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	existingDiscounts, err := s.selectAllDiscounts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	validRow1 := fmt.Sprintf(",Discount %s,1,2,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,%s,,", idutil.ULIDNow(), idutil.ULIDNow())
	validRow2 := fmt.Sprintf(",Discount %s,2,1,12.25,2,2021-12-07,2022-10-07,Remarks,0,,%s,", idutil.ULIDNow(), idutil.ULIDNow())
	validRow3 := fmt.Sprintf("%s,Discount %s,1,2,12000,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,%s,,", existingDiscounts[0].DiscountID.String, idutil.ULIDNow(), idutil.ULIDNow())
	validRow4 := fmt.Sprintf("%s,Discount %s,2,2,12000,2,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,1,,%s,", existingDiscounts[1].DiscountID.String, idutil.ULIDNow(), idutil.ULIDNow())
	invalidEmptyRow1 := fmt.Sprintf(",Discount %s,,3,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 12,0,,,", idutil.ULIDNow())
	invalidEmptyRow2 := fmt.Sprintf("%s,Discount %s,,3,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 12,1,,,", existingDiscounts[2].DiscountID.String, idutil.ULIDNow())
	invalidValueRow1 := fmt.Sprintf(",Discount %s,1,2,12000,2,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,2,,,", idutil.ULIDNow())
	invalidValueRow2 := fmt.Sprintf(",Discount %s,2,1,Two hundreds,2,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,,,", idutil.ULIDNow())
	invalidValueRow3 := fmt.Sprintf(",Discount %s,1,2,12000,,2021-12 23,2022-10-07T00:00:00-07:00,Remarks,0,,,", idutil.ULIDNow())
	invalidValueRow4 := fmt.Sprintf(",Discount %s,2,1,12.25,2,2021-12-07T00:00:00-07:00,2022-10--07,Remarks,0,,,", idutil.ULIDNow())
	invalidValueRow5 := fmt.Sprintf(",Discount %s,3,2,12000,2,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,,,", idutil.ULIDNow())
	invalidValueRow6 := fmt.Sprintf(",Discount %s,2,3,12000,2,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,,,", idutil.ULIDNow())
	invalidValueRow7 := fmt.Sprintf(",Discount %s,1,2,12000,NaN,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,,,", idutil.ULIDNow())

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(fmt.Sprintf(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
			%s
			%s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(fmt.Sprintf(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
			%s
			%s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(fmt.Sprintf(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s`, validRow1, validRow2, validRow3, validRow4, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2, invalidValueRow3, invalidValueRow4, invalidValueRow5, invalidValueRow6, invalidValueRow7)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3, validRow4}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2, invalidValueRow3, invalidValueRow4, invalidValueRow5, invalidValueRow6, invalidValueRow7}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anDiscountInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportDiscountRequest{}
	case "header only":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,student_tag_id_validation,parent_tag_id_validation,discount_tag_id`),
		}
	case "number of column is not equal 10":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				,Discount 1,1,4,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,,,`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				,Discount 1,1,5,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
		}
	case "wrong discount_id column name in header":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(`Number,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,6,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
		}
	case "wrong name column name in header":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(`discount_id,Naming,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,7,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
		}
	case "wrong discount_type column name in header":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(`discount_id,name,Discount Type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,8,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
		}
	case "wrong discount_amount_type column name in header":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(`discount_id,name,discount_type,Discount Amount Type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,9,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
		}
	case "wrong discount_amount_value column name in header":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(`discount_id,name,discount_type,discount_amount_type,Discount Amount Value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,10,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
		}
	case "wrong recurring_valid_duration column name in header":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,Recurring Valid Duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,11,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
		}
	case "wrong available_from column name in header":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,Available From,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,12,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
		}
	case "wrong available_until column name in header":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,Available Until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,13,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
		}
	case "wrong remarks column name in header":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,Description,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,14,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
		}
	case "wrong is_archived column name in header":
		stepState.Request = &pb.ImportDiscountRequest{
			Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,Is Archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,15,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllDiscounts(ctx context.Context) ([]*entities.Discount, error) {
	allEntities := []*entities.Discount{}
	stmt :=
		`
		SELECT 
			discount_id,
			name,
			discount_type,
			discount_amount_type,
			discount_amount_value,
			recurring_valid_duration,
			available_from,
			available_until,
		    remarks,
			is_archived,
			student_tag_id_validation,
			parent_tag_id_validation,
			discount_tag_id
		FROM
			discount
		`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query discount")
	}

	defer rows.Close()
	for rows.Next() {
		e := &entities.Discount{}
		err := rows.Scan(
			&e.DiscountID,
			&e.Name,
			&e.DiscountType,
			&e.DiscountAmountType,
			&e.DiscountAmountValue,
			&e.RecurringValidDuration,
			&e.AvailableFrom,
			&e.AvailableUntil,
			&e.Remarks,
			&e.IsArchived,
			&e.StudentTagIDValidation,
			&e.ParentTagIDValidation,
			&e.DiscountTagID,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan discount")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *suite) insertSomeDiscounts(ctx context.Context) error {
	for i := 0; i < 3; i++ {
		randomStr := idutil.ULIDNow()
		name := database.Text("Discount " + randomStr)

		rand.Seed(time.Now().UnixNano())
		discountType := pb.DiscountType_DISCOUNT_TYPE_COMBO.String()
		recurringValidDuration := sql.NullInt32{
			Int32: int32(rand.Intn(10) + 1),
			Valid: true,
		}

		if rand.Int()%2 == 0 {
			discountType = pb.DiscountType_DISCOUNT_TYPE_REGULAR.String()
			recurringValidDuration = sql.NullInt32{}
		}

		rand.Seed(time.Now().UnixNano())
		discountAmountType := pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String()
		discountAmountValue := 12.25
		if rand.Int()%2 == 0 {
			discountAmountType = pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT.String()
			discountAmountValue = 1225
		}

		availableFrom := time.Now()
		availableUntil := time.Now().AddDate(1, 0, 0)
		remarks := database.Text("Remarks " + randomStr)

		rand.Seed(time.Now().UnixNano())
		isArchived := database.Bool(rand.Int()%2 == 0)

		rand.Seed(time.Now().UnixNano())

		var (
			studentTagIDValidateGen string
			ParentTagIDValidateGen  string
		)
		if rand.Int()%2 == 0 {
			studentTagIDValidateGen = "s" + randomStr
			ParentTagIDValidateGen = ""
		} else {
			studentTagIDValidateGen = ""
			ParentTagIDValidateGen = "p" + randomStr
		}

		stmt := `INSERT INTO discount
		(discount_id, name, discount_type, discount_amount_type, discount_amount_value, recurring_valid_duration, available_from, available_until, remarks, is_archived, created_at, updated_at, student_tag_id_validation, parent_tag_id_validation, discount_tag_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now(), $11, $12, $13)`
		_, err := s.FatimaDBTrace.Exec(ctx, stmt, randomStr, name, discountType, discountAmountType, discountAmountValue, recurringValidDuration, availableFrom, availableUntil, remarks, isArchived, database.Text(studentTagIDValidateGen), database.Text(ParentTagIDValidateGen), pgtype.Text{Status: pgtype.Null})
		if err != nil {
			return fmt.Errorf("cannot insert discount, err: %s", err)
		}
	}
	return nil
}
