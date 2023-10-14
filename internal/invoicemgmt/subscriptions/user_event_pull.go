package subscriptions

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	open_api_svc "github.com/manabie-com/backend/internal/invoicemgmt/services/open_api"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	natsOrg "github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	batchSize = 10
	fetchSize = 500
)

func (rcv *UserEventSubscription) PullSubscribe() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamUser, constants.DurablePaymentDetailsUserCreated),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
			nats.DeliverNew(),
		},
		PullOpt: nats.PullSubscribeOption{
			FetchSize: fetchSize,
			BatchSize: batchSize,
		},
	}

	if err := rcv.JSM.PullSubscribe(constants.SubjectUserCreated, constants.DurablePaymentDetailsUserCreated, rcv.BatchCreateStudentPaymentDetail, opts); err != nil {
		return fmt.Errorf("rcv.JSM.PullSubscribe: %v", err)
	}

	updateOpts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamUser, constants.DurablePaymentDetailsUserUpdated),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
			nats.DeliverNew(),
		},
		PullOpt: nats.PullSubscribeOption{
			FetchSize: fetchSize,
			BatchSize: batchSize,
		},
	}

	if err := rcv.JSM.PullSubscribe(constants.SubjectUserUpdated, constants.DurablePaymentDetailsUserUpdated, rcv.BatchUpdateStudentPaymentDetail, updateOpts); err != nil {
		return fmt.Errorf("rcv.JSM.PullSubscribe: %v", err)
	}

	return nil
}

func (rcv *UserEventSubscription) BatchCreateStudentPaymentDetail(msgs []*natsOrg.Msg) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	for _, m := range msgs {
		var dataInMsg npb.DataInMessage
		var err error
		if err = proto.Unmarshal(m.Data, &dataInMsg); err != nil {
			if ackErr := m.Ack(); ackErr != nil {
				rcv.Logger.Error("msg.Ack", zap.Error(ackErr))
			}
			continue
		}

		ctx = golibs.ResourcePathToCtx(ctx, dataInMsg.ResourcePath)
		// Inject the user ID to context
		claims := interceptors.JWTClaimsFromContext(ctx)
		if claims != nil {
			claims.Manabie.UserID = dataInMsg.UserId
			ctx = interceptors.ContextWithJWTClaims(ctx, claims)
		}

		// Added the checking here since CheckAutoSetConvenienceStoreIsEnabled needs resource path in context
		enableAutoCS, err := rcv.OpenAPIService.CheckAutoSetConvenienceStoreIsEnabled(ctx)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}

		if !enableAutoCS {
			return nil
		}

		req := &upb.EvtUser{}
		if err = proto.Unmarshal(dataInMsg.Payload, req); err != nil {
			return err
		}

		switch req.Message.(type) {
		case *upb.EvtUser_CreateStudent_:
			msg := req.GetCreateStudent()

			if msg.UserAddress == nil {
				continue
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
				return fmt.Errorf("err rcv.OpenAPIModifierService.AutoSetConvenienceStore: %w student_id: %s", err, msg.StudentId)
			}
		default:
			rcv.Logger.Info("User not a student")
		}
	}

	return nil
}

func (rcv *UserEventSubscription) BatchUpdateStudentPaymentDetail(msgs []*natsOrg.Msg) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	for _, m := range msgs {
		var dataInMsg npb.DataInMessage
		var err error
		if err = proto.Unmarshal(m.Data, &dataInMsg); err != nil {
			if ackErr := m.Ack(); ackErr != nil {
				rcv.Logger.Error("msg.Ack", zap.Error(ackErr))
			}
			continue
		}

		ctx = golibs.ResourcePathToCtx(ctx, dataInMsg.ResourcePath)
		// Inject the user ID to context
		claims := interceptors.JWTClaimsFromContext(ctx)
		if claims != nil {
			claims.Manabie.UserID = dataInMsg.UserId
			ctx = interceptors.ContextWithJWTClaims(ctx, claims)
		}

		// Added the checking here since CheckAutoSetConvenienceStoreIsEnabled needs resource path in context
		enableAutoCS, err := rcv.OpenAPIService.CheckAutoSetConvenienceStoreIsEnabled(ctx)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}

		if !enableAutoCS {
			return nil
		}

		req := &upb.EvtUser{}
		if err = proto.Unmarshal(dataInMsg.Payload, req); err != nil {
			return err
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
				return fmt.Errorf("err rcv.OpenAPIModifierService.AutoUpdateBillingAddressInfoAndPaymentDetail: %w student_id: %s", err, msg.StudentId)
			}

		default:
			rcv.Logger.Info("User not a student")
		}
	}

	return nil
}
