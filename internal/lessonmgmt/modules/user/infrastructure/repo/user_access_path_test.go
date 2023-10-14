package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UserAccessPathRepoSqlMock() (*UserAccessPathRepo, *testutil.MockDB) {
	r := &UserAccessPathRepo{}
	return r, testutil.NewMockDB()
}

func TestUserAccessPathRepo_GetLocationAssignedByUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	l, mockDB := UserAccessPathRepoSqlMock()

	args := []interface{}{mock.Anything, mock.Anything, &[]string{"user-id-1", "user-id-2"}}

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		var userID, locationID pgtype.Text
		mockDB.MockScanArray(nil, []string{"user_id", "location_id"}, [][]interface{}{
			{&userID, &locationID},
		})
		lessonIDs, err := l.GetLocationAssignedByUserID(ctx, mockDB.DB, []string{"user-id-1", "user-id-2"})
		assert.Nil(t, err)
		assert.NotNil(t, lessonIDs)
	})
}

func TestUserAccessPathRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now()
	userAccessPaths := []*domain.UserAccessPath{
		{
			UserID:     "user-id-1",
			LocationID: "location-id-1",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	t.Run("bulk upsert successful", func(t *testing.T) {
		r, mockDB := UserAccessPathRepoSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(userAccessPaths); i++ {
			batchResults.On("Exec").Return(cmdTag, nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		err := r.Create(ctx, mockDB.DB, userAccessPaths)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})
}
