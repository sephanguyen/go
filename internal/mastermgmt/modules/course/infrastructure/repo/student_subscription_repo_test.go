package repo

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentSubscriptionRepoWithSqlMock() (*StudentSubscriptionRepo, *testutil.MockDB) {
	r := &StudentSubscriptionRepo{}
	return r, testutil.NewMockDB()
}

type TestCase struct {
	name         string
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestStudentSubscriptionAccessPathRepo_GetLocationActiveStudentSubscriptions(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := StudentSubscriptionRepoWithSqlMock()
	ids := []string{"id"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)

		subs, err := r.GetLocationActiveStudentSubscriptions(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, subs)
	})

	locationIDs := []string{"student-id-1", "student-id-2"}

	t.Run("success with select all fields", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(2).Return(true)
		y := 0
		for i := 0; i < 2; i++ {
			var locationID string
			rows.On("Scan", &locationID).Once().Run(func(args mock.Arguments) {
				reflect.ValueOf(args[0]).Elem().SetString(locationIDs[y])
				y++
			}).Return(nil)
		}
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		rows.On("Err").Once().Return(nil)

		_, err := r.GetLocationActiveStudentSubscriptions(ctx, mockDB.DB, ids)
		assert.Nil(t, err)

	})
}
