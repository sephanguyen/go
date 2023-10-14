package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotificationLocationFilterRepo_BulkUpsert(t *testing.T) {
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
			Req: []*entities.NotificationLocationFilter{
				{
					NotificationID: database.Text("notification-1"),
					LocationID:     database.Text("location-1"),
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

	repo := &NotificationLocationFilterRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.BulkUpsert(ctx, db, testCase.Req.([]*entities.NotificationLocationFilter))
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestNotificationLocationFilterRepo_GetNotificationIDsByLocationIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &NotificationLocationFilterRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	tagIDsReq := database.TextArray([]string{"tag_id-1", "tag_id-2"})
	notiIDsReq := database.TextArray([]string{"notification_id-1", "notification_id-2"})
	notficationIDsRes := []pgtype.Text{database.Text("notification_id-1"), database.Text("notification_id-2")}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, notiIDsReq, tagIDsReq)
		mockDB.MockScanArray(nil, []string{"notification_id"}, [][]interface{}{{&notficationIDsRes[0]}, {&notficationIDsRes[1]}})

		_, err := r.GetNotificationIDsByLocationIDs(ctx, db, notiIDsReq, tagIDsReq)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, notiIDsReq, tagIDsReq)
		mockDB.MockScanArray(nil, []string{"notification_id"}, [][]interface{}{{&notficationIDsRes[0]}, {&notficationIDsRes[1]}})

		notIDs, err := r.GetNotificationIDsByLocationIDs(ctx, db, notiIDsReq, tagIDsReq)
		assert.NoError(t, err)
		for i, notID := range notIDs {
			assert.Equal(t, notID, notficationIDsRes[i].String)
		}
	})
}

func TestNotificationCourseFilterRepo_BulkUpsert(t *testing.T) {
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
			Req: []*entities.NotificationCourseFilter{
				{
					NotificationID: database.Text("notification-1"),
					CourseID:       database.Text("course-1"),
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

	repo := &NotificationCourseFilterRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.BulkUpsert(ctx, db, testCase.Req.([]*entities.NotificationCourseFilter))
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestNotificationCourseFilterRepo_GetNotificationIDsByCourseIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &NotificationCourseFilterRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	tagIDsReq := database.TextArray([]string{"tag_id-1", "tag_id-2"})
	notiIDsReq := database.TextArray([]string{"notification_id-1", "notification_id-2"})
	notficationIDsRes := []pgtype.Text{database.Text("notification_id-1"), database.Text("notification_id-2")}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, notiIDsReq, tagIDsReq)
		mockDB.MockScanArray(nil, []string{"notification_id"}, [][]interface{}{{&notficationIDsRes[0]}, {&notficationIDsRes[1]}})

		_, err := r.GetNotificationIDsByCourseIDs(ctx, db, notiIDsReq, tagIDsReq)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, notiIDsReq, tagIDsReq)
		mockDB.MockScanArray(nil, []string{"notification_id"}, [][]interface{}{{&notficationIDsRes[0]}, {&notficationIDsRes[1]}})

		notIDs, err := r.GetNotificationIDsByCourseIDs(ctx, db, notiIDsReq, tagIDsReq)
		assert.NoError(t, err)
		for i, notID := range notIDs {
			assert.Equal(t, notID, notficationIDsRes[i].String)
		}
	})
}

func TestNotificationClassFilterRepo_BulkUpsert(t *testing.T) {
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
			Req: []*entities.NotificationClassFilter{
				{
					NotificationID: database.Text("notification-1"),
					ClassID:        database.Text("course-1"),
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

	repo := &NotificationClassFilterRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.BulkUpsert(ctx, db, testCase.Req.([]*entities.NotificationClassFilter))
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestNotificationClassFilterRepo_GetNotificationIDsByClassIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &NotificationClassFilterRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	tagIDsReq := database.TextArray([]string{"tag_id-1", "tag_id-2"})
	notiIDsReq := database.TextArray([]string{"notification_id-1", "notification_id-2"})
	notficationIDsRes := []pgtype.Text{database.Text("notification_id-1"), database.Text("notification_id-2")}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, notiIDsReq, tagIDsReq)
		mockDB.MockScanArray(nil, []string{"notification_id"}, [][]interface{}{{&notficationIDsRes[0]}, {&notficationIDsRes[1]}})

		_, err := r.GetNotificationIDsByClassIDs(ctx, db, notiIDsReq, tagIDsReq)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, notiIDsReq, tagIDsReq)
		mockDB.MockScanArray(nil, []string{"notification_id"}, [][]interface{}{{&notficationIDsRes[0]}, {&notficationIDsRes[1]}})

		notIDs, err := r.GetNotificationIDsByClassIDs(ctx, db, notiIDsReq, tagIDsReq)
		assert.NoError(t, err)
		for i, notID := range notIDs {
			assert.Equal(t, notID, notficationIDsRes[i].String)
		}
	})
}

func TestNotificationFilterRepo_SoftDeleteByNotificationID(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()

	testCases := []struct {
		Name           string
		NotificationID string
		ExpcErr        error
		Setup          func(ctx context.Context)
	}{
		{
			Name:           "happy case",
			NotificationID: "notification-1",
			ExpcErr:        nil,
			Setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`1`)), nil, mock.Anything, mock.AnythingOfType("string"), "notification-1")
			},
		},
		{
			Name:           "err conn closed",
			NotificationID: "notification-2",
			ExpcErr:        fmt.Errorf("err db.Exec: %v", puddle.ErrClosedPool),
			Setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, nil, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), "notification-2")
			},
		},
	}

	locationFilterRepo := &NotificationLocationFilterRepo{}
	classFilterRepo := &NotificationClassFilterRepo{}
	courseFilterRepo := &NotificationCourseFilterRepo{}
	ctx := context.Background()
	for _, testCase := range testCases {
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			err := locationFilterRepo.SoftDeleteByNotificationID(ctx, mockDB.DB, testCase.NotificationID)
			if testCase.ExpcErr == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.ExpcErr.Error(), err.Error())
			}
		})
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			err := classFilterRepo.SoftDeleteByNotificationID(ctx, mockDB.DB, testCase.NotificationID)
			if testCase.ExpcErr == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.ExpcErr.Error(), err.Error())
			}
		})
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			err := courseFilterRepo.SoftDeleteByNotificationID(ctx, mockDB.DB, testCase.NotificationID)
			if testCase.ExpcErr == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.ExpcErr.Error(), err.Error())
			}
		})
	}
}
