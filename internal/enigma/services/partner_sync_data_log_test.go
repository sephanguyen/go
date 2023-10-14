package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	entities_enigma "github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_enigma_repo "github.com/manabie-com/backend/mock/enigma/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name        string
	ctx         context.Context
	req         interface{}
	expectedErr error
	setup       func(ctx context.Context)
}

type updateLogStatusRequest struct {
	id     string
	status string
}

type getLogStatusRequest struct {
	signature string
}

func TestPartnerSyncDataLogService_UpdateLogStatus(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	partnerSyncDataLogSplitRepo := new(mock_enigma_repo.MockPartnerSyncDataLogSplitRepo)

	partnerSyncDataLogService := PartnerSyncDataLogService{
		DB:                          db,
		PartnerSyncDataLogSplitRepo: partnerSyncDataLogSplitRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: updateLogStatusRequest{
				id:     mock.Anything,
				status: string(entities_enigma.StatusSuccess),
			},
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("UpdateLogStatus", ctx, db, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "log id empty",
			ctx:  ctx,
			req: updateLogStatusRequest{
				id:     "",
				status: string(entities_enigma.StatusSuccess),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "status unknown",
			ctx:  ctx,
			req: updateLogStatusRequest{
				id:     mock.Anything,
				status: string("unknown"),
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("PartnerSyncDataLogService.UpdateLogStatus: %s", ErrPartnerSyncDataLogStatusUnknown),
		},
		{
			name: "update log status fail",
			ctx:  ctx,
			req: updateLogStatusRequest{
				id:     mock.Anything,
				status: string(entities_enigma.StatusSuccess),
			},
			setup: func(ctx context.Context) {
				partnerSyncDataLogSplitRepo.On("UpdateLogStatus", ctx, db, mock.Anything).Once().Return(fmt.Errorf("failed"))
			},
			expectedErr: fmt.Errorf("PartnerSyncDataLogService.UpdateLogStatus: failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			req := testCase.req.(updateLogStatusRequest)
			err := partnerSyncDataLogService.UpdateLogStatus(testCase.ctx, req.id, req.status)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})

		mock.AssertExpectationsForObjects(t, db, partnerSyncDataLogSplitRepo)
	}
}

func TestPartnerSyncDataLogService_GetLogBySignature(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	partnerSyncDataLogRepo := new(mock_enigma_repo.MockPartnerSyncDataLogRepo)

	partnerSyncDataLogService := PartnerSyncDataLogService{
		DB:                     db,
		PartnerSyncDataLogRepo: partnerSyncDataLogRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: getLogStatusRequest{
				signature: "signature-hash",
			},
			setup: func(ctx context.Context) {
				partnerSyncDataLogRepo.On("GetBySignature", ctx, db, "signature-hash").Once().Return(&entities_enigma.PartnerSyncDataLog{
					Signature: database.Text("signature-hash"),
				}, nil)
			},
		},
		{
			name: "signature empty",
			ctx:  ctx,
			req: getLogStatusRequest{
				signature: "",
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("Signature is empty"),
		},
		{
			name: "get log status fail",
			ctx:  ctx,
			req: getLogStatusRequest{
				signature: "signature-hash",
			},
			setup: func(ctx context.Context) {
				partnerSyncDataLogRepo.On("GetBySignature", ctx, db, "signature-hash").Once().Return(nil, fmt.Errorf("PartnerSyncDataLogService.GetLogBySignature: failed"))
			},
			expectedErr: fmt.Errorf("PartnerSyncDataLogService.GetLogBySignature: failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			req := testCase.req.(getLogStatusRequest)
			log, err := partnerSyncDataLogService.GetLogBySignature(testCase.ctx, req.signature)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, log.Signature.String, "signature-hash")
			}
		})

		mock.AssertExpectationsForObjects(t, db, partnerSyncDataLogRepo)
	}
}
