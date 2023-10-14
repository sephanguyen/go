package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UserDiscountTagRepoWithSqlMock() (*UserDiscountTagRepo, *testutil.MockDB) {
	studentDiscountTrackerRepo := &UserDiscountTagRepo{}
	return studentDiscountTrackerRepo, testutil.NewMockDB()
}

func TestUserDiscountTagRepo_GetDiscountTagsByUserIDAndLocationID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockUserDiscountRepo, mockDB := UserDiscountTagRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.UserDiscountTag{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		discountTags, err := mockUserDiscountRepo.GetDiscountTagsByUserIDAndLocationID(ctx, mockDB.DB, mock.Anything, mock.Anything)
		assert.Nil(t, err)
		assert.Equal(t, e, discountTags[0])
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			pgx.ErrNoRows,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.UserDiscountTag{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		discountTags, err := mockUserDiscountRepo.GetDiscountTagsByUserIDAndLocationID(ctx, mockDB.DB, mock.Anything, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, discountTags)
	})
}

func TestUserDiscountTagRepo_GetDiscountEligibilityOfStudentProduct(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockUserDiscountRepo, mockDB := UserDiscountTagRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.UserDiscountTag{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		discountTags, err := mockUserDiscountRepo.GetDiscountEligibilityOfStudentProduct(ctx, mockDB.DB, mock.Anything, mock.Anything, mock.Anything)
		assert.Nil(t, err)
		assert.Equal(t, e, discountTags[0])
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			pgx.ErrNoRows,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.UserDiscountTag{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		discountTags, err := mockUserDiscountRepo.GetDiscountEligibilityOfStudentProduct(ctx, mockDB.DB, mock.Anything, mock.Anything, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, discountTags)
	})
}

func TestUserDiscountTagRepo_GetDiscountTagsWithActivityOnDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockUserDiscountRepo, mockDB := UserDiscountTagRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.UserDiscountTag{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		discountTags, err := mockUserDiscountRepo.GetDiscountTagsWithActivityOnDate(ctx, mockDB.DB, time.Now())
		assert.Nil(t, err)
		assert.Equal(t, e, discountTags[0])
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			pgx.ErrNoRows,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.UserDiscountTag{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		discountTags, err := mockUserDiscountRepo.GetDiscountTagsWithActivityOnDate(ctx, mockDB.DB, time.Now())
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, discountTags)
	})
}

func TestUserDiscountTagRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.UserDiscountTag{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockUserDiscountRepo, mockDB := UserDiscountTagRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := mockUserDiscountRepo.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert user discount tag fail", func(t *testing.T) {
		mockUserDiscountRepo, mockDB := UserDiscountTagRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := mockUserDiscountRepo.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert UserDiscountTag: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert user discount tag", func(t *testing.T) {
		mockUserDiscountRepo, mockDB := UserDiscountTagRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := mockUserDiscountRepo.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert UserDiscountTag: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestUserDiscountTagRepo_GetActiveDiscountTagIDsByUserID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := UserDiscountTagRepoWithSqlMock()

	rows := mockDB.Rows

	t.Run(constant.HappyCase, func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		record, err := repo.GetActiveDiscountTagIDsByDateAndUserID(ctx, mockDB.DB, time.Now(), "test-student")
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.GetActiveDiscountTagIDsByDateAndUserID(ctx, mockDB.DB, time.Now(), "test-student-1")

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Row scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything).Once().Return(errors.New("test-error"))

		record, err := repo.GetActiveDiscountTagIDsByDateAndUserID(ctx, mockDB.DB, time.Now(), "test-student-2")

		assert.Equal(t, "row.Scan: test-error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
	t.Run("Row no rows result set", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)

		record, err := repo.GetActiveDiscountTagIDsByDateAndUserID(ctx, mockDB.DB, time.Now(), "test-student-3")

		assert.Equal(t, nil, err)
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestUserDiscountTagRepo_SoftDeleteByTypesAndUserID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	singleDiscountType := database.TextArray([]string{"test-type1"})
	multiDiscountType := database.TextArray([]string{"test-type1", "test-type2"})
	studentID := "test"

	mockE := &entities.UserDiscountTag{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase+"single discount type", func(t *testing.T) {
		repo, mockDB := UserDiscountTagRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.SoftDeleteByTypesAndUserID(ctx, mockDB.DB, studentID, singleDiscountType)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.HappyCase+"multi discount type", func(t *testing.T) {
		repo, mockDB := UserDiscountTagRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`2`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.SoftDeleteByTypesAndUserID(ctx, mockDB.DB, studentID, multiDiscountType)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("soft delete user discount tag record fail", func(t *testing.T) {
		repo, mockDB := UserDiscountTagRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.SoftDeleteByTypesAndUserID(ctx, mockDB.DB, studentID, singleDiscountType)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err delete UserDiscountTagRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after soft deleting user discount tag record", func(t *testing.T) {
		repo, mockDB := UserDiscountTagRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.SoftDeleteByTypesAndUserID(ctx, mockDB.DB, studentID, singleDiscountType)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err delete UserDiscountTagRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestUserDiscountTagRepo_GetDiscountTagsByUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockUserDiscountRepo, mockDB := UserDiscountTagRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.UserDiscountTag{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		discountTags, err := mockUserDiscountRepo.GetDiscountTagsByUserID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.Equal(t, e, discountTags[0])
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			pgx.ErrNoRows,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.UserDiscountTag{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		discountTags, err := mockUserDiscountRepo.GetDiscountTagsByUserID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, discountTags)
	})
}
