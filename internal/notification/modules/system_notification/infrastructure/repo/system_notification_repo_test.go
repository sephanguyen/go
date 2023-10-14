package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/model"
	"github.com/manabie-com/backend/mock/testutil"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_SystemNotificationRepo(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()

	repo := &SystemNotificationRepo{}
	testCases := []struct {
		Name               string
		SystemNotification *model.SystemNotification
		Err                error
		Setup              func(ctx context.Context, t *testing.T)
	}{
		{
			Name:               "happy case",
			SystemNotification: &model.SystemNotification{},
			Err:                nil,
			Setup: func(ctx context.Context, t *testing.T) {
				ctx, span := interceptors.StartSpan(ctx, "InsertSystemNotification")
				defer span.End()

				e := &model.SystemNotification{}
				values := database.GetScanFields(e, database.GetFieldNames(e))
				args := []interface{}{ctx, mock.AnythingOfType("string")}
				for range values {
					args = append(args, mock.Anything)
				}
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			tc.Setup(ctx, t)
			err := repo.UpsertSystemNotification(ctx, mockDB.DB, tc.SystemNotification)
			assert.Equal(t, tc.Err, err)
		})
	}
}

func Test_FindSystemNotifications(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()

	repo := &SystemNotificationRepo{}

	eventInDB1 := &model.SystemNotification{}
	eventInDB2 := &model.SystemNotification{}
	database.AllRandomEntity(eventInDB1)
	database.AllRandomEntity(eventInDB2)

	fields, values1 := eventInDB1.FieldMap()
	_, values2 := eventInDB2.FieldMap()

	filter := NewFindSystemNotificationFilter()
	filter.ValidFrom.Set(time.Now())
	filter.UserID.Set("user_id_1")
	filter.Status.Set([]string{npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE.String(), npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NEW.String()})
	filter.Language.Set("en")
	filter.Keyword.Set("keyword")
	filter.Limit.Set(5)
	filter.Offset.Set(0)

	ctx := context.Background()

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, filter.ValidFrom, filter.UserID, filter.Language, filter.Keyword, filter.Status, filter.Limit, filter.Offset)
		mockDB.MockScanArray(nil, fields, [][]interface{}{values1, values2})

		res, err := repo.FindSystemNotifications(ctx, mockDB.DB, &filter)

		assert.Nil(t, err)
		assert.Equal(t, eventInDB1, res[0])
		assert.Equal(t, eventInDB2, res[1])
	})

	t.Run("error query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, filter.ValidFrom, filter.UserID, filter.Language, filter.Keyword, filter.Status, filter.Limit, filter.Offset)

		res, err := repo.FindSystemNotifications(ctx, mockDB.DB, &filter)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("error scan", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, filter.ValidFrom, filter.UserID, filter.Language, filter.Keyword, filter.Status, filter.Limit, filter.Offset)
		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{values1, values2})

		res, err := repo.FindSystemNotifications(ctx, mockDB.DB, &filter)

		assert.Nil(t, res)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func Test_CountSystemNotifications(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()

	repo := &SystemNotificationRepo{}

	ctx := context.Background()

	filter := NewFindSystemNotificationFilter()
	filter.ValidFrom.Set(time.Now())
	filter.UserID.Set("user_id_1")
	filter.Status.Set(nil)
	filter.Language.Set("en")
	filter.Keyword.Set("keyword")

	res1 := &TotalForStatus{
		Status: database.Text(npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE.String()),
		Total:  database.Int8(1),
	}
	res2 := &TotalForStatus{
		Status: database.Text(npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NEW.String()),
		Total:  database.Int8(2),
	}
	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, filter.ValidFrom, filter.UserID, filter.Language, filter.Keyword, filter.Status)
		mockDB.MockScanArray(nil, []string{"status", "total"}, [][]interface{}{{&res1.Status, &res1.Total}, {&res2.Status, &res2.Total}})

		res, err := repo.CountSystemNotifications(ctx, mockDB.DB, &filter)

		assert.Nil(t, err)
		assert.Equal(t, uint32(res1.Total.Int), res[res1.Status.String])
		assert.Equal(t, uint32(res2.Total.Int), res[res2.Status.String])
		assert.Equal(t, uint32(res1.Total.Int+res2.Total.Int), res[npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NONE.String()])
	})

	t.Run("rows scan field error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, filter.ValidFrom, filter.UserID, filter.Language, filter.Keyword, filter.Status)
		mockDB.MockScanArray(nil, []string{"status", "total"}, [][]interface{}{{&res1.Status, &res1.Total}, {&res2.Status, &res2.Total}})

		res, err := repo.CountSystemNotifications(ctx, mockDB.DB, &filter)

		assert.Nil(t, res)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func Test_FindByReferenceID(t *testing.T) {
	t.Parallel()

	repo := &SystemNotificationRepo{}
	mockDB := testutil.NewMockDB()
	refID := "refID"

	testCases := []struct {
		Name               string
		ReferenceID        string
		Err                error
		SystemNotification *model.SystemNotification
		Setup              func(ctx context.Context)
	}{
		{
			Name:               "happy case",
			ReferenceID:        refID,
			Err:                nil,
			SystemNotification: &model.SystemNotification{},
			Setup: func(ctx context.Context) {
				ctx, span := interceptors.StartSpan(ctx, "span")
				defer span.End()

				e := &model.SystemNotification{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockDB.MockQueryRowArgs(t, ctx, mock.Anything, database.Text(refID))
				mockDB.MockRowScanFields(nil, fields, values)
			},
		},
		{
			Name:               "no rows but still pass",
			ReferenceID:        refID,
			Err:                nil,
			SystemNotification: nil,
			Setup: func(ctx context.Context) {
				ctx, span := interceptors.StartSpan(ctx, "span")
				defer span.End()

				e := &model.SystemNotification{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockDB.MockQueryRowArgs(t, ctx, mock.Anything, database.Text(refID))
				mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			tc.Setup(ctx)
			res, err := repo.FindByReferenceID(ctx, mockDB.DB, refID)
			assert.Equal(t, tc.Err, err)
			assert.Equal(t, tc.SystemNotification, res)
		})
	}
}

func TestSystemNotificationRepo_SetStatus(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	repo := &SystemNotificationRepo{}

	t.Run("happy case", func(t *testing.T) {
		systemNotificationID := "sn-1"
		status := npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE
		ctx := context.Background()

		mockDB.DB.On("Exec", mock.Anything, mock.Anything, database.Text(status.String()), database.Text(systemNotificationID)).Once().
			Return(pgconn.CommandTag("1"), nil)

		err := repo.SetStatus(ctx, mockDB.DB, systemNotificationID, status.String())
		assert.Nil(t, err)
	})
}

func TestSystemNotificationRepo_CheckUserBelongToSystemNotification(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	repo := &SystemNotificationRepo{}

	testCases := []struct {
		Name                 string
		Res                  bool
		Err                  error
		UserID               string
		SystemNotificationID string
		Setup                func(ctx context.Context, userID, snID string)
	}{
		{
			Name:                 "case true",
			Res:                  true,
			Err:                  nil,
			UserID:               "user-1",
			SystemNotificationID: "sn-1",
			Setup: func(ctx context.Context, userID, snID string) {
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, database.Text(userID), database.Text(snID))
				id := "sn-1"
				mockDB.MockRowScanFields(nil, []string{"field"}, []interface{}{&id})
			},
		},
		{
			Name:                 "case false",
			Res:                  false,
			Err:                  nil,
			UserID:               "user-1",
			SystemNotificationID: "sn-1",
			Setup: func(ctx context.Context, userID, snID string) {
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, database.Text(userID), database.Text(snID))
				id := ""
				mockDB.MockRowScanFields(nil, []string{"field"}, []interface{}{&id})
			},
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		tc.Setup(ctx, tc.UserID, tc.SystemNotificationID)
		t.Run(tc.Name, func(t *testing.T) {
			isBelong, err := repo.CheckUserBelongToSystemNotification(ctx, mockDB.DB, tc.UserID, tc.SystemNotificationID)
			assert.Equal(t, tc.Err, err)
			assert.Equal(t, tc.Res, isBelong)
		})
	}
}
