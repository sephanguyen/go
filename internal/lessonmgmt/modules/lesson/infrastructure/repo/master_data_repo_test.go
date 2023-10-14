package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/mock/testutil"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func LessonMasterDataRepoWithSqlMock() (*MasterDataRepo, *testutil.MockDB) {
	r := &MasterDataRepo{}
	return r, testutil.NewMockDB()
}

func TestMasterDataRepo_GetLocationByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonMasterDataRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		centerID := "center-id"

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &centerID)
		e := &Location{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(nil, fields, values)

		_, err := r.GetLocationByID(ctx, mockDB.DB, centerID)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
			mockDB.Row,
		)
	})

	t.Run("got error", func(t *testing.T) {
		centerID := "center-id"

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &centerID)
		e := &Location{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(errors.New("error"), fields, values)

		_, err := r.GetLocationByID(ctx, mockDB.DB, centerID)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
			mockDB.Row,
		)
	})
}
