package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/zeus/entities"
	mock_repositories "github.com/manabie-com/backend/mock/zeus/repositories"

	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreateLogs(t *testing.T) {
	t.Parallel()
	t.Run("create activity logs success", func(t *testing.T) {
		t.Parallel()
		req := &npb.ActivityLogEvtCreated{
			UserId:       mock.Anything,
			ActionType:   mock.Anything,
			ResourcePath: mock.Anything,
			RequestAt:    timestamppb.Now(),
			FinishedAt:   timestamppb.Now(),
			Payload:      []byte(mock.Anything),
		}
		activityLogRepo := &mock_repositories.MockActivityLogRepo{}
		activityLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := &CentralizeLogsService{
			ActivityLogRepo: activityLogRepo,
		}
		err := svc.CreateLogs(context.Background(), req)
		assert.Nil(t, err)
	})
}

func TestCreateBulk(t *testing.T) {
	t.Parallel()
	t.Run("create bulk activity logs success", func(t *testing.T) {
		t.Parallel()
		logs := []*entities.ActivityLog{}
		activityLogRepo := &mock_repositories.MockActivityLogRepo{}
		activityLogRepo.On("CreateBulk", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := &CentralizeLogsService{
			ActivityLogRepo: activityLogRepo,
		}
		err := svc.BulkCreateLogs(context.Background(), logs)
		assert.Nil(t, err)
	})
}
