package discount

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/discount/mockdata"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	discountPb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"
)

func (s *suite) thereIsAnExistingStudentWithActiveProducts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentID, _, err := mockdata.InsertStudentWithActiveProducts(ctx, s.FatimaDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.StudentID = studentID

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisStudentHasUserDiscountTagRecords(ctx context.Context, existingStatus, discountTypes string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if existingStatus != "no-existing" {
		discountTypeStr := strings.Split(discountTypes, "/")

		err := s.createUserDiscountTagFromDiscountType(ctx, len(discountTypeStr), discountTypeStr)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aRequestPayloadForUpsertUserDiscountTag(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &discountPb.UpsertStudentDiscountTagRequest{
		StudentId: stepState.StudentID,
	}

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertsUserDiscountTagRecordsForThisStudent(ctx context.Context, discountTypes string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	discountTagIDs := make([]string, 0)
	if strings.TrimSpace(discountTypes) != "" {
		discountTypeStr := strings.Split(discountTypes, "/")

		for i := 0; i < len(discountTypeStr); i++ {
			discountType := s.getDiscountTypeMapping(discountTypeStr[i])
			discountTagID, ok := stepState.DiscountTagTypeAndIDMap[discountType]
			if !ok {
				return StepStateToContext(ctx, stepState), fmt.Errorf("there is no existing discount tag record with discount type: %v", discountType)
			}

			discountTagIDs = append(discountTagIDs, discountTagID)
		}
	}

	if len(discountTagIDs) > 0 {
		stepState.Request.(*discountPb.UpsertStudentDiscountTagRequest).DiscountTagIds = discountTagIDs
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) applyTheUpsertDiscountTagsOnTheStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = discountPb.NewDiscountServiceClient(s.DiscountConn).
		UpsertStudentDiscountTag(contextWithToken(ctx), stepState.Request.(*discountPb.UpsertStudentDiscountTagRequest))

	return StepStateToContext(ctx, stepState), nil
}

type UserDiscountTags []*entities.UserDiscountTag

func (u *UserDiscountTags) Add() database.Entity {
	e := &entities.UserDiscountTag{}
	*u = append(*u, e)

	return e
}

func (s *suite) thisStudentHasCorrectUserDiscountTagRecords(ctx context.Context, discountTypes string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	discountTypeStr := strings.Split(discountTypes, "/")

	fields, _ := (&entities.UserDiscountTag{}).FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM user_discount_tag WHERE user_id = $1 AND deleted_at IS NULL`, strings.Join(fields, ","))
	userDiscountTags := UserDiscountTags{}
	err := database.Select(ctx, s.FatimaDBTrace, query, stepState.StudentID).ScanAll(&userDiscountTags)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(userDiscountTags) != len(discountTypeStr) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected count user discount tag records to retrieved: %d but got: %d", len(discountTypeStr), len(userDiscountTags))
	}

	var foundUserDiscountType bool
	for _, userDiscountTag := range userDiscountTags {
		for i := 0; i < len(discountTypeStr); i++ {
			discountTypeFormatted := s.getDiscountTypeMapping(discountTypeStr[i])

			if userDiscountTag.DiscountType.String == discountTypeFormatted {
				foundUserDiscountType = true
			}
		}
		if !foundUserDiscountType {
			return StepStateToContext(ctx, stepState), fmt.Errorf("student: %v has unexpected discount type record: %v", stepState.StudentID, userDiscountTag.DiscountType.String)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisStudentHasNoUserDiscountTagRecords(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT
			count(*)
		FROM user_discount_tag
		WHERE user_id = $1 AND deleted_at IS NULL
		`

	var count int
	err := s.FatimaDBTrace.QueryRow(ctx, stmt, stepState.StudentID).Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting no user discount tag record for student: %v but got: %d count records", stepState.StudentID, count)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAnInvalidPayloadRequestForUpsertUserDiscountTag(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &discountPb.UpsertStudentDiscountTagRequest{
		StudentId: "",
	}

	if condition == "invalid data" {
		studentID, _, err := mockdata.InsertPreconditionData(ctx, s.FatimaDBTrace)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		req.StudentId = studentID
		req.DiscountTagIds = []string{"test-1", "test-invalid"}
	}

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}
