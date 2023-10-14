package discount

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/discount/mockdata"
	discountPb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) thereIsAnExistingDiscountMasterDataWithDiscountTag(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for i := 0; i < len(DiscountTagTypeToMap); i++ {
		discountTagName := DiscountTagTypeToMap[i]

		_, discountTagID, err := mockdata.InsertOrgDiscount(ctx, s.FatimaDBTrace, discountTagName)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.DiscountTagTypeAndIDMap[discountTagName] = discountTagID
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAStudentThatHasUserDiscountTagRecords(ctx context.Context, recordCount int, discountTypes string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentID, _, err := mockdata.InsertPreconditionData(ctx, s.FatimaDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.StudentID = studentID
	discountTypeStr := strings.Split(discountTypes, "/")

	err = s.createUserDiscountTagFromDiscountType(ctx, recordCount, discountTypeStr)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisUserDiscountTagHasStartDateAndEndDate(ctx context.Context, startDateStr, endDateStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	startDate := getFormattedTimestampDate(startDateStr)
	endDate := getFormattedTimestampDate(endDateStr)
	stmt := `
		UPDATE user_discount_tag
		SET
			start_date = (CASE WHEN discount_type = 'DISCOUNT_TYPE_COMBO' OR discount_type = 'DISCOUNT_TYPE_SIBLING' THEN $1::DATE ELSE NULL END),
			created_at = (CASE WHEN discount_type != 'DISCOUNT_TYPE_COMBO' AND discount_type != 'DISCOUNT_TYPE_SIBLING' THEN $1::DATE ELSE NOW()::DATE END),
			end_date = (CASE WHEN discount_type = 'DISCOUNT_TYPE_COMBO' OR discount_type = 'DISCOUNT_TYPE_SIBLING' THEN $2::DATE ELSE NULL END)
		WHERE user_id = $3`

	_, err := s.FatimaDBTrace.Exec(ctx, stmt, startDate.AsTime(), endDate.AsTime(), stepState.StudentID)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidRequestPayloadForRetrieveUserDiscountTagWithDateToday(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &discountPb.RetrieveActiveStudentDiscountTagRequest{
		StudentId:           stepState.StudentID,
		DiscountDateRequest: timestamppb.Now(), // current day when test is run
	}

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAnInvalidPayloadRequestForRetrieveUserDiscountTag(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &discountPb.RetrieveActiveStudentDiscountTagRequest{
		StudentId:           stepState.StudentID,
		DiscountDateRequest: timestamppb.Now(), // current day when test is run
	}

	switch condition {
	case "empty student":
		req.DiscountDateRequest = timestamppb.Now()
	case "empty date request":
		req.StudentId = ""
	default:
		req = &discountPb.RetrieveActiveStudentDiscountTagRequest{}
	}

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrievesUserDiscountTagForThisStudent(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Response, stepState.ResponseErr = discountPb.NewDiscountServiceClient(s.DiscountConn).
		RetrieveActiveStudentDiscountTag(contextWithToken(ctx), stepState.Request.(*discountPb.RetrieveActiveStudentDiscountTagRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDiscountTagRecordsAreRetrievedSuccessfully(ctx context.Context, discountTypes string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	discountTypeStr := strings.Split(discountTypes, "/")
	response := s.StepState.Response.(*discountPb.RetrieveActiveStudentDiscountTagResponse)

	if response.StudentId != stepState.StudentID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected studentID: %v but got: %v", stepState.StudentID, response.StudentId)
	}

	if len(discountTypeStr) != len(response.DiscountTagDetails) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected discount type records length: %d but got: %d", len(discountTypeStr), len(response.DiscountTagDetails))
	}

	for _, discountDetail := range response.DiscountTagDetails {
		stmt := `
		SELECT
			discount_type
		FROM user_discount_tag
		WHERE user_id = $1 AND discount_tag_id = $2
		`

		var discountTypeFromUserTagDB string
		err := s.FatimaDBTrace.QueryRow(ctx, stmt, stepState.StudentID, discountDetail.DiscountTagId).Scan(&discountTypeFromUserTagDB)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		// check expected discount type
		var foundDiscountType bool
		for i := 0; i < len(discountTypeStr); i++ {
			discountTypeFormatted := s.getDiscountTypeMapping(discountTypeStr[i])

			if discountTypeFromUserTagDB == discountTypeFormatted {
				foundDiscountType = true
			}
		}
		if !foundDiscountType {
			return StepStateToContext(ctx, stepState), fmt.Errorf("record not match with expected discount type: %v for student: %v", discountTypeFromUserTagDB, stepState.StudentID)
		}
		// check discount tag table
		stmt2 := `
		SELECT
			count(*)
		FROM discount_tag
		WHERE discount_tag_id = $1 AND discount_tag_name = $2 and selectable = $3
		`

		var count int
		err = s.FatimaDBTrace.QueryRow(ctx, stmt2, discountDetail.DiscountTagId, discountDetail.DiscountTagName, discountDetail.Selectable).Scan(&count)

		if err != nil || count == 0 {
			return StepStateToContext(ctx, stepState), errors.New("no records found in discount tag table")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsNoUserDiscountTagRecordsRetrieved(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	discountTagDetails := s.StepState.Response.(*discountPb.RetrieveActiveStudentDiscountTagResponse).DiscountTagDetails
	if len(discountTagDetails) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no expected user discount tag records to retrieve but got count: %v", len(discountTagDetails))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsANonExistingStudentRecord(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.StudentID = "non-existing"

	return StepStateToContext(ctx, stepState), nil
}
