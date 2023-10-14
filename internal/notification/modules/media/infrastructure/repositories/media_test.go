package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/notification/modules/media/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/mock"
	"gotest.tools/assert"
)

func TestUpsertMediaBatch(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	mediaRepo := &MediaRepo{}
	testCases := []struct {
		Name  string
		Err   error
		Req   domain.Medias
		Setup func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Req: []*domain.Media{
				{
					MediaID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			Err: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "error send batch",
			Req: []*domain.Media{
				{
					MediaID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
				{
					MediaID: pgtype.Text{String: "2", Status: pgtype.Present},
				},
			},
			Err: fmt.Errorf("UpsertMediaBatch batchResults.Exec: %w", pgx.ErrTxClosed),
			Setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.Setup(ctx)
		err := mediaRepo.UpsertMediaBatch(ctx, db, testCase.Req)
		if testCase.Err != nil {
			assert.Equal(t, testCase.Err.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.Err, err)
		}
	}
}
