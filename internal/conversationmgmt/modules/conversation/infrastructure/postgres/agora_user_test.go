package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/infrastructure/postgres/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAgoraUserRepo_GetByUserIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &AgoraUserRepo{}

	dto1 := &dto.AgoraUserPgDTO{}
	dto2 := &dto.AgoraUserPgDTO{}
	database.AllRandomEntity(dto1)
	database.AllRandomEntity(dto2)

	userIDs := []string{"user-id-1", "user-id-2"}

	t.Run("success", func(t *testing.T) {
		fields, vals1 := dto1.FieldMap()
		_, vals2 := dto2.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{vals1, vals2})
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, userIDs)

		vendorUsers, err := r.GetByUserIDs(ctx, db, userIDs)
		assert.Nil(t, err)
		assert.Equal(t, dto1.ToChatVendorUserDomain(), vendorUsers[0])
		assert.Equal(t, dto2.ToChatVendorUserDomain(), vendorUsers[1])
	})
	t.Run("error scan", func(t *testing.T) {
		fields, vals1 := dto1.FieldMap()
		_, vals2 := dto2.FieldMap()
		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{vals1, vals2})
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, userIDs)

		vendorUsers, err := r.GetByUserIDs(ctx, db, userIDs)
		assert.Nil(t, vendorUsers)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("error query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, userIDs)

		vendorUsers, err := r.GetByUserIDs(ctx, db, userIDs)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Nil(t, vendorUsers)
	})
}

func TestAgoraUserRepo_GetByUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &AgoraUserRepo{}

	dto := &dto.AgoraUserPgDTO{}
	database.AllRandomEntity(dto)
	userID := "user-id"
	t.Run("success", func(t *testing.T) {

		fields, vals := dto.FieldMap()
		mockDB.MockRowScanFields(nil, fields, vals)

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, database.Text(userID))
		retUser, err := r.GetByUserID(ctx, db, userID)
		assert.Nil(t, err)
		assert.Equal(t, dto.ToChatVendorUserDomain(), retUser)
	})
	t.Run("err query row", func(t *testing.T) {
		fields, vals := dto.FieldMap()
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, vals)
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, database.Text(userID))
		retUser, err := r.GetByUserID(ctx, db, userID)
		assert.Nil(t, retUser)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestAgoraUserRepo_GetByVendorUserIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &AgoraUserRepo{}

	dto1 := &dto.AgoraUserPgDTO{}
	dto2 := &dto.AgoraUserPgDTO{}
	database.AllRandomEntity(dto1)
	database.AllRandomEntity(dto2)

	vendorUserIDs := []string{"vendor-user-id-1", "vendor-user-id-2"}

	t.Run("success", func(t *testing.T) {
		fields, vals1 := dto1.FieldMap()
		_, vals2 := dto2.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{vals1, vals2})
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, vendorUserIDs)

		vendorUsers, err := r.GetByVendorUserIDs(ctx, db, vendorUserIDs)
		assert.Nil(t, err)
		assert.Equal(t, dto1.ToChatVendorUserDomain(), vendorUsers[0])
		assert.Equal(t, dto2.ToChatVendorUserDomain(), vendorUsers[1])
	})
	t.Run("error scan", func(t *testing.T) {
		fields, vals1 := dto1.FieldMap()
		_, vals2 := dto2.FieldMap()
		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{vals1, vals2})
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, vendorUserIDs)

		vendorUsers, err := r.GetByVendorUserIDs(ctx, db, vendorUserIDs)
		assert.Nil(t, vendorUsers)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("error query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, vendorUserIDs)

		vendorUsers, err := r.GetByVendorUserIDs(ctx, db, vendorUserIDs)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Nil(t, vendorUsers)
	})
}

func TestAgoraUserRepo_GetByVendorUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &AgoraUserRepo{}

	dto := &dto.AgoraUserPgDTO{}
	database.AllRandomEntity(dto)
	vendorUserID := "vendor-user-id"
	t.Run("success", func(t *testing.T) {

		fields, vals := dto.FieldMap()
		mockDB.MockRowScanFields(nil, fields, vals)

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, database.Text(vendorUserID))
		retUser, err := r.GetByVendorUserID(ctx, db, vendorUserID)
		assert.Nil(t, err)
		assert.Equal(t, dto.ToChatVendorUserDomain(), retUser)
	})
	t.Run("err query row", func(t *testing.T) {
		fields, vals := dto.FieldMap()
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, vals)
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, database.Text(vendorUserID))
		retUser, err := r.GetByVendorUserID(ctx, db, vendorUserID)
		assert.Nil(t, retUser)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}
