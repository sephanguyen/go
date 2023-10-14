package repo

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/model"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_SystemNotificationContent_SoftDeleteBySystemNotificationID(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	repo := &SystemNotificationContentRepo{}
	eventID := "id"
	testCases := []struct {
		Name                 string
		SystemNotificationID string
		Err                  error
		Setup                func(ctx context.Context)
	}{
		{
			Name:                 "happy case",
			SystemNotificationID: eventID,
			Err:                  nil,
			Setup: func(ctx context.Context) {
				ctx, span := interceptors.StartSpan(ctx, "span")
				defer span.End()
				mockDB.DB.On("Exec", ctx, mock.Anything, database.Text(eventID)).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
		{
			Name:                 "case error delete",
			SystemNotificationID: eventID,
			Err:                  fmt.Errorf("failed exec: %+v", pgx.ErrTxClosed),
			Setup: func(ctx context.Context) {
				ctx, span := interceptors.StartSpan(ctx, "span")
				defer span.End()

				mockDB.DB.On("Exec", ctx, mock.Anything, database.Text(eventID)).Once().Return(pgconn.CommandTag([]byte(`1`)), pgx.ErrTxClosed)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			tc.Setup(ctx)

			res := repo.SoftDeleteBySystemNotificationID(ctx, mockDB.DB, tc.SystemNotificationID)
			assert.Equal(t, tc.Err, res)
		})
	}
}

func Test_FindBySystemNotificationIDs(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	repo := &SystemNotificationContentRepo{}
	ids := []string{"id1", "id2"}
	testCases := []struct {
		Name                  string
		SystemNotificationIDs []string
		Err                   error
		Setup                 func(ctx context.Context)
	}{
		{
			Name:                  "happy case",
			SystemNotificationIDs: ids,
			Err:                   nil,
			Setup: func(ctx context.Context) {
				ctx, span := interceptors.StartSpan(ctx, "span")
				defer span.End()
				e := &model.SystemNotificationContent{}
				fieldNames := database.GetFieldNames(e)
				scanFields := database.GetScanFields(e, fieldNames)
				mockDB.MockQueryArgs(t, nil, ctx, mock.Anything, database.TextArray(ids))
				mockDB.MockScanArray(nil, fieldNames, [][]interface{}{
					scanFields,
				})
			},
		},
		{
			Name:                  "case error",
			SystemNotificationIDs: ids,
			Err:                   fmt.Errorf("failed ScanAll: err db.Query: %+v", pgx.ErrTxClosed),
			Setup: func(ctx context.Context) {
				ctx, span := interceptors.StartSpan(ctx, "span")
				defer span.End()

				e := &model.SystemNotificationContent{}
				fieldNames := database.GetFieldNames(e)
				scanFields := database.GetScanFields(e, fieldNames)
				mockDB.MockQueryArgs(t, pgx.ErrTxClosed, ctx, mock.Anything, database.TextArray(ids))
				mockDB.MockScanArray(nil, fieldNames, [][]interface{}{
					scanFields,
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			tc.Setup(ctx)

			_, err := repo.FindBySystemNotificationIDs(ctx, mockDB.DB, tc.SystemNotificationIDs)
			assert.Equal(t, tc.Err, err)
		})
	}
}

func Test_BulkInsertSystemNotificationContents(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	repo := &SystemNotificationContentRepo{}
	testCases := []struct {
		Name    string
		Content model.SystemNotificationContents
		Err     error
		Setup   func(ctx context.Context)
	}{
		{
			Name:    "happy case",
			Content: model.SystemNotificationContents{},
			Err:     nil,
			Setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			tc.Setup(ctx)
			err := repo.BulkInsertSystemNotificationContents(ctx, db, tc.Content)
			assert.Nil(t, err)
		})
	}
}
