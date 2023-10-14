package consumers

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type SubscriberHandler interface {
	Handle(ctx context.Context, msg []byte) (bool, error)
}

type LessonDeletedHandler struct {
	Logger            *zap.Logger
	WrapperConnection *support.WrapperDBConnection
	JSM               nats.JetStreamManagement
	Cfg               configurations.Config
	RecordedVideoRepo infrastructure.RecordedVideoRepo
	MediaModulePort   infrastructure.MediaModulePort
	FileStore         infrastructure.FileStore
}

func (l *LessonDeletedHandler) Handle(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	r := &bpb.EvtLesson{}
	if err := proto.Unmarshal(data, r); err != nil {
		return false, err
	}
	switch m := r.Message.(type) {
	case *bpb.EvtLesson_DeletedLessons_:
		lessonIds := m.DeletedLessons.GetLessonIds()
		if err := l.handle(ctx, lessonIds); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, fmt.Errorf("type message is invalid, expected EvtLesson_DeletedLessons_ but got: %T", r.Message)
}

func (l *LessonDeletedHandler) handle(ctx context.Context, lessonIds []string) error {
	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return err
	}
	if err := database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) error {
		records, err := l.RecordedVideoRepo.ListRecordingByLessonIDs(ctx, tx, lessonIds)
		if err != nil {
			return fmt.Errorf("RecordedVideoRepo.ListRecordingByLessonIDs: %s %w", lessonIds, err)
		}
		if len(records) > 0 {
			// get media to fill video id field
			medias, err := l.MediaModulePort.RetrieveMediasByIDs(ctx, records.GetMediaIDs())
			if err != nil {
				return fmt.Errorf("MediaModulePort.RetrieveMediasByIDs: %v", err)
			}
			if err = records.WithMedias(medias); err != nil {
				return fmt.Errorf("records.WithMedias: %s %w", lessonIds, err)
			}

			if err = l.RecordedVideoRepo.DeleteRecording(ctx, tx, records.GetRecordIDs()); err != nil {
				return fmt.Errorf("RecordedVideoRepo.DeleteRecording: %s %w", lessonIds, err)
			}

			if err = l.MediaModulePort.DeleteMedias(ctx, records.GetMediaIDs()); err != nil {
				return fmt.Errorf("MediaModulePort.DeleteMedias: %s %w", lessonIds, err)
			}

			for _, resource := range records.GetResources() {
				path := filepath.Dir(resource) + "/"
				objects, err := l.FileStore.GetObjectsWithPrefix(ctx, l.Cfg.Agora.BucketName, path, "")
				if err != nil {
					return fmt.Errorf("FileStore.GetObjectsWithPrefix: %s %w", path, err)
				}

				for _, v := range objects {
					if err = l.FileStore.DeleteObject(ctx, l.Cfg.Agora.BucketName, v.Name); err != nil {
						return fmt.Errorf("FileStore.DeleteObject: %s %w", resource, err)
					}
				}
			}
		}
		return nil
	}); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}
