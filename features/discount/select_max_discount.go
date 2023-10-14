package discount

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/discount/mockdata"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	discountPb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"
	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) prepareDataForMaxDiscountSelection(ctx context.Context, discountTag string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	discountType := s.getDiscountTypeMapping(discountTag)
	discountID, discountTagID, err := mockdata.InsertOrgDiscount(ctx, s.FatimaDBTrace, discountType)

	stepState.DiscountTagID = discountTagID
	stepState.DiscountID = discountID

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	updateStudentStatusSubscriptionOptions := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(), nats.AckWait(30 * time.Second),
			nats.MaxDeliver(10),
			nats.Bind(constants.StreamUpdateStudentProduct, constants.DurableUpdateStudentProductCreated),
			nats.DeliverSubject(constants.DeliverUpdateStudentProductCreated),
		},
	}

	handlerUpdateStudentProductSubscription := func(ctx context.Context, data []byte) (bool, error) {
		updateStudentProductLog := &entities.UpdateProductDiscount{}
		err := json.Unmarshal(data, updateStudentProductLog)
		if err != nil {
			return false, err
		}

		stepState.FoundChanForJetStream <- updateStudentProductLog
		return false, nil
	}

	sub, err := s.JSM.QueueSubscribe(constants.SubjectUpdateStudentProductCreated, constants.QueueUpdateStudentProductCreated, updateStudentStatusSubscriptionOptions, handlerUpdateStudentProductSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe UpdateStudentProduct: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) tagIsAddedToStudent(ctx context.Context, userGroup string, discountTag string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	studentID, _, err := mockdata.InsertStudentWithActiveProducts(ctx, s.FatimaDBTrace)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.StudentID = studentID

	discountType := s.getDiscountTypeMapping(discountTag)
	discountTagID := stepState.DiscountTagID

	userDiscountTagEntity, err := generateUserDiscountTag(studentID, discountType, discountTagID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = mockdata.InsertUserDiscountTag(ctx, s.FatimaDBTrace, userDiscountTagEntity)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemSelectsMaxDiscount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()

	userInfo := golibs.UserInfoFromCtx(ctx)
	req := discountPb.AutoSelectHighestDiscountRequest{
		OrganizationId: userInfo.ResourcePath,
	}

	_, err := discountPb.NewInternalServiceClient(s.DiscountConn).
		AutoSelectHighestDiscount(contextWithToken(ctx), &req)

	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		_, err = discountPb.NewInternalServiceClient(s.PaymentConn).
			AutoSelectHighestDiscount(contextWithToken(ctx), &req)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eventIsReceivedForUpdateProduct(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	message := <-stepState.FoundChanForJetStream
	switch v := message.(type) {
	case *entities.UpdateProductDiscount:
		if v.StudentID == "" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to publish update product from discount automation")
		}
		return StepStateToContext(ctx, stepState), nil
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to publish update product from discount automation")
	}
}

func (s *suite) getDiscountTypeMapping(discountType string) string {
	switch discountType {
	case "single parent":
		return paymentPb.DiscountType_DISCOUNT_TYPE_SINGLE_PARENT.String()
	case "employee full time":
		return paymentPb.DiscountType_DISCOUNT_TYPE_EMPLOYEE_FULL_TIME.String()
	case "employee part time":
		return paymentPb.DiscountType_DISCOUNT_TYPE_EMPLOYEE_PART_TIME.String()
	case "family":
		return paymentPb.DiscountType_DISCOUNT_TYPE_FAMILY.String()
	case "combo":
		return paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String()
	case "sibling":
		return paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String()
	default:
		return paymentPb.DiscountType_DISCOUNT_TYPE_NONE.String()
	}
}
