package subscriptions

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type CloudConvert struct {
	Logger *zap.Logger
	DB     database.Ext
	JSM    nats.JetStreamManagement

	MediaRepo interface {
		UpdateConvertedImages(ctx context.Context, db database.QueryExecer, media []*entities.Media) error
	}

	ConversionTaskRepo interface {
		RetrieveResourceURL(ctx context.Context, db database.QueryExecer, jobUUID pgtype.Text) (string, string, error)
		UpdateTasks(ctx context.Context, db database.QueryExecer, tasks []*entities.ConversionTask) error
	}

	ConversionSvc interface {
		UploadPrefixURL() string
	}
}

func (c *CloudConvert) Subscribe() error {
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamCloudConvertJobEvent, constants.DurableCloudConvertJobEventNatsJS),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverCloudConvertJobEvent),
			nats.AckWait(30 * time.Second),
		},
	}

	_, err := c.JSM.QueueSubscribe(constants.SubjectCloudConvertJobEventNatsJS, constants.QueueCloudConvertJobEventNatsJS, option, c.handleCloudConvertJobEvent)
	if err != nil {
		return fmt.Errorf("sub.Subscribe: %w", err)
	}

	return nil
}

func (c *CloudConvert) handleCloudConvertJobEvent(ctx context.Context, data []byte) (bool, error) {
	var cloudData npb.CloudConvertJobData
	if err := proto.Unmarshal(data, &cloudData); err != nil {
		return false, fmt.Errorf("handleCloudConvertJobEvent proto.Unmarshal: %w", err)
	}

	ackAble, err := c.handle(ctx, &cloudData)
	if err != nil {
		return true, fmt.Errorf("c.handle: %w", err)
	}
	if ackAble {
		return false, nil
	}

	return false, nil
}

func (c *CloudConvert) handle(ctx context.Context, data *npb.CloudConvertJobData) (bool, error) {
	resourceURL, resourcePath, err := c.ConversionTaskRepo.RetrieveResourceURL(ctx, c.DB, database.Text(data.JobId))
	if err != nil {
		return true, err
	}

	ctx = golibs.ResourcePathToCtx(ctx, resourcePath)
	status := services.ToConversionTaskStatus(data.JobStatus)

	e := new(entities.ConversionTask)
	e.TaskUUID.Set(data.JobId)
	e.Status.Set(status)
	e.ConversionResponse.Set(data.RawPayload)

	var images []*entities.ConvertedImage
	if status == bpb.ConversionTaskStatus_CONVERSION_TASK_STATUS_FINISHED.String() {
		uploadURL := c.ConversionSvc.UploadPrefixURL()
		for _, f := range data.ConvertedFiles {
			url := fmt.Sprintf("%s/%s", uploadURL, f)
			images = append(images, &entities.ConvertedImage{
				ImageURL: url,
			})
		}
	}

	m := new(entities.Media)
	m.Resource.Set(resourceURL)
	if len(images) == 0 {
		m.ConvertedImages.Set(nil)
	} else {
		m.ConvertedImages.Set(images)
	}

	if err := database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := c.ConversionTaskRepo.UpdateTasks(ctx, tx, []*entities.ConversionTask{e}); err != nil {
			return err
		}
		if err := c.MediaRepo.UpdateConvertedImages(ctx, tx, []*entities.Media{m}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return false, err
	}

	return true, nil
}
