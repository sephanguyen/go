package repositories

import (
	"context"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/tom/domain/core"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_ConversationLocation_BulkUpsert(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	rows := mockDB.Rows
	db := mockDB.DB

	r := &ConversationLocationRepo{}

	locationIDs := []string{"loc-1", "loc-2"}
	var (
		loc1ap  = "loc-2/loc-1"
		loc1    = "loc-1"
		loc2    = "loc-2"
		loc2ap  = "loc-2"
		loc1Ent = core.ConversationLocation{
			LocationID: database.Text(loc1),
			AccessPath: database.Text(loc1ap),
		}
		loc2Ent = core.ConversationLocation{
			LocationID: database.Text(loc2),
			AccessPath: database.Text(loc2ap),
		}
	)
	t.Run("error finding access path for location", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(locationIDs))
		rows.On("Close").Once()

		mockDB.MockScanArray(nil, []string{"access_path", "location_id"}, [][]interface{}{
			{
				&loc1ap,
				&loc1,
			},
		})

		err := r.BulkUpsert(context.Background(), db, []core.ConversationLocation{loc1Ent, loc2Ent})
		assert.Equal(t, "cannot find access path for location loc-2", err.Error())
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(locationIDs))
		rows.On("Close").Once()

		mockDB.MockScanArray(nil, []string{"access_path", "location_id"}, [][]interface{}{
			{
				&loc1ap,
				&loc1,
			},
			{
				&loc2ap,
				&loc2,
			},
		})

		cmdTag := pgconn.CommandTag([]byte(`1`))
		batchResults := &mock_database.BatchResults{}
		db.On("SendBatch", mock.Anything, mock.MatchedBy(func(b *pgx.Batch) bool {
			return b.Len() == 2
		})).Once().Return(batchResults)
		batchResults.On("Exec").Twice().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := r.BulkUpsert(context.Background(), db, []core.ConversationLocation{loc1Ent, loc2Ent})
		assert.NoError(t, err)
	})
}
