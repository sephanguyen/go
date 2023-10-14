package discount

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/discount/mockdata"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	overWriteExisting                 = "overwrite existing"
	wrongIsArchivedColumnNameInHeader = "wrong is_archived column name in header"
)

// specific discount tags to test
var (
	DiscountTagTypeToMap = map[int]string{
		0: paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String(),
		1: paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String(),
		2: paymentPb.DiscountType_DISCOUNT_TYPE_EMPLOYEE_FULL_TIME.String(),
		3: paymentPb.DiscountType_DISCOUNT_TYPE_EMPLOYEE_PART_TIME.String(),
		4: paymentPb.DiscountType_DISCOUNT_TYPE_SINGLE_PARENT.String(),
		5: paymentPb.DiscountType_DISCOUNT_TYPE_FAMILY.String(),
	}
)

func (s *suite) createUserDiscountTagFromDiscountType(ctx context.Context, recordCount int, discountTypeStr []string) error {
	stepState := StepStateFromContext(ctx)

	for i := 0; i < recordCount; i++ {
		discountType := s.getDiscountTypeMapping(discountTypeStr[i])
		discountTagID, ok := stepState.DiscountTagTypeAndIDMap[discountType]
		if !ok {
			return fmt.Errorf("there is no existing discount tag record with discount type: %v", discountType)
		}
		userDiscountTagEntity, err := generateUserDiscountTag(stepState.StudentID, discountType, discountTagID)
		if err != nil {
			return err
		}

		if discountType == paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String() || discountType == paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String() {
			productGroupIDs, _, err := mockdata.InsertProductGroupMappingForSpecialDiscount(ctx, s.FatimaDBTrace, discountType)
			if err != nil {
				return err
			}
			// check and retrieve product id from product group table
			stmt := `SELECT product_id FROM product_group_mapping WHERE product_group_id = $1`
			var productID string
			row := s.FatimaDBTrace.QueryRow(ctx, stmt, productGroupIDs[0])
			err = row.Scan(&productID)
			if err != nil {
				return err
			}

			userDiscountTagEntity.ProductGroupID = database.Text(productGroupIDs[0])
			userDiscountTagEntity.ProductID = database.Text(productID)
		}

		err = mockdata.InsertUserDiscountTag(ctx, s.FatimaDBTrace, userDiscountTagEntity)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *suite) loginsToBackofficeApp(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, user)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func generateUserDiscountTag(studentID, discountType, discountTagID string) (*entities.UserDiscountTag, error) {
	userDiscountTagEntity := &entities.UserDiscountTag{}
	err := multierr.Combine(
		userDiscountTagEntity.UserID.Set(studentID),
		userDiscountTagEntity.DiscountType.Set(discountType),
		userDiscountTagEntity.DiscountTagID.Set(discountTagID),
	)
	if err != nil {
		return userDiscountTagEntity, err
	}

	return userDiscountTagEntity, nil
}

func getFormattedTimestampDate(dateString string) *timestamppb.Timestamp {
	var dateTimestamp *timestamppb.Timestamp
	switch dateString {
	case "TODAY":
		dateTimestamp = timestamppb.New(time.Now().Add(1 * time.Hour))
	case "TODAY+1":
		dateTimestamp = timestamppb.New(time.Now().AddDate(0, 0, 1))
	case "TODAY+2":
		dateTimestamp = timestamppb.New(time.Now().AddDate(0, 0, 2))
	case "TODAY+3":
		dateTimestamp = timestamppb.New(time.Now().AddDate(0, 0, 3))
	case "TODAY-1":
		dateTimestamp = timestamppb.New(time.Now().AddDate(0, 0, -1))
	case "TODAY-2":
		dateTimestamp = timestamppb.New(time.Now().AddDate(0, 0, -2))
	case "TODAY-3":
		dateTimestamp = timestamppb.New(time.Now().AddDate(0, 0, -3))
	}

	return dateTimestamp
}

func (s *suite) checkStudentDiscountTrackerByStudentID(ctx context.Context, userID string) (result bool) {
	var studentID string

	stmt := `SELECT student_id FROM student_discount_tracker WHERE student_id = $1 ORDER BY created_at DESC LIMIT 1`
	row := s.FatimaDBTrace.QueryRow(ctx, stmt, userID)
	err := row.Scan(&studentID)

	if err != nil || studentID == "" {
		return false
	}

	return true
}
