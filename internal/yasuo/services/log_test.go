package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	entities_enigma "github.com/manabie-com/backend/internal/enigma/entities"
	mock_enigma_repo "github.com/manabie-com/backend/mock/enigma/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type updateLogStatusRequest struct {
	id     string
	status string
}

func TestLogService_UpdateLogStatus(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	partnerSyncDataLogSplitRepo := new(mock_enigma_repo.MockPartnerSyncDataLogSplitRepo)

	logService := LogsService{
		DB:                          db,
		PartnerSyncDataLogSplitRepo: partnerSyncDataLogSplitRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: updateLogStatusRequest{
				id:     "1234",
				status: string(entities_enigma.StatusSuccess),
			},
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("UpdateLogStatus", ctx, db, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "update log status fail",
			ctx:  ctx,
			req: updateLogStatusRequest{
				id:     "1234",
				status: string(entities_enigma.StatusSuccess),
			},
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("UpdateLogStatus", ctx, db, mock.Anything).Once().Return(fmt.Errorf("failed"))
			},
			expectedErr: fmt.Errorf("LogsService.UpdateLogStatus: failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			req := testCase.req.(updateLogStatusRequest)
			err := logService.UpdateLogStatus(testCase.ctx, req.id, req.status)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})

		mock.AssertExpectationsForObjects(t, db, partnerSyncDataLogSplitRepo)
	}
}
