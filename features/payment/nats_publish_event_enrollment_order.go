package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) prepareDataValidOrderRequestForEnrollment(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                  true,
		InsertDiscount:             true,
		InsertStudent:              true,
		InsertEnrolledStudent:      false,
		InsertPotentialStudent:     true,
		InsertProductPrice:         true,
		InsertProductLocation:      true,
		InsertLocation:             false,
		InsertProductGrade:         true,
		InsertMaterial:             true,
		InsertBillingSchedule:      true,
		IsTaxExclusive:             false,
		InsertDiscountNotAvailable: false,
		InsertProductOutOfTime:     false,
		InsertProductDiscount:      true,
		BillingScheduleStartDate:   time.Now(),
	}

	defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
	req, err := s.validCaseBilledAtOrderItemsSingleItemExpectedForEnrollment(ctx, defaultOptionPrepareData)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eventPublishedSignalEnrollmentOrderSubmitted(ctx context.Context, result string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	timer := time.NewTimer(time.Minute * 5)
	defer timer.Stop()

	req := stepState.Request.(*pb.CreateOrderRequest)
	for {
		select {
		case data := <-stepState.FoundChanForJetStream:
			orderEventLog, ok := data.(*entities.OrderEventLog)
			if !ok {
				return StepStateToContext(ctx, stepState), fmt.Errorf("fail to parse data to *entities.OrderEventLog")
			}
			if orderEventLog.OrderType != pb.OrderType_ORDER_TYPE_ENROLLMENT.String() && orderEventLog.StudentID != req.StudentId {
				data, err := json.Marshal(orderEventLog)
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("fail to publish event to NATS")
				}
				_, err = connections.JSM.PublishAsyncContext(ctx, constants.SubjectOrderEventLogCreated, data)
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("fail to publish event to NATS")
				}
				continue
			}
			return StepStateToContext(ctx, stepState), nil
			// Check enrollment_status in student table after User squad update it to "Enrolled"
			// Retry because of delaying when bob sink data student to bob
			// Comment out because of waiting for implementing from Usermnt
			// switch result {
			// case "successfully":
			//	if err := try.Do(func(attempt int) (retry bool, err error) {
			//		student, err := s.getStudentByID(ctx, orderEventLog.StudentID)
			//		if err != nil {
			//			return attempt < 3, err
			//		}
			//		if student.EnrollmentStatus.Status == pgtype.Present &&
			//			student.EnrollmentStatus.String != upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String() {
			//			time.Sleep(1 * time.Second)
			//			return attempt < 3, fmt.Errorf(`validate student: expected "enrollment_status": %v but actual is %v`, upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(), student.EnrollmentStatus.String)
			//		}
			//		return false, nil
			//	}); err != nil {
			//		return StepStateToContext(ctx, stepState), fmt.Errorf(`Error when check student student info: %s`, err)
			//	}
			//}
		case <-ctx.Done():
			return StepStateToContext(ctx, stepState), fmt.Errorf("timeout waiting for event to be published")
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out cause of failing")
		}
	}
}

func (s *suite) subscribeCreatedOrderEventLog(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	orderEventLogSubscriptionOptions := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(), nats.AckWait(30 * time.Second),
			nats.MaxDeliver(10),
			nats.Bind(constants.StreamOrderEventLog, constants.DurableOrderEventLogCreated),
			nats.DeliverSubject(constants.DeliverOrderEventLogCreated),
		},
	}

	handlerCreatedOrderEventLogSubscription := func(ctx context.Context, data []byte) (bool, error) {
		orderEventLog := &entities.OrderEventLog{}
		err := json.Unmarshal(data, orderEventLog)
		if err != nil {
			return false, err
		}
		stepState.FoundChanForJetStream <- orderEventLog
		return false, nil
	}

	sub, err := s.JSM.QueueSubscribe(constants.SubjectOrderEventLogCreated, constants.QueueOrderEventLogCreated, orderEventLogSubscriptionOptions, handlerCreatedOrderEventLogSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe OrderEventLog: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getStudentByID(ctx context.Context, studentID string) (student entities.Student, err error) {
	studentFieldNames, studentFieldValues := student.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_id = $1
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentFieldNames, ","),
		student.TableName(),
	)
	row := s.FatimaDBTrace.QueryRow(ctx, stmt, studentID)
	err = row.Scan(studentFieldValues...)
	if err != nil {
		return
	}
	return
}
