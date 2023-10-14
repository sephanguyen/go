package services

import (
	"context"
	"testing"
	"time"

	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestConfigService_UpsertConfig(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	configRepo := &mock_repositories.MockConfigRepo{}
	db := &mock_database.Ext{}
	configService := &ConfigService{
		ConfigRepo: configRepo,
		DB:         db,
	}

	testCases := map[string]TestCase{
		"happy case": {
			req: &UpsertConfig{
				Key:     "key",
				Group:   "group",
				Country: "VN",
				Value:   "value",
			},
			setup: func(ctx context.Context) {
				configRepo.On("Upsert", ctx, db, mock.Anything).Once().Return(nil)
			},
			expectedErr: nil,
		},
		"error case": {
			req: &UpsertConfig{
				Key:     "key",
				Group:   "group",
				Country: "VN",
				Value:   "value",
			},
			setup: func(ctx context.Context) {
				configRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
			expectedErr: status.Error(codes.Internal, "ConfigRepo.Upsert:"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := configService.UpsertConfig(ctx, testCase.req.(*UpsertConfig))
			if testCase.expectedErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, db, configRepo)
		})
	}
}
