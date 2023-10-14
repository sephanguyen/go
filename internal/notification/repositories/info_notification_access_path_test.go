package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInfoNotificationAccessPathRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	notiID := "noti-id-1"
	locationID := "loc-id-1"
	accessPath := "loc-id-1"
	testCases := []struct {
		Name  string
		Ent   *entities.InfoNotificationAccessPath
		Err   error
		SetUp func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Ent: &entities.InfoNotificationAccessPath{
				NotificationID: database.Text(notiID),
				LocationID:     database.Text(locationID),
				AccessPath:     database.Text(accessPath),
			},
			SetUp: func(ctx context.Context) {
				e := &entities.InfoNotification{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, mock.Anything)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
	}

	repo := &InfoNotificationAccessPathRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.Upsert(ctx, db, testCase.Ent)
			if testCase.Err == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestInfoNotificationAccessPathRepo_BulkUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	testCases := []struct {
		Name        string
		Req         interface{}
		ExpectedErr error
		SetUp       func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Req: []*entities.InfoNotificationAccessPath{
				{
					NotificationID: database.Text("noti-1"),
					LocationID:     database.Text("loc-1"),
					AccessPath:     database.Text("access-path-1"),
				},
			},
			ExpectedErr: nil,
			SetUp: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	repo := &InfoNotificationAccessPathRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.BulkUpsert(ctx, db, testCase.Req.([]*entities.InfoNotificationAccessPath))
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestInfoNotificationAccessPathRepo_GetByNotificationIDAndNotInLocationIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &InfoNotificationAccessPathRepo{}

	ent1 := &entities.InfoNotificationAccessPath{}
	ent2 := &entities.InfoNotificationAccessPath{}
	database.AllRandomEntity(ent1)
	database.AllRandomEntity(ent2)
	notiID := "noti_1"
	locationIDs := []string{"loc-1"}
	t.Run("success", func(t *testing.T) {
		fields, vals1 := ent1.FieldMap()
		_, vals2 := ent2.FieldMap()
		// scan twice using values from 2 entities
		mockDB.MockScanArray(nil, fields, [][]interface{}{vals1, vals2})

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text(notiID), database.TextArray(locationIDs))
		notiLocs, err := r.GetByNotificationIDAndNotInLocationIDs(ctx, db, notiID, locationIDs)
		assert.Nil(t, err)
		assert.Equal(t, ent1, notiLocs[0])
		assert.Equal(t, ent2, notiLocs[1])
	})
	t.Run("error query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, database.Text(notiID), database.TextArray(locationIDs))
		notiLocs, err := r.GetByNotificationIDAndNotInLocationIDs(ctx, db, notiID, locationIDs)
		assert.Nil(t, notiLocs)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
	t.Run("error scan", func(t *testing.T) {
		fields, vals := ent1.FieldMap()
		mockDB.MockScanFields(pgx.ErrNoRows, fields, vals)
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text(notiID), database.TextArray(locationIDs))
		notiLocs, err := r.GetByNotificationIDAndNotInLocationIDs(ctx, db, notiID, locationIDs)
		assert.Nil(t, notiLocs)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestInfoNotificationAccessPathRepo_GetByNotificationID(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	notiID := "noti-id"
	repo := &InfoNotificationAccessPathRepo{}
	locationIDs := []string{"loc-1", "loc-2"}
	ctx := context.Background()
	t.Run("happy case", func(t *testing.T) {
		mockCtx, span := interceptors.StartSpan(context.Background(), "TestInfoNotificationAccessPathRepo_GetByNotificationID")
		defer span.End()

		mockDB.MockQueryArgs(t, nil, mockCtx, mock.Anything, notiID)
		mockDB.MockScanArray(nil, []string{"location_id"}, [][]interface{}{
			{
				&locationIDs[0],
			},
			{
				&locationIDs[1],
			},
		})

		locs, err := repo.GetLocationIDsByNotificationID(ctx, mockDB.DB, notiID)
		assert.Nil(t, err)
		assert.Equal(t, locationIDs, locs)
	})
}

func TestInfoNotificationAccessPathRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()

	testCases := []struct {
		Name    string
		Filter  *SoftDeleteNotificationAccessPathFilter
		ExpcErr error
		Setup   func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Filter: &SoftDeleteNotificationAccessPathFilter{
				NotificationIDs: database.TextArray([]string{"noti-1"}),
				LocationIDs:     database.TextArray([]string{"loc-1"}),
			},
			ExpcErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`1`)), nil, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name: "err conn closed",
			Filter: &SoftDeleteNotificationAccessPathFilter{
				NotificationIDs: database.TextArray([]string{"noti-1"}),
				LocationIDs:     database.TextArray([]string{"loc-1"}),
			},
			ExpcErr: fmt.Errorf("err db.Exec: %v", puddle.ErrClosedPool),
			Setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, nil, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name:    "err nil filter",
			Filter:  NewSoftDeleteNotificationAccessPathFilter(),
			ExpcErr: fmt.Errorf("cannot delete notification access path without notification_id"),
			Setup: func(ctx context.Context) {
			},
		},
	}

	infoNotificationAccessPathRepo := &InfoNotificationAccessPathRepo{}
	ctx := context.Background()
	for _, testCase := range testCases {
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			err := infoNotificationAccessPathRepo.SoftDelete(ctx, mockDB.DB, testCase.Filter)
			if testCase.ExpcErr == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.ExpcErr.Error(), err.Error())
			}
		})
	}
}
