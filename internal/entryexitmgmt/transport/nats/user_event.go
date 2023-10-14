package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/configurations"
	"github.com/manabie-com/backend/internal/entryexitmgmt/services"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	natsOrg "github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type UserEventSubscription struct {
	Config           *configurations.Config
	Logger           *zap.Logger
	EntryExitService *services.EntryExitModifierService
	JSM              nats.JetStreamManagement
}

func (rcv *UserEventSubscription) Subscribe() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamUser, constants.DurableEntryExitUserCreated),
			nats.DeliverSubject(constants.DeliverEntryExitUserCreated),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
	}

	// subscribeCreatedStudent sub
	_, err := rcv.JSM.QueueSubscribe(constants.SubjectUserCreated, constants.QueueEntryExitUserCreated, opts, rcv.HandlerNatsMessageCreateStudent)
	if err != nil {
		return fmt.Errorf("rcv.JSM.QueueSubscribe: %v", err)
	}
	return nil
}

var (
	batchSize = 10
)

func (rcv *UserEventSubscription) PullSubscribe() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamUser, constants.DurableEntryExitUserCreated),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
			nats.DeliverNew(),
		},
		PullOpt: nats.PullSubscribeOption{
			FetchSize: 500,
			BatchSize: batchSize,
		},
	}

	return rcv.JSM.PullSubscribe(constants.SubjectUserCreated, constants.DurableEntryExitUserCreated,
		rcv.BatchCreateStudentQR, opts)
}

func (rcv *UserEventSubscription) HandlerNatsMessageCreateStudent(ctx context.Context, raw []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Do not process if internal config for auto generating qr code is disabled
	enableAutoGenQRCode, err := rcv.EntryExitService.CheckAutoGenQRCodeIsEnabled(ctx)
	if err != nil {
		return false, status.Errorf(codes.Internal, err.Error())
	}

	if !enableAutoGenQRCode {
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
		_, err = rcv.EntryExitService.Generate(ctx, msg.StudentId)
		if err != nil {
			return true, fmt.Errorf("err rcv.EntryExitService.Generate: %w", err)
		}
	default:
		rcv.Logger.Info("User not a student")
	}
	return false, nil
}

func (rcv *UserEventSubscription) BatchCreateStudentQR(msgs []*natsOrg.Msg) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	studentIDs := []string{}
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

		// Do not process if internal config for auto generating qr code is disabled
		enableAutoGenQRCode, err := rcv.EntryExitService.CheckAutoGenQRCodeIsEnabled(ctx)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}

		if !enableAutoGenQRCode {
			return nil
		}

		req := &upb.EvtUser{}
		if err = proto.Unmarshal(dataInMsg.Payload, req); err != nil {
			return err
		}

		switch req.Message.(type) {
		case *upb.EvtUser_CreateStudent_:
			msg := req.GetCreateStudent()
			studentIDs = append(studentIDs, msg.StudentId)
		default:
			rcv.Logger.Info("User not a student")
		}
	}

	_, err := rcv.EntryExitService.GenerateBatchQRCodes(ctx, &eepb.GenerateBatchQRCodesRequest{
		StudentIds: studentIDs,
	})
	if err != nil {
		return fmt.Errorf("err rcv.EntryExitService.Generate: %w", err)
	}

	return nil
}
