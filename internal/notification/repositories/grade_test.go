package repositories

import (
	"context"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGradeRepo_GetByPartnerInternalIDs(t *testing.T) {
	t.Parallel()
	repo := GradeRepo{}
	partnerInternalIDs := []string{"id_1", "id_2"}
	orgID := "123"
	expectedMapGrade := make(map[string]string)
	expectedMapGrade["grade_id_1"] = partnerInternalIDs[0]
	t.Run("query err", func(t *testing.T) {
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		db := mockDB.DB
		mockDB.DB.On("Query", ctx, mock.Anything, database.TextArray(partnerInternalIDs), orgID).Once().Return(nil, pgx.ErrTxClosed)

		mapGrade, _ := repo.GetByPartnerInternalIDs(ctx, db, partnerInternalIDs, orgID)
		assert.Equal(t, map[string]string(nil), mapGrade)
	})
	t.Run("happy case", func(t *testing.T) {
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		db := mockDB.DB
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, database.TextArray(partnerInternalIDs), orgID).Once().Return(rows, nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			reflect.ValueOf(args[0]).Elem().SetString("grade_id_1")
			reflect.ValueOf(args[1]).Elem().SetString("id_1")
		}).Return(nil)

		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		rows.On("Err").Once().Return(nil)

		mapGrade, err := repo.GetByPartnerInternalIDs(ctx, db, partnerInternalIDs, orgID)
		assert.Nil(t, err)
		assert.Equal(t, expectedMapGrade, mapGrade)
	})
}

func TestGradeRepo_GetGradesByOrg(t *testing.T) {
	t.Parallel()
	repo := GradeRepo{}
	orgID := "123"
	grades := []string{"grade_1"}
	expectedMapGrade := make(map[string]string)
	expectedMapGrade["grade_id_1"] = grades[0]
	t.Run("query err", func(t *testing.T) {
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		db := mockDB.DB
		mockDB.DB.On("Query", ctx, mock.Anything, orgID).Once().Return(nil, pgx.ErrTxClosed)

		mapGrade, _ := repo.GetGradesByOrg(ctx, db, orgID)
		assert.Equal(t, map[string]string(nil), mapGrade)
	})
	t.Run("happy case", func(t *testing.T) {
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		db := mockDB.DB
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, orgID).Once().Return(rows, nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			reflect.ValueOf(args[0]).Elem().SetString("grade_id_1")
			reflect.ValueOf(args[1]).Elem().SetString(grades[0])
		}).Return(nil)

		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		rows.On("Err").Once().Return(nil)

		mapGrade, err := repo.GetGradesByOrg(ctx, db, orgID)
		assert.Nil(t, err)
		assert.Equal(t, expectedMapGrade, mapGrade)
	})
}
