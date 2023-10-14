package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func SchoolInfoRepoWithSqlMock() (*SchoolInfoRepo, *testutil.MockDB) {
	r := &SchoolInfoRepo{}
	return r, testutil.NewMockDB()
}

func TestSchoolInfoRepo_BulkImport(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := SchoolInfoRepoWithSqlMock()

	// mockE := &entities.SchoolInfo{}
	// _, fieldMap := mockE.FieldMap()

	// args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)
	t.Run("BulkImport school_info success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)

		allSchoolInfo := []*entity.SchoolInfo{}
		for i := 0; i < 10; i++ {
			allSchoolInfo = append(allSchoolInfo, &entity.SchoolInfo{
				ID:           database.Text(fmt.Sprintf("%s%d", idutil.ULIDNow(), i)),
				Name:         database.Text(fmt.Sprintf("School %s%d", idutil.ULIDNow(), i)),
				NamePhonetic: database.Text(fmt.Sprintf("S%s%d", idutil.ULIDNow(), i)),
				LevelID:      database.Text(fmt.Sprintf("school_level_id-%s%d", idutil.ULIDNow(), i)),
				Address:      database.Text(fmt.Sprintf("Address %s%d", idutil.ULIDNow(), i)),
				IsArchived:   database.Bool(false),
			})
			batchResults.On("Exec").Once().Return(cmdTag, nil)
		}

		batchResults.On("Close").Once().Return(nil)
		errs := r.BulkImport(ctx, mockDB.DB, allSchoolInfo)
		assert.Equal(t, len(errs), 0)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
			mockDB.Rows,
		)
	})
	t.Run("BulkImport school_info fail", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)

		allSchoolInfo := []*entity.SchoolInfo{}
		for i := 0; i < 10; i++ {
			allSchoolInfo = append(allSchoolInfo, &entity.SchoolInfo{
				ID:           database.Text(fmt.Sprintf("%s%d", idutil.ULIDNow(), i)),
				Name:         database.Text(fmt.Sprintf("School %s%d", idutil.ULIDNow(), i)),
				NamePhonetic: database.Text(fmt.Sprintf("S%s%d", idutil.ULIDNow(), i)),
				LevelID:      database.Text(fmt.Sprintf("school_level_id-%s%d", idutil.ULIDNow(), i)),
				Address:      database.Text(fmt.Sprintf("Address %s%d", idutil.ULIDNow(), i)),
				IsArchived:   database.Bool(false),
			})
			batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		}
		batchResults.On("Close").Once().Return(nil)
		errs := r.BulkImport(ctx, mockDB.DB, allSchoolInfo)
		assert.Equal(t, len(errs), 10)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestSchoolInfoRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entity.SchoolInfo{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)
	t.Run("Create school_info success", func(t *testing.T) {
		r, mockDB := SchoolInfoRepoWithSqlMock()
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Create(ctx, mockDB.DB, &entity.SchoolInfo{})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Insert school_info fail", func(t *testing.T) {
		r, mockDB := SchoolInfoRepoWithSqlMock()
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrTxClosed, args...)

		err := r.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err create SchoolInfoRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})
	t.Run("no rows affect after create school_info", func(t *testing.T) {
		r, mockDB := SchoolInfoRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.MockExecArgs(t, cmdTag, nil, args...)

		err := r.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err create SchoolInfoRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestSchoolInfoRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entity.SchoolInfo{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)
	t.Run("Update success", func(t *testing.T) {
		r, mockDB := SchoolInfoRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)
		err := r.Update(ctx, mockDB.DB, &entity.SchoolInfo{})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Update school_info fail", func(t *testing.T) {
		r, mockDB := SchoolInfoRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := r.Update(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update SchoolInfoRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after update school_info", func(t *testing.T) {
		r, mockDB := SchoolInfoRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := r.Update(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update SchoolInfoRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestSchoolInfoRepo_GetByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	schoolInfoIDs := database.TextArray([]string{idutil.ULIDNow(), idutil.ULIDNow()})
	_, schoolValues := (&entity.SchoolInfo{}).FieldMap()
	repo, mockDB := SchoolInfoRepoWithSqlMock()

	rows := mockDB.Rows
	t.Run("happy case", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", schoolValues...).Once().Return(nil)
		rows.On("Next").Once().Return(false)

		schools, err := repo.GetByIDs(ctx, mockDB.DB, schoolInfoIDs)
		assert.Nil(t, err)
		assert.NotNil(t, schools)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", schoolValues...).Once().Return(pgx.ErrNoRows)

		_, err := repo.GetByIDs(ctx, mockDB.DB, schoolInfoIDs)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestSchoolInfoRepo_GetBySchoolPartnerIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	schoolPartnerInfoIDs := database.TextArray([]string{idutil.ULIDNow(), idutil.ULIDNow()})
	_, schoolValues := (&entity.SchoolInfo{}).FieldMap()
	repo, mockDB := SchoolInfoRepoWithSqlMock()

	rows := mockDB.Rows
	t.Run("happy case", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", schoolValues...).Once().Return(nil)
		rows.On("Next").Once().Return(false)

		schools, err := repo.GetBySchoolPartnerIDs(ctx, mockDB.DB, schoolPartnerInfoIDs)
		assert.Nil(t, err)
		assert.NotNil(t, schools)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", schoolValues...).Once().Return(pgx.ErrNoRows)

		_, err := repo.GetBySchoolPartnerIDs(ctx, mockDB.DB, schoolPartnerInfoIDs)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
