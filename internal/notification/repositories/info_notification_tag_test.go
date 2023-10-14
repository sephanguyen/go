package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInfoNotificationTagRepo_GetByNotificationIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	entInDB1 := &entities.InfoNotificationTag{}
	entInDB2 := &entities.InfoNotificationTag{}
	database.AllRandomEntity(entInDB1)
	database.AllRandomEntity(entInDB2)

	notiIDs := database.TextArray([]string{entInDB1.NotificationID.String, entInDB1.NotificationTagID.String})

	type TestCase struct {
		Name    string
		NotiIDs []string
		Err     error
		SetUp   func(ctx context.Context, this *TestCase)
	}

	testCases := []TestCase{
		{
			Name:    "happy case",
			NotiIDs: []string{entInDB1.NotificationID.String, entInDB2.NotificationID.String},
			SetUp: func(ctx context.Context, this *TestCase) {

				fields, vals1 := entInDB1.FieldMap()
				_, vals2 := entInDB2.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{vals1, vals2})
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, notiIDs)
			},
		},
	}

	repo := &InfoNotificationTagRepo{}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx, &testCase)
			res, err := repo.GetByNotificationIDs(ctx, db, notiIDs)
			if testCase.Err != nil {
				assert.Equal(t, testCase.Err, err)
			}
			assert.Nil(t, err)
			assert.Equal(t, entInDB1, res[entInDB1.NotificationID.String][0])
			assert.Equal(t, entInDB2, res[entInDB2.NotificationID.String][0])
		})
	}
}

func TestInfoNotificationTagRepoV2_GetNotificationIDsByTagIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &InfoNotificationTagRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	tagIDsReq := database.TextArray([]string{"tag_id-1", "tag_id-2"})
	notficationIDsRes := []pgtype.Text{database.Text("notification_id-1"), database.Text("notification_id-2")}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, tagIDsReq)
		mockDB.MockScanArray(nil, []string{"notification_msg_id"}, [][]interface{}{{&notficationIDsRes[0]}, {&notficationIDsRes[1]}})

		_, err := r.GetNotificationIDsByTagIDs(ctx, db, tagIDsReq)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, tagIDsReq)
		mockDB.MockScanArray(nil, []string{"notification_msg_id"}, [][]interface{}{{&notficationIDsRes[0]}, {&notficationIDsRes[1]}})

		notIDs, err := r.GetNotificationIDsByTagIDs(ctx, db, tagIDsReq)
		assert.NoError(t, err)
		for i, notID := range notIDs {
			assert.Equal(t, notID, notficationIDsRes[i].String)
		}
	})
}

func TestInfoNotificationTagRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()

	testCases := []struct {
		Name    string
		Filter  *SoftDeleteNotificationTagFilter
		ExpcErr error
		Setup   func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Filter: &SoftDeleteNotificationTagFilter{
				NotificationTagIDs: database.TextArray([]string{"noti-1"}),
				NotificationIDs:    database.TextArray([]string{"noti-tag-1"}),
				TagIDs:             database.TextArray([]string{"tag-1"}),
			},
			ExpcErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`1`)), nil, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name: "err conn closed",
			Filter: &SoftDeleteNotificationTagFilter{
				NotificationTagIDs: database.TextArray([]string{"noti-2"}),
				NotificationIDs:    database.TextArray([]string{"noti-tag-2"}),
				TagIDs:             database.TextArray([]string{"tag-2"}),
			},
			ExpcErr: puddle.ErrClosedPool,
			Setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, nil, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name:    "err nil filter",
			Filter:  NewSoftDeleteNotificationTagFilter(),
			ExpcErr: fmt.Errorf("cannot soft delete with nil filter"),
			Setup: func(ctx context.Context) {
			},
		},
	}

	infoNotiTagRepo := &InfoNotificationTagRepo{}
	ctx := context.Background()
	for _, testCase := range testCases {
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			err := infoNotiTagRepo.SoftDelete(ctx, mockDB.DB, testCase.Filter)
			if testCase.ExpcErr == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.ExpcErr.Error(), err.Error())
			}
		})
	}
}

func TestInfoNotificationTagRepo_BulkInsert(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	ifntRepo := InfoNotificationTagRepo{}
	records := utils.GenInfoNotificationTagBulkInsert()
	testCases := []struct {
		Name    string
		Data    []*entities.InfoNotificationTag
		ExpcErr error
		Setup   func(ctx context.Context)
	}{
		{
			Name:    "happy case",
			Data:    records,
			ExpcErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				for range records {
					cmdTag := pgconn.CommandTag([]byte(`1`))
					batchResults.On("Exec").Once().Return(cmdTag, nil)
				}
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name:    "error send batch",
			Data:    records,
			ExpcErr: fmt.Errorf("batchResults.Exec: %w", pgx.ErrTxClosed),
			Setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				for range records {
					cmdTag := pgconn.CommandTag([]byte(`0`))
					batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				}
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	ctx := context.Background()
	for _, testCase := range testCases {
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			err := ifntRepo.BulkUpsert(ctx, mockDB.DB, testCase.Data)
			if testCase.ExpcErr == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.ExpcErr.Error(), err.Error())
			}
		})
	}
}
