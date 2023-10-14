package subscriptions

import (
	"context"
	"fmt"
	"time"

	enigma_entities "github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/yasuo/services"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	nats_org "github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type JprepUserRegistration struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement
	SubsJS []nats_org.Subscription

	UserService interface {
		SyncStudent(ctx context.Context, req []*npb.EventUserRegistration_Student) error
		SyncTeacher(ctx context.Context, req []*npb.EventUserRegistration_Staff) error
	}
	PartnerSyncDataLogService interface {
		UpdateLogStatus(ctx context.Context, id, status string) error
	}
	ClassService interface {
		SyncClassMember(ctx context.Context, req []*npb.EventUserRegistration_Student) error
	}
	ConfigService interface {
		UpsertConfig(ctx context.Context, upsertReq *services.UpsertConfig) error
	}
}

func (j *JprepUserRegistration) Subscribe() error {
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
			nats.DeliverNew(),
		},
	}

	optionClassMember := nats.Option{
		JetStreamOptions: append(option.JetStreamOptions,
			nats.Bind(constants.StreamSyncUserRegistration, constants.DurableSyncClassMember),
			nats.DeliverSubject(constants.DeliverSyncUserRegistrationClassMember)),
	}
	_, err := j.JSM.QueueSubscribe(constants.SubjectUserRegistrationNatsJS,
		constants.QueueSyncClassMember, optionClassMember, j.syncClassMemberHandler)
	if err != nil {
		return fmt.Errorf("syncClassMemberSub.Subscribe: %w", err)
	}

	return nil
}

func (j *JprepUserRegistration) syncClassMemberHandler(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req npb.EventUserRegistration
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("syncClassMemberHandler proto.Unmarshal: %w", err)
	}
	j.Logger.Info("JprepUserRegistration.syncClassMemberHandler",
		zap.String("signature", req.Signature),
	)
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusProcessing)); err != nil {
		return true, fmt.Errorf("JprepUserRegistration.syncClassMemberHandler update log status to processing: %w", err)
	}
	if len(req.Students) == 0 {
		return false, fmt.Errorf("syncClassMemberHandler length of Students = 0")
	}
	if err := nats.ChunkHandler(len(req.Students), constants.MaxRecordProcessPertime, func(start, end int) error {
		return j.ClassService.SyncClassMember(ctx, req.Students[start:end])
	}); err != nil {
		return true, fmt.Errorf("syncClassMemberHandler err SyncClassMember: %w", err)
	}
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusSuccess)); err != nil {
		return true, fmt.Errorf("JprepUserRegistration.syncClassMemberHandler update log status to success: %w", err)
	}

	return false, nil
}
