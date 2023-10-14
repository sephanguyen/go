package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	golibs_nats "github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/taskq/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type PublishImportUserEventsTaskOptions struct {
	ImportUserEventIDs []int64
	ResourcePath       string
}

func PublishImportUserEventsTask(ctx context.Context, db database.Ext, jsm golibs_nats.JetStreamManagement, opts *PublishImportUserEventsTaskOptions) *taskq.Message {
	zapLogger := ctxzap.Extract(ctx).Sugar()

	return taskq.RegisterTask(&taskq.TaskOptions{
		Name: fmt.Sprintf("PublishImportUserEventsTask-%s", idutil.ULIDNow()),
		Handler: func(ctx context.Context) error {
			zapLogger.Info("-----START: Publishing import user events task-----")
			importUserEventRepo := &repository.ImportUserEventRepo{}

			ctx = auth.InjectFakeJwtToken(ctx, opts.ResourcePath)
			importUserEvents, err := importUserEventRepo.GetByIDs(ctx, db, database.Int8Array(opts.ImportUserEventIDs))
			if err != nil {
				return fmt.Errorf("failed to importUserEventRepo.GetByIDs: %v", err)
			}

			zapLogger.Info(len(importUserEvents))

			for _, event := range importUserEvents {
				status := cpb.ImportUserEventStatus_IMPORT_USER_EVENT_STATUS_FINISHED.String()
				pubAck, err := publishUserEvent(ctx, constants.SubjectUserCreated, event, jsm)
				if err != nil {
					zapLogger.Errorf("failed to publishUserEvent: %v", err)
					status = cpb.ImportUserEventStatus_IMPORT_USER_EVENT_STATUS_FAILED.String()
				}
				if pubAck != nil {
					err = event.SequenceNumber.Set(pubAck.Sequence)
					if err != nil {
						zapLogger.Errorf("failed to event.SequenceNumber.Set: %v, import_user_event_id: %v", err, event.ID)
					}
				}
				err = event.Status.Set(status)
				if err != nil {
					zapLogger.Errorf("failed to event.Status.Set: %v, import_user_event_id: %v", err, event.ID)
				}

				// avoid harassing nats-server with time.Sleep
				time.Sleep(100 * time.Millisecond)
			}

			_, err = importUserEventRepo.Upsert(ctx, db, importUserEvents)
			if err != nil {
				return fmt.Errorf("failed to importUserEventRepo.Upsert: %v", err)
			}

			zapLogger.Info("-----DONE: Publishing import user events task-----")

			return nil
		},
	}).WithArgs(ctx)
}

func publishUserEvent(ctx context.Context, eventType string, importUserEvent *entity.ImportUserEvent, jsm golibs_nats.JetStreamManagement) (*nats.PubAck, error) {
	userEvent := &pb.EvtUser{}
	err := protojson.Unmarshal(importUserEvent.Payload.Bytes, userEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to protojson.Unmarshal: %v, import_user_event_id: %v", err, importUserEvent.ID)
	}
	data, err := proto.Marshal(userEvent)
	if err != nil {
		return nil, fmt.Errorf("marshal event %s error, %w", eventType, err)
	}
	pubAck, err := jsm.TracedPublish(ctx, "cronjob publishUserEvent", eventType, data)
	if err != nil {
		return nil, fmt.Errorf("publishUserEvent with %s: s.JSM.Publish failed: %w", eventType, err)
	}

	return pubAck, nil
}
