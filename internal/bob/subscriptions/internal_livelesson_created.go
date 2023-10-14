package subscriptions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type InternalLiveLessonCreated struct {
	JSM    nats.JetStreamManagement
	Logger *zap.Logger
	DB     database.Ext

	LessonRepo interface {
		UpdateRoomID(ctx context.Context, db database.QueryExecer, lessonID, roomID pgtype.Text) error
	}

	WhiteboardSvc interface {
		CreateRoom(context.Context, *whiteboard.CreateRoomRequest) (*whiteboard.CreateRoomResponse, error)
	}
}

func (i *InternalLiveLessonCreated) Subscribe() error {
	// bus := i.BusFactory.GetConn()

	subject := constants.SubjectLessonCreated
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamLesson, constants.DurableInternalLessonCreated),
			nats.DeliverSubject(constants.DeliverLessonCreated),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
	}
	_, err := i.JSM.QueueSubscribe(subject, constants.QueueInternalLessonCreated, opts, i.handleCreateRoom)
	if err != nil {
		return fmt.Errorf("i.JSM.QueueSubscribe: %w", err)
	}

	return nil
}

func (i *InternalLiveLessonCreated) handleCreateRoom(ctx context.Context, data []byte) (bool, error) {
	req := bpb.EvtLesson{}
	if err := req.Unmarshal(data); err != nil {
		i.Logger.Error("proto.Unmarshal", zap.Error(err))
		return true, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	evtCreateLesson := req.GetCreateLessons()
	if evtCreateLesson == nil {
		return true, errors.New("evtCreateLesson is nil")
	}
	ackAble, err := i.handle(ctx, evtCreateLesson)
	if err != nil {
		i.Logger.Error("handleCreateRoom: i.handle", zap.Error(err))
	}
	return ackAble, err
}

func (i *InternalLiveLessonCreated) handle(ctx context.Context, data *bpb.EvtLesson_CreateLessons) (bool, error) {
	err := nats.ChunkHandler(len(data.Lessons), 10, func(start, end int) error {
		var errB error
		for _, l := range data.Lessons[start:end] {
			room, err := i.WhiteboardSvc.CreateRoom(ctx, &whiteboard.CreateRoomRequest{
				Name:     l.LessonId,
				IsRecord: false,
			})

			if err != nil {
				errB = multierr.Append(errB, err)
			} else {
				err = i.LessonRepo.UpdateRoomID(ctx, i.DB, database.Text(l.LessonId), database.Text(room.UUID))
				if err != nil {
					errB = multierr.Append(errB, err)
				}
			}
		}

		return errB
	})
	if err != nil {
		return false, err
	}

	return true, nil
}
