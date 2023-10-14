package subscriptions

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/cloudconvert"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleCloudConvertJobEvent(t *testing.T) {
	t.Parallel()
	t.Run("job id not found", func(t *testing.T) {
		t.Parallel()
		jobID := "id"

		conversionTaskRepo := &mock_repositories.MockConversionTaskRepo{}
		conversionTaskRepo.On("RetrieveResourceURL", mock.Anything, mock.Anything, database.Text(jobID)).Once().Return("", "", pgx.ErrNoRows)

		data := &npb.CloudConvertJobData{
			JobId: jobID,
		}

		s := &CloudConvert{
			ConversionTaskRepo: conversionTaskRepo,
		}

		ackable, err := s.handle(context.Background(), data)
		assert.Equal(t, pgx.ErrNoRows, err)
		assert.True(t, ackable)
	})

	t.Run("job's status is failed", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		db := &mock_database.Ext{}
		tx := &mock_database.Tx{}
		db.On("Begin", mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		jobID := "id"
		filename := idutil.ULIDNow()

		conversionTaskRepo := &mock_repositories.MockConversionTaskRepo{}
		conversionTaskRepo.On("RetrieveResourceURL", mock.Anything, mock.Anything, database.Text(jobID)).Once().Return("https://"+filename, "", nil)

		taskEntity := new(entities.ConversionTask)
		taskEntity.TaskUUID.Set(jobID)
		taskEntity.Status.Set(bpb.ConversionTaskStatus_CONVERSION_TASK_STATUS_FAILED.String())
		taskEntity.ConversionResponse.Set(nil)
		conversionTaskRepo.On("UpdateTasks", mock.Anything, tx, []*entities.ConversionTask{taskEntity}).Once().Return(nil)

		m := new(entities.Media)
		m.Resource.Set("https://" + filename)
		m.ConvertedImages.Set(nil)

		mediaRepo := &mock_repositories.MockMediaRepo{}
		mediaRepo.On("UpdateConvertedImages", mock.Anything, tx, []*entities.Media{m}).Once().Return(nil)

		cloudConvertSvc := &cloudconvert.Service{
			StorageBucket:   "bucket",
			StorageEndpoint: "endpoint",
		}

		data := &npb.CloudConvertJobData{
			JobId:     jobID,
			JobStatus: "job.failed",
		}

		s := &CloudConvert{
			DB:                 db,
			ConversionTaskRepo: conversionTaskRepo,
			ConversionSvc:      cloudConvertSvc,
			MediaRepo:          mediaRepo,
		}

		ackAble, err := s.handle(ctx, data)
		assert.Nil(t, err)
		assert.True(t, ackAble)
	})

	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		db := &mock_database.Ext{}
		tx := &mock_database.Tx{}
		db.On("Begin", mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		jobID := "id"
		filename := idutil.ULIDNow()
		now := time.Now().UTC().Format("2006-01-02")

		conversionTaskRepo := &mock_repositories.MockConversionTaskRepo{}
		conversionTaskRepo.On("RetrieveResourceURL", mock.Anything, mock.Anything, database.Text(jobID)).Once().Return("https://"+filename, "", nil)

		taskEntity := new(entities.ConversionTask)
		taskEntity.TaskUUID.Set(jobID)
		taskEntity.Status.Set(bpb.ConversionTaskStatus_CONVERSION_TASK_STATUS_FINISHED.String())
		taskEntity.ConversionResponse.Set(nil)
		conversionTaskRepo.On("UpdateTasks", mock.Anything, mock.Anything, []*entities.ConversionTask{taskEntity}).Once().Return(nil)

		images := []*entities.ConvertedImage{
			{
				ImageURL: fmt.Sprintf("%s/%s/assignment/%s/images/%s/f1", "endpoint", "bucket", now, filename),
			},
			{
				ImageURL: fmt.Sprintf("%s/%s/assignment/%s/images/%s/f2", "endpoint", "bucket", now, filename),
			},
		}

		m := new(entities.Media)
		m.Resource.Set("https://" + filename)
		m.ConvertedImages.Set(images)

		mediaRepo := &mock_repositories.MockMediaRepo{}
		mediaRepo.On("UpdateConvertedImages", mock.Anything, mock.Anything, []*entities.Media{m}).Once().Return(nil)

		cloudConvertSvc := &cloudconvert.Service{
			StorageBucket:   "bucket",
			StorageEndpoint: "endpoint",
		}

		data := &npb.CloudConvertJobData{
			JobId:      jobID,
			JobStatus:  "job.finished",
			RawPayload: nil,
			ConvertedFiles: []string{
				fmt.Sprintf("assignment/%s/images/%s/f1", now, filename),
				fmt.Sprintf("assignment/%s/images/%s/f2", now, filename),
			},
		}

		s := &CloudConvert{
			DB:                 db,
			ConversionTaskRepo: conversionTaskRepo,
			ConversionSvc:      cloudConvertSvc,
			MediaRepo:          mediaRepo,
		}

		ackAble, err := s.handle(ctx, data)
		assert.Nil(t, err)
		assert.True(t, ackAble)
	})
}
