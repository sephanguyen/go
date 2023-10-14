package debezium

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
)

func TestIncrementalSnapshot(t *testing.T) {
	ctx := context.Background()
	snapshotWaitDuration = time.Millisecond * 10

	testCases := []struct {
		name           string
		dataCollection DataCollection
		err            error
		mockDBFunc     func(DataCollection) database.QueryExecer
	}{
		{
			name:           "happy case insert successfully",
			dataCollection: DataCollection{SourceID: "bob", Tables: []string{"public.a", "public.b", "public.c", "public.d"}},
			err:            nil,
			mockDBFunc: func(dataCollection DataCollection) database.QueryExecer {
				mockDB := testutil.NewMockDB()
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("pgtype.Text"))
				isActive := pgtype.Bool{}
				mockDB.MockRowScanFields(nil, []string{"is_active"}, []interface{}{&isActive})
				tpText := database.Text("execute-snapshot")
				dataCollectionText := database.Text(`{"data-collections": ["public.a","public.b","public.c","public.d"]}`)

				args := append([]interface{}{ctx, mock.AnythingOfType("string")}, mock.AnythingOfType("Text"), tpText, dataCollectionText)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

				return mockDB.DB
			},
		},
		{
			name:           "cannot insert to data collections to trigger incrementalSnapshot",
			dataCollection: DataCollection{SourceID: "bob", Tables: []string{}},
			mockDBFunc: func(dataCollection DataCollection) database.QueryExecer {
				mockDB := testutil.NewMockDB()
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("pgtype.Text"))
				isActive := pgtype.Bool{}
				mockDB.MockRowScanFields(nil, []string{"is_active"}, []interface{}{&isActive})
				tpText := database.Text("execute-snapshot")
				dataCollectionText := database.Text(dataCollection.String())

				args := append([]interface{}{ctx, mock.AnythingOfType("string")}, mock.AnythingOfType("Text"), tpText, dataCollectionText)
				mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

				return mockDB.DB
			},
			err: fmt.Errorf("cannot insert signal to trigger snapshot new synced table"),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := tc.mockDBFunc(tc.dataCollection)
			err := IncrementalSnapshot(ctx, db, "dbz_signal", tc.dataCollection)
			assert.Equal(t, tc.err, err)
		})
	}
}
