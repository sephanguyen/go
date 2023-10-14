package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func SchoolCourseRepoWithSqlMock() (*SchoolCourseRepo, *testutil.MockDB) {
	r := &SchoolCourseRepo{}
	return r, testutil.NewMockDB()
}

func TestSchoolCourseRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entity.SchoolCourse{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)
	t.Run("Create school_course success", func(t *testing.T) {
		r, mockDB := SchoolCourseRepoWithSqlMock()
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Create(ctx, mockDB.DB, &entity.SchoolCourse{})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Insert school_course fail", func(t *testing.T) {
		r, mockDB := SchoolCourseRepoWithSqlMock()
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrTxClosed, args...)

		err := r.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err create SchoolCourseRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})
	t.Run("no rows affect after create school_course", func(t *testing.T) {
		r, mockDB := SchoolCourseRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.MockExecArgs(t, cmdTag, nil, args...)

		err := r.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err create SchoolCourseRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestSchoolInfoRepo_GetByIDsAndSchoolIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	schoolCourseIDs := database.TextArray([]string{idutil.ULIDNow(), idutil.ULIDNow()})
	schoolInfoIDs := database.TextArray([]string{idutil.ULIDNow(), idutil.ULIDNow()})
	_, schoolCourseValues := (&entity.SchoolCourse{}).FieldMap()
	repo, mockDB := SchoolCourseRepoWithSqlMock()

	rows := mockDB.Rows
	t.Run("happy case", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", schoolCourseValues...).Once().Return(nil)
		rows.On("Next").Once().Return(false)

		schools, err := repo.GetByIDsAndSchoolIDs(ctx, mockDB.DB, schoolCourseIDs, schoolInfoIDs)
		assert.Nil(t, err)
		assert.NotNil(t, schools)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", schoolCourseValues...).Once().Return(pgx.ErrNoRows)

		_, err := repo.GetByIDsAndSchoolIDs(ctx, mockDB.DB, schoolCourseIDs, schoolInfoIDs)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestSchoolInfoRepo_GetBySchoolCoursePartnerIDsAndSchoolIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	schoolCoursePartnerIDs := database.TextArray([]string{idutil.ULIDNow(), idutil.ULIDNow()})
	schoolInfoIDs := database.TextArray([]string{idutil.ULIDNow(), idutil.ULIDNow()})
	_, schoolCourseValues := (&entity.SchoolCourse{}).FieldMap()
	repo, mockDB := SchoolCourseRepoWithSqlMock()

	rows := mockDB.Rows
	t.Run("happy case", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", schoolCourseValues...).Once().Return(nil)
		rows.On("Next").Once().Return(false)

		schools, err := repo.GetBySchoolCoursePartnerIDsAndSchoolIDs(ctx, mockDB.DB, schoolCoursePartnerIDs, schoolInfoIDs)
		assert.Nil(t, err)
		assert.NotNil(t, schools)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", schoolCourseValues...).Once().Return(pgx.ErrNoRows)

		_, err := repo.GetBySchoolCoursePartnerIDsAndSchoolIDs(ctx, mockDB.DB, schoolCoursePartnerIDs, schoolInfoIDs)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
