package subscriptions

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/invoicemgmt/configurations"
	open_api_svc "github.com/manabie-com/backend/internal/invoicemgmt/services/open_api"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type UserEventSubscription struct {
	Config         *configurations.Config
	Logger         *zap.Logger
	JSM            nats.JetStreamManagement
	OpenAPIService *open_api_svc.OpenAPIModifierService
}

func (rcv *UserEventSubscription) Subscribe() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamUser, constants.DurablePaymentDetailsUserCreated),
			nats.DeliverSubject(constants.DeliverPaymentDetailsUserCreated),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
	}

	// subscribeCreatedStudent sub
	_, err := rcv.JSM.QueueSubscribe(constants.SubjectUserCreated, constants.QueuePaymentDetailsUserCreated, opts, rcv.HandlerNatsMessageCreateStudent)
	if err != nil {
		return fmt.Errorf("rcv.JSM.QueueSubscribe: %v", err)
	}

	updateStudentOpts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamUser, constants.DurablePaymentDetailsUserUpdated),
			nats.DeliverSubject(constants.DeliverPaymentDetailsUserUpdated),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
	}
	// subscribe update student event
	_, err = rcv.JSM.QueueSubscribe(constants.SubjectUserUpdated, constants.QueuePaymentDetailsUserUpdated, updateStudentOpts, rcv.HandlerNatsMessageUpdateStudent)
	if err != nil {
		return fmt.Errorf("rcv.JSM.QueueSubscribe: %v", err)
	}
	return nil
}

func (rcv *UserEventSubscription) HandlerNatsMessageCreateStudent(ctx context.Context, raw []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Second) // increase to 100s
	defer cancel()

	var err error
	enableAutoCS, err := rcv.OpenAPIService.CheckAutoSetConvenienceStoreIsEnabled(ctx)
	if err != nil {
		return false, status.Errorf(codes.Internal, err.Error())
	}

	if !enableAutoCS {
		return false, nil
	}

	req := &upb.EvtUser{}
	err = proto.Unmarshal(raw, req)
	if err != nil {
		rcv.Logger.Error(err.Error())
		return false, err
	}

	switch req.Message.(type) {
	case *upb.EvtUser_CreateStudent_:
		msg := req.GetCreateStudent()

		if msg.UserAddress == nil {
			return false, nil
		}

		billingAddressInfo := &open_api_svc.BillingAddressInfo{
			StudentID:    msg.StudentId,
			PayerName:    msg.StudentLastName + " " + msg.StudentFirstName,
			PostalCode:   msg.UserAddress.PostalCode,
			PrefectureID: msg.UserAddress.Prefecture,
			City:         msg.UserAddress.City,
			Street1:      msg.UserAddress.FirstStreet,
			Street2:      msg.UserAddress.SecondStreet,
		}

		err = rcv.OpenAPIService.AutoSetConvenienceStore(ctx, billingAddressInfo)
		if err != nil {
			return true, fmt.Errorf("err rcv.OpenAPIModifierService.AutoSetConvenienceStore: %w student_id: %s", err, msg.StudentId)
		}
	default:
		rcv.Logger.Info("User not a student")
	}

	return false, nil
}

func (rcv *UserEventSubscription) HandlerNatsMessageUpdateStudent(ctx context.Context, raw []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var err error
	enableAutoCS, err := rcv.OpenAPIService.CheckAutoSetConvenienceStoreIsEnabled(ctx)
	if err != nil {
		return false, status.Errorf(codes.Internal, err.Error())
	}

	if !enableAutoCS {
		return false, nil
	}

	req := &upb.EvtUser{}
	err = proto.Unmarshal(raw, req)
	if err != nil {
		rcv.Logger.Error(err.Error())
		return false, err
	}

	switch req.Message.(type) {
	case *upb.EvtUser_UpdateStudent_:
		msg := req.GetUpdateStudent()

		if msg.UserAddress == nil {
			msg.UserAddress = &upb.UserAddress{}
		}

		err = rcv.OpenAPIService.AutoUpdateBillingAddressInfoAndPaymentDetail(ctx, &open_api_svc.UpdateBillingAddressEventInfo{
			StudentID:   msg.StudentId,
			PayerName:   msg.StudentLastName + " " + msg.StudentFirstName,
			UserAddress: msg.UserAddress,
		})
		if err != nil {
			return true, fmt.Errorf("err rcv.OpenAPIModifierService.AutoUpdateBillingAddressInfoAndPaymentDetail: %w student_id: %s", err, msg.StudentId)
		}

	default:
		rcv.Logger.Info("User not a student")
	}

	return false, nil
}
